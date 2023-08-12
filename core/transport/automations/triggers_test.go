package automations_test

import (
	"fmt"
	automation "node-herder/transport/automations"
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
	mqtt := mqtt.NewMqttClient(mqttConfig)
	err := mqtt.Connect()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	mqtt.Publish("zigbee2mqtt/bridge/devices", nil)

	time.Sleep(10 * time.Second)

	fmt.Println("finish")
}
