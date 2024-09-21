package utils

import "time"

const (
	UnitSeconds = "seconds"
	UnitMinutes = "minutes"
	UnitHours   = "hours"
	UnitDays    = "days"
)

type Clock interface {
	Now() time.Time
	Sleep(duration time.Duration)
}

type RealClock struct{}

func NewRealClock() Clock {
	return &RealClock{}
}

func (r *RealClock) Now() time.Time {
	return time.Now().UTC()
}

func (r *RealClock) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

type TimeInterval struct {
	Value int    `json:"value"`
	Unit  string `json:"unit"`
}

func newTimeInterval(value int, unit string) *TimeInterval {
	return &TimeInterval{
		Value: value,
		Unit:  unit,
	}
}

func IntervalFromSeconds(value int) *TimeInterval {
	return newTimeInterval(value, UnitSeconds)
}

func IntervalFromMinutes(value int) *TimeInterval {
	return newTimeInterval(value, UnitMinutes)
}

func IntervalFromHours(value int) *TimeInterval {
	return newTimeInterval(value, UnitHours)
}

func IntervalFromDays(value int) *TimeInterval {
	return newTimeInterval(value, UnitDays)
}

func (d *TimeInterval) Duration() time.Duration {
	switch d.Unit {
	case UnitSeconds:
		return time.Duration(d.Value) * time.Second
	case UnitMinutes:
		return time.Duration(d.Value) * time.Minute
	case UnitHours:
		return time.Duration(d.Value) * time.Hour
	case UnitDays:
		return time.Duration(d.Value) * time.Hour * 24
	default:
		return 0
	}
}
