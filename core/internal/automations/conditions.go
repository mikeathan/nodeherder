package automations

import (
	"encoding/json"
	"errors"
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
	Friendlyname     string          `json:"friendlyname"`
	Type             string          `json:"type"`
	Value            any             `json:"value"`
	Constraint       Constraint      `json:"-"`
	RawConstraint    json.RawMessage `json:"constraint"`
	Action           *MqttAction     `json:"action"`
	EqualityOperator string          `json:"equalityoperator"`
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
		utils.LogInfo("Condition evaluated")
		err := m.Action.Run()
		if err != nil {
			utils.LogErrorf("Action failed %s", err.Error())
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

	utils.LogDebugf("TimeDurationCondition scheduled for %v \n", tc.Timestamp.Format(time.RFC3339))
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

func (d *DeviceCondition) MarshalJSON() ([]byte, error) {
	buffer, err := json.Marshal(&d.Constraint)
	if err != nil {
		return nil, err
	}

	deviceCond := struct {
		Friendlyname     string          `json:"friendlyname"`
		Type             string          `json:"type"`
		Value            any             `json:"value"`
		Constraint       Constraint      `json:"-"`
		RawConstraint    json.RawMessage `json:"constraint"`
		Action           *MqttAction     `json:"action"`
		EqualityOperator string          `json:"equalityoperator"`
	}{
		Friendlyname:     d.Friendlyname,
		Type:             d.Type,
		Value:            d.Value,
		RawConstraint:    json.RawMessage(buffer), // fill in raw constraint
		Action:           d.Action,
		EqualityOperator: d.EqualityOperator,
	}

	return json.Marshal(deviceCond)
}

func (d *DeviceCondition) UnmarshalJSON(b []byte) error {
	var deviceCond = struct {
		Friendlyname     string          `json:"friendlyname"`
		Type             string          `json:"type"`
		Value            any             `json:"value"`
		Constraint       Constraint      `json:"-"`
		RawConstraint    json.RawMessage `json:"constraint"`
		Action           *MqttAction     `json:"action"`
		EqualityOperator string          `json:"equalityoperator"`
	}{}

	err := json.Unmarshal(b, &deviceCond)
	if err != nil {
		return err
	}

	var rowConstraint map[string]interface{}
	err = json.Unmarshal(deviceCond.RawConstraint, &rowConstraint)
	if err != nil {
		return err
	}

	switch rowConstraint["type"] {
	case "timer":
		var tc = NewTimerConstraint()
		err = json.Unmarshal(deviceCond.RawConstraint, &tc)
		if err != nil {
			return err
		}
		deviceCond.Constraint = tc
	case "device":
		var dc = NewDeviceConstraint()
		err = json.Unmarshal(deviceCond.RawConstraint, &dc)
		if err != nil {
			return err
		}
		deviceCond.Constraint = dc
	default:
		return errors.New("unknown constraint type")
	}

	d.Friendlyname = deviceCond.Friendlyname
	d.Type = deviceCond.Type
	d.EqualityOperator = deviceCond.EqualityOperator
	d.Action = deviceCond.Action
	d.Constraint = deviceCond.Constraint
	d.Value = deviceCond.Value
	return nil
}
