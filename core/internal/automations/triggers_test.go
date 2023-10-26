package automations_test

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
	"node-herder/mocks"
	"node-herder/models/devices"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestTriggerWithNoConditionsCallsAction(t *testing.T) {
	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}
	testTriggerData := map[string]string{}
	testTriggerData["buttonSwitch1"] = "brightness"
	testTriggerData["buttonSwitch2"] = "state"
	testTriggerData["buttonSwitch3"] = "color_brightness"
	switch1Trigger := createSwitchTriggerWithBindingAction("buttonSwitch1", "brightness", mqtt)
	switch2Trigger := createSwitchTriggerWithBindingAction("buttonSwitch2", "state", mqtt)
	switch3Trigger := createSwitchTriggerWithBindingAction("buttonSwitch3", "color_brightness", mqtt)

	// create device trigger
	deviceTrigger := automations.NewDevice("button switch")
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, switch1Trigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, switch2Trigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, switch3Trigger)

	testCases := []struct {
		triggeredEntity string
		value           any
		result          bool
	}{
		{triggeredEntity: "buttonSwitch1", value: 200.0, result: true},
		{triggeredEntity: "buttonSwitch2", value: true, result: true},
		{triggeredEntity: "door_state", value: true, result: false},
		{triggeredEntity: "co2", value: 70, result: false},
		{triggeredEntity: "quality_index", value: 30.1, result: false},
		{triggeredEntity: "buttonSwitch2", value: false, result: true},
		{triggeredEntity: "buttonSwitch3", value: 16000.0, result: true},
		{triggeredEntity: "buttonSwitch3", value: 5000.9, result: true},
		{triggeredEntity: "temperature", value: 25.1, result: false},
		{triggeredEntity: "humdity", value: 35.1, result: false},
	}

	for _, testCase := range testCases {
		var data = map[string]any{
			"data1":                  false,
			"data2":                  90,
			testCase.triggeredEntity: testCase.value,
		}
		var messageHandler = func(id string, payload []byte) {
			if !testCase.result {
				t.Fatalf("invalid msg received")
			}

			wg.Done()
			var msg map[string]interface{}
			json.Unmarshal([]byte(payload), &msg)
			property := testTriggerData[testCase.triggeredEntity]
			if msg[property] != testCase.value {
				t.Fatalf("invalid msg: received want %s got %s", testCase.value, msg[testCase.triggeredEntity])
			}
		}

		mqtt.OnMessageHandler(messageHandler)
		if testCase.result {
			wg.Add(1)
		}
		device := devices.NewDeviceV2("1")
		device.Exposes = createExposures(data)

		deviceTrigger.EvaluateV2(device)
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()
}

func TestTurnOnAndOffLightFromPresence(t *testing.T) {

	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}

	turnOffTrigger := createTriggerDelayTurnOffLightWithPresenceOff(mqtt, 100*time.Millisecond)
	turnOnTrigger := createTriggerTurnOnLightWithPresenceOnAndLux(mqtt, 30)

	// create device trigger
	deviceTrigger := automations.NewDevice("human sensor")
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOffTrigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOnTrigger)

	device := devices.NewDeviceV2(deviceTrigger.Id)

	testCases := []struct {
		presence   bool
		sleepdelay time.Duration
		lux        any
		result     bool
	}{
		{presence: false, sleepdelay: 200, lux: 30, result: true},
		{presence: false, sleepdelay: 100, lux: 30, result: false},
		{presence: false, sleepdelay: 100, lux: 30, result: false},
		{presence: true, sleepdelay: 100, lux: 30, result: true},
		{presence: false, sleepdelay: 200, lux: 30, result: true},
		{presence: true, sleepdelay: 100, lux: 30, result: true},
		{presence: false, sleepdelay: 200, lux: 30, result: true},
		{presence: true, sleepdelay: 100, lux: 30.1, result: false},
		{presence: true, sleepdelay: 100, lux: 29.9, result: true},
		{presence: false, sleepdelay: 200, lux: 30, result: true},
		{presence: true, sleepdelay: 100, lux: 30.1, result: false},
		{presence: true, sleepdelay: 100, lux: 29, result: true},

		{presence: true, sleepdelay: 100, lux: 7, result: false},
		{presence: true, sleepdelay: 100, lux: 15, result: false},
	}

	for _, testCase := range testCases {
		var data = map[string]any{
			"presence": testCase.presence,
			"lux":      testCase.lux,
		}

		var messageHandler = func(id string, payload []byte) {

			wg.Done()

			action := turnOnTrigger.Action
			if !strings.HasPrefix(id, action.FriendlyName) {
				t.Fatalf("invalid received topic: want %s got %s", action.FriendlyName, id)
			}

			data := unpackJsonToMap(string(payload))
			if data == nil {
				t.Fatalf("error unpacking json")
			}
			value, ok := data[action.Property]
			if !ok {
				t.Fatalf("property not %s found in payload", action.Property)
			}
			if value != testCase.presence {
				t.Fatalf("value mismatch: want %v got %v", testCase.presence, value)
			}
		}

		mqtt.OnMessageHandler(messageHandler)
		if testCase.result {
			wg.Add(1)
		}

		device.Exposes = createExposures(data)
		deviceTrigger.EvaluateV2(device)

		time.Sleep(testCase.sleepdelay * time.Millisecond)
	}

	wg.Wait()
}

func TestEqualityChecks(t *testing.T) {
	testCases := []struct {
		op     string
		value1 any
		value2 any
		result bool
	}{
		{op: ">", value1: 1.1, value2: 1, result: true},
		{op: ">=", value1: 1, value2: 1, result: true},
		{op: "<", value1: 1, value2: 1.1, result: true},
		{op: "<=", value1: 1, value2: 1, result: true},
		{op: "=", value1: 2, value2: 2, result: true},

		{op: ">", value1: 1, value2: 1.1, result: false},
		{op: ">=", value1: 1, value2: 1.1, result: false},
		{op: "<", value1: 1.1, value2: 1, result: false},
		{op: "<=", value1: 1.1, value2: 1, result: false},
		{op: "=", value1: 1, value2: 2, result: false},
	}

	for _, testCase := range testCases {
		res := automations.EqualityOperators[testCase.op](testCase.value1, testCase.value2)
		if res != testCase.result {
			t.Fatalf("operation result mismatch: want %v got %v in  %v %s %v", testCase.result, res, testCase.value1, testCase.op, testCase.value2)
		}
	}
}

func unpackJsonToMap(value string) map[string]any {
	var payload map[string]any
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		fmt.Print("ERROR unpacking ", value)
	}
	return payload
}

func createTriggerTurnOnLightWithPresenceOnAndLux(mqtt mqtt.MqttClient, lux any) *automations.Trigger {
	// action = turn off light
	turnOnAction := &automations.MqttAction{}
	turnOnAction.FriendlyName = "Attic light"
	turnOnAction.Type = "light"
	turnOnAction.Property = "state"
	turnOnAction.Data = true
	turnOnAction.Delay = 0
	turnOnAction.Client = mqtt

	// Turn on sensor trigger
	turnOnTrigger := &automations.Trigger{}
	turnOnTrigger.Name = "presence"
	turnOnTrigger.Action = turnOnAction

	// condition = presence = off && lux <= 30
	turnOnCondition := &automations.Condition{}
	turnOnCondition.Name = "presence"
	turnOnCondition.EqualityOperator = "="
	turnOnCondition.Value = true

	luxCondition := &automations.Condition{}
	luxCondition.Name = "lux"
	luxCondition.EqualityOperator = "<="
	luxCondition.Value = lux

	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, turnOnCondition)
	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, luxCondition)

	return turnOnTrigger
}

func createSwitchTriggerWithBindingAction(triggerName string, actionProp string, mqtt mqtt.MqttClient) *automations.Trigger {
	// action = turn off light
	brightnessAction := &automations.MqttAction{}
	brightnessAction.FriendlyName = "Attic light"
	brightnessAction.Type = "light"
	brightnessAction.Property = actionProp
	brightnessAction.Client = mqtt

	// Turn off sensor trigger
	button1Trigger := &automations.Trigger{}
	button1Trigger.Name = triggerName
	button1Trigger.Action = brightnessAction

	return button1Trigger
}

func createTriggerDelayTurnOffLightWithPresenceOff(mqtt mqtt.MqttClient, delay time.Duration) *automations.Trigger {
	// action = turn off light
	turnOffAction := &automations.MqttAction{}
	turnOffAction.FriendlyName = "Attic light"
	turnOffAction.Type = "light"
	turnOffAction.Property = "state"
	turnOffAction.Data = false
	turnOffAction.Delay = delay
	turnOffAction.Client = mqtt

	// Turn off sensor trigger
	turnOffTrigger := &automations.Trigger{}
	turnOffTrigger.Name = "presence"
	turnOffTrigger.Action = turnOffAction

	// condition = presence == false
	turnOffCondition := &automations.Condition{}
	turnOffCondition.Name = "presence"
	turnOffCondition.EqualityOperator = "="
	turnOffCondition.Value = false

	turnOffTrigger.Conditions = append(turnOffTrigger.Conditions, turnOffCondition)

	return turnOffTrigger
}

func createExposures(data map[string]interface{}) map[string]*devices.Entity {
	var entities = make(map[string]*devices.Entity)
	for key, value := range data {

		newEntity := createEntity(key, "", value, "", "", nil)
		entities[key] = newEntity
	}
	return entities
}

func createEntity(name string, description string, data any, unit string, dataType string, props map[string]any) *devices.Entity {
	if props == nil {
		props = make(map[string]any)
	}
	newEntity := &devices.Entity{}
	newEntity.Data = data
	newEntity.Name = name
	newEntity.Unit = unit
	newEntity.Description = description
	newEntity.Properties = props
	return newEntity
}
