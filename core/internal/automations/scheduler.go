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

func NewTimeSchedule() *TimeSchedule {
	return &TimeSchedule{
		Enabled: false,
		Start:   "",
		End:     "",
	}
}

type Scheduler struct {
	enabled       bool
	ctx           context.Context
	cronScheduler *gocron.Scheduler
	jobs          []*gocron.Job
}

func NewScheduler(ctx context.Context) *Scheduler {

	return &Scheduler{
		enabled:       false,
		ctx:           ctx,
		cronScheduler: gocron.NewScheduler(time.UTC),
		jobs:          []*gocron.Job{},
	}
}

func (s *Scheduler) AddJob(timeString string, action func() error) error {
	time, err := parserTime(timeString)
	if err != nil {
		utils.LogError("Error parsing start time:", err)
		return err
	}

	name := fmt.Sprintf("Job %s", timeString)
	job, err := s.cronScheduler.Name(name).Every(1).Day().At(time).Do(func() {
		select {

		case <-s.ctx.Done():

			utils.LogInfo("Scheduler context cancel requested")
			s.Stop()

			return

		default:
			utils.LogInfo("Executing job ", time)

			err = action()
			if err != nil {
				utils.LogErrorf("Job %s failed: %s", timeString, err.Error())
			}
		}

	})

	if err != nil {
		utils.LogErrorf("Job %s failed to schedule: %s", timeString, err.Error())
		return err
	}
	s.jobs = append(s.jobs, job)

	return nil
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
	if len(s.jobs) == 0 {
		utils.LogInfo("No jobs scheduled")
		return
	}

	s.cronScheduler.StartAsync()

	utils.LogInfo("Scheduler started")
}

func (s *Scheduler) Stop() error {

	if !s.cronScheduler.IsRunning() {
		return fmt.Errorf("scheduler is not running")
	}

	if len(s.jobs) == 0 {
		return fmt.Errorf("no jobs scheduled")
	}

	s.cronScheduler.Stop()

	for _, j := range s.jobs {
		if j.IsRunning() {
			return fmt.Errorf("job %s is still running", j.Error())
		}

	}

	utils.LogInfo("Scheduler stopped")

	return nil
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
