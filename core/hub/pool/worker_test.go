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

	return currentTime.Format("03:04:05.99999")
}

type mockJob struct {
	id string
	wg *sync.WaitGroup
}

func createJob(id string, wg *sync.WaitGroup) *mockJob {
	return &mockJob{id: id, wg: wg}

}
func (m *mockJob) OnFailure(err error) {
	log(fmt.Sprintf("Job: %s error: %s", m.id, err.Error()))

}

func (m *mockJob) Execute() error {
	log(fmt.Sprintf("processing job: %s", m.id))
	time.Sleep(1 * time.Second)

	m.wg.Done()
	return nil
}

func addNewActivity(id int, numOfTasks int, workerPool *pool.WorkerPool, wg *sync.WaitGroup) {
	for i := 0; i < numOfTasks; i++ {
		var name = fmt.Sprintf("device %d", id)

		job := createJob(name, wg)
		log(fmt.Sprintf("adding job: %s", name))
		workerPool.AddWorkNonBlocking(job)
	}
}

func TestAsyncProcessing(t *testing.T) {
	wg := &sync.WaitGroup{}

	ctx, _ := context.WithCancel(context.Background())
	worker := pool.NewWorkerPool(1, ctx)
	worker.Start()

	numOfActivies := 2
	numOfTasks := 5
	totalJobs := numOfActivies * numOfActivies
	wg.Add(totalJobs)

	for i := 0; i < numOfActivies; i++ {
		go addNewActivity(i, numOfTasks, worker, wg)
	}

	wg.Wait()
	log("finished")
}
