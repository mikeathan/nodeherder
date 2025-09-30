package automations_test

import (
	"encoding/json"
	"fmt"
	"node-herder/internal/automations"
	"node-herder/internal/mqtt"
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/repository"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"sync"
	"testing"
	"time"
)

func TestTriggerWithNoConditionsCallsAction(t *testing.T) {
	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}
	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}

	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	triggerId := "button switch"

	exposeNames := []string{"brightness", "state", "color_brightness"}
	entities := []*devices.Entity{}
	for _, e := range exposeNames {
		if e == "state" {
			e := utils_test.CreateEntity(e, "binary", false)
			entities = append(entities, e)

		} else {
			e := utils_test.CreateNumericEntity(e, nil)
			entities = append(entities, e)
		}
	}

	device := utils_test.CreateDeviceWithExposes(triggerId, "dial device", entities)
	deviceBridgeList := utils_test.CreateBridgeInfoList([]*devices.Device{device})
	registrar.RegisterBridge(deviceBridgeList)

	testTriggerData := map[string]string{}
	testTriggerData["buttonSwitch1"] = "brightness"
	testTriggerData["buttonSwitch2"] = "state"
	testTriggerData["buttonSwitch3"] = "color_brightness"

	switch1Trigger := createSwitchTriggerWithBindingAction(triggerId, registrar, "buttonSwitch1", "brightness", 120, mqtt)
	switch2Trigger := createSwitchTriggerWithBindingAction(triggerId, registrar, "buttonSwitch2", "state", true, mqtt)
	switch3Trigger := createSwitchTriggerWithBindingAction(triggerId, registrar, "buttonSwitch3", "color_brightness", 250, mqtt)

	// create device trigger
	deviceTrigger := automations.NewDevice(triggerId)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, switch1Trigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, switch2Trigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, switch3Trigger)

	testCases := []struct {
		triggeredEntity string
		value           any
		result          bool
	}{
		{triggeredEntity: "buttonSwitch1", value: 120.0, result: true},
		{triggeredEntity: "buttonSwitch2", value: true, result: true},
		{triggeredEntity: "door_state", value: true, result: false},
		{triggeredEntity: "co2", value: 70, result: false},
		{triggeredEntity: "quality_index", value: 30.1, result: false},
		{triggeredEntity: "buttonSwitch3", value: 250.0, result: true},
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
		device.Exposes = createExposures(data)

		deviceTrigger.Evaluate(automations.NewDeviceEvent(device))
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()
}

func TestAutomationwithMultipleTriggerActions(t *testing.T) {
	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}
	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}
	triggerId := "Attic light"

	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	exposeNames := []string{"brightness", "color_temperature", "color_brightness"}
	entities := []*devices.Entity{}
	for _, e := range exposeNames {
		e := utils_test.CreateNumericEntity(e, nil)
		entities = append(entities, e)

	}
	device := utils_test.CreateDeviceWithExposes(triggerId, "sensor device", entities)
	deviceBridgeList := utils_test.CreateBridgeInfoList([]*devices.Device{device})
	registrar.RegisterBridge(deviceBridgeList)

	trigger := createTriggerwithMultipleActions(triggerId, registrar, mqtt, "button1", []string{"brightness", "color_temperature", "color_brightness"}, []any{120, 250, 200})
	trigger.Conditions = append(trigger.Conditions, utils_test.NewExposeCondition("button1", "pressed", "="))

	automation := automations.NewDevice(triggerId)
	automation.Triggers = append(automation.Triggers, trigger)

	testCases := []struct {
		triggeredEntity string
		value           any
		result          bool
	}{
		{triggeredEntity: "button1", value: "pressed", result: true},
		{triggeredEntity: "button1", value: "release", result: false},
		{triggeredEntity: "button1", value: "pressed", result: true},
	}

	for _, testCase := range testCases {
		var data = map[string]any{
			"some_data1":             false,
			"some_data2":             90,
			testCase.triggeredEntity: testCase.value,
		}
		var messageHandler = func(id string, payload []byte) {

			for _, action := range trigger.Actions {

				fmt.Println(action)
				data := unpackJsonToMap(string(payload))
				if data == nil {
					t.Fatalf("error unpacking json")
				}
				// triggerAction, ok := action.(*automations.MqttTriggerAction)
				// if !ok {
				// 	t.Fatalf("invalid action type")
				// }

			}

			wg.Done()
		}

		mqtt.OnMessageHandler(messageHandler)

		if testCase.result {
			wg.Add(1)
		}
		device.Exposes = createExposures(data)

		automation.Evaluate(automations.NewDeviceEvent(device))
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()
}

func TestHandleMultipleSameValueTriggerWithDelay(t *testing.T) {

	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}
	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}

	triggerId := "human sensor"
	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	device := utils_test.CreatePresenceDevice(triggerId, "sensor device", "presence", false)

	deviceBridgeList := utils_test.CreateBridgeInfoList([]*devices.Device{device})

	registrar.RegisterBridge(deviceBridgeList)

	turnOffTrigger := createTriggerDelayTurnOffLight(triggerId, registrar, mqtt, utils.IntervalFromSeconds(3))
	turnOnTrigger := createTriggerTurnOnLightWithPresenceOnAndLux(triggerId, registrar, mqtt, 30)

	// create device trigger
	deviceTrigger := automations.NewDevice(triggerId)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOffTrigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOnTrigger)

	testCases := []struct {
		presence bool
		lux      any
		result   bool
	}{
		{presence: false, lux: 30, result: true},
		{presence: false, lux: 15, result: false},
		{presence: false, lux: 6, result: false},
	}

	for _, testCase := range testCases {
		var data = map[string]any{
			"presence": testCase.presence,
			"lux":      testCase.lux,
		}

		var messageHandler = func(id string, payload []byte) {

			for _, action := range turnOnTrigger.Actions {

				data := unpackJsonToMap(string(payload))
				if data == nil {
					t.Fatalf("error unpacking json")
				}
				triggerAction, ok := action.(*automations.MqttTriggerAction)
				if !ok {
					t.Fatalf("invalid action type")
				}

				for _, expose := range triggerAction.Exposes {
					value, ok := data[expose.Name]
					if !ok {
						t.Fatalf("property not %s found in payload", expose.Name)
					}
					if value != testCase.presence {
						t.Fatalf("value mismatch: want %v got %v", testCase.presence, value)
					}
				}
			}

			wg.Done()
		}

		mqtt.OnMessageHandler(messageHandler)
		if testCase.result {
			wg.Add(1)
		}

		device.Exposes = createExposures(data)
		deviceTrigger.Evaluate(automations.NewDeviceEvent(device))

		time.Sleep(500 * time.Millisecond)
	}

	wg.Wait()
}

func TestTurnOnAndOffLightFromPresence(t *testing.T) {

	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}

	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}

	id := "human sensor"
	device := utils_test.CreatePresenceDevice(id, "sensor device", "presence", false)
	deviceBridgeList := utils_test.CreateBridgeInfoList([]*devices.Device{device})

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(deviceBridgeList)
	turnOnTrigger := createTriggerTurnOnLightWithPresenceOnAndLux(id, registrar, mqtt, 30)
	turnOffTrigger := createTriggerDelayTurnOffLightWithPresenceOff(id, registrar, mqtt, utils.IntervalFromMilliseconds(100))

	// create device trigger
	deviceTrigger := automations.NewDevice(id)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOffTrigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOnTrigger)

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

			if !testCase.result {
				t.Fatalf("invalid msg received")
			}

			for _, action := range turnOnTrigger.Actions {

				data := unpackJsonToMap(string(payload))
				if data == nil {
					t.Fatalf("error unpacking json")
				}
				triggerAction, ok := action.(*automations.MqttTriggerAction)
				if !ok {
					t.Fatalf("invalid action type")
				}

				for _, expose := range triggerAction.Exposes {
					value, ok := data[expose.Name]
					if !ok {
						t.Fatalf("property not %s found in payload", expose.Name)
					}
					if value != testCase.presence {
						t.Fatalf("value mismatch: want %v got %v", testCase.presence, value)
					}
				}
			}

			wg.Done()
		}

		mqtt.OnMessageHandler(messageHandler)
		if testCase.result {
			wg.Add(1)
		}

		device.Exposes = createExposures(data)
		deviceTrigger.Evaluate(automations.NewDeviceEvent(device))

		time.Sleep(testCase.sleepdelay * time.Millisecond)
	}

	wg.Wait()
}

func TestActionWithTimerRangeConditionLightFromPresence(t *testing.T) {

	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}
	mockClock := &mocks.MockClock{}

	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}

	id := "human sensor"
	device := utils_test.CreatePresenceDevice(id, "sensor device", "presence", false)
	deviceBridgeList := utils_test.CreateBridgeInfoList([]*devices.Device{device})

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(deviceBridgeList)
	// create turn on trigger
	turnOnTrigger := createTriggerTurnOnLight(id, registrar, mqtt)

	// initialize turn on condition with timer
	onTimeRange := automations.NewTimeRange("11:00", "17:00")
	turnOnCondition := utils_test.NewExposeCondition("presence", true, "=")
	turnOnTimerCondition := utils_test.NewTimeCondition(onTimeRange, mockClock)
	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, turnOnCondition)
	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, turnOnTimerCondition)

	// create turn off trigger
	turnOffTrigger := createTriggerDelayTurnOffLight(id, registrar, mqtt, nil)

	// initialize turn off condition with timer
	offTimeRange := automations.NewTimeRange("09:00", "13:25")
	turnOffCondition := utils_test.NewExposeCondition("presence", false, "=")
	turnOffTimerCondition := utils_test.NewTimeCondition(offTimeRange, mockClock)
	turnOffTrigger.Conditions = append(turnOffTrigger.Conditions, turnOffCondition)
	turnOffTrigger.Conditions = append(turnOffTrigger.Conditions, turnOffTimerCondition)

	// create device trigger
	deviceTrigger := automations.NewDevice(id)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOnTrigger)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOffTrigger)

	testCases := []struct {
		presence   bool
		sleepdelay time.Duration
		timeNow    time.Time
		result     bool
	}{
		{presence: true, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(11, 0, 0), result: true},
		{presence: false, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(12, 15, 0), result: true},
		{presence: true, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(17, 1, 0), result: false},
		{presence: true, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(13, 24, 0), result: true},
		{presence: false, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(6, 25, 0), result: false},
		{presence: false, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(9, 1, 0), result: true},
		{presence: true, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(10, 59, 0), result: false},
		{presence: true, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(12, 25, 0), result: true},
		{presence: false, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(8, 59, 0), result: false},
		{presence: false, sleepdelay: 200, timeNow: utils_test.CreateTimeFrom(11, 25, 0), result: true},
	}

	for i, testCase := range testCases {
		var data = map[string]any{
			"presence": testCase.presence,
		}

		if i == 3 {
			fmt.Println("test case debug")
		}
		mockClock.SetMockTime(testCase.timeNow)
		var messageHandler = func(id string, payload []byte) {

			if !testCase.result {
				t.Fatalf("invalid msg received")
			}

			for _, action := range turnOnTrigger.Actions {

				data := unpackJsonToMap(string(payload))
				if data == nil {
					t.Fatalf("error unpacking json")
				}
				triggerAction, ok := action.(*automations.MqttTriggerAction)
				if !ok {
					t.Fatalf("invalid action type")
				}
				if !ok {
					t.Fatalf("invalid action type")
				}

				for _, expose := range triggerAction.Exposes {
					value, ok := data[expose.Name]
					if !ok {
						t.Fatalf("property not %s found in payload", expose.Name)
					}
					if value != testCase.presence {
						t.Fatalf("value mismatch: want %v got %v", testCase.presence, value)
					}
					if value != testCase.presence {
						t.Fatalf("value mismatch: want %v got %v", testCase.presence, value)
					}
				}
			}

			wg.Done()
		}

		mqtt.OnMessageHandler(messageHandler)
		if testCase.result {
			wg.Add(1)
		}

		device.Exposes = createExposures(data)

		fmt.Println("emiting presence", testCase.presence)
		deviceTrigger.Evaluate(automations.NewDeviceEvent(device))

		time.Sleep(testCase.sleepdelay * time.Millisecond)
	}

	wg.Wait()
}

func TestManualTriggerTurnsOnLight(t *testing.T) {

	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}

	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}

	id := "Light attic"
	device := utils_test.CreatePresenceDevice(id, "light device", "presence", false)
	deviceBridgeList := utils_test.CreateBridgeInfoList([]*devices.Device{device})

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(deviceBridgeList)
	turnOnTrigger := createTriggerTurnOnLight(id, registrar, mqtt)

	// create device trigger
	deviceTrigger := automations.NewDevice(id)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOnTrigger)

	testCases := []struct {
		presence bool
		result   bool
	}{
		{presence: true, result: true},
		{presence: false, result: true},
		{presence: true, result: true},
	}
	for _, testCase := range testCases {
		var data = map[string]any{
			"presence": testCase.presence,
		}

		var messageHandler = func(id string, payload []byte) {
			wg.Done()
			for _, action := range turnOnTrigger.Actions {

				responseData := unpackJsonToMap(string(payload))
				if responseData == nil {
					t.Fatalf("error unpacking json")
				}
				triggerAction, ok := action.(*automations.MqttTriggerAction)
				if !ok {
					t.Fatalf("invalid action type")
				}
				if !ok {
					t.Fatalf("invalid action type")
				}
				for _, expose := range triggerAction.Exposes {
					value, ok := responseData[expose.Name]
					if !ok {
						t.Fatalf("property not %s found in payload", expose.Name)
					}
					if value != expose.Data {
						t.Fatalf("value mismatch: want %v got %v", expose.Data, value)
					}
				}

			}
		}

		mqtt.OnMessageHandler(messageHandler)
		wg.Add(1)

		device.Exposes = createExposures(data)
		deviceTrigger.EvaluateTrigger(automations.NewDeviceEvent(device), turnOnTrigger.Name)
		wg.Wait()
	}
}

func TestManualTriggerWithScheduleTurnsOnLight(t *testing.T) {

	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}
	mockClock := &mocks.MockClock{}
	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}

	id := "Light attic"
	device := utils_test.CreatePresenceDevice(id, "Attic Light", "state", false)
	deviceBridgeList := utils_test.CreateBridgeInfoList([]*devices.Device{device})

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(deviceBridgeList)
	// create turn on trigger
	onTimeRange := automations.NewTimeRange("11:00", "17:00")

	turnOnTrigger := createDeviceTriggerWithScheduleTurnOnLight(id, onTimeRange, registrar, mqtt, mockClock)
	deviceTrigger := automations.NewDevice(id)
	deviceTrigger.Triggers = append(deviceTrigger.Triggers, turnOnTrigger)

	testCases := []struct {
		timeNow time.Time
		result  bool
	}{
		{timeNow: utils_test.CreateTimeFrom(10, 59, 0), result: false},
		{timeNow: utils_test.CreateTimeFrom(11, 0, 0), result: true},
		{timeNow: utils_test.CreateTimeFrom(12, 15, 0), result: true},
		{timeNow: utils_test.CreateTimeFrom(13, 36, 0), result: true},
		{timeNow: utils_test.CreateTimeFrom(14, 7, 10), result: true},
		{timeNow: utils_test.CreateTimeFrom(17, 0, 0), result: true},
		{timeNow: utils_test.CreateTimeFrom(17, 0, 1), result: false},
		{timeNow: utils_test.CreateTimeFrom(18, 10, 30), result: false},
	}

	for _, testCase := range testCases {
		mockClock.SetMockTime(testCase.timeNow)

		var data = map[string]any{
			"presence": true,
		}

		var messageHandler = func(id string, payload []byte) {
			wg.Done()
		}

		mqtt.OnMessageHandler(messageHandler)
		if testCase.result {
			wg.Add(1)
		}

		device.Exposes = createExposures(data)
		deviceTrigger.EvaluateTrigger(automations.NewDeviceEvent(device), turnOnTrigger.Name)

		wg.Wait()
	}
}

func TestSwitch(t *testing.T) {

	// todo
}

func TestEqualityChecks(t *testing.T) {
	testCases := []struct {
		op     utils.EqualityOperator
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
		res := utils.EqualityOperators[testCase.op](testCase.value1, testCase.value2)
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

func createTriggerTurnOnLightWithPresenceOnAndLux(id string, registrar services.DeviceRegistrar, mqtt mqtt.MqttClient, lux any) *automations.DeviceTrigger {
	// action = turn off light
	turnOnAction := automations.NewTriggerAction()
	turnOnAction.Id = id
	turnOnAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: "presence",
			Data: true,
		},
	}
	turnOnAction.Delay = nil
	turnOnAction.Configure(registrar, mqtt)

	// Turn on sensor trigger
	turnOnTrigger := automations.NewDeviceTrigger("presence")
	turnOnTrigger.Actions = []automations.MqttAction{turnOnAction}

	// condition = presence = off && lux <= 30
	turnOnCondition := utils_test.NewExposeCondition("presence", true, "=")
	luxCondition := utils_test.NewExposeCondition("lux", lux, "<=")

	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, turnOnCondition)
	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, luxCondition)

	return turnOnTrigger
}

func createDeviceTriggerWithScheduleTurnOnLight(id string, timeRange *automations.TimeRange, registrar services.DeviceRegistrar, mqtt mqtt.MqttClient, clock utils.Clock) *automations.DeviceTrigger {

	// Turn on sensor trigger
	turnOnTrigger := automations.NewDeviceTrigger(id)

	// action = turn off light
	turnOnAction := automations.NewTriggerAction()
	turnOnAction.Id = id
	turnOnAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: "state",
			Data: true,
		},
	}
	turnOnAction.Delay = nil
	turnOnAction.Configure(registrar, mqtt)
	turnOnTrigger.Actions = []automations.MqttAction{turnOnAction}

	// schedule condition
	scheduleCondition := utils_test.NewTimeCondition(timeRange, clock)
	turnOnTrigger.Conditions = append(turnOnTrigger.Conditions, scheduleCondition)

	return turnOnTrigger
}

func createTriggerDelayTurnOffLightWithPresenceOff(id string, registrar services.DeviceRegistrar, mqtt mqtt.MqttClient, delay *utils.TimeInterval) *automations.DeviceTrigger {

	turnOffTrigger := createTriggerDelayTurnOffLight(id, registrar, mqtt, delay)

	turnOffCondition := utils_test.NewExposeCondition("presence", false, "=")

	turnOffTrigger.Conditions = append(turnOffTrigger.Conditions, turnOffCondition)

	return turnOffTrigger
}

func createTriggerTurnOnLight(id string, registrar services.DeviceRegistrar, mqtt mqtt.MqttClient) *automations.DeviceTrigger {
	// action = turn off light
	turnOnAction := automations.NewTriggerAction()
	turnOnAction.Id = id
	turnOnAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: "presence",
			Data: true,
		},
	}
	turnOnAction.Delay = nil
	turnOnAction.Configure(registrar, mqtt)

	// Turn on sensor trigger
	turnOnTrigger := automations.NewDeviceTrigger("presence")
	turnOnTrigger.Actions = []automations.MqttAction{turnOnAction}

	// condition = presence = off

	return turnOnTrigger
}

func createSwitchTriggerWithBindingAction(id string, registrar services.DeviceRegistrar, triggerName string, actionProp string, actionData any, mqtt mqtt.MqttClient) *automations.DeviceTrigger {
	// action = turn off light
	brightnessAction := automations.NewTriggerAction()
	brightnessAction.Id = id
	brightnessAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: actionProp,
			Data: actionData,
		},
	}
	brightnessAction.Configure(registrar, mqtt)

	// Turn off sensor trigger
	button1Trigger := automations.NewDeviceTrigger(triggerName)
	button1Trigger.Actions = []automations.MqttAction{brightnessAction}

	return button1Trigger
}

func createTriggerwithMultipleActions(id string, registrar services.DeviceRegistrar, mqtt mqtt.MqttClient, triggerName string, actions []string, data []any) *automations.DeviceTrigger {
	brightnessAction := automations.NewTriggerAction()
	brightnessAction.Id = id
	for i, action := range actions {
		brightnessAction.Exposes = append(brightnessAction.Exposes, &automations.MqttTriggerActionExpose{
			Name: action,
			Data: data[i],
		})
	}

	err := brightnessAction.Configure(registrar, mqtt)
	if err != nil {
		fmt.Println("[ERROR] configuring action", err)
		return nil
	}

	trigger := automations.NewDeviceTrigger(triggerName)
	trigger.Actions = []automations.MqttAction{brightnessAction}
	trigger.Conditions = []automations.Condition{}

	return trigger
}

func createTriggerDelayTurnOffLight(id string, registrar services.DeviceRegistrar, mqtt mqtt.MqttClient, delay *utils.TimeInterval) *automations.DeviceTrigger {
	// action = turn off light
	turnOffAction := automations.NewTriggerAction()
	turnOffAction.Id = id
	turnOffAction.Exposes = []*automations.MqttTriggerActionExpose{
		{
			Name: "presence",
			Data: false,
		},
	}

	turnOffAction.Delay = delay
	err := turnOffAction.Configure(registrar, mqtt)
	if err != nil {
		fmt.Println("error configuring delay", err)
	}

	// Turn off sensor trigger
	turnOffTrigger := automations.NewDeviceTrigger("presence")
	turnOffTrigger.Actions = []automations.MqttAction{turnOffAction}

	return turnOffTrigger
}

func createExposures(data map[string]interface{}) map[string]*devices.Entity {
	var entities = make(map[string]*devices.Entity)
	for key, value := range data {

		newEntity := createEntity(key, "", value, "", nil)
		entities[key] = newEntity
	}
	return entities
}

func createEntity(name string, description string, data any, unit string, attributes map[string]any) *devices.Entity {
	if attributes == nil {
		attributes = make(map[string]any)
	}

	newEntity := &devices.Entity{}
	newEntity.Attributes = map[string]any{}
	newEntity.Values = map[string]any{}
	newEntity.Data = devices.NewEntityData(data)
	newEntity.Name = name
	newEntity.Unit = unit
	newEntity.Description = description
	newEntity.Attributes = attributes
	return newEntity
}
