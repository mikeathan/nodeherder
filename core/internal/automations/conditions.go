package automations

import (
	"fmt"
	"time"
)

// todo: new trigger for device which will hold state

type TimerConstrain struct {
	Duration time.Duration `json:"duration"`
	procFunc func()
	reset    chan bool
	exit     chan bool
}

func (t *TimerConstrain) run() {

	t.reset <- true

	go func() {
		defer close(t.exit)
		defer close(t.reset)

		for {

			timestamp := time.Now().Add(t.Duration)
			diff := time.Until(timestamp).Seconds()
			ticker := *time.NewTicker(time.Duration(diff) * time.Second)
			select {
			case <-ticker.C:
				// do stuff
			case <-t.reset:
				fmt.Println("resetting timer constrain")
				continue
			case <-t.exit:
				fmt.Println("stopping timer constrain")
				return
			}
		}
	}()
}

type MqttCondition struct {
	Friendlyname string        `json:"friendlyname"`
	Type         string        `json:"type"`
	Value        any           `json:"value"`
	Constrains   []interface{} `json:"constrains"` // TODO
	Action       *MqttAction   `json:"action"`
	cache        map[string]any
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

	if value == m.Value {

		// for _, c := range m.Constrains {
		// 	timer, ok := c.(TimerConstrain)
		// 	if ok {
		// 		timer.procFunc = func() {

		// 			// make sure value hasnt changed while we are waiting
		// 			if m.cache[m.Type] == m.Value {
		// 				m.Action.Run()
		// 			}
		// 		}
		// 		timer.run()
		// 		return //????
		// 	}
		// }

		err := m.Action.Run()
		if err != nil {
			fmt.Printf("device sensor action failed %s", err.Error())
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
