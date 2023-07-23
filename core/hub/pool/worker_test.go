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
	id   string
	task func()
	wg   *sync.WaitGroup
}

func createJobWithFunc(id string, task func(), wg *sync.WaitGroup) *mockJobWithFunc {
	return &mockJobWithFunc{id: id, task: task, wg: wg}

}
func (m *mockJobWithFunc) OnFailure(err error) {
	log(fmt.Sprintf("Job: %s error: %s", m.id, err.Error()))

}

func (m *mockJobWithFunc) Execute() error {
	log(fmt.Sprintf("processing job: %s", m.id))
	m.task()
	m.wg.Done()
	return nil
}

func TestAsyncFuncProcessing(t *testing.T) {
	wg := &sync.WaitGroup{}

	ctx, _ := context.WithCancel(context.Background())
	worker := pool.NewWorkerPool(1, ctx)
	worker.Start()

	numOfActivies := 2
	numOfTasks := 5
	totalJobs := numOfActivies * numOfActivies
	wg.Add(totalJobs)

	for i := 0; i < numOfActivies; i++ {
		id := i
		var name = fmt.Sprintf("device %d", id)

		go func() {
			for j := 0; j < numOfTasks; j++ {

				task := func() {
					mockTaskProcess(name, nil)
				}

				job := createJobWithFunc(name, task, wg)
				//log(fmt.Sprintf("adding job: %s", name))
				worker.AddWorkNonBlocking(job)
			}
		}()
	}

	wg.Wait()

	log("finished")
}

func mockTaskProcess(name string, payload interface{}) {
	log(fmt.Sprintf("processing task: %s", name))
	time.Sleep(100 * time.Millisecond)
}
