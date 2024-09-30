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
	"node-herder/models/metrics"
	"node-herder/models/settings"
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

	wg.Add(numTriggers)
	for i := 0; i < numTriggers; i++ {

		prevValue, _ := light.Exposes["brightness"].Data.(float64)

		action_time := 10 + (i * 2)
		payload = map[string]any{"action": "dial_rotate_left_slow", "action_direction": "left", "action_time": action_time, "action_type": "step"}
		mqtt.Publish(dialDevice.FriendlyName, payload)

		time.Sleep(50 * time.Millisecond)

		// assert.  calculate expected value
		wantvalue := prevValue + (float64(action_time) * dialRotateSlowTrigger.Action.Data.(float64))
		gotValue, _ := light.Exposes["brightness"].Data.(float64)

		wantvalue = math.Min(wantvalue, max)

		if gotValue != wantvalue {
			t.Fatalf("brightness value mismatch. want %v got %v", wantvalue, gotValue)
		}

		wg.Done()
	}

	wg.Wait()
}

func TestHubTriggersRemoteLogger(t *testing.T) {

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

	// enable remote hook
	utils.EnableRemoteLoggerHook(true)

	TODO
	// find a way to test the remote logger
	mqtt.Publish("bridge/devices", deviceBridgeList)
	time.Sleep(500 * time.Millisecond) // give it time to configure bridgeInfo
}

func TestProcessorTriggersAutomationsStoresMetricsForNewDeviceNotInBridge(t *testing.T) {
	mqtt := &mocks.MockMqttClient{}
	ws := &mocks.NopWsServer{}

	// setup device
	device1Expose1 := utils_test.CreateEnumEntity("action", utils_test.CreateDialActionEnums())
	device1Expose2 := utils_test.CreateNumericEntity("action_time", 0)
	dialDevice := utils_test.CreateDeviceWithExposes("x01111111", "Dial button", []*devices.Entity{device1Expose1, device1Expose2})

	device2Expose1 := utils_test.CreateEntity("brightness", "numeric", nil)
	device2Expose2 := utils_test.CreateEnumEntity("color_temp", utils_test.CreateColorTempPresets())
	lightDevice := utils_test.CreateDeviceWithExposes("", "Attic light", []*devices.Entity{device2Expose1, device2Expose2})

	// setup bridgeInfo List
	devices := []*devices.Device{dialDevice}
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
	cfg.RateLimit = 10 // 10 ms
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
		cfg.RateLimit = (id + 1) * 2 // 10 ms
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
		if configs[id].RateLimit != cfg.RateLimit {
			t.Fatalf("rateLimit mismatch want %v got %v", configs[id].RateLimit, cfg.RateLimit)
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
	cfg.RateLimit = 10 // 10 ms
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

func newMockBroadcastEventHub(mockBroadcastEvent func(eventName string, data interface{}) error) ws.EventHub {
	return &mocks.MockEventHub{MockBroadcastEvent: mockBroadcastEvent}
}
