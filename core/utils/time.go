package utils

import (
	"time"
)

const (
	UnitMilliseconds = "milliseconds"
	UnitSeconds      = "seconds"
	UnitMinutes      = "minutes"
	UnitHours        = "hours"
	UnitDays         = "days"
)

type Clock interface {
	Now() time.Time
	Sleep(duration time.Duration)
	CompareWithNow(t time.Time, operator EqualityOperator) (bool, error)
	IsInRange(start time.Time, end time.Time) bool
}

type RealClock struct{}

func NewRealClock() Clock {
	return &RealClock{}
}

func (r *RealClock) CompareWithNow(t time.Time, operator string) (bool, error) {

	now := r.Now()
	toTime := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	return CompareTimeRange(now, toTime, operator)
}

func (r *RealClock) IsInRange(from time.Time, to time.Time) bool {
	now := r.Now()
	fromTime := time.Date(now.Year(), now.Month(), now.Day(), from.Hour(), from.Minute(), from.Second(), from.Nanosecond(), from.Location())
	toTime := time.Date(now.Year(), now.Month(), now.Day(), to.Hour(), to.Minute(), to.Second(), to.Nanosecond(), to.Location())

	res1, _ := CompareTimeRange(now, fromTime, ">=")
	res2, _ := CompareTimeRange(now, toTime, "<=")

	return res1 && res2
}

func CompareTimeRange(from time.Time, to time.Time, operator EqualityOperator) bool {

	switch operator {
	case Equals:
		return from.Equal(to)
	case LessThan:
		return from.Before(to)
	case LessThanEqual:
		return from.Before(to) || from.Equal(to)
	case GreaterThan:
		return from.After(to)
	case GreaterThanEqual:
		return from.After(to) || from.Equal(to)

	}
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

func IntervalFromMilliseconds(value int) *TimeInterval {
	return newTimeInterval(value, UnitMilliseconds)
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
	case UnitMilliseconds:
		return time.Duration(d.Value) * time.Millisecond
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
