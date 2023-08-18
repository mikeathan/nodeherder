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

func Load(bridgeDevices []*bridge.BridgeDevice, mqtClient mqtt.MqttClient) error {

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

	return nil
}

func createMockConfiguration() *configuration {

	trigger := newTimerTrigger()
	trigger.Type = "timer"
	trigger.Name = "Turn on light"

	// new condition
	tc := &TimeDurationCondition{}
	tc.Duration = 20 * time.Second
	trigger.Conditions = append(trigger.Conditions, tc)

	// new action
	ma := &MqttAction{}
	ma.client = nil // todo: setup client

	ma.Friendlyname = "Hive light 1"
	ma.Property = "state"
	ma.Type = "light"
	ma.Value = false

	trigger.Actions = append(trigger.Actions, ma)

	conf := newConfiguration()
	conf.triggers = append(conf.triggers, trigger)
	return conf
}
