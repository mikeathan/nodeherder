package automations

import (
	"errors"
	"node-herder/models/bridge"
	"node-herder/transport/mqtt"
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

type Loader struct {
}

func Load(bridgeDevices []*bridge.BridgeDevice) (*Loader, error) {

	var mqtClient mqtt.MqttClient // todo: setup, this is just empty

	// fake input data
	config := createMockConfiguration()

	for _, trigger := range config.triggers {
		switch trigger.(type) {
		case TimerTrigger:
			break
		default:
			return nil, errors.New("unsupported trigger type ")
		}

		timerTrigger := trigger.(TimerTrigger)
		err := timerTrigger.configure(bridgeDevices, mqtClient)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func createMockConfiguration() *configuration {

	trigger := newTimerTrigger()
	trigger.Type = "timer"
	trigger.Name = "Turn on light"

	// new condition
	tc := &TimeDurationCondition{}
	tc.Duration = 5 * time.Second
	tc.Repeat = true
	trigger.Conditions = append(trigger.Conditions, tc)

	// new action
	ma := &MqttAction{}
	ma.client = nil // todo: setup client

	ma.Friendlyname = "Hive light 1"
	ma.Property = "state"
	ma.Type = "light"
	ma.Value = true

	trigger.Actions = append(trigger.Actions, ma)

	conf := newConfiguration()
	conf.triggers = append(conf.triggers, trigger)
	return conf
}
