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
	wg      *sync.WaitGroup
}

func createJobWithFunc(id string, eventId int, task func(), wg *sync.WaitGroup) *mockJobWithFunc {
	return &mockJobWithFunc{id: id, eventId: eventId, task: task, wg: wg}

}
func (m *mockJobWithFunc) OnFailure(err error) {
	fmt.Printf("Job: %s error: %s", m.id, err.Error())
}

func (m *mockJobWithFunc) Execute() error {
	defer m.wg.Done()
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
					time.Sleep(10 * time.Millisecond)
				}

				job := createJobWithFunc(name, eventId, task, wg)
				worker.AddWorkNonBlocking(job)
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
	wg := &sync.WaitGroup{}
	//ticker := time.NewTicker(1 * time.Second)
	ctx, cancelCtx := context.WithCancel(context.Background())
	worker := pool.NewWorkerPool(1, ctx)
	worker.Start()
	//cancelled := false
	numOfTasks := 10
	wg.Add(numOfTasks)
	go func() {

		for j := 0; j < numOfTasks; j++ {
			eventId := j
			task := func() {
				time.Sleep(10 * time.Second)
			}

			job := createJobWithFunc("test", eventId, task, wg)
			fmt.Println("adding", j)
			worker.AddWorkNonBlocking(job)
		}
	}()

	go func() {
		time.Sleep(1 * time.Second)
		cancelCtx()
		//cancelled = true
		//ticker.Reset(time.Microsecond)
	}()

	worker.Wait()

}
