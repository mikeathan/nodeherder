package automations_test

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
	"node-herder/mocks"
	"node-herder/models/devices"
	"strings"
	"testing"
	"time"
)

// https://www.home-assistant.io/docs/automation/basics/
// https://www.home-assistant.io/docs/automation/editor/
// mosquitto_pub -h 192.168.179:1883 -u sinkhole -P mqtt2023 -t 'zigbee2mqtt/Hive light 1/set' -m '{ "state": "ON" }'

func TestMqttActionSuccess(t *testing.T) {

	deviceName := "device1"
	sensorProperty := "state"
	newValue := true

	var data = map[string]any{
		"presence":    true,
		"temperature": 20.5,
		"humidity":    50.2,
		"lux":         8,
	}

	var messageHandler = func(id string, payload []byte) {
		fmt.Println("Received message", id, string(payload))
		if !strings.HasPrefix(id, deviceName) {
			t.Fatalf("invalid received topic: want %s got %s", deviceName, id)
		}

		data := unpackJsonToMap(string(payload))
		if data == nil {
			t.Fatalf("error unpacking json")
		}
		value, ok := data[sensorProperty]
		if !ok {
			t.Fatalf("property not %s found in payload", sensorProperty)
		}
		if value != newValue {
			t.Fatalf("value mismatch: want %v got %v", newValue, value)
		}
	}

	mqtt := &mocks.MockMqttClient{}
	mqtt.OnMessageHandler(messageHandler)

	trigger := newMockMqttTrigger(deviceName, true)
	condition := newMockMqttCondition("presence", true)
	action := newMockMqttAction(deviceName, sensorProperty, "light", newValue)
	action.Client = mqtt
	condition.Action = action
	trigger.Conditions = append(trigger.Conditions, condition)

	trigger.Evaluate(data)
}

func unpackJsonToMap(value string) map[string]any {
	var payload map[string]any
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		fmt.Print("ERROR unpacking ", value)
	}
	return payload

}
func newMockMqttAction(friendlyName string, property string, actionType string, value any) *automations.MqttAction {
	action := &automations.MqttAction{}
	action.Friendlyname = friendlyName
	action.Property = property
	action.Type = actionType
	action.Value = value

	return action
}

func newMockMqttCondition(sensor string, value any) *automations.MqttCondition {
	condition := &automations.MqttCondition{}
	condition.Friendlyname = sensor
	condition.Type = sensor
	condition.Value = value

	return condition

}
func newMockMqttTrigger(deviceName string, enabled bool) *automations.MqttTrigger {

	trigger := &automations.MqttTrigger{}
	trigger.DeviceName = deviceName
	trigger.Description = fmt.Sprintf("automation for device %s", trigger.DeviceName)
	trigger.Enabled = enabled

	return trigger
}

func TestMqttAction(t *testing.T) {
	t.Skip()

	mqttConfig := mqtt.MqttConfig{
		Username: "sinkhole",
		Password: "mqtt2023",
		Broker:   "192.168.50.179:1883",
	}
	mqtt := mqtt.NewMqttClient(mqttConfig)
	topics := []string{
		"bridge/devices",
		"bridge/logging",
	}

	for _, t := range topics {
		mqtt.AddTopic(t)
	}
	automation := automations.NewEngine(mqtt)
	var messageHandler = func(id string, payload []byte) {
		if id == "bridge/devices" {

			devices, err := devices.LoadBridgeDevices(payload)
			if err != nil {
				fmt.Println("error: ", err.Error())
			}

			err = automation.Load(devices)
			if err != nil {
				fmt.Println("error loading automations: ", err.Error())
			}

			// fmt.Println(device)
		} else if id == "bridge/logging" {

			fmt.Println(string(payload))
			// handle error ?
			//{"level":"error",
			//"message":"Publish 'set' 'state' to 'Hive light 1' failed: 'Error: Command 0x70ac08fffefafeca/1 genOnOff.off({}, {\"sendWhen\":\"immediate\",\"timeout\":10000,\"disableResponse\":false,\"disableRecovery\":false,\"disableDefaultResponse\":false,\"direction\":0,\"srcEndpoint\":null,\"reservedBits\":0,\"manufacturerCode\":null,\"transactionSequenceNumber\":null,\"writeUndiv\":false}) failed (Data request failed with error: 'MAC no ack' (233))'"}
		} else {
			//automation.HandleDevice(nil)
		}
	}

	mqtt.OnMessageHandler(messageHandler)
	err := mqtt.Connect()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	mqtt.Publish("zigbee2mqtt/bridge/devices", nil)

	time.Sleep(300 * time.Second)

	fmt.Println("finish")
}
