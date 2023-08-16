package automations_test

import (
	"fmt"
	"node-herder/models/automations"
	automation "node-herder/models/automations"
	"node-herder/models/bridge"
	"node-herder/transport/mqtt"
	"testing"
	"time"
)

func TestTriggers(t *testing.T) {

	tt := automation.TriggerFromDuration(5 * time.Second)
	tt.Repeat = false
	trigger := new(automation.TimerTrigger)
	trigger.WithCondition(tt)
	trigger.WithAction("Action triggered!!!!!!!!!!!1")
	trigger.Process()

	// https://www.home-assistant.io/docs/automation/basics/
	// https://www.home-assistant.io/docs/automation/editor/
}

func TestMqttAction(t *testing.T) {
	mqttConfig := mqtt.MqttConfig{
		Username: "sinkhole",
		Password: "mqtt2023",
		Broker:   "192.168.50.179:1883",
		Topics: []string{
			"bridge/devices",
		},
	}

	var messageHandler = func(id string, payload []byte) {
		if id == "bridge/devices" {

			devices, err := bridge.Parse(payload)
			if err != nil {
				fmt.Println("error: ", err.Error())
			}
			automations.Load(devices)
			// device, err := bridge.FindByExposeType(payload, "light")
			// if err != nil {
			// 	fmt.Println("error: ", err.Error())
			// }
			// fmt.Println(device)
		}

	}

	mqtt := mqtt.NewMqttClient(mqttConfig)
	mqtt.OnMessageHandler(messageHandler)
	err := mqtt.Connect()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	mqtt.Publish("zigbee2mqtt/bridge/devices", nil)

	time.Sleep(10 * time.Second)

	fmt.Println("finish")
}
