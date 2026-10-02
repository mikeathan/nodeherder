package automations_test

import (
	"context"
	"errors"
	"node-herder/internal/automations"
	"node-herder/mocks"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type resetCountingClock struct {
	utils.Clock
	resets atomic.Int32
}

func (c *resetCountingClock) AfterFunc(d time.Duration, fn func()) utils.Timer {
	return &resetCountingTimer{Timer: c.Clock.AfterFunc(d, fn), resets: &c.resets}
}

type resetCountingTimer struct {
	utils.Timer
	resets *atomic.Int32
}

func (t *resetCountingTimer) Reset(d time.Duration) bool { t.resets.Add(1); return t.Timer.Reset(d) }

func TestSchedulerStopRestartCommittedCallback(t *testing.T) {
	mock := mocks.NewMockClock(func() time.Time { return time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC) })
	clock := &resetCountingClock{Clock: mock}
	scheduler := automations.NewScheduler(clock, context.Background())
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	if err := scheduler.At("12:00:01").Every(time.Second).Do(func() error {
		if calls.Add(1) == 1 {
			close(entered)
			<-release
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := scheduler.Start(); err != nil {
		t.Fatal(err)
	}
	defer scheduler.Stop()
	var once sync.Once
	defer once.Do(func() { close(release) })
	go func() { defer close(done); mock.Advance(time.Second) }()
	awaitReload(t, entered)
	if err := scheduler.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := scheduler.Start(); err != nil {
		t.Fatal(err)
	}
	once.Do(func() { close(release) })
	awaitReload(t, done)
	if clock.resets.Load() != 0 {
		t.Fatal("old completion rearmed replacement timer")
	}
	mock.Advance(time.Second)
	if calls.Load() != 2 || clock.resets.Load() != 1 {
		t.Fatal("replacement run lost or duplicated")
	}
	if err := scheduler.Stop(); err != nil {
		t.Fatal(err)
	}
	mock.Advance(10 * time.Second)
	if calls.Load() != 2 {
		t.Fatal("stopped scheduler kept firing")
	}
}

func TestPreparedScheduleCanceledBeforeActivation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	clock := mocks.NewMockClock(func() time.Time { return time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC) })
	handler := automations.NewAutomationScheduler(automations.WithContext(ctx), automations.WithSchedulerClock(clock), automations.WithAutomationsFuncs())
	a := reloadRecipe(false)
	a.Schedules = []*automations.TimeSchedule{{StartAt: "12:00:01", Type: "enable"}}
	change, err := handler.Prepare(a)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := handler.Prepare(a); !errors.Is(err, automations.ErrAutomationUpdateInProgress) {
		t.Fatalf("overlapping schedule prepare not rejected: %v", err)
	}
	cancel()
	if err := change.Activate(); err == nil {
		t.Fatal("canceled scheduler activated")
	}
	change.Rollback()
	change.Rollback()
	if _, err := handler.Prepare(a); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled parent preparation accepted: %v", err)
	}
	clock.Advance(48 * time.Hour)
	if handler.IsRunning(a) {
		t.Fatal("failed activation retained scheduler")
	}
}

func TestSchedulerPartialStartDoesNotLeaveTimers(t *testing.T) {
	clock := mocks.NewMockClock(func() time.Time { return time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC) })
	scheduler := automations.NewScheduler(clock, context.Background())
	var calls int
	if err := scheduler.At("12:00:01").Every(time.Second).Do(func() error { calls++; return nil }); err != nil {
		t.Fatal(err)
	}
	if err := scheduler.At("invalid").Every(time.Second).Do(func() error { calls++; return nil }); err == nil {
		t.Fatal("invalid job accepted")
	}
	if err := scheduler.Start(); err == nil {
		t.Fatal("partial start error hidden")
	}
	clock.Advance(48 * time.Hour)
	if calls != 0 || scheduler.IsRunning() {
		t.Fatal("failed start left runnable timers")
	}
}

func TestAddTimeSchedule(t *testing.T) {

	wg := sync.WaitGroup{}
	wg.Add(4)

	order := []int{1, 2, 1, 2}
	done := make(chan int, 4)
	loc, _ := time.LoadLocation("Europe/London")
	now := time.Now().In(loc)
	start := now.Add(500 * time.Millisecond)
	end := now.Add(2000 * time.Millisecond)

	schedules := utils_test.CreateTimeSchedules(start, end)
	s := automations.NewScheduler(utils.NewRealClock(), context.Background())
	defer s.Stop()
	var mu sync.Mutex
	doneCount := 0

	// add start job
	err := s.Name("Start job").At(schedules[0].StartAt).Every(time.Second * 4).Do(func() error {
		mu.Lock()
		defer mu.Unlock()
		if doneCount < 4 {
			done <- 1
			wg.Done()
			doneCount++
		}
		return nil
	})

	if err != nil {
		t.Errorf("failed to add start job %s", err.Error())
	}

	err = s.Name("End job").At(schedules[1].StartAt).Every(time.Second * 4).Do(func() error {
		mu.Lock()
		defer mu.Unlock()
		if doneCount < 4 {
			done <- 2
			wg.Done()
			doneCount++
		}
		return nil
	})

	if err != nil {
		t.Errorf("failed to add End job %s", err.Error())
	}

	err = s.Start()
	if err != nil {
		t.Errorf("failed to start scheduler %s", err.Error())
	}

	if !waitTimeout(&wg, 10*time.Second) {
		t.Errorf("failed to execute jobs")
	}

	for i := 0; i < 4; i++ {
		id := <-done

		if order[i] != id {
			t.Errorf("job %d executed before job %d", id, order[i])
		}
	}
}

func TestStopTimeSchedule(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/London")
	now := time.Now().In(loc)
	start := now.Add(2000 * time.Millisecond)
	end := start.Add(2000 * time.Millisecond)

	schedules := utils_test.CreateTimeSchedules(start, end)
	s := automations.NewScheduler(utils.NewRealClock(), context.Background())

	// add start job
	err := s.Name("Start job").At(schedules[0].StartAt).Every(time.Second * 1).Do(func() error {
		t.Errorf("Start job executed")

		return nil
	})

	if err != nil {
		t.Errorf("failed to add start job %s", err.Error())
	}

	// add end job
	err = s.Name("End job").At(schedules[1].StartAt).Every(time.Second * 1).Do(func() error {
		t.Errorf("End job executed")

		return nil
	})
	if err != nil {
		t.Errorf("failed to add end job %s", err.Error())
	}

	err = s.Start()
	if err != nil {
		t.Errorf("failed to start scheduler %s", err.Error())
	}

	time.Sleep(200 * time.Millisecond)

	err = s.Stop()

	if err != nil {
		t.Errorf("failed to stop scheduler %s", err.Error())
	}
}

func TestTimeScheduleContextCancellation(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	loc, _ := time.LoadLocation("Europe/London")
	now := time.Now().In(loc)
	start := now.Add(2000 * time.Millisecond)
	end := start.Add(2000 * time.Millisecond)

	schedules := utils_test.CreateTimeSchedules(start, end)
	s := automations.NewScheduler(utils.NewRealClock(), ctx)

	// add start job
	err := s.Name("Start job").At(schedules[0].StartAt).Every(time.Second * 1).Do(func() error {
		t.Errorf("Start job executed")

		return nil
	})

	if err != nil {
		t.Errorf("failed to add start job %s", err.Error())
	}

	// add end job
	err = s.Name("End job").At(schedules[1].StartAt).Every(time.Second * 1).Do(func() error {
		t.Errorf("End job executed")

		return nil
	})
	if err != nil {
		t.Errorf("failed to add end job %s", err.Error())
	}

	s.Start()

	time.Sleep(100 * time.Millisecond)

	cancel()

	time.Sleep(100 * time.Millisecond)

	<-ctx.Done()

	if s.IsRunning() {
		t.Errorf("scheduler is still running")
	}
}

func TestTimeScheduleCSupportFileFormats(t *testing.T) {

	ctx := context.Background()

	testCase := []struct {
		format string
	}{
		{
			format: "15:04:05.000",
		},
		{
			format: "15:04:05",
		},
		{
			format: "15:04",
		},
	}

	for _, tc := range testCase {
		loc, _ := time.LoadLocation("Europe/London")
		now := time.Now().In(loc)
		start := now.Add(2000 * time.Millisecond)
		end := start.Add(2000 * time.Millisecond)

		schedules := utils_test.CreateTimeSchedulesWithTimeFormat(start, end, tc.format)
		s := automations.NewScheduler(utils.NewRealClock(), ctx)
		err := s.Name("Start job").At(schedules[0].StartAt).Every(time.Second * 1).Do(func() error {
			t.Errorf("Start job executed")

			return nil
		})
		if err != nil {
			t.Errorf("failed to add start job %s", err.Error())
		}

		err = s.Name("End job").At(schedules[1].StartAt).Every(time.Second * 1).Do(func() error {
			t.Errorf("End job executed")

			return nil
		})
		if err != nil {
			t.Errorf("failed to add end job %s", err.Error())
		}

		s.Stop()
	}
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		return true
	case <-time.After(timeout):
		return false
	}
}
