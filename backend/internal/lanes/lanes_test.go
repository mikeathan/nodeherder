package lanes_test

import (
	"context"
	"errors"
	"node-herder/internal/lanes"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const testTimeout = 5 * time.Second

// blockLane occupies key with a running task until the returned func is called.
func blockLane(t *testing.T, e *lanes.Executor, key string) (release func()) {
	t.Helper()
	running, unblock := make(chan struct{}), make(chan struct{})
	if err := e.Submit(key, func() { close(running); <-unblock }); err != nil {
		t.Fatal(err)
	}
	select {
	case <-running:
	case <-time.After(testTimeout):
		t.Fatal("blocking task did not start")
	}
	var once sync.Once
	return func() { once.Do(func() { close(unblock) }) }
}

func shutdown(t *testing.T, e *lanes.Executor) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

func TestTasksForOneKeyRunInSubmissionOrder(t *testing.T) {
	const n = 5000
	e := lanes.New(lanes.WithCapacity(n))
	var got []int
	for i := 0; i < n; i++ {
		if err := e.Submit("device", func() { got = append(got, i) }); err != nil {
			t.Fatal(err)
		}
	}
	shutdown(t, e)
	if len(got) != n {
		t.Fatalf("ran %d tasks, want %d", len(got), n)
	}
	for i, v := range got {
		if v != i {
			t.Fatalf("position %d ran task %d", i, v)
		}
	}
}

// AC-03: a slow key does not delay other keys.
func TestSlowKeyDoesNotBlockOtherKeys(t *testing.T) {
	e := lanes.New()
	release := blockLane(t, e, "device A")
	defer release()

	done := make(chan struct{})
	if err := e.Submit("device B", func() { close(done) }); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(testTimeout):
		t.Fatal("device B waited for device A")
	}
	release()
	shutdown(t, e)
}

// NFR-04: beyond capacity a task is rejected and counted; queued tasks all run.
func TestFullLaneRejectsAndCountsWithoutDropping(t *testing.T) {
	const capacity = 10
	e := lanes.New(lanes.WithCapacity(capacity))
	release := blockLane(t, e, "dial")

	var ran atomic.Int32
	for i := 0; i < capacity; i++ {
		if err := e.Submit("dial", func() { ran.Add(1) }); err != nil {
			t.Fatalf("task %d: %v", i, err)
		}
	}
	err := e.Submit("dial", func() { t.Error("rejected task ran") })
	if !errors.Is(err, lanes.ErrLaneFull) {
		t.Fatalf("got %v, want ErrLaneFull", err)
	}
	if got := e.Rejected(); got != 1 {
		t.Fatalf("rejected count %d, want 1", got)
	}
	if err := e.Submit("light", func() {}); err != nil {
		t.Fatalf("other key affected by full lane: %v", err)
	}

	release()
	shutdown(t, e)
	if got := ran.Load(); got != capacity {
		t.Fatalf("ran %d queued tasks, want %d", got, capacity)
	}
}

func TestPanickingTaskDoesNotStopLane(t *testing.T) {
	e := lanes.New()
	var ran atomic.Bool
	if err := e.Submit("device", func() { panic("boom") }); err != nil {
		t.Fatal(err)
	}
	if err := e.Submit("device", func() { ran.Store(true) }); err != nil {
		t.Fatal(err)
	}
	shutdown(t, e)
	if !ran.Load() {
		t.Fatal("task after a panic did not run")
	}
}

// NFR-02: idle keys keep no goroutine and no map entry.
func TestIdleLanesReleaseGoroutines(t *testing.T) {
	e := lanes.New()
	before := runtime.NumGoroutine()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		if err := e.Submit(string(rune('a'+i)), wg.Done); err != nil {
			t.Fatal(err)
		}
	}
	wg.Wait()

	deadline := time.Now().Add(testTimeout)
	for (e.Busy() != 0 || runtime.NumGoroutine() > before) && time.Now().Before(deadline) {
		runtime.Gosched()
	}
	if busy := e.Busy(); busy != 0 {
		t.Fatalf("%d lanes still registered while idle", busy)
	}
	if n := runtime.NumGoroutine(); n > before {
		t.Fatalf("goroutines %d after idle, %d before", n, before)
	}
}

func TestDoReturnsTaskResult(t *testing.T) {
	e := lanes.New()
	want := errors.New("trigger failed")
	if err := e.Do(context.Background(), "device", func() error { return want }); !errors.Is(err, want) {
		t.Fatalf("got %v, want %v", err, want)
	}
	if err := e.Do(context.Background(), "device", func() error { panic("boom") }); err == nil {
		t.Fatal("panicking task returned nil")
	}
	shutdown(t, e)
}

// AC-10: a deadline that expires while queued returns ErrNotStarted and the
// task never runs.
func TestDoDeadlineBeforeStartNeverRuns(t *testing.T) {
	e := lanes.New()
	release := blockLane(t, e, "device")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	var ran atomic.Bool
	err := e.Do(ctx, "device", func() error { ran.Store(true); return nil })
	// DeadlineExceeded can only come from the queued-then-abandoned branch: the
	// context was live when Do queued the task.
	if !errors.Is(err, lanes.ErrNotStarted) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v, want ErrNotStarted wrapping context.DeadlineExceeded", err)
	}

	release()
	shutdown(t, e)
	if ran.Load() {
		t.Fatal("abandoned task ran")
	}
}

// Once started, Do waits for the task instead of reporting it as not started.
func TestDoWaitsForStartedTask(t *testing.T) {
	e := lanes.New()
	ctx, cancel := context.WithCancel(context.Background())
	started, finish := make(chan struct{}), make(chan struct{})
	result := make(chan error, 1)
	go func() {
		result <- e.Do(ctx, "device", func() error { close(started); <-finish; return nil })
	}()
	<-started
	cancel()
	close(finish)
	if err := <-result; err != nil {
		t.Fatalf("started task reported %v", err)
	}
	shutdown(t, e)
}

func TestDoOnFullLaneReturnsErrLaneFull(t *testing.T) {
	e := lanes.New(lanes.WithCapacity(1))
	release := blockLane(t, e, "device")
	if err := e.Submit("device", func() {}); err != nil {
		t.Fatal(err)
	}
	err := e.Do(context.Background(), "device", func() error { return nil })
	if !errors.Is(err, lanes.ErrLaneFull) {
		t.Fatalf("got %v, want ErrLaneFull", err)
	}
	release()
	shutdown(t, e)
}

// AC-14: shutdown drains queued work, then refuses new work.
func TestShutdownDrainsThenRejects(t *testing.T) {
	e := lanes.New()
	var ran atomic.Int32
	for i := 0; i < 100; i++ {
		if err := e.Submit("device", func() { ran.Add(1) }); err != nil {
			t.Fatal(err)
		}
	}
	shutdown(t, e)
	if got := ran.Load(); got != 100 {
		t.Fatalf("ran %d tasks before shutdown returned, want 100", got)
	}
	if err := e.Submit("device", func() {}); !errors.Is(err, lanes.ErrClosed) {
		t.Fatalf("got %v, want ErrClosed", err)
	}
	if err := e.Do(context.Background(), "device", func() error { return nil }); !errors.Is(err, lanes.ErrClosed) {
		t.Fatalf("Do after shutdown: got %v, want ErrClosed", err)
	}
}

// AC-14: a shutdown deadline discards queued tasks and returns promptly.
func TestShutdownDeadlineDiscardsQueuedTasks(t *testing.T) {
	e := lanes.New()
	release := blockLane(t, e, "device")
	var ran atomic.Bool
	if err := e.Submit("device", func() { ran.Store(true) }); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := e.Shutdown(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
	release()
	shutdown(t, e)
	if ran.Load() {
		t.Fatal("discarded task ran")
	}
}
