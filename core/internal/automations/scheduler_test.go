package automations_test

import (
	"context"
	"node-herder/internal/automations"
	utils_test "node-herder/testing"
	"sync"
	"testing"
	"time"
)

func TestAddTimeSchedule(t *testing.T) {

	wg := sync.WaitGroup{}
	wg.Add(4)

	order := []int{1, 2, 1, 2}
	done := make(chan int, 4)
	now := time.Now().UTC()
	start := now.Add(500 * time.Millisecond)
	end := now.Add(2000 * time.Millisecond)

	schedules := utils_test.CreateTimeSchedule(start, end)
	s := automations.NewScheduler(context.Background())
	// add start job
	err := s.Name("Start job").At(schedules[0].StartAt).Every(time.Second * 4).Do(func() error {
		done <- 1
		wg.Done()
		return nil
	})

	if err != nil {
		t.Errorf("failed to add start job %s", err.Error())
	}

	err = s.Name("End job").At(schedules[1].StartAt).Every(time.Second * 4).Do(func() error {
		done <- 2
		wg.Done()
		return nil
	})

	if err != nil {
		t.Errorf("failed to add End job %s", err.Error())
	}

	err = s.Start()
	if err != nil {
		t.Errorf("failed to start scheduler %s", err.Error())
	}

	if !waitTimeout(&wg, 60*time.Second) {
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

	now := time.Now().UTC()
	start := now.Add(2000 * time.Millisecond)
	end := start.Add(2000 * time.Millisecond)

	schedules := utils_test.CreateTimeSchedule(start, end)
	s := automations.NewScheduler(context.Background())

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
	now := time.Now().UTC()
	start := now.Add(2000 * time.Millisecond)
	end := start.Add(2000 * time.Millisecond)

	schedules := utils_test.CreateTimeSchedule(start, end)
	s := automations.NewScheduler(ctx)

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
		now := time.Now().UTC()
		start := now.Add(2000 * time.Millisecond)
		end := start.Add(2000 * time.Millisecond)

		schedules := utils_test.CreateTimeScheduleWithTimeFormat(start, end, tc.format)
		s := automations.NewScheduler(ctx)
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
