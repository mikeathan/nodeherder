package automations_test

import (
	"context"
	"fmt"
	"node-herder/internal/automations"
	"sync"
	"testing"
	"time"
)

func TestAddTimeSchedule(t *testing.T) {

	wg := sync.WaitGroup{}
	wg.Add(2)

	now := time.Now().UTC()
	start := now.Add(1000 * time.Millisecond)
	end := start.Add(500 * time.Millisecond)

	ts := &automations.TimeSchedule{
		Start: start.Format("15:04:05"),
		End:   end.Format("15:04:05"),
	}
	s := automations.NewScheduler(context.Background())

	// add start job
	err := s.AddJob(ts.Start, func() error {
		fmt.Println("Start job executed")
		wg.Done()

		return nil
	})

	if err != nil {
		t.Errorf("failed to add start job %s", err.Error())
	}

	// add end job
	err = s.AddJob(ts.End, func() error {
		fmt.Println("End job executed")
		wg.Done()

		return nil
	})
	if err != nil {
		t.Errorf("failed to add end job %s", err.Error())
	}

	s.Start()

	if !waitTimeout(&wg, 5*time.Second) {
		t.Errorf("failed to execute jobs")
	}
}

func TestStopTimeSchedule(t *testing.T) {

	now := time.Now().UTC()
	start := now.Add(2000 * time.Millisecond)
	end := start.Add(2000 * time.Millisecond)

	ts := &automations.TimeSchedule{
		Start: start.Format("15:04:05"),
		End:   end.Format("15:04:05"),
	}
	s := automations.NewScheduler(context.Background())

	// add start job
	err := s.AddJob(ts.Start, func() error {
		t.Errorf("Start job executed")

		return nil
	})

	if err != nil {
		t.Errorf("failed to add start job %s", err.Error())
	}

	// add end job
	err = s.AddJob(ts.End, func() error {
		t.Errorf("End job executed")

		return nil
	})
	if err != nil {
		t.Errorf("failed to add end job %s", err.Error())
	}

	s.Start()

	time.Sleep(200 * time.Millisecond)

	err = s.Stop()

	if err != nil {
		t.Errorf("failed to stop scheduler %s", err.Error())
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
