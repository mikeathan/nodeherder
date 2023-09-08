package automations

import (
	"fmt"
	"node-herder/utils"
	"time"
)

var Equalityoperators = map[string]func(any, any) bool{
	"=": func(v1 any, v2 any) bool {
		return v1 == v2
	},
	">=": func(v1 any, v2 any) bool {
		return toFloat(v1) >= toFloat(v2)
	},
	"<=": func(v1 any, v2 any) bool {
		return toFloat(v1) <= toFloat(v2)
	},
	">": func(v1 any, v2 any) bool {
		return toFloat(v1) > toFloat(v2)
	},
	"<": func(v1 any, v2 any) bool {
		return toFloat(v1) < toFloat(v2)
	},
}

type DeviceCondition struct {
	Friendlyname     string      `json:"friendlyname"`
	Type             string      `json:"type"`
	Value            any         `json:"value"`
	Constraint       Constraint  `json:"constraint"` // TODO
	Action           *MqttAction `json:"action"`
	EqualityOperator string      `json:"equalityoperator"`
	cache            map[string]any
}

func NewDeviceCondition() *DeviceCondition {
	return &DeviceCondition{cache: make(map[string]any), EqualityOperator: "="}
}

func (m *DeviceCondition) getValue(sensor string) (any, bool) {
	if value, ok := m.cache[sensor]; ok {
		return value, true
	}

	return nil, false
}

func (m *DeviceCondition) isConditionMatched(data map[string]any) bool {
	value, ok := data[m.Type]
	if !ok {
		utils.LogDebugf("sensor type %s not in input payload \n", m.Type)
		return false
	}

	return Equalityoperators[m.EqualityOperator](m.Value, value)
}

func (m *DeviceCondition) isConditionMatchedFromCache() bool {
	value, ok := m.cache[m.Type]
	if !ok {
		utils.LogDebugf("sensor type %s not in cached  payload \n", m.Type)
		return false
	}

	return Equalityoperators[m.EqualityOperator](m.Value, value)
}

func (m *DeviceCondition) Evaluate(data map[string]any) {

	m.cache = data // cache any values, we need them for constraints

	if m.Constraint != nil {
		m.Constraint.Evaluate(m)

	} else if m.isConditionMatched(data) {
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
