package automations_test

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
	"node-herder/mocks"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestTurnOnAndOffLightFromPresence(t *testing.T) {

	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}

	turnOffTrigger := createTriggerDelayTurnOffLightWithPresenceOff(mqtt, 100*time.Millisecond)
	turnOnTrigger := createTriggerTurnOnLightWithPresenceOnAndLux(mqtt, 30)

	// create device trigger
	deviceTrigger := automations.NewDeviceTrigger("human sensor")
	deviceTrigger.SensorTriggers = make(map[string][]*automations.SensorTrigger)
	deviceTrigger.SensorTriggers[turnOffTrigger.Name] = append(deviceTrigger.SensorTriggers[turnOffTrigger.Name], turnOffTrigger)
	deviceTrigger.SensorTriggers[turnOnTrigger.Name] = append(deviceTrigger.SensorTriggers[turnOnTrigger.Name], turnOnTrigger)

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

			action := turnOnTrigger.ActionRunner
			if !strings.HasPrefix(id, action.Friendlyname) {
				t.Fatalf("invalid received topic: want %s got %s", action.Friendlyname, id)
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

		deviceTrigger.Evaluate(data)

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
		res := automations.Equalityoperators[testCase.op](testCase.value1, testCase.value2)
		if res != testCase.result {
			t.Fatalf("operation result mismatch: want %v got %v in  %v %s %v", testCase.result, res, testCase.value1, testCase.op, testCase.value2)
		}
	}
}

func TestExportToFile(t *testing.T) {
	mqtt := &mocks.MockMqttClient{}

	turnOffTrigger := createTriggerDelayTurnOffLightWithPresenceOff(mqtt, 100*time.Millisecond)
	turnOnTrigger := createTriggerTurnOnLightWithPresenceOnAndLux(mqtt, 30)

	// create device trigger
	deviceTrigger := automations.NewDeviceTrigger("human sensor")
	deviceTrigger.Description = "test human sensor automation"
	deviceTrigger.SensorTriggers = make(map[string][]*automations.SensorTrigger)
	deviceTrigger.SensorTriggers[turnOffTrigger.Name] = append(deviceTrigger.SensorTriggers[turnOffTrigger.Name], turnOffTrigger)
	deviceTrigger.SensorTriggers[turnOnTrigger.Name] = append(deviceTrigger.SensorTriggers[turnOnTrigger.Name], turnOnTrigger)

	err := deviceTrigger.Save("temp1", true)
	if err != nil {
		t.Fatalf("ERROR saving trigger %s", err.Error())
	}

	newTrigger, err := automations.LoadTrigger("temp1")
	if err != nil {
		t.Fatalf("ERROR laoding trigger from file %s", err.Error())
	}

	if newTrigger.Name != deviceTrigger.Name {
		t.Fatalf("ERROR Name mismatch")
	}

	if newTrigger.Description != deviceTrigger.Description {
		t.Fatalf("ERROR Description mismatch")
	}
	if newTrigger.Enabled != deviceTrigger.Enabled {
		t.Fatalf("ERROR Description mismatch")
	}

	for sensor, triggers := range deviceTrigger.SensorTriggers {
		newTriggers := newTrigger.SensorTriggers[sensor]

		for tidx, trigger := range triggers {

			newTrigger := newTriggers[tidx]
			if trigger.Name != newTrigger.Name {
				t.Fatalf("ERROR Trigger.Name mismatch")
			}
			if trigger.ActionRunner.Friendlyname != newTrigger.ActionRunner.Friendlyname {
				t.Fatalf("ERROR Action.Friendlyname mismatch")
			}

			if trigger.ActionRunner.Property != newTrigger.ActionRunner.Property {
				t.Fatalf("ERROR Action.Property mismatch")
			}

			if trigger.ActionRunner.Type != newTrigger.ActionRunner.Type {
				t.Fatalf("ERROR Action.Type mismatch")
			}
			if trigger.ActionRunner.Delay != newTrigger.ActionRunner.Delay {
				t.Fatalf("ERROR Action.Delay mismatch")
			}

			for cidx, condition := range trigger.Conditions {

				newCondition := newTriggers[tidx].Conditions[cidx]

				if condition.EqualityOperator != newCondition.EqualityOperator {
					t.Fatalf("ERROR Condition.EqualityOperator mismatch")
				}
				if condition.Name != newCondition.Name {
					t.Fatalf("ERROR Condition.Name mismatch")
				}

				if condition.Value != newCondition.Value {
					t.Fatalf("ERROR Condition.Value mismatch")
				}
			}
		}
	}

	err = automations.DeleteTrigger("temp1")

	if err != nil {
		t.Fatalf("ERROR deleting file%s", err.Error())
	}
}

// func TestExportTriggerToFile(t *testing.T) {

// 	deviceName := "device 1"
// 	luxSensor := "lux"
// 	presenceSensor := "presence"
// 	actionProperty := "state"
// 	offConditionValue := false
// 	offActionValue := false
// 	onActionValue := true
// 	onConditionValue := true
// 	luxConstraintValue := 30.1
// 	constraintOp := "<="
// 	timerValue := 50 * time.Millisecond

// 	mqtt := &mocks.MockMqttClient{}

// 	trigger := newMockMqttTrigger(deviceName, true)

// 	// sensor off condition
// 	offCondition := createConditionWithTimerConstraint(deviceName, presenceSensor, offConditionValue, actionProperty, offActionValue, timerValue, mqtt)
// 	trigger.Conditions = append(trigger.Conditions, offCondition)

// 	// sensor on condition
// 	turnOnCondition := createTriggerConditionWithSensorConstraint(deviceName, presenceSensor, onConditionValue, actionProperty, onActionValue, luxSensor, luxConstraintValue, constraintOp, mqtt)
// 	trigger.Conditions = append(trigger.Conditions, turnOnCondition)

// 	err := trigger.Save("temp1", true)
// 	if err != nil {
// 		t.Fatalf("ERROR saving trigger %s", err.Error())
// 	}

// 	newTrigger, err := automations.LoadTrigger("temp1")
// 	if err != nil {
// 		t.Fatalf("ERROR laoding trigger from file %s", err.Error())
// 	}

// 	if newTrigger.Name != trigger.Name {
// 		t.Fatalf("ERROR Name mismatch")
// 	}
// 	if newTrigger.Description != trigger.Description {
// 		t.Fatalf("ERROR Description mismatch")
// 	}
// 	if newTrigger.Enabled != trigger.Enabled {
// 		t.Fatalf("ERROR Description mismatch")
// 	}

// 	for idx, c := range trigger.Conditions {
// 		nc := newTrigger.Conditions[idx]
// 		if c.EqualityOperator != nc.EqualityOperator {
// 			t.Fatalf("ERROR Condition.EqualityOperator mismatch")
// 		}
// 		if c.Friendlyname != nc.Friendlyname {
// 			t.Fatalf("ERROR Condition.Friendlyname mismatch")
// 		}
// 		if c.Type != nc.Type {
// 			t.Fatalf("ERROR Condition.Type mismatch")
// 		}
// 		if c.Value != nc.Value {
// 			t.Fatalf("ERROR Condition.Value mismatch")
// 		}

// 		if nc.Constraint != nil {
// 			if otc, ok := nc.Constraint.(*automations.TimerConstraint); ok {
// 				ntc := c.Constraint.(*automations.TimerConstraint)
// 				if ntc.Duration != otc.Duration {
// 					t.Fatalf("ERROR TimerConstraint.Duration mismatch")
// 				}
// 				if ntc.Type != otc.Type {
// 					t.Fatalf("ERROR TimerConstraint.Type mismatch")
// 				}
// 			}

// 			if odc, ok := nc.Constraint.(*automations.DeviceConstraint); ok {
// 				ndc := c.Constraint.(*automations.DeviceConstraint)
// 				if ndc.EqualityOperator != odc.EqualityOperator {
// 					t.Fatalf("ERROR DeviceConstraint.EqualityOperator mismatch")
// 				}
// 				if ndc.Sensor != odc.Sensor {
// 					t.Fatalf("ERROR DeviceConstraint.Sensor mismatch")
// 				}
// 				if ndc.Value != odc.Value {
// 					t.Fatalf("ERROR DeviceConstraint.Value mismatch")
// 				}
// 				if ndc.Type != odc.Type {
// 					t.Fatalf("ERROR DeviceConstraint.Type mismatch")
// 				}
// 			}
// 		}

// 		if c.Action.Friendlyname != nc.Action.Friendlyname {
// 			t.Fatalf("ERROR Action.Friendlyname mismatch")
// 		}
// 		if c.Action.Property != nc.Action.Property {
// 			t.Fatalf("ERROR Action.Property mismatch")
// 		}
// 		if c.Action.Value != nc.Action.Value {
// 			t.Fatalf("ERROR Action.Value mismatch")
// 		}
// 		if c.Action.Type != nc.Action.Type {
// 			t.Fatalf("ERROR Action.Type mismatch")
// 		}
// 	}

// 	err = automations.DeleteTrigger("temp1")
// 	if err != nil {
// 		t.Fatalf("ERROR deleting file%s", err.Error())
// 	}
// }

func unpackJsonToMap(value string) map[string]any {
	var payload map[string]any
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		fmt.Print("ERROR unpacking ", value)
	}
	return payload
}

func createTriggerTurnOnLightWithPresenceOnAndLux(mqtt mqtt.MqttClient, lux any) *automations.SensorTrigger {
	// action = turn off light
	turnOnAction := &automations.MqttAction{}
	turnOnAction.Friendlyname = "Attic light"
	turnOnAction.Type = "light"
	turnOnAction.Property = "state"
	turnOnAction.Value = true
	turnOnAction.Delay = 0
	turnOnAction.Client = mqtt

	// Turn on sensor trigger
	turnOnTrigger := &automations.SensorTrigger{}
	turnOnTrigger.Name = "presence"
	turnOnTrigger.ActionRunner = turnOnAction

	// condition = presence = off && lux <= 30
	turnOnCondition := &automations.SensorCondition{}
	turnOnCondition.Name = "presence"
	turnOnCondition.EqualityOperator = "="
	turnOnCondition.Value = true

	luxCondition := &automations.SensorCondition{}
	luxCondition.Name = "lux"
	luxCondition.EqualityOperator = "<="
	luxCondition.Value = lux

	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, turnOnCondition)
	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, luxCondition)

	return turnOnTrigger
}

func createTriggerDelayTurnOffLightWithPresenceOff(mqtt mqtt.MqttClient, delay time.Duration) *automations.SensorTrigger {
	// action = turn off light
	turnOffAction := &automations.MqttAction{}
	turnOffAction.Friendlyname = "Attic light"
	turnOffAction.Type = "light"
	turnOffAction.Property = "state"
	turnOffAction.Value = false
	turnOffAction.Delay = delay
	turnOffAction.Client = mqtt

	// Turn off sensor trigger
	turnOffTrigger := &automations.SensorTrigger{}
	turnOffTrigger.Name = "presence"
	turnOffTrigger.ActionRunner = turnOffAction

	// condition = presence == false
	turnOffCondition := &automations.SensorCondition{}
	turnOffCondition.Name = "presence"
	turnOffCondition.EqualityOperator = "="
	turnOffCondition.Value = false

	turnOffTrigger.Conditions = append(turnOffTrigger.Conditions, turnOffCondition)

	return turnOffTrigger
}
