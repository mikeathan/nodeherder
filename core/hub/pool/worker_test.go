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

type mockJobWithFunc struct {
	id      string
	eventId int
	task    func()
}

func createJobWithFunc(id string, eventId int, task func()) *mockJobWithFunc {
	return &mockJobWithFunc{id: id, eventId: eventId, task: task}

}
func (m *mockJobWithFunc) OnFailure(err error) {
	fmt.Printf("Job: %s error: %s", m.id, err.Error())
}

func (m *mockJobWithFunc) Execute() error {

	m.task()
	return nil
}

func TestAsyncFuncProcessingAllJobs(t *testing.T) {
	wg := &sync.WaitGroup{}
	ticker := time.NewTicker(10 * time.Second)
	ctx, _ := context.WithCancel(context.Background())
	worker := pool.NewWorkerPool(1, ctx)
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
				task := func() {
					defer wg.Done()
					time.Sleep(10 * time.Millisecond)
				}

				job := createJobWithFunc(name, eventId, task)
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

		ctx, cancelCtx := context.WithCancel(context.Background())
		worker := pool.NewWorkerPool(expectedFinishedJobs, ctx)
		worker.Start()
		numOfTasks := 10

		go func() {
			for j := 0; j < numOfTasks; j++ {
				eventId := j
				func() {
					task := func() {
						select {
						case <-ctx.Done():
							return
						default:
							break
						}

						for i := 0; i < 10; i++ {
							time.Sleep(100 * time.Millisecond)
						}
						fmt.Printf("Job: %d finished\n", eventId)
						finishedJobs++
					}

					job := createJobWithFunc("test", eventId, task)
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
