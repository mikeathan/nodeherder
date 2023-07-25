package pool

import (
	"context"
	"fmt"
	"sync"
)

type Task interface {
	Execute() error
	OnFailure(error)
}

type WorkerPool struct {
	numOfWorkers int
	queue        chan Task
	ctx          context.Context
	quit         chan bool
	wg           *sync.WaitGroup
}

func NewWorkerPool(numOfWorkers int, ctx context.Context) *WorkerPool {
	return &WorkerPool{
		numOfWorkers: numOfWorkers,
		ctx:          ctx,
		queue:        make(chan Task),
		quit:         make(chan bool),
		wg:           &sync.WaitGroup{},
	}
}

func (w *WorkerPool) Start() {
	go w.startWorkers()
}

func (w *WorkerPool) startWorkers() {

	defer close(w.quit)

	w.wg.Add(w.numOfWorkers)

	for i := 0; i < w.numOfWorkers; i++ {
		workerId := i

		go w.worker(workerId, w.wg)
	}

	w.wg.Wait()
}

func (w *WorkerPool) worker(workerId int, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-w.quit:
			fmt.Printf("stopping worker %d with quit channel tasks channel\n", workerId)
			return
		case <-w.ctx.Done():
			fmt.Printf("Cancelled worker. Error: %v\n", w.ctx.Err())
			return
		case task, ok := <-w.queue:
			if !ok {
				fmt.Printf("stopping worker %d with closed tasks channel\n", workerId)
				return
			}
			if err := task.Execute(); err != nil {
				task.OnFailure(err)
			}
		}
	}
}
func (w *WorkerPool) Wait() {
	<-w.quit
}

func (w *WorkerPool) Stop() {
	//close(w.quit)
	w.wg.Wait()
}

func (w *WorkerPool) AddTask(task Task) {

	select {
	case <-w.quit:
		fmt.Println("queue is closed. cannot add task")
		return
	case <-w.ctx.Done():
		fmt.Printf("Cancelled worker. Error: %v\n", w.ctx.Err())
		return
	default:
		w.queue <- task
	}

}

func (w *WorkerPool) AddWorkNonBlocking(task Task) {

	go w.AddTask(task)
}
