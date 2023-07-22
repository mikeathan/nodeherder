package pool

import (
	"context"
)

type Task struct {
}
type WorkerPool struct {
	numOfWorkers int
	items        chan *Task
	ctx          context.Context
}

func NewWorkerPool(numOfWorkers int, ctx context.Context) *WorkerPool {
	return &WorkerPool{numOfWorkers: numOfWorkers, ctx: ctx, items: make(chan *Task)}
}

func (w *WorkerPool) Start() {
	for i := 0; i < w.numOfWorkers; i++ {
		select {
		case <-w.items:
		

	}
}

func (w *WorkerPool) Stop() {
}

func (w *WorkerPool) AddTask(task *Task) {
}
