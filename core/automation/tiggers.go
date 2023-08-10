package automation

import (
	"fmt"
	"time"
)

type TimerCondition struct {
	Timestamp time.Time
	Repeat    bool
}

type Trigger interface {
	Name() string
	WithCondition(cond string)
	WithAction(action string)
	Process() error
	Trigger() error
}

type TimerTrigger struct {
	cond   TimerCondition
	action string
}

func (t TimerTrigger) Trigger() error {

	fmt.Printf("Trigger Action: %s\n", t.action)
	return nil
}

func (t TimerTrigger) Name() string {
	return "Timer"
}

func (t *TimerTrigger) WithCondition(cond TimerCondition) {
	t.cond = cond
	fmt.Printf("Sceduled for %v \n", t.cond.Timestamp)
}

func (t *TimerTrigger) WithAction(action string) {
	t.action = action
}

func (t *TimerTrigger) Process() error {

	for {
		diff := time.Until(t.cond.Timestamp).Seconds()
		ticker := *time.NewTicker(time.Duration(diff) * time.Second)
		<-ticker.C

		err := t.Trigger()
		if err != nil {
			fmt.Println(err)
		}

		if !t.cond.Repeat {
			fmt.Println("timer exit")
			return nil
		}

		// timer needs resetting

		t.cond.Timestamp = getTomorrow(t.cond.Timestamp)
		fmt.Printf("Sceduled for %v \n", t.cond.Timestamp)
	}

}

func getTomorrow(ts time.Time) time.Time {
	return time.Date(ts.Year(), ts.Month(), ts.Day()+1, ts.Hour(), ts.Minute(), 0, 0, ts.Location())
}
