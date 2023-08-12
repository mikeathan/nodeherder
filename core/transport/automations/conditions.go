package automations

import (
	"fmt"
	"time"
)

type Condition interface {
}

type TimerCondition struct {
	_timestamp time.Time
	_duration  time.Duration
	Repeat     bool
	IsDateTime bool
}

func FromDuration(d time.Duration) *TimerCondition {
	c := &TimerCondition{
		_duration: d,
	}
	return c
}

func FromTime(t time.Time) *TimerCondition {
	c := &TimerCondition{
		_timestamp: t,
		IsDateTime: true,
	}
	return c
}

func (tc *TimerCondition) GetSchedule() time.Time {
	if !tc.IsDateTime {
		tc._timestamp = time.Now().Add(tc._duration)
	} else {
		if time.Now().Sub(tc._timestamp) < 0 {
			tc._timestamp = getTomorrow(tc._timestamp)
		}
	}

	fmt.Printf("Sceduled for %v \n", tc._timestamp.Format(time.RFC3339))
	return tc._timestamp
}

func getTomorrow(ts time.Time) time.Time {
	return time.Date(ts.Year(), ts.Month(), ts.Day()+1, ts.Hour(), ts.Minute(), 0, 0, ts.Location())
}
