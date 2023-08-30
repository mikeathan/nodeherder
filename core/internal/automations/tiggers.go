package automations

import (
	"errors"
	"fmt"
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
	Type      string      `json:"trigger"`
	Name      string      `json:"name"`
	Condition interface{} `json:"condition"`
	Action    *MqttAction `json:"action"`
	Enabled   bool        `json:"enabled"`
}

func newTimerTrigger() *TimerTrigger {
	return &TimerTrigger{
		Type:    "timer",
		Name:    "",
		Action:  &MqttAction{},
		Enabled: true,
	}
}

func (t *MqttTrigger) configure(bridgeDevices []*devices.BridgeDevice, client mqtt.MqttClient) error {
	fmt.Println("Loading mqtt trigger:", t.Name)
	if t.Type != "mqtt" {
		return errors.New("unsupported mqtt type ")
	}

	// validate conditions
	_, ok := t.Condition.(*MqttCondition)
	if !ok {
		return errors.New("unsupported condition type ")
	}

	// validate actions
	err := validateAction(bridgeDevices, t.Action, client)
	if err != nil {
		return err
	}

	return nil
}

func (t *MqttTrigger) Evaluate(device *devices.Device) {

	if !t.Enabled {
		fmt.Printf("%s is disabled\n", t.Name)
		return
	}

	mc := t.Condition.(MqttCondition)
	mc.Evaluate(device)
}

func (t *TimerTrigger) configure(bridgeDevices []*devices.BridgeDevice, client mqtt.MqttClient) error {
	if t.Type != "timer" {
		return fmt.Errorf("unsupported trigger type %s", t.Type)
	}

	// valdate conditions
	_, ok := t.Condition.(*TimeDurationCondition)
	if !ok {
		_, ok = t.Condition.(*TimestampCondition)
		if !ok {
			return errors.New("unsupported condition type ")
		}
	}

	// validate actions
	err := validateAction(bridgeDevices, t.Action, client)

	if err != nil {
		return err
	}

	t.run()
	return nil
}

// validate actions
func validateAction(bridgeDevices []*devices.BridgeDevice, action *MqttAction, client mqtt.MqttClient) error {

	for _, device := range bridgeDevices {
		if device.FriendlyName == action.Friendlyname {

			if action.Type != "light" {
				return fmt.Errorf("unsupported action type %s", action.Type)
			}

			if device.Disabled {
				return errors.New("device is disabled ")
			}
			action.client = client

			// TODO: refactor. no need to keep looping once we found our value
			// need to fix state, convert it to expected one: { "state": "ON" }'
			for _, expose := range device.Definition.Exposes {
				for _, feature := range expose.Features {
					if feature.Property == action.Property { // state property only!

						// for now we only support "state" property
						if action.Value == true {
							action.Value = feature.ValueOn
						} else {
							action.Value = feature.ValueOff
						}
						break
					}
				}
			}
		}
	}
	return nil
}

func (t *TimerTrigger) Trigger() error {

	fmt.Printf("Trigger Action %s\n", t.Action.Friendlyname)
	return t.Action.Run()
}

func (t *TimerTrigger) run() {

	if !t.Enabled {
		fmt.Printf("%s is disabled\n", t.Name)
		return
	}

	tc := t.Condition.(TimerCondition)
	go func() {
		for {

			//
			// todo: allow enable/disable triggers
			//
			diff := time.Until(tc.GetSchedule()).Seconds()
			ticker := *time.NewTicker(time.Duration(diff) * time.Second)
			select {
			case <-ticker.C:
				err := t.Trigger()
				if err != nil {
					fmt.Println(err)
				}

				if !tc.IsRepeat() {
					fmt.Println("timer finished. no repeat. exiting")
					return
				}
			}
		}
	}()
}
