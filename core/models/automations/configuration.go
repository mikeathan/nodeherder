package automations

import (
	"errors"
	"node-herder/internal/mqtt"
	"node-herder/models/devices"
	"time"
)

type MqttTrigger struct {
	Type      string      `json:"trigger"`
	Name      string      `json:"name"`
	Condition interface{} `json:"condition"`
	Action    *MqttAction `json:"action"`
	Enabled   bool        `json:"enabled"`
}

func newMqttTrigger() *MqttTrigger {
	return &MqttTrigger{
		Type:    "mqtt",
		Name:    "",
		Action:  &MqttAction{},
		Enabled: true,
	}
}

type TimerTrigger struct {
	Type       string        `json:"trigger"`
	Name       string        `json:"name"`
	Conditions []interface{} `json:"conditions"`
	Actions    []*MqttAction `json:"actions"`
	Enabled    bool          `json:"enabled"`
}

func newTimerTrigger() *TimerTrigger {
	return &TimerTrigger{
		Type:       "timer",
		Name:       "",
		Conditions: []interface{}{},
		Actions:    []*MqttAction{},
		Enabled:    true,
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

	return nil
}

func newMockMqttTrigger() *MqttTrigger {

	trigger := newMqttTrigger()
	trigger.Name = "Turn on light 1 mqtt automation"
	trigger.Enabled = false

	// new condition
	mcOn := &MqttCondition{}
	mcOn.Friendlyname = "Human presence"
	mcOn.Type = "presence"
	mcOn.Value = true

	// new action
	ma := &MqttAction{}
	ma.client = nil

	ma.Friendlyname = "Attic light"
	ma.Property = "state"
	ma.Type = "light"
	ma.Value = false

	//////////////////////////////////////////////////////////////
	trigger2 := newMqttTrigger()
	trigger2.Name = "Turn off light 1 mqtt automation"
	trigger2.Enabled = false

	mcOff := &MqttCondition{}
	mcOff.Friendlyname = "Human presence"
	mcOff.Type = "presence"
	mcOn.Value = false

	// TODO:
	// add timer condition

	//example sensor says falss and start timer, after eg 15 min call action to turn off light

	// new action
	ma2 := &MqttAction{}
	ma2.client = nil

	ma2.Friendlyname = "Attic light"
	ma2.Property = "state"
	ma2.Type = "light"
	ma2.Value = false

	////////////////////////////////////////////////////

	return nil
}
func createMockConfiguration() *configuration {

	trigger := newTimerTrigger()
	trigger.Name = "Turn on light 1 timer automation"
	trigger.Enabled = false

	// new condition
	tc := &TimeDurationCondition{}
	tc.Duration = 5 * time.Second
	trigger.Conditions = append(trigger.Conditions, tc)

	// new action
	ma := &MqttAction{}
	ma.client = nil

	ma.Friendlyname = "Attic light"
	ma.Property = "state"
	ma.Type = "light"
	ma.Value = false

	trigger.Actions = append(trigger.Actions, ma)

	conf := newConfiguration()
	conf.triggers = append(conf.triggers, trigger)
	return conf
}
