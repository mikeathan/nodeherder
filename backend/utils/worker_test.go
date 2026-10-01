package utils_test

import (
	"context"
	"fmt"
	"node-herder/utils"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s \n", name, elapsed)
}

type mockTask struct {
	Id         string
	EventId    int
	procesFunc func(task utils.Task) error
}

func (m *mockTask) Process() error {
	return m.procesFunc(m)
}

func (m *mockTask) OnFailure(err error) {
	fmt.Printf("Job: %s EventId: %d Error: %s", m.Id, m.EventId, err.Error())
}

func createTask(id string, eventId int, processFunc func(task utils.Task) error) utils.Task {
	return &mockTask{Id: id, EventId: eventId, procesFunc: processFunc}
}

func TestAsyncFuncProcessingAllJobs(t *testing.T) {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	procesFunc := func(task utils.Task) error {

		mockTask := task.(*mockTask)
		fmt.Printf("Processing Id: %s  EventId: %d \n", mockTask.Id, mockTask.EventId)
		time.Sleep(10 * time.Millisecond)
		wg.Done()
		return nil
	}

	worker := utils.NewWorkerPool(1, ctx)
	worker.Run()
	t.Cleanup(func() {
		cancel()
		worker.Wait()
	})

	numOfActivies := 3
	numOfTasks := 10
	totalJobs := numOfActivies * numOfTasks
	wg.Add(totalJobs)
	startTime := time.Now()

	for i := 0; i < numOfActivies; i++ {

		id := i
		name := fmt.Sprintf("device %d", id)

		go func() {

			for j := 0; j < numOfTasks; j++ {
				eventId := j

				job := createTask(name, eventId, procesFunc)
				worker.AddTask(job)
			}
		}()
	}

	completed := make(chan struct{})
	go func() {
		wg.Wait()
		timeTrack(startTime, "")
		close(completed)
	}()

	select {
	case <-completed:
	case <-time.After(10 * time.Second):
		t.Fatal("Failed to process all tasks")
	}
}

func TestCancelContextStopsWorker(t *testing.T) {

	testCases := []struct {
		numOfJobs int
	}{
		{numOfJobs: 1},
		{numOfJobs: 2},
		{numOfJobs: 3},
	}
	for _, testCase := range testCases {
		var expectedFinishedJobs = testCase.numOfJobs
		var finishedJobs atomic.Int32
		started := make(chan struct{}, expectedFinishedJobs)
		release := make(chan struct{})
		var releaseOnce sync.Once
		releaseJobs := func() { releaseOnce.Do(func() { close(release) }) }

		procesFunc := func(task utils.Task) error {
			mockTask := task.(*mockTask)
			started <- struct{}{}
			<-release
			fmt.Printf("Job: %d finished\n", mockTask.EventId)
			finishedJobs.Add(1)
			return nil
		}

		ctx, cancelCtx := context.WithCancel(context.Background())
		worker := utils.NewWorkerPool(expectedFinishedJobs, ctx)
		worker.Run()
		t.Cleanup(func() {
			cancelCtx()
			releaseJobs()
			worker.Wait()
		})
		numOfTasks := 10

		producerDone := make(chan struct{})
		go func() {
			defer close(producerDone)
			for j := 1; j <= numOfTasks; j++ {
				eventId := j
				job := createTask("test", eventId, procesFunc)
				if err := worker.AddTask(job); err != nil {
					return
				}
			}
		}()

		deadline := time.NewTimer(5 * time.Second)
		for i := 0; i < expectedFinishedJobs; i++ {
			select {
			case <-started:
			case <-deadline.C:
				t.Fatal("Workers did not start their initial tasks")
			}
		}
		deadline.Stop()
		cancelCtx()
		// Drain the blocked producer before releasing tasks, so cancellation cannot
		// race with another submission to a newly available worker.
		select {
		case <-producerDone:
		case <-time.After(5 * time.Second):
			t.Fatal("Cancellation did not unblock task submission")
		}
		releaseJobs()
		worker.Wait()

		if got := int(finishedJobs.Load()); got != expectedFinishedJobs {
			t.Errorf("Not matching num of finished jobs: got %d want %d", got, expectedFinishedJobs)
		}
		fmt.Println("finish ")
	}

}
