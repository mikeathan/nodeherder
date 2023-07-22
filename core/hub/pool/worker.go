package pool

import (
	"context"
	"fmt"
	"time"
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
}

func NewWorkerPool(numOfWorkers int, ctx context.Context) *WorkerPool {
	return &WorkerPool{
		numOfWorkers: numOfWorkers,
		ctx:          ctx,
		queue:        make(chan Task),
		quit:         make(chan bool),
	}
}

func (w *WorkerPool) Start() {
	go w.startWorkers()
}

func (w *WorkerPool) startWorkers() {
	//for i := 0; i < w.numOfWorkers; i++ {
	for {
		select {
		case <-w.quit:
			log(fmt.Sprintf("stopping worker %d with quic channel tasks channel\n", w.numOfWorkers))
			return
		case <-w.ctx.Done():
			log(fmt.Sprintf("Cancelled worker. Error: %v\n", w.ctx.Err()))
			return
		case task, ok := <-w.queue:
			if !ok {
				log(fmt.Sprintf("stopping worker %d with closed tasks channel\n", w.numOfWorkers))
				return
			}
			if err := task.Execute(); err != nil {
				task.OnFailure(err)
			}
		}
	}

	//}
}
func (w *WorkerPool) Stop() {
	close(w.quit)
}

func (w *WorkerPool) AddTask(task Task) {

	w.queue <- task
}

func (w *WorkerPool) AddWorkNonBlocking(task Task) {
	go w.AddTask(task)
}

func log(message string) {
	fmt.Printf("[%s] %s\n", getNow(), message)
}

func getNow() string {
	currentTime := time.Now()

	return currentTime.Format("03:04:05.99999")
}
