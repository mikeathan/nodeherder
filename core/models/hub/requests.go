package hub

import (
	"errors"
	"node-herder/utils"
	"time"
)

type Request interface {
	ID() string
	Process(value any) error
}

type ActiveStateTimerRequest struct {
	id       string
	timer    *utils.ActiveStateTimer
	duration time.Duration
}

func NewActiveStateTimerRequest(id string, duration time.Duration, timer *utils.ActiveStateTimer) Request {
	return &ActiveStateTimerRequest{id: id, duration: duration, timer: timer}
}

func (r *ActiveStateTimerRequest) ID() string {
	return r.id
}

func (r *ActiveStateTimerRequest) Process(value any) error {

	if active, ok := value.(bool); ok {
		if active {
			return r.timer.Start(r.duration)
		}
		return r.timer.Stop(true)
	}

	return errors.New("invalid value type")
}
