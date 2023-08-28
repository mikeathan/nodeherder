package automations

import (
	"fmt"
	"node-herder/models/devices"
	"time"
)

type MqttCondition struct {
	Friendlyname string        `json:"friendlyname"`
	Type         string        `json:"type"`
	Value        any           `json:"value"`
	Constrains   []interface{} `json:"constrains"` // TODO
}

func (m *MqttCondition) Evaluate(device *devices.Device) bool {

	value, ok := device.Sensors[m.Type]
	if !ok {
		fmt.Printf("type %s not exists in %s\n", m.Type, device.Id)
		return false
	}
	// TODO: validate
	// m.Constrains
	return value == m.Value
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
