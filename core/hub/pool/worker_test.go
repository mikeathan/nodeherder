package pool_test

import (
	"context"
	"fmt"
	"node-herder/hub/pool"
	"sync"
	"testing"
	"time"
)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s \n", name, elapsed)
}

type mockTask struct {
	Id      string
	EventId int
}

func (m *mockTask) OnFailure(err error) {
	fmt.Printf("Job: %s EventId: %d Error: %s", m.Id, m.EventId, err.Error())
}

func createTask(id string, eventId int) pool.Task {
	return &mockTask{Id: id, EventId: eventId}
}

func TestAsyncFuncProcessingAllJobs(t *testing.T) {
	wg := &sync.WaitGroup{}
	ticker := time.NewTicker(10 * time.Second)
	ctx, _ := context.WithCancel(context.Background())

	procesFunc := func(task pool.Task) error {

		mockTask := task.(*mockTask)
		fmt.Printf("Processing Id: %s  EventId: %d \n", mockTask.Id, mockTask.EventId)
		time.Sleep(10 * time.Millisecond)
		wg.Done()
		return nil
	}

	worker := pool.NewWorkerPool(1, ctx, procesFunc)
	worker.Start()

	var processed = false
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

				job := createTask(name, eventId)
				worker.AddTask(job)
			}
		}()
	}

	go func() {
		wg.Wait()
		timeTrack(startTime, "")
		processed = true
		ticker.Reset(time.Microsecond)
	}()

	<-ticker.C
	if !processed {
		t.Error("Failed to process all tasks")
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
		var finishedJobs = 0

		procesFunc := func(task pool.Task) error {
			mockTask := task.(*mockTask)
			for i := 0; i < 10; i++ {
				time.Sleep(100 * time.Millisecond)
			}
			fmt.Printf("Job: %d finished\n", mockTask.EventId)
			finishedJobs++
			return nil
		}

		ctx, cancelCtx := context.WithCancel(context.Background())
		worker := pool.NewWorkerPool(expectedFinishedJobs, ctx, procesFunc)
		worker.Start()
		numOfTasks := 10

		go func() {
			for j := 1; j <= numOfTasks; j++ {
				eventId := j
				func() {
					job := createTask("test", eventId)
					fmt.Println("adding", eventId)
					worker.AddTask(job)
				}()
			}
		}()

		go func() {
			time.Sleep(1 * time.Second)
			cancelCtx()
		}()
		worker.Wait()

		if finishedJobs != expectedFinishedJobs {
			t.Errorf("Not matching num of finished jobs: got %d want %d", finishedJobs, expectedFinishedJobs)
		}
		fmt.Println("finish ")
	}

}
