package automations

import (
	"errors"
	"fmt"
	"node-herder/models/bridge"
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

	config := createMockConfiguration()

	for _, trigger := range config.triggers {
		switch trigger.(type) {
		case TimerTrigger:
			break
		default:
			return nil, errors.New("unsupported trigger type ")
		}
		timerTrigger := trigger.(TimerTrigger)
		fmt.Println("loading:", timerTrigger.Name)
		if timerTrigger.Type != "timer" {
			return nil, errors.New("unsupported trigger type ")
		}
		// valdate conditions
		for _, cond := range timerTrigger.Conditions {
			switch cond.(type) {
			case TimeDurationCondition:
				break
			case TimestampCondition:
				break
			default:
				return nil, errors.New("unsupported condition type ")
			}
		}

		//st,_ok:=trigger.(SceduleTrigger)
		// validate actions
		for _, action := range timerTrigger.Actions {
			for _, device := range bridgeDevices {
				if device.FriendlyName == action.Friendlyname {

					if action.Type != "light" {
						return nil, errors.New("unsupported action type ")
					}

					if device.Disabled {
						return nil, errors.New("device is disabled ")
					}
					action.client = nil // TODO: setup
				}
			}
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
