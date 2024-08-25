package utils

import "time"

type Clock interface {
	Now() time.Time
	Sleep(duration time.Duration)
}

type RealClock struct{}

func (r *RealClock) Now() time.Time {
	return time.Now().UTC()
}

func (r *RealClock) Sleep(duration time.Duration) {
	time.Sleep(duration)
}
