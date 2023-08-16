package automations

import (
	"fmt"
	"time"
)

func (t SceduleTrigger) Trigger() error {

	fmt.Printf("Trigger Action: %s\n", t.action)
	return nil
}

func (t *TimerTrigger) Process() error {

	for {

		diff := time.Until(t.cond.GetSchedule()).Seconds()
		ticker := *time.NewTicker(time.Duration(diff) * time.Second)
		<-ticker.C

		err := t.Trigger()
		if err != nil {
			fmt.Println(err)
		}

		if !t.cond.IsRepeat() {
			fmt.Println("timer exit")
			return nil
		}
	}
}
