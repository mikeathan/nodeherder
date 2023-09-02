package automations

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

type TimerConstraint struct {
	Duration time.Duration `json:"duration"`
	stop     chan bool
	mut      sync.RWMutex
	sem      *semaphore.Weighted
}

func NewTimerConstraint() *TimerConstraint {
	return &TimerConstraint{stop: make(chan bool), sem: semaphore.NewWeighted(1)}

}

type DeviceConstraint struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

func (c *DeviceConstraint) Evaluate() {

}

func (c *DeviceConstraint) Reset() {

}

func (t *TimerConstraint) Reset() {
	go func() {
		defer t.mut.Unlock()

		t.mut.Lock()
		t.stop <- true
	}()
}

func (t *TimerConstraint) run(procFunc func()) {

	defer t.mut.Unlock()
	t.mut.Lock()

	// check if its already running
	if ok := t.sem.TryAcquire(1); !ok {
		fmt.Println("time constraint is running")
		return
	}

	go func() {

		fmt.Println("time constraint started")

		ticker := *time.NewTicker(getDurationFromNow(t) * time.Millisecond)

		defer func() {
			t.sem.Release(1)
			ticker.Stop()
		}()

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

}

func getDurationFromNow(t *TimerConstraint) time.Duration {

	timestamp := time.Now().Add(t.Duration)
	diff := time.Until(timestamp).Milliseconds()

	return time.Duration(diff)
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

	value, ok := data[m.Type]
	if !ok {
		fmt.Printf("[DEBUG] sensor type %s not in input payload \n", m.Type)
		return
	}

	m.cache[m.Type] = value // cache any values, we might use them for any constraints

	if value == m.Value {

		if m.Constraint != nil {
			timer, ok := m.Constraint.(*TimerConstraint)
			if ok {
				procFunc := func() {
					m.Action.Run()
				}
				timer.run(procFunc)
				return
			}
			deviceContraint, ok := m.Constraint.(*DeviceConstraint)
			if ok {
				if sensorValue, ok := m.cache[deviceContraint.Type]; ok {
					if deviceContraint.Value != sensorValue {
						return
					}
				}
			}
		}

		err := m.Action.Run()
		if err != nil {
			fmt.Printf("device sensor action failed %s", err.Error())
		}
	} else {

		// reset any existing constraint state
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
