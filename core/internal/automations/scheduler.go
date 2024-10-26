package automations

import (
	"context"
	"fmt"
	"node-herder/utils"
	"time"

	"github.com/go-co-op/gocron"
)

type TimeSchedule struct {
	Enabled bool   `json:"enabled"`
	Start   string `json:"start"`
	End     string `json:"end"`
}

type Scheduler struct {
	startTime     time.Time
	endTime       time.Time
	enabled       bool
	ctx           context.Context
	cronScheduler *gocron.Scheduler
}

func NewScheduler(timeSchedule *TimeSchedule, ctx context.Context) (*Scheduler, error) {
	start, err := parserTime(timeSchedule.Start)
	if err != nil {
		utils.LogError("Error parsing start time:", err)
		return nil, err
	}

	end, err := parserTime(timeSchedule.End)
	if err != nil {
		utils.LogError("Error parsing end time:", err)
		return nil, err
	}

	return &Scheduler{
		startTime:     start,
		endTime:       end,
		enabled:       false,
		ctx:           ctx,
		cronScheduler: gocron.NewScheduler(time.UTC),
	}, nil
}

// TODO:
// move to utils
func parserTime(timeString string) (time.Time, error) {
	layout := "15:04:05"
	t, err := time.Parse(layout, timeString)

	if err != nil {
		fmt.Println("Error parsing time:", err)
		return time.Time{}, err
	}

	return t, nil
}

func (s *Scheduler) SetEnabled(enabled bool) {
	s.enabled = enabled
}
func (s *Scheduler) IsEnabled() bool {
	return s.enabled
}

func (s *Scheduler) Start() {

	// TODO: use enabled/disabled logic

	if s.cronScheduler.IsRunning() {
		utils.LogInfo("Scheduler already running")
		return
	}
	// configure start job
	_, err := s.cronScheduler.Every(1).Day().At(s.startTime).Do(func() {
		utils.LogInfo("Daily job executed at ", s.startTime)
	})

	if err != nil {
		utils.LogError("Error scheduling start job:", err)
		return
	}

	// configure end job
	_, err = s.cronScheduler.Every(1).Day().At(s.endTime).Do(func() {
		utils.LogInfo("Daily job executed at ", s.endTime)
	})
	if err != nil {
		utils.LogError("Error scheduling end job:", err)
		return
	}

	go func() {
		defer s.Stop()

		utils.LogInfo("Scheduler started")
		s.cronScheduler.StartAsync()

		select {
		case <-s.ctx.Done():
			utils.LogInfo("Scheduler context cancellation")
			return
		}

	}()

}

func (s *Scheduler) Stop() {

	if !s.cronScheduler.IsRunning() {
		utils.LogInfo("Scheduler is not running")
		return
	}

	s.cronScheduler.Stop()
	utils.LogInfo("Scheduler stopped")
}

// type DaySchedule struct {
// 	Start int `json:"start"`
// 	End   int `json:"end"`
// }
// day schedule
// start day
// end day

// time schedule
// start time
// end time
