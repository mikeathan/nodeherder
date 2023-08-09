package automation

import "errors"

type Trigger interface {
	Trigger() error
	Name() string
}

type TimerTrigger struct {
}

func (t TimerTrigger) Trigger() error {
	return errors.New("not implemented")
}

func (t TimerTrigger) Name() string {
	return "Timer"
}
