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

func TestMqttEvaluateSuccessfulCondition(t *testing.T) {

	wantHits := 2
	onValue := true
	offValue := false
	deviceName := "device1"
	sensorProperty := "someproperty"

	var data = map[string]any{
		"presence":    true,
		"temperature": 20.5,
		"humidity":    50.2,
		"lux":         8,
	}

	testCases := []struct {
		sensor   string
		newValue any
		want     bool
	}{
		{sensor: "presence", newValue: true, want: true},
		{sensor: "presence", newValue: false, want: false},
		{sensor: "temperature", newValue: 18.1, want: false},
		{sensor: "lux", newValue: 15, want: false},
		{sensor: "humidity", newValue: 70.3, want: false},
	}

	// Setup

	mqtt := &mocks.MockMqttClient{}

	trigger := newMockMqttTrigger(deviceName, true)

	// create first condition
	turnOnCondition := newMockMqttCondition("presence", true)
	onAction := newMockMqttAction(deviceName, sensorProperty, "light", onValue)

	trigger.Conditions = append(trigger.Conditions, turnOnCondition)
	onAction.Client = mqtt
	turnOnCondition.Action = onAction

	// create second condition
	turnOffCondition := newMockMqttCondition("presence", false)
	offAction := newMockMqttAction(deviceName, sensorProperty, "light", offValue)

	trigger.Conditions = append(trigger.Conditions, turnOffCondition)
	offAction.Client = mqtt
	turnOffCondition.Action = offAction

	// Evaluation
	numOfHits := 0
	for _, testCase := range testCases {

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
			if value != testCase.want {
				t.Fatalf("value mismatch: want %v got %v", testCase.want, value)
			}
			numOfHits++
		}

		mqtt.OnMessageHandler(messageHandler)
		data[testCase.sensor] = testCase.newValue
		trigger.Evaluate(data)
	}

	time.Sleep(100 * time.Millisecond)
	if numOfHits != wantHits {
		t.Fatalf("hist value mismatch: want %v got %v", wantHits, numOfHits)
	}

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
