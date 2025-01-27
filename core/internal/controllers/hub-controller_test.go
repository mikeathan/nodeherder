package controllers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"node-herder/internal/automations"
	"node-herder/internal/controllers"
	"node-herder/internal/ws"
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/models/hub"
	"node-herder/models/logging"
	"node-herder/models/metrics"
	"node-herder/models/settings"
	"node-herder/store"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"sync"
	"testing"
	"time"
)

const device1BatterySource = `{"id":"device 1","conn":"mqtt","power_source":"battery","humidity":92.49999999999999,"temperature":19.000000000000004,"availability":"online","last_seen":"2023-07-20T19:48:35+01:00","linkquality":47,"battery":98}`
const device2 = `{"battery":98, "humidity":71.2,  "linkquality":36.1,"temperature":17.1,"voltage":2999}`
const device3NoLastSeen = `{"id":"device 1","conn":"mqtt","power_source":"battery","humidity":91.12,"temperature":19.000000000000004,"availability":"online","linkquality":47,"battery":67}`

func createMockPayload() map[string]interface{} {
	return map[string]interface{}{
		"battery":     98,
		"humidity":    71.2,
		"last_seen":   time.Now().Format(time.RFC3339),
		"linkquality": 36.1,
		"temperature": 17.1,
	}
}

func TestProcessorTriggerScheduledAutomation(t *testing.T) {

	mqtt := &mocks.MockMqttClient{}
	ws := &mocks.NopWsServer{}

	deviceAutomation := utils_test.CreateDoorContactWithAlarmTriggerAutomation("x01111111", "x02222222", mqtt)

	now := time.Now().UTC()
	start := now.Add(500 * time.Millisecond)
	end := now.Add(1500 * time.Millisecond)

	deviceAutomation.Schedules = utils_test.CreateTimeSchedules(start, end)
	automationStorage := mocks.NewMockAutomationStorage([]*automations.Device{deviceAutomation})
	// setup device
	alarmDevice := utils_test.CreateAlarmDevice("x02222222", "alarm device", false)
	doorSensorDevice := utils_test.CreateDoorSensorDevice("x01111111", "front door sensor", false)
	// setup bridgeInfo List
	devices := []*devices.Device{doorSensorDevice, alarmDevice}
	deviceBridgeList := utils_test.CreateBridgeInfoList(devices)

	// register hub		//todo

	store := utils_test.CreateStore()
	hub := controllers.RegisterHubController(ws, store, mqtt, context.Background())

	hub.WithAutomationStorage(automationStorage) // overide storage

	//  publish deviceBridgeList to configure hub with devices
	mqtt.Publish("bridge/devices", deviceBridgeList)

	time.Sleep(600 * time.Millisecond)

	numOfEvents := 2
	for i := 0; i < numOfEvents; i++ {

		payload := map[string]any{"contact": true}
		mqtt.Publish(doorSensorDevice.FriendlyName, payload)
		time.Sleep(500 * time.Millisecond)

		// alarm should be trigger only when schedule is due
		if i == 0 {
			alarm, _ := store.FindDeviceById("x02222222")
			if alarm.Exposes["alarm"].Data != true {
				t.Errorf("alarm should be ON when door sensor triggers")
			}

			// reset alarm
			payload = map[string]any{"alarm": false}
			mqtt.Publish(alarmDevice.FriendlyName, payload)
			time.Sleep(200 * time.Millisecond)

			if alarm.Exposes["alarm"].Data != false {
				t.Errorf("alarm should be off ")
			}

			// reset contact
			payload = map[string]any{"contact": false}
			mqtt.Publish(doorSensorDevice.FriendlyName, payload)
			time.Sleep(200 * time.Millisecond)

			contact, _ := store.FindDeviceById("x01111111")
			if contact.Exposes["contact"].Data != false {
				t.Errorf("contact should be off ")
			}

			time.Sleep(3 * time.Second)

		} else {

			time.Sleep(500 * time.Millisecond)

			//  alarm should not be triggered as schedule is not due.
			alarm, _ := store.FindDeviceById("x02222222")
			if alarm.Exposes["alarm"].Data != false {
				t.Errorf("alarm should be OFF - schedule end should disable automation")
			}
		}
	}
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
	deviceAutomation.Triggers = []*automations.Trigger{dialRotateSlowTrigger, btn1PressTrigger, btn2PressTrigger}
	automationStorage := mocks.NewMockAutomationStorage([]*automations.Device{deviceAutomation})

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
	hub := controllers.RegisterHubController(ws, store, mqtt, context.Background())
	hub.WithAutomationStorage(automationStorage) // overide storage
	//  publish deviceBridgeList to configure hub with devices
	mqtt.Publish("bridge/devices", deviceBridgeList)
	time.Sleep(500 * time.Millisecond) // give it time to configure bridgeInfo
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
	tasks := []store.Task{store.DefaultRemoteLoggerTask()}

	store, cleanup, err := utils_test.CreateStoreWithTasks(tasks)
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

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

	controllers.RegisterHubController(ws, store, mqtt, context.Background())

	// find a way to test the remote logger
	mqtt.Publish("bridge/devices", deviceBridgeList)
	time.Sleep(500 * time.Millisecond) // give it time to configure bridgeInfo

	expectedEnabledMessages := 5
	testCases := []bool{true, false, true, false, true, false, true, false, true, false}
	for _, enabled := range testCases {

		logger := settings.NewLoggerConfig(enabled)
		store.SaveLoggerConfig(logger)
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

		return nil
	}

	// register mock remote logger
	remoteLogEmitter := mocks.NewMockRemoteLoggerEmitter(handler)
	utils.RegisterRemoteLoggerHook(remoteLogEmitter)

	controllers.RegisterHubController(ws, store, mqtt, context.Background())

	// find a way to test the remote logger
	mqtt.Publish("bridge/devices", deviceBridgeList)
	time.Sleep(500 * time.Millisecond) // give it time to configure bridgeInfo

	utils.EnableRemoteLoggerHook(true)

	payload := map[string]any{"brightness": 10.0, "color_temp": 100}
	mqtt.Publish(lightDevice.FriendlyName, payload)

	time.Sleep(100 * time.Millisecond)

	// publish dial button device
	payload = map[string]any{"action": "button_2_hold"}
	mqtt.Publish(dialDevice.FriendlyName, payload)

	time.Sleep(100 * time.Millisecond)

	utils.EnableRemoteLoggerHook(false)

	if index != len(expectedRemoteLogMessages) {
		t.Errorf("expected %d log messages, got %d", len(expectedRemoteLogMessages), index)
	}
}

func TestProcessorTriggersAutomationsStoresMetricsForNewDeviceNotInBridge(t *testing.T) {
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

	controllers.RegisterHubController(ws, store, mqtt, context.Background())

	mqtt.Publish("bridge/devices", deviceBridgeList)
	time.Sleep(500 * time.Millisecond) // give it time to configure bridgeInfo
	//  SETUP END

	// note:
	// update dial button device - This should NOT be stored as metrics
	payload := map[string]any{"action": "button_2_hold"}
	mqtt.Publish(dialDevice.FriendlyName, payload)

	// note:
	// publish new device that is not in Bridge - eg via HTTP . it will be registered here and generate new Id
	payload = map[string]any{"brightness": 10.0, "color_temp": 100}
	mqtt.Publish(lightDevice.FriendlyName, payload)

	time.Sleep(500 * time.Millisecond)
	d, err := store.FindDeviceByFriendlyName("Attic light")
	if err != nil {
		t.Fatalf("device not found. err %v ", err)
	}

	// enable metrics for light device. use its new Id
	cfg := settings.NewDeviceConfig(d.Id)
	cfg.MetricsEnabled = true
	cfg.RateLimit = utils.IntervalFromMilliseconds(10)
	store.SaveDeviceConfig(cfg)

	// note:
	// publish new device again- This SHOULD be stored as metrics NOW
	payload = map[string]any{"brightness": 20.0, "color_temp": 110.0}
	mqtt.Publish(d.FriendlyName, payload)
	time.Sleep(5 * time.Second)

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

func TestHubCreatesNewDeviceConfigurationsForNewDevices(t *testing.T) {

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

	controllers.RegisterHubController(ws, store, mqtt, context.Background())

	mqtt.Publish("bridge/devices", deviceBridgeList)
	time.Sleep(500 * time.Millisecond) // give it time to configure bridgeInfo
	//  SETUP END

	configs := []*settings.DeviceConfig{}
	for id, device := range devices {
		cfg, err := store.FindDeviceConfig(device.Id)
		if err != nil {
			t.Fatalf("device not found. err %v ", err)
		}
		// update values and store for assertions
		rt := (id + 1) * 2
		cfg.RateLimit = utils.IntervalFromMilliseconds(rt)
		cfg.Disabled = true
		cfg.MetricsEnabled = true
		store.SaveDeviceConfig(cfg)
		configs = append(configs, cfg)
	}

	// assert new stored values
	for id, device := range devices {
		cfg, err := store.FindDeviceConfig(device.Id)
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
	}
}

func TestProcessorTriggersAutomationsStoresMetricsForExistingDevice(t *testing.T) {

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

	controllers.RegisterHubController(ws, store, mqtt, context.Background())

	mqtt.Publish("bridge/devices", deviceBridgeList)
	time.Sleep(500 * time.Millisecond) // give it time to configure bridgeInfo
	//  SETUP END

	cfg, err := store.FindDeviceConfig("x01111111")
	if err != nil {
		t.Fatalf("device not found. err %v ", err)
	}

	// enable metrics for dial device
	cfg.MetricsEnabled = true
	cfg.RateLimit = utils.IntervalFromMilliseconds(10)
	store.SaveDeviceConfig(cfg)

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

func TestProcessorAddsNewDevice(t *testing.T) {

	name := "device 1"
	store := utils_test.CreateStore()

	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}

	controllers.RegisterHubController(ws, store, mqtt, context.Background())
	mqtt.Publish(name, []byte(device1BatterySource))

	time.Sleep(500 * time.Millisecond)

	id := utils.HashName(name)
	device, err := store.FindDeviceById(id)
	if err != nil {
		t.Fatalf(err.Error())
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
	controllers.RegisterHubController(ws, store, mqtt, context.Background())
	mqtt.Publish("device1", []byte(device1BatterySource))
	mqtt.Publish("device2", []byte(device2))
	mqtt.Publish("device2", []byte(device1BatterySource))

	time.Sleep(100 * time.Millisecond)
	name := "device2"
	id := utils.HashName(name)

	device, err := store.FindDeviceById(id)
	if err != nil {
		t.Fatalf(err.Error())
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
	controllers.RegisterHubController(ws, store, mqtt, context.Background())
	mqtt.Publish(name, []byte(device3NoLastSeen))

	want := time.Now().Format(time.RFC3339)
	time.Sleep(100 * time.Millisecond)
	id := utils.HashName(name)

	device, err := store.FindDeviceById(id)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if device.FriendlyName != name {
		t.Fatalf("want %s got %s", name, device.Id)
	}

	if device.Properties["last_seen"] == nil {
		t.Fatalf("want %s got %s", "last_seen", "nil")
	}

	if device.Properties["last_seen"] != want {
		t.Fatalf("want %s got %s", want, device.Properties["last_seen"])
	}
}

func TestProcessorHandlesBridgePermitJoinwithActiveStateTimer(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	callbackCounter := 0

	// we dont need tasks here just using it as it using valid settings repo
	tasks := []store.Task{}
	store, cleanup, err := utils_test.CreateStoreWithTasks(tasks)
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

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

			err = store.SaveBridgePermitJoin(value)
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

	controllers.RegisterHubController(eventHub, store, mqtt, context.Background())

	req := settings.NewBridgeConfig()
	req.PermitJoin = true
	req.TimeExpireAt.Value = 1 // 1 second expiration

	eventHub.Broadcast(ws.BridgePermitJoin, req)

	time.Sleep(200 * time.Millisecond)

	// assert bridge permit join is set to true, from initial request callback
	bridgeConfig, err := store.LoadBridgeConfig()
	if err != nil {
		t.Fatalf("error loading bridge config %s", err.Error())
	}
	if bridgeConfig.PermitJoin != true {
		t.Fatalf("want %v got %v", true, bridgeConfig.PermitJoin)
	}

	wg.Wait()

	//  assert bridge permit join is set to false, from timeout callback
	bridgeConfig, err = store.LoadBridgeConfig()
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
	tasks := []store.Task{}
	store, cleanup, err := utils_test.CreateStoreWithTasks(tasks)
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()

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

			err = store.SaveBridgePermitJoin(value)
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

	controllers.RegisterHubController(eventHub, store, mqtt, context.Background())

	req := settings.NewBridgeConfig()
	req.PermitJoin = true
	req.TimeExpireAt.Value = 120 // 2 min second so we can reject second request

	// send first request
	eventHub.Broadcast(ws.BridgePermitJoin, req)

	time.Sleep(500 * time.Millisecond)

	// assert bridge permit join is set to true, from initial request callback
	bridgeConfig, err := store.LoadBridgeConfig()
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
	bridgeConfig, err = store.LoadBridgeConfig()
	if err != nil {
		t.Fatalf("error loading bridge config %s", err.Error())
	}
	if bridgeConfig.PermitJoin != true {
		t.Fatalf("want %v got %v", true, bridgeConfig.PermitJoin)
	}

	wg.Wait()

}

func TestNewDeviceValuesAreBroadcastedOnly(t *testing.T) {
	name := "device1"
	var payload = createMockPayload()

	testCases := []struct {
		key       string
		value     any
		broadcast bool
	}{
		{key: "temperature", value: 15.6, broadcast: true},
		{key: "temperature", value: 15.6, broadcast: false},
		{key: "temperature", value: 18.5, broadcast: true},
		{key: "humidity", value: 70.3, broadcast: true},
		{key: "humidity", value: 70.3, broadcast: false},
		{key: "linkquality", value: 120, broadcast: false},
		{key: "linkquality", value: 14, broadcast: false},
		{key: "battery", value: 70, broadcast: false},
		{key: "temperature", value: 18.5, broadcast: false},
		{key: "temperature", value: 21, broadcast: true},
		{key: "temperature", value: 21, broadcast: false},
		{key: "temperature", value: 21, broadcast: false},
	}

	var messageBroadcasted = false
	broadcast := func(eventName string, data interface{}) error {
		messageBroadcasted = true
		return nil
	}

	ws := newMockBroadcastEventHub(broadcast)
	store := utils_test.CreateStore()

	mqtt := &mocks.MockMqttClient{}

	controllers.RegisterHubController(ws, store, mqtt, context.Background())
	for idx, testCase := range testCases {
		// reset
		messageBroadcasted = false
		// use test case for updating sensor values
		payload[testCase.key] = testCase.value
		data, err := json.Marshal(payload)
		if err != nil {
			panic(err)
		}
		mqtt.Publish(name, []byte(data))

		time.Sleep(100 * time.Millisecond)
		if testCase.broadcast != messageBroadcasted {
			t.Fatalf("idx %d,key %s, value %v, broadcast want %v got %v", idx, testCase.key, testCase.value, testCase.broadcast, messageBroadcasted)
		}
	}
}

func TestDevicesBroadcastDeviceEvent(t *testing.T) {
	var payload = createMockPayload()

	testCases := []struct {
		key   string
		value any
	}{
		{key: "temperature", value: 15.6},
		{key: "temperature", value: 20.1},
		{key: "humidity", value: 61.2},
		{key: "humidity", value: 54.8},
		{key: "lux", value: 599.0},
		{key: "human_presence", value: true},
		{key: "buttonswitch1", value: 10},
		{key: "buttonswitch1", value: 11},
		{key: "lux", value: 90},
		{key: "buttonswitch2", value: true},
	}
	eventIdx := 0
	expectedEventNames := []string{
		"deviceAdded",
		"deviceUpdated",
		"deviceAdded",
		"deviceUpdated",
		"deviceAdded",
		"deviceAdded",
		"deviceAdded",
		"deviceUpdated",
		"deviceUpdated",
		"deviceAdded",
	}

	broadcast := func(eventName string, data interface{}) error {
		expectedEvent := expectedEventNames[eventIdx]
		fmt.Println(eventName, data)
		if eventName != expectedEvent {
			t.Fatalf("invalid broadcasted event:  want %s got %s", expectedEvent, eventName)
		}
		return nil
	}

	ws := newMockBroadcastEventHub(broadcast)
	store := utils_test.CreateStore()

	mqtt := &mocks.MockMqttClient{}

	controllers.RegisterHubController(ws, store, mqtt, context.Background())
	for _, testCase := range testCases {
		payload[testCase.key] = testCase.value
		data, err := json.Marshal(payload)
		if err != nil {
			panic(err)
		}
		mqtt.Publish(testCase.key, []byte(data))

		time.Sleep(100 * time.Millisecond)
		eventIdx++
	}
}

// // TODO: test if presence is converted to 0 and 1
// func TestDevicePackageData(t *testing.T) {

// 	p := &devices.UpdatePackage{Id: "x0111", LastSeen: time.Now().String(), Data: make(map[string]any), Properties: make(map[string]any)}

// 	p.Data["temperature"] = 15.6
// 	p.Data["humidity"] = 61.2
// 	p.Data["lux"] = 599.0
// 	p.Data["human_presence"] = "on"

// 	temp, err := json.Marshal(p.Data)
// 	if err == nil {
// 		fmt.Println(string(temp))
// 	} else {
// 		fmt.Println(err.Error())
// 	}
// }

func TestAvailabilityStatusIsUpdated(t *testing.T) {

	name := "device 1"
	store := utils_test.CreateStore()

	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	hub := controllers.RegisterHubController(ws, store, mqtt, context.Background())
	hub.DeviceAvailabilityTimeoutOverride = 1

	mqtt.Publish(name, []byte(device1BatterySource))
	time.Sleep(100 * time.Millisecond)

	id := utils.HashName(name)
	device, err := store.FindDeviceById(id)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if device.Properties["availability"] != "online" {
		t.Fatalf("want online got offline")
	}

	time.Sleep(1100 * time.Millisecond)
	if device.Properties["availability"] != "offline" {
		t.Fatalf("want offline got online")
	}

	mqtt.Publish(name, []byte(device1BatterySource))
	time.Sleep(200 * time.Millisecond)

	id = utils.HashName(name)
	device1, _ := store.FindDeviceById(id)

	if device1.Properties["availability"] != "online" {
		t.Fatalf("want online got offline")
	}
}

func TestAvailabilityIsDisposed(t *testing.T) {

	name := "device 1"
	store := utils_test.CreateStore()

	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	hub := controllers.RegisterHubController(ws, store, mqtt, context.Background())
	hub.DeviceAvailabilityTimeoutOverride = 1

	mqtt.Publish(name, []byte(device1BatterySource))
	time.Sleep(100 * time.Millisecond)

	id := utils.HashName(name)
	device, err := store.FindDeviceById(id)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if device.Properties["availability"] != "online" {
		t.Fatalf("want online got offline")
	}

	device.Dispose()
	time.Sleep(100 * time.Millisecond)

	if device.Properties["availability"] != "offline" {
		t.Fatalf("want offline got online")
	}
}

func createMockDialAndLightDevices(dialName string, lightName string) []*devices.Device {

	device1Expose1 := utils_test.CreateEnumEntity("action", utils_test.CreateDialActionEnums())
	device1Expose2 := utils_test.CreateNumericEntity("action_time", 0)
	dialDevice := utils_test.CreateDeviceWithExposes(dialName, "Dial button", []*devices.Entity{device1Expose1, device1Expose2})

	device2Expose1 := utils_test.CreateEntity("brightness", "numeric", nil)
	device2Expose2 := utils_test.CreateEnumEntity("color_temp", utils_test.CreateColorTempPresets())
	lightDevice := utils_test.CreateDeviceWithExposes(lightName, "Attic light", []*devices.Entity{device2Expose1, device2Expose2})

	return []*devices.Device{dialDevice, lightDevice}
}

func newMockBroadcastEventHub(mockBroadcastEvent func(eventName string, data interface{}) error) ws.EventHub {
	return &mocks.MockEventHub{MockBroadcastEvent: mockBroadcastEvent}
}
