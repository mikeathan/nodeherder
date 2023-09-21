package automations

import (
	"encoding/json"
	"errors"
	"node-herder/utils"
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
	cacheData        map[string]any
}

// examples
// sensor name = condition 1 & condiiton 1  && condtion n.... = action
// presence = (true) && (lux <= 30) = turn on
// presence = (false) && timer condition = turn off
// presence = true = turn on
// presence = false = turn off

type SensorCondition struct {
	Name             string `json:"name"`
	Value            any    `json:"value"`
	EqualityOperator string `json:"equalityoperator"`
}

func (s *SensorCondition) Evaluate(data map[string]any) bool {

	value, ok := data[s.Name]
	if !ok {
		utils.LogDebugf("sensor %s not found in payload", s.Name)
		return false
	}

	if Equalityoperators[s.EqualityOperator](s.Value, value) {
		return true
	}

	return false
}

type DeviceContext struct {
	data map[string]any
}

func (d *DeviceContext) Get(name string) any {
	return d.data[name]
}

func (d *DeviceContext) Set(name string, value any) {
	d.data[name] = value
}

func NewDeviceContext() *DeviceContext {
	return &DeviceContext{data: map[string]any{}}
}

type DeviceTrigger struct {
	Name           string                      `json:"name"`
	Description    string                      `json:"description"`
	Enabled        bool                        `json:"enabled"`
	SensorTriggers map[string][]*SensorTrigger `json:"sensor_triggers"`
	deviceContext  *DeviceContext
}

func NewDeviceTrigger(name string) *DeviceTrigger {
	return &DeviceTrigger{
		Name:           name,
		Description:    "",
		Enabled:        false,
		SensorTriggers: make(map[string][]*SensorTrigger),
		deviceContext:  NewDeviceContext(),
	}
}

func (d *DeviceTrigger) Evaluate(data map[string]any) bool {
	for sensor := range data {
		if triggers, ok := d.SensorTriggers[sensor]; ok {
			for _, trigger := range triggers {
				d.processTrigger(trigger, data)
			}
		}
	}
	return false
}

func (d *DeviceTrigger) processTrigger(trigger *SensorTrigger, data map[string]any) {

	currValue := d.deviceContext.Get(trigger.Name)
	for _, c := range trigger.Conditions {

		isMatched := c.Evaluate(data)
		if !isMatched {
			trigger.ActionRunner.Stop()
			return
		}

		// avoid calling action again for sensor if value hasnt changed
		if trigger.Name == c.Name && currValue == c.Value {
			return
		}
	}

	trigger.ActionRunner.Execute(func() {
		// on success callback
		// update sensor current value
		d.deviceContext.Set(trigger.Name, data[trigger.Name])
	})

}

type SensorTrigger struct {
	Name          string             `json:"name"`
	Conditions    []*SensorCondition `json:"-"`
	RawConditions json.RawMessage    `json:"conditions"`
	ActionRunner  *ActionRunner      `json:"action"`
}

func newSensorTrigger(name string) *SensorTrigger {
	return &SensorTrigger{}
}

func NewDeviceCondition() *DeviceCondition {
	return &DeviceCondition{cacheData: make(map[string]any), EqualityOperator: "="}
}

func (m *DeviceCondition) conditionWithConstraintIsMatched(c *DeviceConstraint) bool {

	return false
}

func (m *DeviceCondition) conditionFromCacheIsMatched() bool {

	return false
}

func (m *DeviceCondition) Evaluate(data map[string]any) {

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
