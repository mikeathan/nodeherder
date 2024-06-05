package utils

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type WorkerTask struct {
	action func() error
}

func NewWorkerTask(action func() error) *WorkerTask {
	return &WorkerTask{
		action: action,
	}
}

func (t *WorkerTask) OnFailure(err error) {
	// TODO: maybe do somethng with the error
	LogErrorf("WorkerTask  Error: %s", err.Error())
}

func (t *WorkerTask) Process() error {
	return t.action()
}

type Task interface {
	OnFailure(error)
	Process() error
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

func (w *WorkerPool) WithContext(ctx context.Context) {
	w.ctx = ctx
}

func (w *WorkerPool) Run() {
	go w.startWorkers()
}

func (w *WorkerPool) startWorkers() {
	defer close(w.queue)
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
			LogInfof("stopping worker %d with quit channel tasks channel", workerId)
			return
		case <-w.ctx.Done():
			LogInfof("Cancelled worker. Error: %v", w.ctx.Err())
			return
		case task, ok := <-w.queue:
			if !ok {
				LogInfof("stopping worker %d with closed tasks channel", workerId)
				return
			}
			if err := task.Process(); err != nil {
				task.OnFailure(err)
			}
		}
	}
}
func (w *WorkerPool) Wait() {
	<-w.quit
}

func (w *WorkerPool) Stop() {
	close(w.quit)
}

func (w *WorkerPool) AddTask(task Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint("not accepting new jobs: ", r))
		}
	}()

	select {
	case <-w.quit:
		return errors.New("not accepting new jobs")
	case <-w.ctx.Done():
		return fmt.Errorf("cancelled worker. Error: %v", w.ctx.Err())
	case w.queue <- task:
		return nil
	}
}

func (w *WorkerPool) AddWorkNonBlocking(task Task) {

	go w.AddTask(task)
}
