// Package lanes runs tasks in FIFO order per key, with different keys running
// concurrently.
//
// Ownership and lifecycle (BE-03): an Executor owns one goroutine per busy key.
// The goroutine starts with the first task submitted to an idle key, runs that
// key's tasks one at a time, and exits, removing the key, once its queue is
// empty; an idle key costs nothing. Each queue holds at most the configured
// capacity; a task beyond it is rejected with ErrLaneFull, logged and counted,
// never dropped silently. A panicking task is recovered and logged, and the
// lane continues. Shutdown stops intake and waits for queued tasks to finish.
package lanes

import (
	"context"
	"errors"
	"fmt"
	"node-herder/utils"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// DefaultCapacity bounds the queued tasks per key (spec 005, Q-03).
	DefaultCapacity = 1000
	// depthWarnSteps: a warning is logged each time a queue grows by another
	// 1/depthWarnSteps of its capacity.
	depthWarnSteps = 10
	// defaultSlowWait logs a warning once per busy period when a task waited longer.
	defaultSlowWait = time.Second
)

var (
	// ErrLaneFull reports that the key's queue is at capacity; the task was not queued.
	ErrLaneFull = errors.New("lane queue is full")
	// ErrClosed reports that the executor is shut down; the task was not queued.
	ErrClosed = errors.New("lane executor is shut down")
	// ErrNotStarted reports that Do's context ended before its task started; the
	// task will not run.
	ErrNotStarted = errors.New("lane task did not start before the deadline")
)

// Executor runs tasks in per-key FIFO lanes. The zero value is not usable; use New.
type Executor struct {
	name     string
	capacity int
	slowWait time.Duration

	mu     sync.Mutex
	lanes  map[string]*lane
	closed bool
	wg     sync.WaitGroup

	rejected atomic.Uint64
}

type lane struct {
	queue      []queuedTask
	warnedWait bool
}

type queuedTask struct {
	run      func()
	enqueued time.Time
}

// Option configures an Executor.
type Option func(*Executor)

// WithCapacity sets the maximum number of queued tasks per key.
func WithCapacity(capacity int) Option {
	return func(e *Executor) {
		if capacity > 0 {
			e.capacity = capacity
		}
	}
}

// WithName labels the executor in log messages.
func WithName(name string) Option {
	return func(e *Executor) {
		e.name = name
	}
}

// New returns an Executor with DefaultCapacity unless configured otherwise.
func New(opts ...Option) *Executor {
	e := &Executor{
		name:     "lanes",
		capacity: DefaultCapacity,
		slowWait: defaultSlowWait,
		lanes:    make(map[string]*lane),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Submit queues task on key's lane without blocking. Tasks for one key run in
// submission order; tasks for different keys may run concurrently.
func (e *Executor) Submit(key string, task func()) error {
	now := time.Now()
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return ErrClosed
	}
	l, busy := e.lanes[key]
	if busy && len(l.queue) >= e.capacity {
		e.mu.Unlock()
		rejected := e.rejected.Add(1)
		utils.LogErrorf("%s: lane %q full (%d queued); rejected task, %d rejected in total", e.name, key, e.capacity, rejected)
		return fmt.Errorf("%w: %s", ErrLaneFull, key)
	}
	if !busy {
		l = &lane{}
		e.lanes[key] = l
		e.wg.Add(1)
		go e.run(key, l)
	}
	l.queue = append(l.queue, queuedTask{run: task, enqueued: now})
	depth := len(l.queue)
	e.mu.Unlock()

	if step := max(e.capacity/depthWarnSteps, 1); depth%step == 0 {
		utils.LogWarnf("%s: lane %q has %d queued tasks (capacity %d)", e.name, key, depth, e.capacity)
	}
	return nil
}

// Do runs fn on key's lane and returns its error. If ctx ends before fn starts,
// fn never runs and Do returns an error wrapping ErrNotStarted; once fn has
// started, Do waits for it to finish. A full lane returns ErrLaneFull.
// Callers must pass a ctx with a deadline: if Shutdown discards the queued task,
// Do returns only when ctx ends. Calling Do for key K from a task running on
// lane K deadlocks.
func (e *Executor) Do(ctx context.Context, key string, fn func() error) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%w: %w", ErrNotStarted, err)
	}
	const (
		pending int32 = iota
		started
		abandoned
	)
	var state atomic.Int32
	done := make(chan error, 1)

	err := e.Submit(key, func() {
		if !state.CompareAndSwap(pending, started) {
			return
		}
		var err error
		defer func() {
			if r := recover(); r != nil {
				utils.LogErrorf("%s: lane %q task panicked: %v\n%s", e.name, key, r, debug.Stack())
				err = fmt.Errorf("lane %q task panicked: %v", key, r)
			}
			done <- err
		}()
		err = fn()
	})
	if err != nil {
		return err
	}

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		if state.CompareAndSwap(pending, abandoned) {
			return fmt.Errorf("%w: %w", ErrNotStarted, ctx.Err())
		}
		return <-done
	}
}

// Shutdown stops accepting tasks and waits for queued tasks to finish. If ctx
// ends first, tasks not yet started are discarded (logged with their count),
// running tasks are left to finish, and ctx's error is returned.
func (e *Executor) Shutdown(ctx context.Context) error {
	e.mu.Lock()
	e.closed = true
	e.mu.Unlock()

	drained := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(drained)
	}()

	select {
	case <-drained:
		return nil
	case <-ctx.Done():
	}

	e.mu.Lock()
	discarded := 0
	for _, l := range e.lanes {
		discarded += len(l.queue)
		l.queue = nil
	}
	e.mu.Unlock()
	if discarded > 0 {
		utils.LogErrorf("%s: shutdown deadline reached; discarded %d queued tasks", e.name, discarded)
	}
	return ctx.Err()
}

// Rejected returns how many tasks were rejected because a lane was full.
func (e *Executor) Rejected() uint64 {
	return e.rejected.Load()
}

// Busy returns the number of keys with a running lane.
func (e *Executor) Busy() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.lanes)
}

func (e *Executor) run(key string, l *lane) {
	defer e.wg.Done()
	for {
		e.mu.Lock()
		if len(l.queue) == 0 {
			delete(e.lanes, key)
			e.mu.Unlock()
			return
		}
		next := l.queue[0]
		l.queue[0] = queuedTask{}
		l.queue = l.queue[1:]
		warn := !l.warnedWait && time.Since(next.enqueued) >= e.slowWait
		if warn {
			l.warnedWait = true
		}
		e.mu.Unlock()

		if warn {
			utils.LogWarnf("%s: lane %q task waited %s before starting", e.name, key, time.Since(next.enqueued))
		}
		e.execute(key, next.run)
	}
}

func (e *Executor) execute(key string, task func()) {
	defer func() {
		if r := recover(); r != nil {
			utils.LogErrorf("%s: lane %q task panicked: %v\n%s", e.name, key, r, debug.Stack())
		}
	}()
	task()
}
