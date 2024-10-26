package automations_test

import (
	"context"
	"fmt"
	"node-herder/internal/automations"
	"testing"
	"time"
)

func TestAddTimeSchedule(t *testing.T) {
	now := time.Now()
	start := now.Add(5 * time.Second)
	end := start.Add(5 * time.Second)

	startTime := start.Format("15:04:05")
	endTime := end.Format("15:04:05")

	fmt.Println(startTime, endTime)
	ts := &automations.TimeSchedule{
		Start: startTime,
		End:   endTime,
	}
	s, err := automations.NewScheduler(ts, context.Background())
	if err != nil {
		t.Errorf("expected %v, got %v", nil, err)
	}

	s.Start()

	time.Sleep(1 * time.Minute)

}
