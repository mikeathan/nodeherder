package automations

import (
	"errors"
	"node-herder/internal/mqtt"
	"node-herder/models/devices"
	"time"
)

type TimerTrigger struct {
	Type       string        `json:"trigger"`
	Name       string        `json:"name"`
	Conditions []interface{} `json:"conditions"`
	Actions    []*MqttAction `json:"actions"`
}

func newTimerTrigger() *TimerTrigger {
	return &TimerTrigger{
		Type:       "",
		Name:       "",
		Conditions: []interface{}{},
		Actions:    []*MqttAction{},
	}
}

type configuration struct {
	triggers []interface{}
}

func newConfiguration() *configuration {
	return &configuration{
		triggers: []interface{}{},
	}
}

func IsConfigured() bool {
	return _configured
}

var _configured = false

func Load(bridgeDevices []*devices.BridgeDevice, mqtClient mqtt.MqttClient) error {

	// fake input data
	config := createMockConfiguration()

	for _, trigger := range config.triggers {

		timerTrigger, ok := trigger.(*TimerTrigger)
		if !ok {
			return errors.New("unsupported trigger type ")
		}

		err := timerTrigger.configure(bridgeDevices, mqtClient)
		if err != nil {
			return err
		}
	}

	_configured = true
	return nil
}

func createMockConfiguration() *configuration {

	trigger := newTimerTrigger()
	trigger.Type = "timer"
	trigger.Name = "Turn on light"

	// new condition
	tc := &TimeDurationCondition{}
	tc.Duration = 5 * time.Second
	trigger.Conditions = append(trigger.Conditions, tc)

	// new action
	ma := &MqttAction{}
	ma.client = nil // todo: setup client

	ma.Friendlyname = "Attic light"
	ma.Property = "state"
	ma.Type = "light"
	ma.Value = false

	trigger.Actions = append(trigger.Actions, ma)

	conf := newConfiguration()
	conf.triggers = append(conf.triggers, trigger)
	return conf
}
