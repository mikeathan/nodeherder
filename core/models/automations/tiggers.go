package automations

import (
	"errors"
	"fmt"
	"node-herder/models/bridge"
	"node-herder/transport/mqtt"
	"time"
)

func (t *TimerTrigger) configure(bridgeDevices []*bridge.BridgeDevice, client mqtt.MqttClient) error {
	fmt.Println("Loading trigger:", t.Name)
	if t.Type != "timer" {
		return errors.New("unsupported trigger type ")
	}

	// valdate conditions
	for _, cond := range t.Conditions {

		_, ok := cond.(*TimeDurationCondition)
		if ok {
			continue
		}
		_, ok = cond.(*TimestampCondition)
		if ok {
			continue
		}

		return errors.New("unsupported condition type ")
	}

	// validate actions
	for _, action := range t.Actions {
		for _, device := range bridgeDevices {
			if device.FriendlyName == action.Friendlyname {

				if action.Type != "light" {
					return errors.New("unsupported action type ")
				}

				if device.Disabled {
					return errors.New("device is disabled ")
				}
				action.client = client

				// need to fix state, convert it to expected one: { "state": "ON" }'
			}
		}
	}

	t.run()
	return nil
}

func (t *TimerTrigger) Trigger() error {

	for _, action := range t.Actions {

		fmt.Printf("Trigger Action %s\n", action.Friendlyname)
		err := action.Run()
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	}
	return nil
}

func (t *TimerTrigger) run() {

	// c: interface conversion: *automations.TimeDurationCondition is not automations.TimerCondition: missing method IsRepeat

	for _, cond := range t.Conditions {

		tc := cond.(TimerCondition)
		go func() {
			for {

				diff := time.Until(tc.GetSchedule()).Seconds()
				ticker := *time.NewTicker(time.Duration(diff) * time.Second)
				<-ticker.C

				err := t.Trigger()
				if err != nil {
					fmt.Println(err)
				}

				if !tc.IsRepeat() {
					fmt.Println("timer finished. no repeat. exiting")
					return
				}
			}
		}()
	}
}
