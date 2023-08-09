package automation

import (
	"fmt"
	"time"
)

type Trigger interface {
	Name() string
	WithCondition(cond string)
	WithAction(action string)
	Process() error
	Trigger() error
}

type TimerTrigger struct {
	cond   string
	action string
}

func (t TimerTrigger) Trigger() error {

	fmt.Printf("Trigger Action: %s\n", t.action)
	return nil
}

func (t TimerTrigger) Name() string {
	return "Timer"
}

func (t *TimerTrigger) WithCondition(cond string) {
	t.cond = cond
}
func (t *TimerTrigger) WithAction(action string) {
	t.action = action
}

func (t *TimerTrigger) Process() error {
	ticker := *time.NewTicker(1 * time.Second)
	<-ticker.C

	return t.Trigger()
}
