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

func TestMqttEvaluateSuccessfulCondition(t *testing.T) {

	wantHits := 2
	onValue := true
	offValue := false
	deviceName := "device1"
	sensorProperty := "someproperty"

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

		var data = map[string]any{
			testCase.sensor:           testCase.newValue,
			"some_sensor_data":        true,
			"some sensor data 2":      20.5,
			"more sensor data 4":      50.2,
			"even more sensor data 4": 8,
		}

		trigger.Evaluate(data)
	}

	time.Sleep(100 * time.Millisecond)
	if numOfHits != wantHits {
		t.Fatalf("hist value mismatch: want %v got %v", wantHits, numOfHits)
	}
}

func TestMqttConditionWithTimerConstraint(t *testing.T) {

	testSensor := "presence"
	deviceName := "device1"
	sensorProperty := "state"
	offValue := false
	conditonValue := false
	mqtt := &mocks.MockMqttClient{}

	trigger := newMockMqttTrigger(deviceName, true)
	offCondition := createConditionWithTimerConstraint(deviceName, testSensor, conditonValue, sensorProperty, offValue, 100*time.Millisecond, mqtt)
	trigger.Conditions = append(trigger.Conditions, offCondition)

	wg := &sync.WaitGroup{}

	// NOTES:
	// timer constraint delay is 100 ms
	// newValue = false is when it should trigger no presence and start timer for ending message
	testCases := []struct {
		sensor   string
		delay    time.Duration
		newValue any
		result   bool
	}{
		{sensor: "presence", delay: 50 * time.Millisecond, newValue: true, result: false},
		{sensor: "presence", delay: 70 * time.Millisecond, newValue: false, result: false}, // expect to fail as timeout is too short and we have new event next cancelling us
		{sensor: "presence", delay: 50 * time.Millisecond, newValue: true, result: false},
		{sensor: "presence", delay: 50 * time.Millisecond, newValue: true, result: false},
		{sensor: "presence", delay: 50 * time.Millisecond, newValue: false, result: false},
		{sensor: "presence", delay: 110 * time.Millisecond, newValue: false, result: true},
		{sensor: "presence", delay: 50 * time.Millisecond, newValue: true, result: false},
		{sensor: "presence", delay: 110 * time.Millisecond, newValue: false, result: true},
		{sensor: "presence", delay: 50 * time.Millisecond, newValue: true, result: false},
		{sensor: "presence", delay: 90 * time.Millisecond, newValue: false, result: false},
		{sensor: "presence", delay: 100 * time.Millisecond, newValue: false, result: true},
	}

	for _, testCase := range testCases {
		var data = map[string]any{
			testCase.sensor:           testCase.newValue,
			"some_sensor_data":        true,
			"some sensor data 2":      20.5,
			"more sensor data 4":      50.2,
			"even more sensor data 4": 8,
		}

		if testCase.result {
			wg.Add(1)
		}
		var messageHandler = func(id string, payload []byte) {
			if !testCase.result {
				t.Fatalf("operation result mismatch: want false got %v ", testCase.result)
			}
			wg.Done()
		}

		mqtt.OnMessageHandler(messageHandler)

		trigger.Evaluate(data)
		time.Sleep(testCase.delay)
	}

	wg.Wait()
}

func TestMqttConditionWithSensorConstraint(t *testing.T) {

	testSensor := "presence"
	deviceName := "device1"
	sensor1Property := "state"
	sensor2Property := "lux"

	mqtt := &mocks.MockMqttClient{}

	wg := &sync.WaitGroup{}

	testCases := []struct {
		op              string
		occupancy       bool
		newValue        any
		constraintValue any
		result          bool
	}{
		{op: ">", occupancy: true, newValue: 1.1, constraintValue: 1, result: true},
		{op: ">=", occupancy: true, newValue: 1, constraintValue: 1, result: true},
		{op: "<", occupancy: true, newValue: 1, constraintValue: 1.1, result: true},
		{op: "<=", occupancy: true, newValue: 1, constraintValue: 1, result: true},
		{op: "=", occupancy: true, newValue: 2, constraintValue: 2, result: true},

		{op: ">", occupancy: true, newValue: 1, constraintValue: 1.1, result: false},
		{op: ">=", occupancy: true, newValue: 1, constraintValue: 1.1, result: false},
		{op: "<", occupancy: true, newValue: 1.1, constraintValue: 1, result: false},
		{op: "<=", occupancy: true, newValue: 1.1, constraintValue: 1, result: false},
		{op: "=", occupancy: true, newValue: 1, constraintValue: 2, result: false},
	}
	for _, testCase := range testCases {
		if testCase.result {
			wg.Add(1)
		}

		trigger := newMockMqttTrigger(deviceName, true)
		turnOnCondition := createTriggerConditionWithSensorConstraint(deviceName, testSensor, true, sensor1Property, testCase.occupancy, sensor2Property, testCase.constraintValue, testCase.op, mqtt)
		trigger.Conditions = append(trigger.Conditions, turnOnCondition)

		var messageHandler = func(id string, payload []byte) {

			if !testCase.result {
				t.Fatalf("operation result mismatch: want false got %v ", testCase.result)
			}
			wg.Done()
		}

		mqtt.OnMessageHandler(messageHandler)
		var data = map[string]any{
			"presence": testCase.occupancy,
			"lux":      testCase.newValue,
		}

		trigger.Evaluate(data)
		time.Sleep(200 * time.Millisecond)
	}
	wg.Wait()
}

func TestOneConditionWithSensorConstraintAndOneConditionWithTimerConstraint(t *testing.T) {

	deviceName := "device 1"
	luxSensor := "lux"
	presenceSensor := "presence"
	actionProperty := "state"
	offConditionValue := false
	offActionValue := false
	onActionValue := true
	onConditionValue := true
	luxConstraintValue := 30
	constraintOp := "<="
	timerValue := 50 * time.Millisecond

	mqtt := &mocks.MockMqttClient{}

	trigger := newMockMqttTrigger(deviceName, true)

	// sensor off condition
	offCondition := createConditionWithTimerConstraint(deviceName, presenceSensor, offConditionValue, actionProperty, offActionValue, timerValue, mqtt)
	trigger.Conditions = append(trigger.Conditions, offCondition)

	// sensor on condition
	turnOnCondition := createTriggerConditionWithSensorConstraint(deviceName, presenceSensor, onConditionValue, actionProperty, onActionValue, luxSensor, luxConstraintValue, constraintOp, mqtt)
	trigger.Conditions = append(trigger.Conditions, turnOnCondition)
	wg := &sync.WaitGroup{}
	testCases := []struct {
		occupancy bool
		luxValue  any
		result    bool
	}{
		{occupancy: true, luxValue: 35, result: false},
		{occupancy: true, luxValue: 30, result: true},   // turn on light
		{occupancy: false, luxValue: 100, result: true}, // turn off light
		{occupancy: true, luxValue: 29.9, result: true}, // turn on light
		{occupancy: true, luxValue: 31, result: false},
		{occupancy: true, luxValue: 100, result: false},
		{occupancy: true, luxValue: 10, result: true},  // turn on light
		{occupancy: false, luxValue: 10, result: true}, // turn off light
	}

	// NOTES:
	// if presence is true and lux is below 30 lux, send message to turn on light
	// if presene is off, start timer for 50 ms, send  message to turn off light

	for _, testCase := range testCases {
		if testCase.result {
			wg.Add(1)
		}
		var messageHandler = func(id string, payload []byte) {

			if !testCase.result {
				t.Fatalf("operation result mismatch: want true got %v ", testCase.result)
			}
			wg.Done()
		}

		mqtt.OnMessageHandler(messageHandler)
		var data = map[string]any{
			"presence": testCase.occupancy,
			"lux":      testCase.luxValue,
		}

		trigger.Evaluate(data)
		time.Sleep(100 * time.Millisecond)
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

func unpackJsonToMap(value string) map[string]any {
	var payload map[string]any
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		fmt.Print("ERROR unpacking ", value)
	}
	return payload

}

func createTriggerConditionWithSensorConstraint(deviceName string, sensor string, conditionValue any, actionProperty string, actionValue any, constraintProperty string, constraintValue any, constraintOp string, mqtt mqtt.MqttClient) *automations.DeviceCondition {

	// create condition
	condition := newMockMqttCondition(sensor, conditionValue)
	action := newMockMqttAction(deviceName, actionProperty, "light", actionValue)

	action.Client = mqtt
	condition.Action = action

	//
	// add device constrains
	deviceConstraint := &automations.DeviceConstraint{}
	deviceConstraint.Type = constraintProperty
	deviceConstraint.Value = constraintValue
	deviceConstraint.EqualityOperator = constraintOp
	condition.Constraint = deviceConstraint
	return condition
}
func createConditionWithTimerConstraint(deviceName string, sensor string, conditionValue any, actionProperty string, actionValue any, constraintDuration time.Duration, mqtt mqtt.MqttClient) *automations.DeviceCondition {

	// create condition
	condition := newMockMqttCondition(sensor, conditionValue)
	action := newMockMqttAction(deviceName, actionProperty, "light", actionValue)

	action.Client = mqtt
	condition.Action = action

	//
	// add timer constrains
	timerConstraint := automations.NewTimerConstraint()
	timerConstraint.Duration = constraintDuration
	condition.Constraint = timerConstraint

	return condition
}
func newMockMqttAction(friendlyName string, property string, actionType string, value any) *automations.MqttAction {
	action := &automations.MqttAction{}
	action.Friendlyname = friendlyName
	action.Property = property
	action.Type = actionType
	action.Value = value

	return action
}

func newMockMqttCondition(sensor string, value any) *automations.DeviceCondition {
	condition := automations.NewDeviceCondition()
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
