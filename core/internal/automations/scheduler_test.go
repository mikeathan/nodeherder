package automations_test

import (
	"context"
	"fmt"
	"node-herder/internal/automations"
	"sync"
	"testing"
	"time"
)

func TestAddTimeScheduleTEMP(t *testing.T) {

	wg := sync.WaitGroup{}
	wg.Add(2)

	order := []int{1, 2, 1, 2}
	done := make(chan int, 4)
	now := time.Now().UTC()
	start := now.Add(500 * time.Millisecond)
	end := now.Add(1500 * time.Millisecond)

	ts := &automations.TimeSchedule{
		Start: start.Format("15:04:05.000"),
		End:   end.Format("15:04:05.000"),
	}

	s := automations.NewScheduler(context.Background())
	// add start job
	err := s.Name("Start job").At(ts.Start).Every(time.Second * 2).Do(func() error {
		fmt.Println("Start job executed")

		done <- 1
		wg.Done()
		return nil
	})

	if err != nil {
		t.Errorf("failed to add start job %s", err.Error())
	}

	err = s.Name("End job").At(ts.End).Every(time.Second * 2).Do(func() error {
		fmt.Println("End job executed")

		done <- 2
		wg.Done()
		return nil
	})

	if err != nil {
		t.Errorf("failed to add End job %s", err.Error())
	}

	s.Start()

	if !waitTimeout(&wg, 60*time.Second) {
		t.Errorf("failed to execute jobs")
	}

	for i := 0; i < 2; i++ {
		id := <-done

		if order[i] != id {
			t.Errorf("job %d executed before job %d", id, order[i])
		}
	}
}

// func TestAddTimeSchedule(t *testing.T) {

// 	wg := sync.WaitGroup{}
// 	wg.Add(2)

// 	now := time.Now().UTC()
// 	start := now.Add(1000 * time.Millisecond)
// 	end := start.Add(500 * time.Millisecond)

// 	ts := &automations.TimeSchedule{
// 		Start: start.Format("15:04:05"),
// 		End:   end.Format("15:04:05"),
// 	}
// 	s := automations.NewScheduler(context.Background())

// 	// add start job
// 	err := s.AddJob(ts.Start, func() error {
// 		fmt.Println("Start job executed")
// 		wg.Done()

// 		return nil
// 	})

// 	if err != nil {
// 		t.Errorf("failed to add start job %s", err.Error())
// 	}

// 	// add end job
// 	err = s.AddJob(ts.End, func() error {
// 		fmt.Println("End job executed")
// 		wg.Done()

// 		return nil
// 	})
// 	if err != nil {
// 		t.Errorf("failed to add end job %s", err.Error())
// 	}

// 	s.Start()

// 	if !waitTimeout(&wg, 5*time.Second) {
// 		t.Errorf("failed to execute jobs")
// 	}
// }

// func TestStopTimeSchedule(t *testing.T) {

// 	now := time.Now().UTC()
// 	start := now.Add(2000 * time.Millisecond)
// 	end := start.Add(2000 * time.Millisecond)

// 	ts := &automations.TimeSchedule{
// 		Start: start.Format("15:04:05"),
// 		End:   end.Format("15:04:05"),
// 	}
// 	s := automations.NewScheduler(context.Background())

// 	// add start job
// 	err := s.AddJob(ts.Start, func() error {
// 		t.Errorf("Start job executed")

// 		return nil
// 	})

// 	if err != nil {
// 		t.Errorf("failed to add start job %s", err.Error())
// 	}

// 	// add end job
// 	err = s.AddJob(ts.End, func() error {
// 		t.Errorf("End job executed")

// 		return nil
// 	})
// 	if err != nil {
// 		t.Errorf("failed to add end job %s", err.Error())
// 	}

// 	s.Start()

// 	time.Sleep(200 * time.Millisecond)

// 	err = s.Stop()

// 	if err != nil {
// 		t.Errorf("failed to stop scheduler %s", err.Error())
// 	}
// }

// func TestTimeScheduleContextCancellation(t *testing.T) {

// 	ctx, cancel := context.WithCancel(context.Background())
// 	now := time.Now().UTC()
// 	start := now.Add(2000 * time.Millisecond)
// 	end := start.Add(2000 * time.Millisecond)

// 	ts := &automations.TimeSchedule{
// 		Start: start.Format("15:04:05"),
// 		End:   end.Format("15:04:05"),
// 	}
// 	s := automations.NewScheduler(ctx)

// 	// add start job
// 	err := s.AddJob(ts.Start, func() error {
// 		t.Errorf("Start job executed")

// 		return nil
// 	})

// 	if err != nil {
// 		t.Errorf("failed to add start job %s", err.Error())
// 	}

// 	// add end job
// 	err = s.AddJob(ts.End, func() error {
// 		t.Errorf("End job executed")

// 		return nil
// 	})
// 	if err != nil {
// 		t.Errorf("failed to add end job %s", err.Error())
// 	}

// 	s.Start()

// 	time.Sleep(500 * time.Millisecond)

// 	cancel()

// 	<-ctx.Done()
// }

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
