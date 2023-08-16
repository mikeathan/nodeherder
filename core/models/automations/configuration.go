package automations

import (
	"errors"
	"fmt"
	"node-herder/models/bridge"
	"time"
)

type timestampCondition struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Repeat    bool      `json:"repeat"`
}

type timeDurationCondition struct {
	Type     string        `json:"type"`
	Duration time.Duration `json:"duration"`
	Repeat   bool          `json:"repeat"`
}

type trigger struct {
	Type       string        `json:"trigger"`
	Name       string        `json:"name"`
	Conditions []interface{} `json:"conditions"`
	Actions    []*MqttAction `json:"actions"`
}

func newTrigger() *trigger {
	return &trigger{
		Type:       "",
		Name:       "",
		Conditions: []interface{}{},
		Actions:    []*MqttAction{},
	}
}

type configuration struct {
	triggers []*trigger
}

func newConfiguration() *configuration {
	return &configuration{
		triggers: []*trigger{},
	}
}

type Loader struct {
}

func Load(bridgeDevices []*bridge.BridgeDevice) (*Loader, error) {

	config := createMockConfiguration()

	for _, trigger := range config.triggers {
		fmt.Println("loading:", trigger.Name)
		if trigger.Type != "timer" {
			return nil, errors.New("unsupported trigger type ")
		}

		// load conditions
		var conds []*TimerCondition
		for _, cond := range trigger.Conditions {
			var ac *TimerCondition
			if td, ok := cond.(timeDurationCondition); ok {
				ac = TriggerFromDuration(td.Duration)
				ac.Repeat = td.Repeat
			} else if tt, ok := cond.(timestampCondition); ok {
				ac = TriggerFromTime(tt.Timestamp)
				ac.Repeat = tt.Repeat

			} else {
				return nil, errors.New("unsupported condition type ")
			}

			conds = append(conds, ac)
		}

		// load actions
		for _, action := range trigger.Actions {

			for _, device := range bridgeDevices {
				if device.FriendlyName == action.Friendlyname {

					// found device
					// find type and state

					if action.Type != "light" {
						return nil, errors.New("unsupported action type ")
					}

					if device.Disabled {
						return nil, errors.New("device is disabled ")
					}

					// mayb we don need all that
					// if we assume that config is correct just use the data
					// they should have been populated from us

					// for _, expose := range device.Definition.Exposes {

					// 	if expose.Type == action.Type {
					// 		for _, feature := range expose.Features {
					// 			if feature.Property == action.Name {
					// 				if action.Value == true {
					// 					valueOn := feature.ValueOn
					// 				} else {
					// 					valueOff := feature.ValueOff
					// 				}
					// 			}
					// 		}
					// 	}
					// }
				}
			}
		}

	}

	return nil, nil
}

func createMockConfiguration() *configuration {

	trigger := newTrigger()
	trigger.Type = "timer"
	trigger.Name = "Turn on light"

	// new condition
	tc := &timeDurationCondition{}
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
