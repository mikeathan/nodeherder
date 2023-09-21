package automations

import (
	"node-herder/utils"
)

func toFloat(value any) float32 {
	switch v := value.(type) {
	case int:
		return float32(v)
	case float64:
		return float32(v)
	case float32:
		return float32(v)
	default:
		return float32(0)
	}
}

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

	if Equalityoperators[s.EqualityOperator](value, s.Value) {
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
	Name         string             `json:"name"`
	Conditions   []*SensorCondition `json:"conditions"`
	ActionRunner *MqttAction        `json:"action"`
}

func newSensorTrigger(name string) *SensorTrigger {
	return &SensorTrigger{}
}
