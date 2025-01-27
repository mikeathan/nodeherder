package automations_test

import (
	"fmt"
	"node-herder/internal/automations"
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/devices"
	utils_test "node-herder/testing"
	"sync"
	"testing"
	"time"
)

func createBridgeInfoes() []*devices.BridgeInfo {

	dev1 := &devices.BridgeInfo{}
	dev1.IeeeAddress = "0x123456"
	dev1.Definition.Description = "Mocking human sensor"
	dev1.FriendlyName = "human sensor"
	dev1.Type = "EndDevice"
	dev1.PowerSource = "battery"
	dev1.Disabled = false
	dev1.InterviewCompleted = true

	e := devices.BridgeExpose{}
	e.Name = "presence"
	e.Property = "presence"
	e.Type = "binary"
	e.Description = "determines if presence has been detected"

	dev1.Definition.Exposes = append(dev1.Definition.Exposes, e)

	//
	dev2 := &devices.BridgeInfo{}
	dev2.IeeeAddress = "0x56789"
	dev2.Definition.Description = "Mocking Attic light"
	dev2.FriendlyName = "Attic light"
	dev2.Type = "EndDevice"
	dev2.PowerSource = "mains"
	dev2.Disabled = false
	dev2.InterviewCompleted = true

	e2 := devices.BridgeExpose{}
	e2.Type = "light"
	b2f1 := devices.BridgeInfoFeature{}
	b2f1.Description = "On/off state of this light"
	b2f1.Name = "state"
	b2f1.Property = "state"
	b2f1.Type = "binary"
	b2f1.ValueOff = "OFF"
	b2f1.ValueOn = "ON"

	b2f2 := devices.BridgeInfoFeature{}
	b2f2.Description = "Brightness of this light"
	b2f2.Name = "brightness"
	b2f2.Property = "brightness"
	b2f2.Type = "numeric"
	b2f2.ValueMax = 255
	b2f2.ValueMin = 0

	e2.Features = append(e2.Features, b2f1)
	e2.Features = append(e2.Features, b2f2)

	dev2.Definition.Exposes = append(dev2.Definition.Exposes, e2)

	return []*devices.BridgeInfo{dev1, dev2}
}

// func TestExportAutomationsFromFile(t *testing.T) {

// 	mqtt := &mocks.MockMqttClient{}
// 	store := utils_test.CreateStore()
// 	eventHub := &mocks.MockEventHub{}
// 	registrar := services.NewHubRegisterService(store, eventHub, 30000)

// 	dev1 := createMockDevice("0x56789", "livingroom", "brightness", nil, 0.0, 255.0)
// 	dev2 := createMockDevice("0x56789", "humansensor", "left_click", nil, 0.0, 255.0)

// 	bridgeinfos := createBridgeInfoes()
// 	registrar.RegisterBridge(bridgeinfos, 60)
// 	registrar.Register("livingroom", dev1)
// 	registrar.Register("humansensor", dev2)

// 	storage := mocks.NewMockAutomationStorage([]*automations.Device{})

// 	engine := automations.NewEngine([]automations.AutomationHandler{}, registrar, mqtt)
// 	engine.WithStorage(storage)
// 	engine.Initialize()

// 	turnOffTrigger := createTriggerDelayTurnOffLightWithPresenceOff(mqtt, 100*time.Millisecond)
// 	turnOffTrigger.Actions[0].Id = "0x56789"

// 	turnOnTrigger := createTriggerTurnOnLightWithPresenceOnAndLux(mqtt, 30.1)
// 	turnOnTrigger.Actions[0].Id = "0x56789"

// 	// create device trigger
// 	inputDeviceTriggers := []*automations.Device{}
// 	deviceTrigger1 := automations.NewDevice("0x123456")
// 	deviceTrigger1.FriendlyName = "humansensor"
// 	deviceTrigger1.Description = "test human sensor automation"
// 	deviceTrigger1.Triggers = append(deviceTrigger1.Triggers, turnOffTrigger)
// 	deviceTrigger1.Triggers = append(deviceTrigger1.Triggers, turnOnTrigger)
// 	inputDeviceTriggers = append(inputDeviceTriggers, deviceTrigger1)

// 	for _, d := range inputDeviceTriggers {
// 		err := engine.Add(d)
// 		if err != nil {
// 			t.Fatalf("ERROR adding trigger %s", err.Error())
// 		}
// 	}

// 	outputDeviceTriggers := engine.GetAllTriggers()
// 	if len(outputDeviceTriggers) != len(inputDeviceTriggers) {
// 		t.Fatalf("ERROR size mismatch. want %v got %d", len(inputDeviceTriggers), len(outputDeviceTriggers))
// 	}
// 	for dIdx, outDeviceTrigger := range outputDeviceTriggers {
// 		inputDeviceTrigger := inputDeviceTriggers[dIdx]
// 		if outDeviceTrigger.Id != inputDeviceTrigger.Id {
// 			t.Fatalf("ERROR Id mismatch")
// 		}
// 		if outDeviceTrigger.FriendlyName != inputDeviceTrigger.FriendlyName {
// 			t.Fatalf("ERROR Name mismatch")
// 		}
// 		if outDeviceTrigger.Description != inputDeviceTrigger.Description {
// 			t.Fatalf("ERROR Description mismatch")
// 		}
// 		if outDeviceTrigger.Enabled != inputDeviceTrigger.Enabled {
// 			t.Fatalf("ERROR Description mismatch")
// 		}

// 		for tidx, outputTrigger := range outDeviceTrigger.Triggers {
// 			inputTrigger := inputDeviceTrigger.Triggers[tidx]

// 			if outputTrigger.Name != inputTrigger.Name {
// 				t.Fatalf("ERROR Trigger.Name mismatch")
// 			}

// 			// check actions
// 			for aidx, outputAction := range outputTrigger.Actions {
// 				inputAction := inputTrigger.Actions[aidx]

// 				if outputAction.Id != inputAction.Id {
// 					t.Fatalf("ERROR Action.Id mismatch")
// 				}
// 				if outputAction.FriendlyName != inputAction.FriendlyName {
// 					t.Fatalf("ERROR Action.Friendlyname mismatch")
// 				}

// 				if outputAction.Property != inputAction.Property {
// 					t.Fatalf("ERROR Action.Property mismatch")
// 				}

// 				if outputAction.Delay != inputAction.Delay {
// 					t.Fatalf("ERROR Action.Delay mismatch")
// 				}
// 			}

// 			for cidx, outputCondition := range outputTrigger.Conditions {

// 				inputCondition := inputTrigger.Conditions[cidx]

// 				if outputCondition.EqualityOperator != inputCondition.EqualityOperator {
// 					t.Fatalf("ERROR Condition.EqualityOperator mismatch")
// 				}
// 				if outputCondition.Name != inputCondition.Name {
// 					t.Fatalf("ERROR Condition.Name mismatch")
// 				}

// 				if outputCondition.Value != inputCondition.Value {
// 					t.Fatalf("ERROR Condition.Value mismatch")
// 				}
// 			}
// 		}
// 	}

// 	for _, inputDeviceTrigger := range inputDeviceTriggers {
// 		err := engine.Delete(inputDeviceTrigger.Id)
// 		if err != nil {
// 			t.Fatalf("ERROR deleting file %v .Error %s", inputDeviceTrigger.Id, err.Error())
// 		}
// 	}

// 	inputDeviceTriggers = engine.GetAllTriggers()
// 	if len(inputDeviceTriggers) != 0 {
// 		t.Fatalf("ERROR triggers found. expecting empty")
// 	}
// }

func TestEngineAutomationUpdateShouldNotResetScheduler(t *testing.T) {

	wg := sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}

	mqtt.OnMessageHandler(func(topic string, payload []byte) {
		fmt.Printf("Received message on topic %s\n", topic)
	})

	store := utils_test.CreateStore()
	eventHub := &mocks.MockEventHub{}

	alarmDevice := utils_test.CreateAlarmDevice("x02222222", "alarm device", false)
	doorSensorDevice := utils_test.CreateDoorSensorDevice("x01111111", "front door sensor", false)

	// setup bridgeInfo List
	devices := []*devices.Device{doorSensorDevice, alarmDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices)
	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	registrar.RegisterBridge(deviceBridgeList, 60)

	deviceAutomation := utils_test.CreateDoorContactWithAlarmTriggerAutomation("x01111111", "x02222222", mqtt)

	now := time.Now().UTC()
	start := now.Add(500 * time.Millisecond)
	end := now.Add(1 * time.Hour)

	deviceAutomation.Schedules = utils_test.CreateTimeSchedules(start, end)
	storage := mocks.NewMockAutomationStorage([]*automations.Device{deviceAutomation})

	wg.Add(1) // we expect only one event to be triggered

	scheduleHandler := automations.NewAutomationScheduler(
		automations.WithScheduleFunc("enable", func(automation *automations.Device) error {
			automation.Enabled = true
			wg.Done()
			return nil
		}),
		automations.WithScheduleFunc("disable", func(automation *automations.Device) error {
			automation.Enabled = false
			return nil
		}),
	)

	engine := automations.NewEngine([]automations.AutomationHandler{scheduleHandler}, registrar, mqtt)
	engine.WithStorage(storage)
	engine.Initialize()

	// make sure automation is disabled on startup
	if deviceAutomation.Enabled {
		t.Fatalf("ERROR automation is enabled")
	}

	time.Sleep(500 * time.Millisecond)

	// automation must be enabled from scheduler
	if !deviceAutomation.Enabled {
		t.Fatalf("ERROR automation is disable")
	}

	// update automation but keep the same schedule
	engine.Add(deviceAutomation)
	time.Sleep(50 * time.Millisecond)

	// assert that scheduler is still running
	if !scheduleHandler.IsRunning(deviceAutomation) {
		t.Fatalf("ERROR scheduler is not running")
	}
	// we expect only one event to be triggered and release the wait group
	// any more events triggered will cause the test to fail
	wg.Wait()
}

func TestEngineAutomationUpdateShouldResetAndTriggerAgainScheduler(t *testing.T) {

	wg := sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}

	mqtt.OnMessageHandler(func(topic string, payload []byte) {
		fmt.Printf("Received message on topic %s\n", topic)
	})

	store := utils_test.CreateStore()
	eventHub := &mocks.MockEventHub{}

	alarmDevice := utils_test.CreateAlarmDevice("x02222222", "alarm device", false)
	doorSensorDevice := utils_test.CreateDoorSensorDevice("x01111111", "front door sensor", false)

	// setup bridgeInfo List
	devices := []*devices.Device{doorSensorDevice, alarmDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices)
	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	registrar.RegisterBridge(deviceBridgeList, 60)

	deviceAutomation := utils_test.CreateDoorContactWithAlarmTriggerAutomation("x01111111", "x02222222", mqtt)

	now := time.Now().UTC()
	start := now.Add(500 * time.Millisecond)
	end := now.Add(1 * time.Hour)

	deviceAutomation.Schedules = utils_test.CreateTimeSchedules(start, end)
	storage := mocks.NewMockAutomationStorage([]*automations.Device{deviceAutomation})

	// we expect only 2 event to be triggered
	// first event is on startup
	// second event is on automation schedule update
	wg.Add(2)

	scheduleHandler := automations.NewAutomationScheduler(
		automations.WithScheduleFunc("enable", func(automation *automations.Device) error {
			automation.Enabled = true
			wg.Done()
			return nil
		}),
		automations.WithScheduleFunc("disable", func(automation *automations.Device) error {
			automation.Enabled = false
			return nil
		}),
	)

	engine := automations.NewEngine([]automations.AutomationHandler{scheduleHandler}, registrar, mqtt)
	engine.WithStorage(storage)
	engine.Initialize()

	// make sure automation is disabled on startup
	if deviceAutomation.Enabled {
		t.Fatalf("ERROR automation is enabled")
	}

	time.Sleep(500 * time.Millisecond)

	// automation must be enabled from scheduler
	if !deviceAutomation.Enabled {
		t.Fatalf("ERROR automation is disable")
	}

	// update automation and change schedule
	now = time.Now().UTC()
	start = now.Add(50 * time.Millisecond)
	deviceAutomation.Schedules[0] = utils_test.CreateTimeSchedule(start)

	engine.Add(deviceAutomation)
	time.Sleep(50 * time.Millisecond)

	// assert that scheduler is still running
	if !scheduleHandler.IsRunning(deviceAutomation) {
		t.Fatalf("ERROR scheduler is not running")
	}
	// we expect only 2 events to be triggered and release the wait group
	// any more events triggered will cause the test to fail
	wg.Wait()
}

func TestEngineAutomationUpdateShouldResetScheduler(t *testing.T) {

	wg := sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}

	mqtt.OnMessageHandler(func(topic string, payload []byte) {
		fmt.Printf("Received message on topic %s\n", topic)
	})

	store := utils_test.CreateStore()
	eventHub := &mocks.MockEventHub{}

	alarmDevice := utils_test.CreateAlarmDevice("x02222222", "alarm device", false)
	doorSensorDevice := utils_test.CreateDoorSensorDevice("x01111111", "front door sensor", false)

	// setup bridgeInfo List
	devices := []*devices.Device{doorSensorDevice, alarmDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices)
	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	registrar.RegisterBridge(deviceBridgeList, 60)

	deviceAutomation := utils_test.CreateDoorContactWithAlarmTriggerAutomation("x01111111", "x02222222", mqtt)

	now := time.Now().UTC()
	start := now.Add(500 * time.Millisecond)
	end := now.Add(1 * time.Hour)

	deviceAutomation.Schedules = utils_test.CreateTimeSchedules(start, end)
	storage := mocks.NewMockAutomationStorage([]*automations.Device{deviceAutomation})

	// we expect only 1 event to be triggered
	// first event is on startup
	// after we update the automation schedule, the second event ashould not be trigger on time
	wg.Add(1)

	scheduleHandler := automations.NewAutomationScheduler(
		automations.WithScheduleFunc("enable", func(automation *automations.Device) error {
			automation.Enabled = true
			wg.Done()
			return nil
		}),
		automations.WithScheduleFunc("disable", func(automation *automations.Device) error {
			automation.Enabled = false
			return nil
		}),
	)

	engine := automations.NewEngine([]automations.AutomationHandler{scheduleHandler}, registrar, mqtt)
	engine.WithStorage(storage)
	engine.Initialize()

	// make sure automation is disabled on startup
	if deviceAutomation.Enabled {
		t.Fatalf("ERROR automation is enabled")
	}

	time.Sleep(500 * time.Millisecond)

	// automation must be enabled from scheduler
	if !deviceAutomation.Enabled {
		t.Fatalf("ERROR automation is disable")
	}

	// update automation and change schedule
	now = time.Now().UTC()
	start = now.Add(1000 * time.Millisecond)
	deviceAutomation.Schedules[0] = utils_test.CreateTimeSchedule(start)

	engine.Add(deviceAutomation)
	time.Sleep(50 * time.Millisecond)

	// assert that scheduler is still running
	if !scheduleHandler.IsRunning(deviceAutomation) {
		t.Fatalf("ERROR scheduler is not running")
	}
	// we expect only 1 event to be triggered and release the wait group
	// any more events triggered will cause the test to fail
	wg.Wait()
}

func TestEngineAutomationUpdateShouldStopScheduler(t *testing.T) {

	wg := sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}

	mqtt.OnMessageHandler(func(topic string, payload []byte) {
		fmt.Printf("Received message on topic %s\n", topic)
	})

	store := utils_test.CreateStore()
	eventHub := &mocks.MockEventHub{}

	alarmDevice := utils_test.CreateAlarmDevice("x02222222", "alarm device", false)
	doorSensorDevice := utils_test.CreateDoorSensorDevice("x01111111", "front door sensor", false)

	// setup bridgeInfo List
	devices := []*devices.Device{doorSensorDevice, alarmDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices)
	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	registrar.RegisterBridge(deviceBridgeList, 60)

	deviceAutomation := utils_test.CreateDoorContactWithAlarmTriggerAutomation("x01111111", "x02222222", mqtt)

	now := time.Now().UTC()
	start := now.Add(500 * time.Millisecond)
	end := now.Add(1 * time.Hour)

	deviceAutomation.Schedules = utils_test.CreateTimeSchedules(start, end)
	storage := mocks.NewMockAutomationStorage([]*automations.Device{deviceAutomation})

	// we expect only 1 event to be triggered
	// first event is on startup
	// after we update the automation schedule, the second event ashould not be trigger on time
	wg.Add(1)

	scheduleHandler := automations.NewAutomationScheduler(
		automations.WithScheduleFunc("enable", func(automation *automations.Device) error {
			automation.Enabled = true
			wg.Done()
			return nil
		}),
		automations.WithScheduleFunc("disable", func(automation *automations.Device) error {
			automation.Enabled = false
			return nil
		}),
	)

	engine := automations.NewEngine([]automations.AutomationHandler{scheduleHandler}, registrar, mqtt)
	engine.WithStorage(storage)
	engine.Initialize()

	// make sure automation is disabled on startup
	if deviceAutomation.Enabled {
		t.Fatalf("ERROR automation is enabled")
	}

	time.Sleep(500 * time.Millisecond)

	// automation must be enabled from scheduler
	if !deviceAutomation.Enabled {
		t.Fatalf("ERROR automation is disable")
	}

	// update automation remove schedules
	deviceAutomation.Schedules = []*automations.TimeSchedule{}

	engine.Add(deviceAutomation)
	time.Sleep(50 * time.Millisecond)

	// assert that scheduler is not running
	if scheduleHandler.IsRunning(deviceAutomation) {
		t.Fatalf("ERROR scheduler is not running")
	}
	// we expect only 1 event to be triggered and release the wait group
	// any more events triggered will cause the test to fail
	wg.Wait()
}

func TestEngineSchedulerConfiguresAutomation(t *testing.T) {

	wg := sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}

	mqtt.OnMessageHandler(func(topic string, payload []byte) {
		fmt.Printf("Received message on topic %s\n", topic)
	})
	store := utils_test.CreateStore()
	eventHub := &mocks.MockEventHub{}

	alarmDevice := utils_test.CreateAlarmDevice("x02222222", "alarm device", false)
	doorSensorDevice := utils_test.CreateDoorSensorDevice("x01111111", "front door sensor", false)

	// setup bridgeInfo List
	devices := []*devices.Device{doorSensorDevice, alarmDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices)
	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	registrar.RegisterBridge(deviceBridgeList, 60)

	deviceAutomation := utils_test.CreateDoorContactWithAlarmTriggerAutomation("x01111111", "x02222222", mqtt)

	now := time.Now().UTC()
	start := now.Add(500 * time.Millisecond)
	end := now.Add(1500 * time.Millisecond)

	deviceAutomation.Schedules = utils_test.CreateTimeSchedules(start, end)
	storage := mocks.NewMockAutomationStorage([]*automations.Device{deviceAutomation})

	scheduleHandler := automations.NewAutomationScheduler(
		automations.WithScheduleFunc("enable", func(automation *automations.Device) error {
			automation.Enabled = true
			wg.Done()
			return nil
		}),
		automations.WithScheduleFunc("disable", func(automation *automations.Device) error {
			automation.Enabled = false
			wg.Done()

			return nil
		}),
	)
	engine := automations.NewEngine([]automations.AutomationHandler{scheduleHandler}, registrar, mqtt)
	engine.WithStorage(storage)
	engine.Initialize()

	wg.Add(1)

	a, _ := engine.Load(deviceAutomation.Id)

	// make sure automation is disabled when we have scheduler enabled
	if a.Enabled {
		t.Fatalf("ERROR automation is enabled (initial state)")
	}

	// trigger the automation
	doorSensorDevice.Exposes["contact"].Data = true
	engine.HandleDevice(doorSensorDevice)
	time.Sleep(50 * time.Millisecond)

	// scheduler is enable so any event will enable automation
	wg.Wait()

	a, _ = engine.Load(deviceAutomation.Id)

	// scheduler should have enabled automation
	if !a.Enabled {
		t.Fatalf("ERROR automation is not enabled")
	}

	// wait until end of schedule to disable automation
	wg.Add(1)
	wg.Wait()

	// trigger the automation
	doorSensorDevice.Exposes["contact"].Data = true
	engine.HandleDevice(doorSensorDevice)
	time.Sleep(50 * time.Millisecond)

	a, _ = engine.Load(deviceAutomation.Id)

	// scheduler should have disabled automation
	if a.Enabled {
		t.Fatalf("ERROR automation is enabled")
	}
}
