package controllers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"node-herder/internal/automations"
	"node-herder/internal/controllers"
	"node-herder/internal/services"
	"node-herder/internal/ws"
	"node-herder/mocks"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/hub"
	"node-herder/models/logging"
	"node-herder/models/metrics"
	"node-herder/models/settings"
	"node-herder/repository"
	"node-herder/store"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"
)

const device1BatterySource = `{"id":"device 1","conn":"mqtt","power_source":"battery","humidity":92.49999999999999,"temperature":19.000000000000004,"availability":"online","last_seen":"2023-07-20T19:48:35+01:00","linkquality":47,"battery":98}`
const device2 = `{"battery":98, "humidity":71.2,  "linkquality":36.1,"temperature":17.1,"voltage":2999}`
const device3NoLastSeen = `{"id":"device 1","conn":"mqtt","power_source":"battery","humidity":91.12,"temperature":19.000000000000004,"availability":"online","linkquality":47,"battery":67}`

func TestDoorTriggersDoorAlarmAutomation(t *testing.T) {
	mqtt := &mocks.MockMqttClient{}
	ws := &mocks.NopWsServer{}

	deviceAutomation := utils_test.CreateDoorContactDurationWithAlarmTriggerAutomation("x01111111", "x02222222", mqtt)

	automationStorage := mocks.NewMockAutomationStorage[automations.Automation]([]automations.Automation{deviceAutomation})

	// setup device
	alarmDevice := utils_test.CreateAlarmDeviceWithDuration("x02222222", "alarm device", false, 1)
	doorSensorDevice := utils_test.CreateDoorSensorDevice("x01111111", "front door sensor", false)

	// setup bridgeInfo List
	devices := []*devices.Device{doorSensorDevice, alarmDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices)

	// register hub
	store := utils_test.CreateStore()
	hub := controllers.RegisterHubController(ws, store, mqtt)

	hub.WithAutomationStorage(automationStorage)

	//  publish deviceBridgeList to configure hub with devices
	mqtt.Publish("bridge/devices", deviceBridgeList)

	time.Sleep(100 * time.Millisecond)

	testCases := []struct {
		name           string
		value          bool
		expectedResult bool
	}{
		{"contact", true, true},
		{"contact", false, false},
		{"contact", true, true},
	}

	for _, testCase := range testCases {

		expose := testCase.name
		value := testCase.value
		expectedResult := testCase.expectedResult
		payload := map[string]any{expose: value}
		mqtt.Publish(doorSensorDevice.FriendlyName, payload)
		time.Sleep(50 * time.Millisecond)

		alarm, _ := store.FindDeviceById("x02222222")
		if alarm.Exposes["alarm"].Data != expectedResult {
			t.Errorf("alarm should be %v when door sensor triggers", expectedResult)
		}
		if alarm.Exposes["duration"].Data != float64(2) {
			t.Errorf("alarm duration should be 2 got %v", alarm.Exposes["duration"].Data)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestManualTriggerTurnsOnLightAutomation(t *testing.T) {

	ws := &mocks.NopWsServer{}
	//wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}

	eventHub := &mocks.MockEventHub{}

	id := "Light attic"
	name := "light device"

	// create mqtt response for device with specified data
	// to simulate loop feedback
	mqtt.AddResponse(name, map[string]any{"state": true})

	entity := utils_test.CreateEntity("state", bridge.BinaryDataType, false)
	device := utils_test.CreateDeviceWithExposes(id, name, []*devices.Entity{entity})
	deviceBridgeList := utils_test.CreateBridgeInfoList([]*devices.Device{device})

	store := utils_test.CreateStore()
	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(deviceBridgeList)
	toggleLightTrigger := utils_test.CreateTriggerToggleLight(id, registrar, mqtt)

	// create device trigger automation without condition
	lightAutomation := automations.NewDevice(id)
	lightAutomation.Enabled = true
	lightAutomation.Triggers = append(lightAutomation.Triggers, toggleLightTrigger)

	automationStorage := mocks.NewMockAutomationStorage[automations.Automation]([]automations.Automation{lightAutomation})

	hub := controllers.RegisterHubController(ws, store, mqtt)
	hub.WithAutomationStorage(automationStorage)

	//  publish deviceBridgeList to configure hub with devices
	mqtt.Publish("bridge/devices", deviceBridgeList)

	time.Sleep(100 * time.Millisecond)

	payload := map[string]any{"state": "TOGGLE"}

	mqtt.Publish(name, payload)

	time.Sleep(5 * time.Minute)

}
func TestProcessorTriggerScheduledAutomation(t *testing.T) {

	mqtt := &mocks.MockMqttClient{}
	ws := &mocks.NopWsServer{}

	deviceAutomation := utils_test.CreateDoorContactDurationWithAlarmTriggerAutomation("x01111111", "x02222222", mqtt)

	clock := mocks.NewMockClock(func() time.Time {
		return time.Now().UTC()
	})
	now := time.Now().UTC()
	start := now.Add(1000 * time.Millisecond)
	end := now.Add(3000 * time.Millisecond)

	deviceAutomation.Schedules = utils_test.CreateTimeSchedules(start, end)
	automationStorage := mocks.NewMockAutomationStorage[automations.Automation]([]automations.Automation{deviceAutomation})

	// setup device
	alarmDevice := utils_test.CreateAlarmDeviceWithDuration("x02222222", "alarm device", false, 2)
	doorSensorDevice := utils_test.CreateDoorSensorDevice("x01111111", "front door sensor", false)
	// setup bridgeInfo List
	devices := []*devices.Device{doorSensorDevice, alarmDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices)

	// register hub
	store := utils_test.CreateStore()

	// overide automation handlers
	wg := sync.WaitGroup{}

	automationHandlers := []automations.AutomationHandler{
		automations.NewAutomationScheduler(
			automations.WithContext(context.Background()),
			automations.WithSchedulerClock(clock),
			automations.WithCustomScheduleFuncs(map[string]func(automations.Automation) error{
				"enable": func(a automations.Automation) error {
					a.SetEnabled(true)
					wg.Done()
					return nil
				},
				"disable": func(a automations.Automation) error {
					a.SetEnabled(false)
					wg.Done()
					return nil
				},
			})),
	}

	wg.Add(2)

	hub := controllers.RegisterHubController(ws, store, mqtt, controllers.WithAutomationHandlers(automationHandlers))

	hub.WithAutomationStorage(automationStorage) // overide storage

	//  publish deviceBridgeList to configure hub with devices
	mqtt.Publish("bridge/devices", deviceBridgeList)

	time.Sleep(100 * time.Millisecond)

	testCases := []struct {
		name                 string
		value                bool
		expectedResult       bool
		sleepBeforeNextEvent time.Duration
	}{
		{"contact", true, true, 1000 * time.Millisecond},
		{"contact", false, false, 500 * time.Millisecond},
		{"contact", true, false, 1500 * time.Millisecond}, // this should not trigger automation chain as it should be disabled by schedule
	}

	for idx, testCase := range testCases {

		expose := testCase.name
		value := testCase.value
		expectedResult := testCase.expectedResult

		clock.Advance(testCase.sleepBeforeNextEvent)

		payload := map[string]any{expose: value}

		mqtt.Publish(doorSensorDevice.FriendlyName, payload)
		time.Sleep(20 * time.Millisecond)

		// alarm should be triggered only when schedule is due
		alarm, _ := store.FindDeviceById("x02222222")
		if alarm.Exposes["alarm"].Data != expectedResult {
			t.Errorf("case %d: alarm should be %v when door sensor triggers. got %v", idx, expectedResult, alarm.Exposes["alarm"].Data)
		}
	}

	wg.Wait()
}

func TestProcessorTriggersStepActionDialAutomations(t *testing.T) {

	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}
	ws := &mocks.NopWsServer{}

	// SETUP START
	// setup automations
	dialRotateSlowTrigger := utils_test.CreateDialTriggerStepActionBrightness("x02222222", "x01111111", "dial_rotate_left_slow", mqtt)
	btn1PressTrigger := utils_test.CreateDialTriggerActionsBrightness("x02222222", "button_1_press", mqtt)
	btn2PressTrigger := utils_test.CreateDialTriggerActionsBrightness("x02222222", "button_2_press", mqtt)

	deviceAutomation := automations.NewDevice("human sensor")
	deviceAutomation.Id = "x01111111"
	deviceAutomation.FriendlyName = "dial button"
	deviceAutomation.Enabled = true
	deviceAutomation.Triggers = automations.TriggerList{dialRotateSlowTrigger, btn1PressTrigger, btn2PressTrigger}
	automationStorage := mocks.NewMockAutomationStorage[automations.Automation]([]automations.Automation{deviceAutomation})

	// setup device
	device1Expose1 := utils_test.CreateEnumEntity("action", utils_test.CreateDialActionEnums())
	device1Expose2 := utils_test.CreateNumericEntity("action_time", 0)
	dialDevice := utils_test.CreateDeviceWithExposes("x01111111", "Dial button", []*devices.Entity{device1Expose1, device1Expose2})

	device2Expose1 := utils_test.CreateEntity("brightness", "numeric", nil)
	device2Expose2 := utils_test.CreateEnumEntity("color_temp", utils_test.CreateColorTempPresets())
	lightDevice := utils_test.CreateDeviceWithExposes("x02222222", "Attic light", []*devices.Entity{device2Expose1, device2Expose2})

	// setup bridgeInfo List
	devices := []*devices.Device{dialDevice, lightDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices) // NEED TO FIX, currently i make all devices features which is not right!!!!

	// register hub
	store := utils_test.CreateStore()
	hub := controllers.RegisterHubController(ws, store, mqtt)
	hub.WithAutomationStorage(automationStorage) // overide storage
	//  publish deviceBridgeList to configure hub with devices
	mqtt.Publish("bridge/devices", deviceBridgeList)
	time.Sleep(100 * time.Millisecond) // give it time to configure bridgeInfo
	//  SETUP END

	// publish light device
	payload := map[string]any{"brightness": 10.0, "color_temp": 100}
	mqtt.Publish(lightDevice.FriendlyName, payload)

	// publish dial button device
	payload = map[string]any{"action": "button_2_hold"} // this event shouldnt trigger autonation as is not in automation condition
	mqtt.Publish(dialDevice.FriendlyName, payload)

	time.Sleep(50 * time.Millisecond)

	light, _ := store.FindDeviceById("x02222222")
	max := light.Exposes["brightness"].Attributes["max"].(float64)

	numTriggers := 10

	rotateStepAction := dialRotateSlowTrigger.Actions[0].(*automations.MqttStepAction)

	wg.Add(numTriggers)
	for i := 0; i < numTriggers; i++ {

		prevValue, _ := light.Exposes["brightness"].Data.(float64)

		action_time := 10 + (i * 2)
		payload = map[string]any{"action": "dial_rotate_left_slow", "action_direction": "left", "action_time": action_time, "action_type": "step"}
		mqtt.Publish(dialDevice.FriendlyName, payload)

		time.Sleep(50 * time.Millisecond)

		// assert. calculate expected value
		wantvalue := prevValue + (float64(action_time) * rotateStepAction.Data.(float64))
		gotValue, _ := light.Exposes["brightness"].Data.(float64)

		wantvalue = math.Min(wantvalue, max)

		if gotValue != wantvalue {
			t.Fatalf("brightness value mismatch. want %v got %v", wantvalue, gotValue)
		}

		wg.Done()
	}

	wg.Wait()
}

func TestHubEnableRemoteLogger(t *testing.T) {
	mqtt := &mocks.MockMqttClient{}
	ws := &mocks.NopWsServer{}

	// cleanup any previous remote logger hooks
	utils.RemoveRemoteLoggerHook()

	// setup device
	devices := createMockDialAndLightDevices("x01111111", "x02222222")
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices) // NEED TO FIX, currently i make all devices features which is not right!!!!

	// register hub
	tasks := []settings.Task{store.DefaultRemoteLoggerTask()}

	store, cleanup, err := utils_test.CreateStoreWithTasks(tasks)
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	appCfg := store.AppConfig()
	expectedEventName := "logger"

	index := 0
	handler := func(eventName string, data interface{}) error {

		message, err := json.Marshal(data)
		if err != nil {
			t.Errorf("Failed to marshal message: %v", err.Error())
		}

		if eventName != expectedEventName {
			t.Errorf("event name is not correct want: %s got: %s", expectedEventName, eventName)
		}
		var logMessage logging.LogMessage
		err = json.Unmarshal(message, &logMessage)

		if err != nil {
			t.Errorf("Failed to unmarshal message: %v", err.Error())
		}

		if logMessage.Level == "debug" {
			t.Errorf("log level is debug and is unsupported	")
		}

		expectedLogLevel := "info"
		expectedLogMessage := "Remote hook enabled: true"
		if logMessage.Message != expectedLogMessage {
			t.Errorf("log message is not correct want: %s got: %s", expectedLogMessage, logMessage.Message)
		}

		if logMessage.Level != expectedLogLevel {
			t.Errorf("log level is not correct want: %s got: %s", expectedLogLevel, logMessage.Level)
		}
		index++

		return nil
	}

	// register mock remote logger
	remoteLogEmitter := mocks.NewMockRemoteLoggerEmitter(handler)
	utils.RegisterRemoteLoggerHook(remoteLogEmitter)

	controllers.RegisterHubController(ws, store, mqtt)

	// find a way to test the remote logger
	mqtt.Publish("bridge/devices", deviceBridgeList)
	time.Sleep(100 * time.Millisecond) // give it time to configure bridgeInfo

	expectedEnabledMessages := 5
	testCases := []bool{true, false, true, false, true, false, true, false, true, false}
	for _, enabled := range testCases {

		logger := settings.NewLoggerConfig(enabled)
		appCfg.SaveLoggerConfig(logger)
		if enabled {
			utils.LogInfo("Remote hook enabled: true")
		}
		time.Sleep(100 * time.Millisecond)
	}

	if index != expectedEnabledMessages {
		t.Errorf("expected %d messages got %d", expectedEnabledMessages, index)
	}
}

func TestHubTriggersRemoteLogger(t *testing.T) {

	mqtt := &mocks.MockMqttClient{}
	ws := &mocks.NopWsServer{}
	wg := sync.WaitGroup{}
	// cleanup any previous remote logger hooks
	utils.RemoveRemoteLoggerHook()

	// setup device
	devices := createMockDialAndLightDevices("x01111111", "x02222222")
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices) // NEED TO FIX, currently i make all devices features which is not right!!!!

	dialDevice := devices[0]
	lightDevice := devices[1]

	// register hub
	store, cleanup, err := utils_test.CreateFileStore()
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	expectedEventName := "logger"
	expectedRemoteLogMessages := []logging.LogMessage{
		{Level: "info", Message: "device [x02222222] Attic light is online"},
		{Level: "info", Message: "device [x01111111] Dial button is online"},
	}

	index := 0
	handler := func(eventName string, data interface{}) error {

		message, err := json.Marshal(data)
		if err != nil {
			t.Errorf("Failed to marshal message: %v", err.Error())
		}

		if eventName != expectedEventName {
			t.Errorf("event name is not correct want: %s got: %s", expectedEventName, eventName)
		}
		var logMessage logging.LogMessage
		err = json.Unmarshal(message, &logMessage)

		if err != nil {
			t.Errorf("Failed to unmarshal message: %v", err.Error())
		}

		if logMessage.Level == "debug" {
			t.Errorf("log level is debug and is unsupported	")
		}

		expectedRemoteLogMessage := expectedRemoteLogMessages[index]
		if logMessage.Message != expectedRemoteLogMessage.Message {
			t.Errorf("log message is not correct want: %s got: %s", expectedRemoteLogMessage.Message, logMessage.Message)
		}

		if logMessage.Level != expectedRemoteLogMessage.Level {
			t.Errorf("log level is not correct want: %s got: %s", expectedRemoteLogMessage.Level, logMessage.Level)
		}
		index++

		wg.Done()
		return nil
	}

	// register mock remote logger
	remoteLogEmitter := mocks.NewMockRemoteLoggerEmitter(handler)
	utils.RegisterRemoteLoggerHook(remoteLogEmitter)

	controllers.RegisterHubController(ws, store, mqtt)

	// find a way to test the remote logger
	mqtt.Publish("bridge/devices", deviceBridgeList)
	time.Sleep(100 * time.Millisecond) // give it time to configure bridgeInfo

	utils.EnableRemoteLoggerHook(true)

	wg.Add(1)
	// publish light device
	payload := map[string]any{"brightness": 10.0, "color_temp": 100}
	mqtt.Publish(lightDevice.FriendlyName, payload)
	wg.Wait()

	wg.Add(1)
	// publish dial button device
	payload = map[string]any{"action": "button_2_hold"}
	mqtt.Publish(dialDevice.FriendlyName, payload)
	wg.Wait()

	utils.EnableRemoteLoggerHook(false)

	if index != len(expectedRemoteLogMessages) {
		t.Errorf("expected %d log messages, got %d", len(expectedRemoteLogMessages), index)
	}
}

func TestProcessorStoresMetricsForNewNonBridgeDevice(t *testing.T) {
	mqtt := &mocks.MockMqttClient{}
	ws := &mocks.NopWsServer{}

	// setup device
	allDevices := createMockDialAndLightDevices("x01111111", "0x02222222")

	dialDevice := allDevices[0]
	lightDevice := allDevices[1]
	deviceBridgeList := utils_test.CreateBridgeInfoList([]*devices.Device{dialDevice})

	// register hub
	store, cleanup, err := utils_test.CreateFileStore()
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	appCfg := store.AppConfig()
	controllers.RegisterHubController(ws, store, mqtt)

	mqtt.Publish("bridge/devices", deviceBridgeList)
	time.Sleep(100 * time.Millisecond) // give it time to configure bridgeInfo
	//  SETUP END

	// note:
	// update dial button device - This should NOT be stored as metrics
	payload := map[string]any{"action": "button_2_hold"}
	mqtt.Publish(dialDevice.FriendlyName, payload)
	time.Sleep(500 * time.Millisecond)

	// note:
	// publish new device that is not in Bridge - eg via HTTP . it will be registered here and generate new Id
	payload = map[string]any{"brightness": 10.0, "color_temp": 100}
	mqtt.Publish(lightDevice.FriendlyName, payload)

	time.Sleep(500 * time.Millisecond)
	d, err := store.FindDeviceByFriendlyName("Attic light")
	if err != nil {
		t.Fatalf("device not found. err %v ", err)
	}

	// enable metrics for light device.
	cfg, err := appCfg.GetDeviceConfig(d.Id)
	if err != nil {
		t.Fatalf("device not found. err %v ", err)
	}
	cfg.MetricsEnabled = true
	cfg.RateLimit = utils.IntervalFromMilliseconds(10)
	appCfg.SetDeviceConfigOverrides(cfg)
	time.Sleep(500 * time.Millisecond)

	// note:
	// publish new device again- This SHOULD be stored as metrics NOW
	payload = map[string]any{"brightness": 20.0, "color_temp": 110.0}
	mqtt.Publish(d.FriendlyName, payload)
	time.Sleep(500 * time.Millisecond)

	from := time.Now().Add(-time.Minute * 2).UTC()
	to := time.Now().UTC()
	lightMetrics, err := store.ViewMetrics(d, from, to)
	if err != nil {
		t.Fatalf("ViewMetrics failed. err %v ", err)
	}

	// assert exposes
	if len(lightMetrics.Exposes) != 2 {
		t.Fatalf("size mismatch want %v got %v", 2, len(lightMetrics.Exposes))
	}

	for idx, gotExpose := range lightMetrics.Exposes {
		if gotExpose.GetType() == "numeric" {
			numericExpose := metrics.ToNumericExposeResults(gotExpose)

			// first index expected to be brightness
			if idx == 0 {
				if numericExpose.Name != "brightness" {
					t.Fatalf("name mismatch want brightness got %v", numericExpose.Name)
				}
				// assert birghtness values
				if len(numericExpose.Data) != 1 {
					t.Fatalf("size mismatch want %v got %v", 1, len(numericExpose.Data))
				}

				if numericExpose.Data[0].Y != 20.0 {
					t.Fatalf("name mismatch want brightness value 20.0 got %v", numericExpose.Data[0].Y)
				}

				//second index expected to be color_temp
			} else if idx == 1 {
				if numericExpose.Name != "color_temp" {
					t.Fatalf("name mismatch want color_temp got %v", numericExpose.Name)
				}

				if len(numericExpose.Data) != 1 {
					t.Fatalf("size mismatch want %v got %v", 1, len(numericExpose.Data))
				}
				if numericExpose.Data[0].Y != 110.0 {
					t.Fatalf("name mismatch want color_temp value 110.0 got %v", numericExpose.Data[0].Y)
				}
			} else {
				t.Errorf("invalid expose type %v: ", gotExpose.GetType())
			}
		}
	}

}

func TestHubSaveDeviceConfigOverrides(t *testing.T) {

	mqtt := &mocks.MockMqttClient{}
	ws := &mocks.NopWsServer{}

	// setup device
	device1Expose1 := utils_test.CreateEnumEntity("action", utils_test.CreateDialActionEnums())
	device1Expose2 := utils_test.CreateNumericEntity("action_time", 0)
	dialDevice := utils_test.CreateDeviceWithExposes("x01111111", "Dial button", []*devices.Entity{device1Expose1, device1Expose2})

	device2Expose1 := utils_test.CreateEntity("brightness", "numeric", nil)
	device2Expose2 := utils_test.CreateEnumEntity("color_temp", utils_test.CreateColorTempPresets())
	lightDevice := utils_test.CreateDeviceWithExposes("x02222222", "Attic light", []*devices.Entity{device2Expose1, device2Expose2})

	// setup bridgeInfo List
	devices := []*devices.Device{dialDevice, lightDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices) // NEED TO FIX, currently i make all devices features which is not right!!!!

	// register hub
	store, cleanup, err := utils_test.CreateFileStore()
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	appCfg := store.AppConfig()
	controllers.RegisterHubController(ws, store, mqtt)

	mqtt.Publish("bridge/devices", deviceBridgeList)
	time.Sleep(100 * time.Millisecond) // give it time to configure bridgeInfo
	//  SETUP END

	configs := []*settings.DeviceConfig{}
	for id, device := range devices {
		cfg, err := appCfg.GetDeviceConfig(device.Id)
		if err != nil {
			t.Fatalf("device not found. err %v ", err)
		}
		// update values and store for assertions
		rt := (id + 1) * 2
		cfg.RateLimit = utils.IntervalFromMilliseconds(rt)
		cfg.Disabled = true
		cfg.MetricsEnabled = true

		// set debounce overrides
		cfg.DebounceOverrides = map[string]*utils.TimeInterval{}
		for _, expose := range device.Exposes {
			cfg.DebounceOverrides[expose.Name] = utils.IntervalFromMinutes(id + 1)
		}

		appCfg.SetDeviceConfigOverrides(cfg)
		configs = append(configs, cfg)
	}

	// assert new stored values
	for id, device := range devices {
		cfg, err := appCfg.GetDeviceConfig(device.Id)
		if err != nil {
			t.Fatalf("device not found. err %v ", err)
		}
		if configs[id].Disabled != cfg.Disabled {
			t.Fatalf("disabled mismatch want %v got %v", configs[id].Disabled, cfg.Disabled)
		}
		if configs[id].RateLimit.Unit != cfg.RateLimit.Unit {
			t.Fatalf("rateLimit.Unit mismatch want %v got %v", configs[id].RateLimit.Unit, cfg.RateLimit.Unit)
		}
		if configs[id].RateLimit.Value != cfg.RateLimit.Value {
			t.Fatalf("rateLimit.Value mismatch want %v got %v", configs[id].RateLimit.Value, cfg.RateLimit.Value)
		}
		if configs[id].MetricsEnabled != cfg.MetricsEnabled {
			t.Fatalf("metricsEnabled mismatch want %v got %v", configs[id].MetricsEnabled, cfg.MetricsEnabled)
		}
		if len(configs[id].DebounceOverrides) != len(cfg.DebounceOverrides) {
			t.Fatalf("debounceOverrides mismatch want %v got %v", len(configs[id].DebounceOverrides), len(cfg.DebounceOverrides))
		}
		for name, debounce := range configs[id].DebounceOverrides {
			if debounce.Unit != cfg.DebounceOverrides[name].Unit {
				t.Fatalf("debounceOverrides.Unit mismatch want %v got %v", debounce.Unit, cfg.DebounceOverrides[name].Unit)
			}
			if debounce.Value != cfg.DebounceOverrides[name].Value {
				t.Fatalf("debounceOverrides.Value mismatch want %v got %v", debounce.Value, cfg.DebounceOverrides[name].Value)
			}
		}
	}
}

func TestHubDeletesDeviceConfigOverride(t *testing.T) {

	mqtt := &mocks.MockMqttClient{}
	ws := &mocks.NopWsServer{}

	// setup device
	device1Expose1 := utils_test.CreateEnumEntity("action", utils_test.CreateDialActionEnums())
	device1Expose2 := utils_test.CreateNumericEntity("action_time", 0)
	dialDevice := utils_test.CreateDeviceWithExposes("x01111111", "Dial button", []*devices.Entity{device1Expose1, device1Expose2})

	device2Expose1 := utils_test.CreateEntity("brightness", "numeric", nil)
	device2Expose2 := utils_test.CreateEnumEntity("color_temp", utils_test.CreateColorTempPresets())
	lightDevice := utils_test.CreateDeviceWithExposes("x02222222", "Attic light", []*devices.Entity{device2Expose1, device2Expose2})

	// setup bridgeInfo List
	devices := []*devices.Device{dialDevice, lightDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices) // NEED TO FIX, currently i make all devices features which is not right!!!!

	// register hub
	store, cleanup, err := utils_test.CreateFileStore()
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	appCache := store.AppConfig()
	controllers.RegisterHubController(ws, store, mqtt)

	mqtt.Publish("bridge/devices", deviceBridgeList)
	time.Sleep(100 * time.Millisecond) // give it time to configure bridgeInfo
	//  SETUP END

	for id, device := range devices {
		cfg, err := appCache.GetDeviceConfig(device.Id)
		if err != nil {
			t.Fatalf("device not found. err %v ", err)
		}
		// update values and store for assertions
		rt := (id + 1) * 2
		cfg.RateLimit = utils.IntervalFromMilliseconds(rt)
		cfg.Disabled = true
		cfg.MetricsEnabled = true
		cfg.DebounceOverrides = map[string]*utils.TimeInterval{}
		eIdx := 0
		for _, expose := range device.Exposes {
			eIdx++
			cfg.DebounceOverrides[expose.Name] = utils.IntervalFromMinutes(eIdx)
		}
		appCache.SetDeviceConfigOverrides(cfg)
		time.Sleep(100 * time.Millisecond)

	}

	// assert config override exists
	cfg, err := appCache.GetDeviceConfig(dialDevice.Id)
	if err != nil {
		t.Fatalf("device not found. err %v ", err)
	}
	time.Sleep(100 * time.Millisecond)

	// expected device config values to match with expected overrides values
	expectedDebounceUnit := "minutes"
	expectedMilliseconds := 2
	expectedDisabled := true
	expectedMetricsEnabled := true
	if cfg.Disabled != expectedDisabled && cfg.MetricsEnabled != expectedMetricsEnabled && cfg.RateLimit.Value != expectedMilliseconds {
		t.Fatalf("invalid config override. want expectedMilliseconds %v got %v", expectedMilliseconds, cfg.RateLimit.Value)
		t.Fatalf("invalid config override. want expectedDisabled %v got %v", expectedDisabled, cfg.Disabled)
		t.Fatalf("invalid config override. want expectedMetricsEnabled %v got %v", expectedMetricsEnabled, cfg.MetricsEnabled)
	}

	eIdx := 0
	for name := range dialDevice.Exposes {
		eIdx++
		debounce := cfg.DebounceOverrides[name]
		if debounce.Unit != expectedDebounceUnit {
			t.Fatalf("debounceOverrides.Unit mismatch want %v got %v", expectedDebounceUnit, debounce.Unit)
		}
		expectedValue := eIdx
		if debounce.Value != expectedValue {
			t.Fatalf("debounceOverrides.Value mismatch want %v got %v", expectedValue, debounce.Value)
		}
	}

	err = appCache.DeleteDeviceConfigOverrides(dialDevice.Id)
	if err != nil {
		t.Fatalf("device not found. err %v ", err)
	}

	cfg, err = appCache.GetDeviceConfig(dialDevice.Id)
	if err != nil {
		t.Fatalf("device not found. err %v ", err)
	}

	// load appconfig to gt device defaults
	app, err := appCache.LoadAppConfig()
	if err != nil {
		t.Fatalf("app not found. err %v ", err)
	}
	// assert device config has default values
	defaults := app.Hub.Devices.Defaults
	if cfg.Disabled != defaults.Disabled && cfg.MetricsEnabled != defaults.MetricsEnabled && cfg.RateLimit.Value != defaults.RateLimit.Value {
		t.Fatalf("invalid config override. want expectedMilliseconds %v got %v", defaults.RateLimit.Value, cfg.RateLimit.Value)
		t.Fatalf("invalid config override. want expectedDisabled %v got %v", defaults.Disabled, cfg.Disabled)
		t.Fatalf("invalid config override. want expectedMetricsEnabled %v got %v", defaults.MetricsEnabled, cfg.MetricsEnabled)
	}

	if cfg.DebounceOverrides != nil {
		t.Fatalf("debounceOverrides should be nil")
	}

	if !reflect.DeepEqual(cfg.DefaultDebounceByCategory, defaults.DefaultDebounceByCategory) {
		t.Fatalf("invalid config override. want expectedDefaultDebounceByCategory %v got %v", defaults.DefaultDebounceByCategory, cfg.DefaultDebounceByCategory)
	}
}

func TestHubSaveDeviceConfigDefaults(t *testing.T) {

	mqtt := &mocks.MockMqttClient{}
	ws := &mocks.NopWsServer{}

	// setup device
	device1Expose1 := utils_test.CreateEnumEntity("action", utils_test.CreateDialActionEnums())
	device1Expose2 := utils_test.CreateNumericEntity("action_time", 0)
	dialDevice := utils_test.CreateDeviceWithExposes("x01111111", "Dial button", []*devices.Entity{device1Expose1, device1Expose2})

	device2Expose1 := utils_test.CreateEntity("brightness", "numeric", nil)
	device2Expose2 := utils_test.CreateEnumEntity("color_temp", utils_test.CreateColorTempPresets())
	lightDevice := utils_test.CreateDeviceWithExposes("x02222222", "Attic light", []*devices.Entity{device2Expose1, device2Expose2})

	// setup bridgeInfo List
	devices := []*devices.Device{dialDevice, lightDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices) // NEED TO FIX, currently i make all devices features which is not right!!!!

	// register hub
	store, cleanup, err := utils_test.CreateFileStore()
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	controllers.RegisterHubController(ws, store, mqtt)

	mqtt.Publish("bridge/devices", deviceBridgeList)
	time.Sleep(100 * time.Millisecond) // give it time to configure bridgeInfo
	//  SETUP END

	appCache := store.AppConfig()
	app, err := appCache.LoadAppConfig()
	if err != nil {
		t.Fatalf("app not found. err %v ", err)
	}
	// assert that device config has expcted default values
	expectedDefaults := settings.DefaultDeviceConfig()
	deviceDefaults := app.Hub.Devices.Defaults

	if expectedDefaults.Disabled != deviceDefaults.Disabled && expectedDefaults.MetricsEnabled != deviceDefaults.MetricsEnabled && expectedDefaults.RateLimit.Value != deviceDefaults.RateLimit.Value {
		t.Fatalf("invalid config override. want expectedMilliseconds %v got %v", expectedDefaults.RateLimit.Value, deviceDefaults.RateLimit.Value)
		t.Fatalf("invalid config override. want expectedDisabled %v got %v", expectedDefaults.Disabled, deviceDefaults.Disabled)
		t.Fatalf("invalid config override. want expectedMetricsEnabled %v got %v", expectedDefaults.MetricsEnabled, deviceDefaults.MetricsEnabled)
	}

	if !reflect.DeepEqual(expectedDefaults.DebounceOverrides, deviceDefaults.DebounceOverrides) {
		t.Fatalf("invalid config override. want expectedDebounceOverrides %v got %v", expectedDefaults.DebounceOverrides, deviceDefaults.DebounceOverrides)
	}
	if !reflect.DeepEqual(expectedDefaults.DefaultDebounceByCategory, deviceDefaults.DefaultDebounceByCategory) {
		t.Fatalf("invalid config override. want expectedDefaultDebounceByCategory %v got %v", expectedDefaults.DefaultDebounceByCategory, deviceDefaults.DefaultDebounceByCategory)
	}

	// update device config defaults
	expectedNewdDeviceDeufalts := &settings.DeviceConfig{
		Disabled:       true,
		MetricsEnabled: true,
		RateLimit:      utils.IntervalFromMilliseconds(500),
		DefaultDebounceByCategory: map[bridge.ExposeCategory]*utils.TimeInterval{
			bridge.ConfigCategory:      utils.IntervalFromMilliseconds(1000),
			bridge.MeasurementCategory: utils.IntervalFromMinutes(6),
		},
	}

	appCache.SetDeviceConfigDefaults(expectedNewdDeviceDeufalts)

	// assert that device config has expcted updated default values
	app, err = appCache.LoadAppConfig()
	if err != nil {
		t.Fatalf("app not found. err %v ", err)
	}
	deviceDefaults = app.Hub.Devices.Defaults
	if expectedNewdDeviceDeufalts.Disabled != deviceDefaults.Disabled && expectedNewdDeviceDeufalts.MetricsEnabled != deviceDefaults.MetricsEnabled && expectedNewdDeviceDeufalts.RateLimit.Value != deviceDefaults.RateLimit.Value {
		t.Fatalf("invalid config override. want expectedMilliseconds %v got %v", expectedNewdDeviceDeufalts.RateLimit.Value, deviceDefaults.RateLimit.Value)
		t.Fatalf("invalid config override. want expectedDisabled %v got %v", expectedNewdDeviceDeufalts.Disabled, deviceDefaults.Disabled)
		t.Fatalf("invalid config override. want expectedMetricsEnabled %v got %v", expectedNewdDeviceDeufalts.MetricsEnabled, deviceDefaults.MetricsEnabled)
	}

	if !reflect.DeepEqual(expectedNewdDeviceDeufalts.DebounceOverrides, deviceDefaults.DebounceOverrides) {
		t.Fatalf("invalid config override. want expectedDebounceOverrides %v got %v", expectedNewdDeviceDeufalts.DebounceOverrides, deviceDefaults.DebounceOverrides)
	}

	if !reflect.DeepEqual(expectedNewdDeviceDeufalts.DefaultDebounceByCategory, deviceDefaults.DefaultDebounceByCategory) {
		t.Fatalf("invalid config override. want expectedDefaultDebounceByCategory %v got %v", expectedNewdDeviceDeufalts.DefaultDebounceByCategory, deviceDefaults.DefaultDebounceByCategory)
	}
}

func TestProcessorStoresMetricsForExistingDevice(t *testing.T) {

	wg := &sync.WaitGroup{}
	mqtt := &mocks.MockMqttClient{}
	ws := &mocks.NopWsServer{}

	// setup device
	device1Expose1 := utils_test.CreateEnumEntity("action", utils_test.CreateDialActionEnums())
	device1Expose2 := utils_test.CreateNumericEntity("action_time", 0)
	dialDevice := utils_test.CreateDeviceWithExposes("x01111111", "Dial button", []*devices.Entity{device1Expose1, device1Expose2})

	device2Expose1 := utils_test.CreateEntity("brightness", "numeric", nil)
	device2Expose2 := utils_test.CreateEnumEntity("color_temp", utils_test.CreateColorTempPresets())
	lightDevice := utils_test.CreateDeviceWithExposes("x02222222", "Attic light", []*devices.Entity{device2Expose1, device2Expose2})

	// setup bridgeInfo List
	devices := []*devices.Device{dialDevice, lightDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices) // NEED TO FIX, currently i make all devices features which is not right!!!!

	// register hub
	store, cleanup, err := utils_test.CreateFileStore()
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	appCfg := store.AppConfig()
	controllers.RegisterHubController(ws, store, mqtt)

	mqtt.Publish("bridge/devices", deviceBridgeList)
	time.Sleep(100 * time.Millisecond) // give it time to configure bridgeInfo
	//  SETUP END

	cfg, err := appCfg.GetDeviceConfig("x01111111")
	if err != nil {
		t.Fatalf("device not found. err %v ", err)
	}

	// enable metrics for dial device
	cfg.MetricsEnabled = true
	cfg.RateLimit = utils.IntervalFromMilliseconds(10)
	appCfg.SetDeviceConfigOverrides(cfg)

	// publish light device
	payload := map[string]any{"brightness": 10.0, "color_temp": 100}
	mqtt.Publish(lightDevice.FriendlyName, payload)

	// // publish dial button device
	payload = map[string]any{"action": "button_2_hold"}
	mqtt.Publish(dialDevice.FriendlyName, payload)

	time.Sleep(50 * time.Millisecond)

	numTriggers := 5

	wg.Add(numTriggers)
	for i := 0; i < numTriggers; i++ {

		action_time := float64(10 + (i * 2))
		payload = map[string]any{"action": "dial_rotate_left_slow", "action_direction": "left", "action_time": action_time, "action_type": "step"}
		//payload = map[string]any{"action_time": action_time}
		mqtt.Publish(dialDevice.FriendlyName, payload)

		time.Sleep(100 * time.Millisecond)

		wg.Done()
	}

	wg.Wait()

	from := time.Now().Add(-time.Minute)
	to := time.Now()
	dialMetrics, err := store.ViewMetrics(dialDevice, from, to)
	if err != nil {
		t.Fatalf("ViewMetrics failed. err %v ", err)
	}

	for idx, gotExpose := range dialMetrics.Exposes {
		if gotExpose.GetType() == "numeric" {
			numericExpose := metrics.ToNumericExposeResults(gotExpose)
			if numericExpose.Name != "action_time" {
				t.Fatalf("name mismatch want action_time got %v", numericExpose.Name)
			}
			if len(numericExpose.Data) != numTriggers {
				t.Fatalf("size mismatch want %v got %v", numTriggers, len(numericExpose.Data))
			}

			for i, v := range numericExpose.Data {
				action_time := float32(10 + (i * 2))

				if v.Y != action_time {
					t.Fatalf("value mismatch want %v got %v", action_time, v.X)
				}
			}

		} else if gotExpose.GetType() == "enum" {

			timeRangeExport := metrics.ToTimeRangeExposeResults(gotExpose)
			if timeRangeExport.Name != "action" {
				t.Fatalf("name mismatch want action got %v", timeRangeExport.Name)
			}
			if idx == 0 {
				if timeRangeExport.Data[0].X != "button_2_hold" {
					t.Fatalf("size mismatch want %v got %v", "button_2_hold", timeRangeExport.Data[0].X)
				}
				if timeRangeExport.Data[1].X != "dial_rotate_left_slow" {
					t.Fatalf("size mismatch want %v got %v", "dial_rotate_left_slow", timeRangeExport.Data[1].X)
				}

			}
			continue
		} else {
			t.Errorf("invalid expose type %v: ", gotExpose.GetType())
		}
	}

	// assert second expose results
	_, err = store.ViewMetrics(lightDevice, from, to)
	if err == nil {
		t.Fatalf("found light device metrics. It should not be stored")
	}
}

func TestImportDashboardGroupsMessage(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	// we dont need tasks here just using it as it using valid settings repo
	bridgeInfoFile := filepath.Join("../../../docs", "device_bridge.json")
	data, err := os.ReadFile(bridgeInfoFile)
	if err != nil {
		t.Fatal("Error reading file:", err)
		return
	}
	bridgeInfoes, err := devices.LoadBridgeDevices(data)
	if err != nil {
		t.Fatal("Error parsing bridge info data:", err)
		return
	}

	store, cleanup, err := utils_test.CreateFileStore()
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	cfg := store.AppConfig()
	mqtt := &mocks.MockMqttClient{}

	eventHub := mocks.NewMockEventHub()

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(bridgeInfoes)

	broadcastHandler := func(eventName string, data interface{}) error {

		if eventName != ws.ImportDashboardGroups {
			t.Fatalf("invalid event name want %v got %v", ws.ImportDashboardGroups, eventName)
			return fmt.Errorf("invalid event name %v", eventName)
		}

		req := map[string]*settings.DashboardGroup{}
		bytes, _ := json.Marshal(data)
		err := json.Unmarshal(bytes, &req)
		if err != nil {
			return fmt.Errorf("OnImportDashboardGroups failed. Invalid payload type : %v ", err.Error())
		}

		// validate request
		for _, group := range req {
			for _, devGroup := range group.DeviceGroup {
				_, err := registrar.LookupById(devGroup.DeviceId)

				if err != nil {
					t.Fatalf("OnImportDashboardGroups failed. Invalid device id %v : %v ", devGroup.DeviceId, err.Error())
					return fmt.Errorf("OnImportDashboardGroups failed. Invalid expose id : %v ", err.Error())
				}
			}

		}
		err = cfg.ImportDashboardGroups(req)
		wg.Done()

		return err
	}

	eventHub.SetMockBroadcastEvent(broadcastHandler)
	controllers.RegisterHubController(eventHub, store, mqtt)

	// create new expose group
	newGroup := settings.NewDashboardGroup("living room group")
	newGroup.AddDeviceExpose("0x00158d0005a23c38", "brightness")
	newGroup.AddDeviceExpose("0x001788010d7d9d3f", "action")
	newGroup.AddDeviceExpose("0xa4c13894070052fc", "presence")
	newGroup.AddDeviceExpose("0xa4c13894070052fc", "illuminance")

	newGroup2 := settings.NewDashboardGroup("attic room group")
	newGroup2.AddDeviceExpose("0x00124b0029207763", "temperature")
	newGroup2.AddDeviceExpose("0xa4c1389b273366c3", "alarm")
	newGroup2.AddDeviceExpose("0xa4c1381b6fd53fc4", "energy")

	wantDashboardGroups := map[string]*settings.DashboardGroup{
		"living room group": newGroup,
		"attic room group":  newGroup2,
	}
	eventHub.Broadcast(ws.ImportDashboardGroups, wantDashboardGroups)
	wg.Wait()

	c, _ := cfg.LoadAppConfig()

	if len(c.Hub.DashboardGroups) != 2 {
		t.Fatalf("want %v got %v", 2, len(c.Hub.DashboardGroups))
	}

	utils_test.CompareDashboardGroups(t, c.Hub.DashboardGroups, wantDashboardGroups)
}

func TestSaveDashboardGroupIsValidated(t *testing.T) {

	wg := sync.WaitGroup{}
	wg.Add(1)
	// we dont need tasks here just using it as it using valid settings repo
	bridgeInfoFile := filepath.Join("../../../docs", "device_bridge.json")
	data, err := os.ReadFile(bridgeInfoFile)
	if err != nil {
		t.Fatal("Error reading file:", err)
		return
	}
	bridgeInfoes, err := devices.LoadBridgeDevices(data)
	if err != nil {
		t.Fatal("Error parsing bridge info data:", err)
		return
	}

	store, cleanup, err := utils_test.CreateFileStore()
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	cfg := store.AppConfig()
	mqtt := &mocks.MockMqttClient{}

	eventHub := mocks.NewMockEventHub()

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(bridgeInfoes)

	// for now replicate the hub-controller logic handling this until we can mock the event hub
	broadcastHandler := func(eventName string, data interface{}) error {

		if eventName != ws.SaveDashboardGroup {
			t.Fatalf("invalid event name want %v got %v", ws.SaveDashboardGroup, eventName)
			return fmt.Errorf("invalid event name %v", eventName)
		}

		req := &settings.DashboardGroup{}
		bytes, _ := json.Marshal(data)
		err := json.Unmarshal(bytes, &req)
		if err != nil {
			return fmt.Errorf("OnSaveDashboardGroup failed. Invalid payload type : %v ", err.Error())
		}

		// validate request
		for id := range req.DeviceGroup {
			_, err := registrar.LookupById(id)

			if err != nil {
				t.Fatalf("OnSaveDashboardGroup failed. Invalid device id : %v ", err.Error())
				return fmt.Errorf("OnSaveEOnSaveDashboardGroupxposeGroup failed. Invalid expose id : %v ", err.Error())
			}
		}

		err = cfg.SaveDashboardGroup(req)
		if err != nil {
			t.Fatalf("OnSaveDashboardGroup failed. %v ", err.Error())
			return fmt.Errorf("OnSaveDashboardGroup failed. %v ", err.Error())
		}

		wg.Done()
		return nil
	}
	eventHub.SetMockBroadcastEvent(broadcastHandler)
	controllers.RegisterHubController(eventHub, store, mqtt)

	// create new expose group
	newGroup := settings.NewDashboardGroup("living room group")
	newGroup.AddDeviceExpose("0x00158d0005a23c38", "brightness")
	newGroup.AddDeviceExpose("0x001788010d7d9d3f", "action")
	newGroup.AddDeviceExpose("0xa4c13894070052fc", "presence")
	newGroup.AddDeviceExpose("0xa4c13894070052fc", "illuminance")

	eventHub.Broadcast(ws.SaveDashboardGroup, newGroup)
	wg.Wait()

	c, _ := cfg.LoadAppConfig()

	if len(c.Hub.DashboardGroups) != 1 {
		t.Fatalf("want %v got %v", 1, len(c.Hub.DashboardGroups))
	}

	for _, group := range c.Hub.DashboardGroups {
		if group.Name != newGroup.Name {
			t.Fatalf("want %v got %v", newGroup.Name, group.Name)
		}
		for id, expose := range group.DeviceGroup {

			if newGroup.DeviceGroup[id].DeviceId != expose.DeviceId {
				t.Fatalf("want %v got %v", expose.DeviceId, newGroup.DeviceGroup[id].DeviceId)
			}
		}
	}
}

func TestRenameDashboardGroup(t *testing.T) {

	wg := sync.WaitGroup{}
	wg.Add(1)
	// we dont need tasks here just using it as it using valid settings repo
	bridgeInfoFile := filepath.Join("../../../docs", "device_bridge.json")
	data, err := os.ReadFile(bridgeInfoFile)
	if err != nil {
		t.Fatal("Error reading file:", err)
		return
	}
	bridgeInfoes, err := devices.LoadBridgeDevices(data)
	if err != nil {
		t.Fatal("Error parsing bridge info data:", err)
		return
	}

	store, cleanup, err := utils_test.CreateFileStore()
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	cfg := store.AppConfig()
	newGroup := settings.NewDashboardGroup("living room group")
	newGroup.AddDeviceExpose("0x00158d0005a23c38", "brightness")
	newGroup.AddDeviceExpose("0x001788010d7d9d3f", "action")
	newGroup.AddDeviceExpose("0xa4c13894070052fc", "presence")
	newGroup.AddDeviceExpose("0xa4c13894070052fc", "illuminance")

	cfg.SaveDashboardGroup(newGroup)
	mqtt := &mocks.MockMqttClient{}

	eventHub := mocks.NewMockEventHub()

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(bridgeInfoes)

	// for now replicate the hub-controller logic handling this until we can mock the event hub
	broadcastHandler := func(eventName string, data interface{}) error {

		if eventName != ws.RenameDashboardGroup {
			t.Fatalf("invalid event name want %v got %v", ws.RenameDashboardGroup, eventName)
			return fmt.Errorf("invalid event name %v", eventName)
		}

		req := &devices.DashboardGroupRenameRequest{}
		bytes, _ := json.Marshal(data)
		err := json.Unmarshal(bytes, &req)
		if err != nil {
			return fmt.Errorf("OnRenameDashboardGroup failed. Invalid payload type : %v ", err.Error())
		}

		dashgroup, err := cfg.RenameDashboardGroup(req.OldName, req.NewName)
		if err != nil {
			t.Fatalf("OnRenameDashboardGroup failed. %v ", err.Error())
			return fmt.Errorf("OnRenameDashboardGroup failed. %v ", err.Error())
		}

		if dashgroup.Name != req.NewName {
			t.Fatalf("OnRenameDashboardGroup failed. name mismatch want %v got %v ", req.NewName, dashgroup.Name)
		}

		for id, expose := range dashgroup.DeviceGroup {

			if newGroup.DeviceGroup[id].DeviceId != expose.DeviceId {
				t.Fatalf("want %v got %v", expose.DeviceId, newGroup.DeviceGroup[id].DeviceId)
			}
		}

		wg.Done()
		return nil
	}

	eventHub.SetMockBroadcastEvent(broadcastHandler)
	controllers.RegisterHubController(eventHub, store, mqtt)

	req := &devices.DashboardGroupRenameRequest{}
	req.OldName = "living room group"
	req.NewName = "new living room group"

	eventHub.Broadcast(ws.RenameDashboardGroup, req)
	wg.Wait()

	c, _ := cfg.LoadAppConfig()

	if len(c.Hub.DashboardGroups) != 1 {
		t.Fatalf("want %v got %v", 1, len(c.Hub.DashboardGroups))
	}

	for _, group := range c.Hub.DashboardGroups {
		if group.Name != "new living room group" {
			t.Fatalf("want %v got %v", "new living room group", group.Name)
		}

	}

}

func TestDeleteDashboardGroupRemovesGroup(t *testing.T) {

	wg := sync.WaitGroup{}
	wg.Add(1)
	// we dont need tasks here just using it as it using valid settings repo
	bridgeInfoFile := filepath.Join("../../../docs", "device_bridge.json")
	data, err := os.ReadFile(bridgeInfoFile)
	if err != nil {
		t.Fatal("Error reading file:", err)
		return
	}
	bridgeInfoes, err := devices.LoadBridgeDevices(data)
	if err != nil {
		t.Fatal("Error parsing bridge info data:", err)
		return
	}

	store, cleanup, err := utils_test.CreateFileStore()
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	cfg := store.AppConfig()
	newGroup := settings.NewDashboardGroup("living room group")
	newGroup.AddDeviceExpose("0x00158d0005a23c38", "brightness")
	newGroup.AddDeviceExpose("0x001788010d7d9d3f", "action")
	newGroup.AddDeviceExpose("0xa4c13894070052fc", "presence")
	newGroup.AddDeviceExpose("0xa4c13894070052fc", "illuminance")

	cfg.SaveDashboardGroup(newGroup)
	mqtt := &mocks.MockMqttClient{}

	eventHub := mocks.NewMockEventHub()

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(bridgeInfoes)

	// for now replicate the hub-controller logic handling this until we can mock the event hub
	broadcastHandler := func(eventName string, p interface{}) error {

		if eventName != ws.DeleteDashboardGroup {
			t.Fatalf("invalid event name want %v got %v", ws.SaveDashboardGroup, eventName)
			return fmt.Errorf("invalid event name %v", eventName)
		}

		bytes, _ := json.Marshal(p)
		payload := make(map[string]interface{})
		err := json.Unmarshal(bytes, &payload)

		if err != nil {
			return fmt.Errorf("OnDeleteExposeGroup failed. Invalid payload type : %v ", err.Error())
		}

		id, ok := payload["groupName"].(string)
		if !ok {
			return fmt.Errorf("OnDeleteExposeGroup failed. Invalid payload type missing group id")
		}
		err = cfg.DeleteDashboardGroup(id)
		if err != nil {
			return fmt.Errorf("OnDeleteExposeGroup failed. %v ", err.Error())
		}

		wg.Done()
		return nil
	}
	eventHub.SetMockBroadcastEvent(broadcastHandler)
	controllers.RegisterHubController(eventHub, store, mqtt)

	payload := map[string]string{
		"groupName": "living room group",
	}

	eventHub.Broadcast(ws.DeleteDashboardGroup, payload)
	wg.Wait()

	c, _ := cfg.LoadAppConfig()

	if len(c.Hub.DashboardGroups) != 0 {
		t.Fatalf("want %v got %v", 0, len(c.Hub.DashboardGroups))
	}
}

func TestProcessorAddsNewDevice(t *testing.T) {

	name := "device 1"
	store := utils_test.CreateStore()

	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}

	controllers.RegisterHubController(ws, store, mqtt)
	mqtt.Publish(name, []byte(device1BatterySource))

	time.Sleep(500 * time.Millisecond)

	id := utils.HashName(name)
	device, err := store.FindDeviceById(id)
	if err != nil {
		t.Fatalf("FindDeviceById failed. err %v ", err)
	}
	if device == nil {
		t.Fatalf("want %s got %s", name, "nil")
	}

	if device.FriendlyName != name {
		t.Fatalf("want %s got %s", name, device.Id)
	}
}

func TestProcessorUpdatesExistingDevice(t *testing.T) {

	store := utils_test.CreateStore()

	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	controllers.RegisterHubController(ws, store, mqtt)
	mqtt.Publish("device1", []byte(device1BatterySource))
	mqtt.Publish("device2", []byte(device2))
	mqtt.Publish("device2", []byte(device1BatterySource))

	time.Sleep(100 * time.Millisecond)
	name := "device2"
	id := utils.HashName(name)

	device, err := store.FindDeviceById(id)
	if err != nil {
		t.Fatalf("FindDeviceById failed. err %v ", err)
	}
	if device == nil {
		t.Fatalf("want %s got %s", "device", "nil")
	}

	if device.FriendlyName != name {
		t.Fatalf("want %s got %s", name, device.Id)
	}
}

func TestProcessorHandlesDeviceNoLastSeen(t *testing.T) {

	name := "device1"

	store := utils_test.CreateStore()

	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	controllers.RegisterHubController(ws, store, mqtt)
	mqtt.Publish(name, []byte(device3NoLastSeen))

	want := time.Now().Format(time.RFC3339)
	time.Sleep(100 * time.Millisecond)
	id := utils.HashName(name)

	device, err := store.FindDeviceById(id)
	if err != nil {
		t.Fatalf("FindDeviceById failed. err %v ", err)
	}

	if device.FriendlyName != name {
		t.Fatalf("want %s got %s", name, device.Id)
	}

	if device.LastSeen == "" {
		t.Fatalf("want %s got %s", "last_seen", "nil")
	}

	if device.LastSeen != want {
		t.Fatalf("want %s got %s", want, device.LastSeen)
	}

}

func TestProcessorHandlesBridgePermitJoinwithActiveStateTimer(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	callbackCounter := 0

	// we dont need tasks here just using it as it using valid settings repo
	tasks := []settings.Task{}
	store, cleanup, err := utils_test.CreateStoreWithTasks(tasks)
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	cfg := store.AppConfig()
	mqtt := &mocks.MockMqttClient{}

	eventHub := mocks.NewMockEventHub()
	broadcastHandler := func(eventName string, data interface{}) error {

		if eventName != ws.BridgePermitJoin {
			return nil
		}
		req := settings.BridgeConfig{}
		bytes, _ := json.Marshal(data)
		err := json.Unmarshal(bytes, &req)
		if err != nil {
			t.Fatalf("failed to unmarshal payload %v", err)
			return err
		}

		f := func(value bool) error {

			// we are expecting to hit it twice, once initally and another one from the timeout
			expectedValue := true
			if callbackCounter == 1 {
				expectedValue = false
			}
			if callbackCounter > 1 {
				t.Fatalf("callbackCounter should be less than 2 got %v", callbackCounter)
				return fmt.Errorf("failed")
			}

			if value != expectedValue {
				t.Fatalf("want %v got %v", expectedValue, value)
				return fmt.Errorf("failed")
			}

			err = cfg.SaveBridgePermitJoin(value)
			callbackCounter++
			wg.Done()

			return err
		}

		mqttReq := hub.NewBridgePermitJoinRequest(&req, f)
		eventHub.Context().Enqueue(mqttReq)

		response := controllers.NewBridgeResponse()
		response.Status = "ok"
		response.Transaction = mqttReq.TransactionId
		response.Data["value"] = req.PermitJoin
		jsonPayload, _ := json.Marshal(response)

		mqtt.Publish("bridge/response/permit_join", jsonPayload)

		// simulate bridge response for permit join

		return nil
	}
	eventHub.SetMockBroadcastEvent(broadcastHandler)

	controllers.RegisterHubController(eventHub, store, mqtt)

	req := settings.NewBridgeConfig()
	req.PermitJoin = true
	req.TimeExpireAt.Value = 1 // 1 second expiration

	eventHub.Broadcast(ws.BridgePermitJoin, req)

	time.Sleep(200 * time.Millisecond)

	// assert bridge permit join is set to true, from initial request callback
	bridgeConfig, err := cfg.LoadBridgeConfig()
	if err != nil {
		t.Fatalf("error loading bridge config %s", err.Error())
	}
	if bridgeConfig.PermitJoin != true {
		t.Fatalf("want %v got %v", true, bridgeConfig.PermitJoin)
	}

	wg.Wait()

	//  assert bridge permit join is set to false, from timeout callback
	bridgeConfig, err = cfg.LoadBridgeConfig()
	if err != nil {
		t.Fatalf("error loading bridge config %s", err.Error())
	}
	if bridgeConfig.PermitJoin != false {
		t.Fatalf("want %v got %v", false, bridgeConfig.PermitJoin)
	}
}

func TestProcessorHandlesBridgePermitJoinRejectRequestWhenActive(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	callbackCounter := 0

	// we dont need tasks here just using it as it using valid settings repo
	tasks := []settings.Task{}
	store, cleanup, err := utils_test.CreateStoreWithTasks(tasks)
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	cfg := store.AppConfig()
	mqtt := &mocks.MockMqttClient{}

	eventHub := mocks.NewMockEventHub()
	broadcastHandler := func(eventName string, data interface{}) error {

		if eventName != ws.BridgePermitJoin {
			if callbackCounter == 3 && eventName == ws.OperationFailed {
				// we are expecting to hit it twice. counter is 3 as we have send success event from first call
				// second we should get back a failure event
				// as the event cannot start since its active
				wg.Done()
				return nil
			}

			return nil
		}
		req := settings.BridgeConfig{}
		bytes, _ := json.Marshal(data)
		err := json.Unmarshal(bytes, &req)
		if err != nil {
			t.Fatalf("failed to unmarshal payload %v", err)
			return err
		}

		f := func(value bool) error {

			if callbackCounter > 1 {
				t.Fatalf("callbackCounter should be less than 2 got %v", callbackCounter)
				return fmt.Errorf("failed")
			}

			if value != true {
				t.Fatalf("callback error: want %v got %v", true, value)
				return fmt.Errorf("failed")
			}

			err = cfg.SaveBridgePermitJoin(value)
			callbackCounter++
			wg.Done()

			return err
		}

		mqttReq := hub.NewBridgePermitJoinRequest(&req, f)

		eventHub.Context().Enqueue(mqttReq)

		response := controllers.NewBridgeResponse()
		response.Status = "ok"
		response.Transaction = mqttReq.TransactionId
		response.Data["value"] = req.PermitJoin
		jsonPayload, _ := json.Marshal(response)

		mqtt.Publish("bridge/response/permit_join", jsonPayload)

		callbackCounter++
		return nil
	}
	eventHub.SetMockBroadcastEvent(broadcastHandler)

	controllers.RegisterHubController(eventHub, store, mqtt)

	req := settings.NewBridgeConfig()
	req.PermitJoin = true
	req.TimeExpireAt.Value = 120 // 2 min second so we can reject second request

	// send first request
	eventHub.Broadcast(ws.BridgePermitJoin, req)

	time.Sleep(500 * time.Millisecond)

	// assert bridge permit join is set to true, from initial request callback
	bridgeConfig, err := cfg.LoadBridgeConfig()
	if err != nil {
		t.Fatalf("error loading bridge config %s", err.Error())
	}
	if bridgeConfig.PermitJoin != true {
		t.Fatalf("error loading BridgeConfig PermitJoin: want %v got %v", true, bridgeConfig.PermitJoin)
	}

	wg.Wait()

	// send second request while first one is active
	eventHub.Broadcast(ws.BridgePermitJoin, req)
	wg.Add(1)

	//  assert bridge permit join is still set to true,
	bridgeConfig, err = cfg.LoadBridgeConfig()
	if err != nil {
		t.Fatalf("error loading bridge config %s", err.Error())
	}
	if bridgeConfig.PermitJoin != true {
		t.Fatalf("want %v got %v", true, bridgeConfig.PermitJoin)
	}

	wg.Wait()
}

func TestNewDeviceExposeValuesAreBroadcastedOnly(t *testing.T) {

	bridgeInfoFile := filepath.Join("../../../docs", "device_bridge.json")
	data, err := os.ReadFile(bridgeInfoFile)
	if err != nil {
		t.Fatal("Error reading file:", err)
		return
	}
	bridgeInfoes, err := devices.LoadBridgeDevices(data)
	if err != nil {
		t.Fatal("Error parsing bridge info data:", err)
		return
	}

	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(bridgeInfoes)

	appConfig := store.AppConfig()
	config, err := appConfig.GetDeviceConfig("0xa4c13894070052fc")
	if err != nil {
		t.Fatalf("error loading device config %s", err.Error())
	}

	config.DebounceOverrides["illuminance"] = utils.IntervalFromMilliseconds(500)
	appConfig.SetDeviceConfigOverrides(config)

	wg := &sync.WaitGroup{}

	presenceDeviceName := "Living room presence sensor"

	testCases := []struct {
		deviceName string
		key        string
		value      any
		broadcast  bool
	}{
		{deviceName: presenceDeviceName, key: "target_distance", value: 13.4, broadcast: true},
		{deviceName: presenceDeviceName, key: "target_distance", value: 12.1, broadcast: false}, // diagnostics debounced
		{deviceName: presenceDeviceName, key: "target_distance", value: 16.3, broadcast: false}, // diagnostics debounced
		{deviceName: presenceDeviceName, key: "target_distance", value: 20.3, broadcast: false}, // diagnostics debounced

		{deviceName: presenceDeviceName, key: "presence", value: true, broadcast: true},
		{deviceName: presenceDeviceName, key: "presence", value: true, broadcast: false},
		{deviceName: presenceDeviceName, key: "presence", value: false, broadcast: true},
		{deviceName: presenceDeviceName, key: "presence", value: false, broadcast: false},
		{deviceName: presenceDeviceName, key: "illuminance", value: 10, broadcast: true},
		{deviceName: presenceDeviceName, key: "illuminance", value: 143, broadcast: false}, // debounced
		{deviceName: presenceDeviceName, key: "illuminance", value: 142, broadcast: false}, // debounced
		{deviceName: presenceDeviceName, key: "illuminance", value: 129, broadcast: true},
		{deviceName: presenceDeviceName, key: "presence", value: false, broadcast: false},
		{deviceName: presenceDeviceName, key: "presence", value: true, broadcast: true},
		{deviceName: presenceDeviceName, key: "illuminance", value: 321, broadcast: true},
		{deviceName: presenceDeviceName, key: "radar_sensitivity", value: 5, broadcast: true},
		{deviceName: presenceDeviceName, key: "radar_sensitivity", value: 5, broadcast: false},
		{deviceName: presenceDeviceName, key: "illuminance", value: 22, broadcast: true},
		{deviceName: presenceDeviceName, key: "illuminance", value: 21, broadcast: false},
	}

	broadcastHandler := func(eventName string, data interface{}) error {

		wg.Done()
		return nil
	}

	mqtt := &mocks.MockMqttClient{}
	eventHub.SetMockBroadcastEvent(broadcastHandler)

	controllers.RegisterHubController(eventHub, store, mqtt)

	for _, testCase := range testCases {

		lastSeen := time.Now().Format(time.RFC3339)
		payload := map[string]interface{}{}
		payload[testCase.key] = testCase.value
		payload["last_seen"] = lastSeen

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload %v", err)
		}

		if testCase.broadcast {
			wg.Add(1)
		}

		mqtt.Publish(testCase.deviceName, []byte(payloadBytes))

		time.Sleep(200 * time.Millisecond)
		wg.Wait()

	}
}

func TestAvailabilityStatusIsUpdated(t *testing.T) {

	name := "device 1"
	store := utils_test.CreateStore()

	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	hub := controllers.RegisterHubController(ws, store, mqtt)
	hub.DeviceAvailabilityTimeoutOverride = 1

	mqtt.Publish(name, []byte(device1BatterySource))
	time.Sleep(100 * time.Millisecond)

	id := utils.HashName(name)
	device, err := store.FindDeviceById(id)
	if err != nil {
		t.Fatalf("FindDeviceById failed. err %v ", err)
	}

	if device.Availability != devices.OnlineAvailability {
		t.Fatalf("want online got offline")
	}

	time.Sleep(1100 * time.Millisecond)
	if device.Availability != devices.OfflineAvailability {
		t.Fatalf("want offline got online")
	}

	mqtt.Publish(name, []byte(device1BatterySource))
	time.Sleep(200 * time.Millisecond)

	id = utils.HashName(name)
	device1, _ := store.FindDeviceById(id)

	if device1.Availability != devices.OnlineAvailability {
		t.Fatalf("want online got offline")
	}
}

func TestAvailabilityIsDisposed(t *testing.T) {

	name := "device 1"
	store := utils_test.CreateStore()

	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	hub := controllers.RegisterHubController(ws, store, mqtt)
	hub.DeviceAvailabilityTimeoutOverride = 1

	mqtt.Publish(name, []byte(device1BatterySource))
	time.Sleep(100 * time.Millisecond)

	id := utils.HashName(name)
	device, err := store.FindDeviceById(id)
	if err != nil {
		t.Fatalf("FindDeviceById failed. err %v ", err)
	}

	if device.Availability != devices.OnlineAvailability {
		t.Fatalf("want online got offline")
	}

	time.Sleep(1500 * time.Millisecond)

	if device.Availability != devices.OfflineAvailability {
		t.Fatalf("want offline got online")
	}
}

func TestHub_DeviceConfigDefaults_DisableDevices(t *testing.T) {

	mqtt := &mocks.MockMqttClient{}
	ws := &mocks.NopWsServer{}

	// setup device
	device1Expose1 := utils_test.CreateEnumEntity("action", utils_test.CreateDialActionEnums())
	device1Expose2 := utils_test.CreateNumericEntity("action_time", 0)
	dialDevice := utils_test.CreateDeviceWithExposes("x01111111", "Dial button", []*devices.Entity{device1Expose1, device1Expose2})

	device2Expose1 := utils_test.CreateEntity("brightness", "numeric", nil)
	device2Expose2 := utils_test.CreateEnumEntity("color_temp", utils_test.CreateColorTempPresets())
	lightDevice := utils_test.CreateDeviceWithExposes("x02222222", "Attic light", []*devices.Entity{device2Expose1, device2Expose2})

	// setup bridgeInfo List
	devices := []*devices.Device{dialDevice, lightDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices) // NEED TO FIX, currently i make all devices features which is not right!!!!

	// register hub
	store, cleanup, err := utils_test.CreateFileStore()
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

	controllers.RegisterHubController(ws, store, mqtt)

	mqtt.Publish("bridge/devices", deviceBridgeList)
	time.Sleep(100 * time.Millisecond) // give it time to configure bridgeInfo
	//  SETUP END

	payload := map[string]any{"brightness": 10.0, "color_temp": 100}
	mqtt.Publish(lightDevice.FriendlyName, payload)
	time.Sleep(100 * time.Millisecond)

	// // publish dial button device
	payload = map[string]any{"action": "button_2_hold"}
	mqtt.Publish(dialDevice.FriendlyName, payload)
	time.Sleep(100 * time.Millisecond)
	d, _ := store.FindDeviceById("x01111111")
	if d.Exposes["action"].Data != "button_2_hold" {
		t.Errorf("expected dial device action to be button_2_hold, got %s", d.Exposes["action"].Data)
	}

	appCache := store.AppConfig()
	// create defaults and set devices disabled
	defaults := settings.DefaultDeviceConfig()
	defaults.Disabled = true
	appCache.SetDeviceConfigDefaults(defaults)
	time.Sleep(100 * time.Millisecond)

	payload = map[string]any{"action": "button_1_hold"}
	mqtt.Publish(dialDevice.FriendlyName, payload)
	time.Sleep(100 * time.Millisecond)

	d, _ = store.FindDeviceById("x01111111")
	if d.Exposes["action"].Data == "button_1_hold" {
		t.Errorf("expected dial device action to be disabled, got %s", d.Exposes["action"].Data)
	}
}

func createMockDialAndLightDevices(dialName string, lightName string) []*devices.Device {

	device1Expose1 := utils_test.CreateEnumEntity("action", utils_test.CreateDialActionEnums())
	device1Expose2 := utils_test.CreateNumericEntity("action_time", 0)
	device1Expose1.Category = bridge.MeasurementCategory
	device1Expose2.Category = bridge.MeasurementCategory

	dialDevice := utils_test.CreateDeviceWithExposes(dialName, "Dial button", []*devices.Entity{device1Expose1, device1Expose2})

	device2Expose1 := utils_test.CreateEntity("brightness", "numeric", nil)
	device2Expose2 := utils_test.CreateEnumEntity("color_temp", utils_test.CreateColorTempPresets())
	device2Expose1.Category = bridge.MeasurementCategory
	device2Expose2.Category = bridge.MeasurementCategory

	lightDevice := utils_test.CreateDeviceWithExposes(lightName, "Attic light", []*devices.Entity{device2Expose1, device2Expose2})

	return []*devices.Device{dialDevice, lightDevice}
}
