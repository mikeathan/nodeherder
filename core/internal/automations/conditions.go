package automations

import (
	"fmt"
	"sync"
	"time"
)

// todo: new trigger for device which will hold state

type TimerConstraint struct {
	Duration time.Duration `json:"duration"`
	stop     chan bool
	mut      sync.RWMutex
}

func NewTimerConstraint() *TimerConstraint {
	return &TimerConstraint{stop: make(chan bool)}

}
func (t *TimerConstraint) Reset() {
	go func() {
		t.mut.Lock()
		t.stop <- true
		t.mut.Unlock()
	}()
}

func (t *TimerConstraint) run(procFunc func()) {

	t.mut.Lock()
	go func() {

		fmt.Println("time constraint started")

		timestamp := time.Now().Add(t.Duration)
		diff := time.Until(timestamp).Seconds()
		ticker := *time.NewTicker(time.Duration(diff) * time.Second)

		defer ticker.Stop()

		select {
		case <-ticker.C:

			procFunc()
			fmt.Println("timer constraint finished")
			return

		case <-t.stop:

			fmt.Println("timer constraint stopped")
			return
		}
	}()
	t.mut.Unlock()
}

type MqttCondition struct {
	Friendlyname string      `json:"friendlyname"`
	Type         string      `json:"type"`
	Value        any         `json:"value"`
	Constraint   interface{} `json:"constraint"` // TODO
	Action       *MqttAction `json:"action"`
	cache        map[string]any
}

func NewMqttCondition() *MqttCondition {
	return &MqttCondition{cache: make(map[string]any)}
}

func (m *MqttCondition) Evaluate(data map[string]any) {

	// TOOD:
	value, ok := data[m.Type]
	if !ok {
		fmt.Printf("[DEBUG] sensor type %s not exists \n", m.Type)
		return
	}
	m.cache[m.Type] = value
	// TODO: validate
	// m.Constrains

	// example
	// if presence if false AND timeout is 10 min => execute action = turn off light
	//
	// presence = false, start timer,  end of timer, chech if values is till same => execute action
	// presence = true, stop timer
	// presence = false, while time is running (shoulnt happen)

	if value == m.Value {

		if m.Constraint != nil {
			timer, ok := m.Constraint.(*TimerConstraint)
			if ok {
				procFunc := func() {
					// make sure value hasnt changed while we are waiting, dont need that ????
					if m.cache[m.Type] == m.Value {
						m.Action.Run()
					}
				}
				timer.run(procFunc)
				return
			}
		}

		err := m.Action.Run()
		if err != nil {
			fmt.Printf("device sensor action failed %s", err.Error())
		}
	} else {

		// reset any state we might have set during a previous match
		if m.Constraint != nil {
			timer, ok := m.Constraint.(*TimerConstraint)
			if ok {
				timer.Reset()
			}
		}
	}
}

type TimerCondition interface {
	GetSchedule() time.Time
	IsRepeat() bool
}

type TimestampCondition struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Repeat    bool      `json:"repeat"`
}

type TimeDurationCondition struct {
	Type      string        `json:"type"`
	Duration  time.Duration `json:"duration"`
	Repeat    bool          `json:"repeat"`
	Timestamp time.Time
}

func (tc *TimeDurationCondition) GetSchedule() time.Time {
	tc.Timestamp = time.Now().Add(tc.Duration)

	fmt.Printf("Sceduled for %v \n", tc.Timestamp.Format(time.RFC3339))
	return tc.Timestamp
}

func (tc *TimestampCondition) GetSchedule() time.Time {

	if time.Since(tc.Timestamp) < 0 {
		tc.Timestamp = getTomorrow(tc.Timestamp)
	}

	fmt.Printf("Sceduled for %v \n", tc.Timestamp.Format(time.RFC3339))
	return tc.Timestamp
}

func (tc *TimeDurationCondition) IsRepeat() bool {
	return tc.Repeat
}

func (tc *TimestampCondition) IsRepeat() bool {
	return tc.Repeat
}

func getTomorrow(ts time.Time) time.Time {
	return time.Date(ts.Year(), ts.Month(), ts.Day()+1, ts.Hour(), ts.Minute(), 0, 0, ts.Location())
}
