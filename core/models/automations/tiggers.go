package automations

import (
	"errors"
	"fmt"
	"node-herder/models/bridge"
	"node-herder/transport/mqtt"
	"time"
)

func (t *TimerTrigger) configure(bridgeDevices []*bridge.BridgeDevice, client mqtt.MqttClient) error {
	fmt.Println("loading:", t.Name)
	if t.Type != "timer" {
		return errors.New("unsupported trigger type ")
	}

	// valdate conditions
	for _, cond := range t.Conditions {
		switch cond.(type) {
		case TimeDurationCondition:
			break
		case TimestampCondition:
			break
		default:
			return errors.New("unsupported condition type ")
		}
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
			}
		}
	}

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

func (t *TimerTrigger) Process() {

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
