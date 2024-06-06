package controllers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"node-herder/internal/controllers"
	"node-herder/internal/ws"
	"node-herder/mocks"
	utils_test "node-herder/testing"
	"node-herder/utils"
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
 TODO: test automations triggers in the hub 
// to confirm the worker taks works correctly


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
