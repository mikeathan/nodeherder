package automations

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
)

type TimeSchedule struct {
	Enabled bool   `json:"enabled"`
	Start   string `json:"start"`
	End     string `json:"end"`
}

type Scheduler struct {
	start   time.Time
	end     time.Time
	enabled bool
}

func NewScheduler(timeSchedule *TimeSchedule) (*Scheduler, error) {

	start, err := parserTime(timeSchedule.Start)
	if err != nil {
		fmt.Println("Error parsing start time:", err)
		return nil, err
	}

	end, err := parserTime(timeSchedule.End)
	if err != nil {
		fmt.Println("Error parsing end time:", err)
		return nil, err
	}
	return &Scheduler{start: start, end: end}, nil
}

// move to utils
func parserTime(timeString string) (time.Time, error) {
	layout := "15:04"
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

	s, err := gocron.NewScheduler()
	if err != nil {
		// handle error
	}

	TODO
}

func (s *Scheduler) Stop() {
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
