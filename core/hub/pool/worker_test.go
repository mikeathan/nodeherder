package pool_test

import (
	"context"
	"fmt"
	"node-herder/hub/pool"
	"sync"
	"testing"
	"time"
)

func log(message string) {
	fmt.Printf("[%s] %s\n", getNow(), message)
}

func getNow() string {
	currentTime := time.Now()

	return currentTime.Format("15:04:05,000") // "15:04:05,000"
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
	log(fmt.Sprintf("Job: %s error: %s", m.id, err.Error()))
}

func (m *mockJobWithFunc) Execute() error {
	defer m.wg.Done()
	m.task()
	return nil
}

func TestAsyncFuncProcessing(t *testing.T) {
	wg := &sync.WaitGroup{}
	ticker := time.NewTicker(50 * time.Second)
	ctx, _ := context.WithCancel(context.Background())
	worker := pool.NewWorkerPool(1, ctx)
	worker.Start()
	var processed = false
	numOfActivies := 2
	numOfTasks := 5
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
					//log(fmt.Sprintf("PROCESSING: %s eventId: %d", name, eventId))
					time.Sleep(500 * time.Millisecond)
				}

				job := createJobWithFunc(name, eventId, task, wg)
				//log(fmt.Sprintf("ADDING:  %s eventId: %d", name, eventId))
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

	log("waiting for completion")
	<-ticker.C
	if !processed {
		t.Error("Failed to process all tasks")
	}

}
func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s \n", name, elapsed)
}
