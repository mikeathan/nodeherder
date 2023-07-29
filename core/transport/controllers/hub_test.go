package controllers_test

import (
	"context"
	"encoding/json"
	"node-herder/mocks"
	repository "node-herder/repository/devices"
	"node-herder/transport/controllers"
	"node-herder/transport/ws"
	"testing"
	"time"
)

const device1BatterySource = `{"id":"device 1","conn":"mqtt","power_source":"battery","sensors":{"humidity":92.49999999999999,"temperature":19.000000000000004},"stats":{"availability":"online","last_seen":"2023-07-20T19:48:35+01:00","linkquality":47,"battery":98}}`
const device2 = `{"battery":98, "humidity":71.2,  "linkquality":36.1,"temperature":17.1,"voltage":2999}`
const device3NoLastSeen = `{"id":"device 1","conn":"mqtt","power_source":"battery","sensors":{"humidity":92.49999999999999,"temperature":19.000000000000004},"stats":{"availability":"online","linkquality":47,"battery":98}}`

func createMockPayload() map[string]interface{} {
	return map[string]interface{}{
		"battery":     98,
		"humidity":    71.2,
		"last_seen":   time.Now().Format(time.RFC3339),
		"linkquality": 36.1,
		"temperature": 17.1,
	}
}

func TestProcessorAddsNewDevice(t *testing.T) {

	id := "device 1"
	repo := repository.NewMemoryDeviceRepo()
	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	controllers.RegisterHubController(ws, mqtt, repo, context.Background())
	mqtt.PublishMessage(id, []byte(device1BatterySource))

	time.Sleep(100 * time.Millisecond)
	device, err := repo.FindDevice(id)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if device == nil {
		t.Fatalf("want %s got %s", id, "nil")
	}
	if device.Id != id {
		t.Fatalf("want %s got %s", id, device.Id)
	}
}

func TestProcessorUpdatesExistingDevice(t *testing.T) {

	repo := repository.NewMemoryDeviceRepo()
	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	controllers.RegisterHubController(ws, mqtt, repo, context.Background())
	mqtt.PublishMessage("device1", []byte(device1BatterySource))
	mqtt.PublishMessage("device2", []byte(device2))
	mqtt.PublishMessage("device2", []byte(device1BatterySource))

	time.Sleep(100 * time.Millisecond)
	id := "device2"
	device, err := repo.FindDevice(id)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if device == nil {
		t.Fatalf("want %s got %s", "device", "nil")
	}
	if device.Id != id {
		t.Fatalf("want %s got %s", id, device.Id)
	}
}

func TestProcessorHandlesDeviceNoLastSeen(t *testing.T) {

	id := "device1"

	repo := repository.NewMemoryDeviceRepo()
	ws := &mocks.NopWsServer{}
	mqtt := &mocks.MockMqttClient{}
	controllers.RegisterHubController(ws, mqtt, repo, context.Background())
	mqtt.PublishMessage(id, []byte(device3NoLastSeen))

	want := time.Now().Format(time.RFC3339)
	time.Sleep(100 * time.Millisecond)
	device, err := repo.FindDevice(id)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if device.Id != id {
		t.Fatalf("want %s got %s", id, device.Id)
	}

	if device.Stats["last_seen"] == nil {
		t.Fatalf("want %s got %s", "last_seen", "nil")
	}

	if device.Stats["last_seen"] != want {
		t.Fatalf("want %s got %s", want, device.Stats["last_seen"])
	}
}

func TestOnlyNewPayloadIsBroadcasted(t *testing.T) {
	id := "device1"
	var payload = createMockPayload()

	var messageBroadcasted = false
	broadcast := func(eventName string, data interface{}) error {
		messageBroadcasted = true
		return nil
	}
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

	ws := newMockBroadcastEventHub(broadcast)
	repo := repository.NewMemoryDeviceRepo()
	mqtt := &mocks.MockMqttClient{}
	controllers.RegisterHubController(ws, mqtt, repo, context.Background())
	for idx, testCase := range testCases {
		// reset
		messageBroadcasted = false
		// use test case for updating sensor values
		payload[testCase.key] = testCase.value
		data, err := json.Marshal(payload)
		if err != nil {
			panic(err)
		}
		mqtt.PublishMessage(id, []byte(data))

		time.Sleep(100 * time.Millisecond)
		if testCase.broadcast != messageBroadcasted {
			t.Fatalf("idx %d,key %s, value %v, broadcast want %v got %v", idx, testCase.key, testCase.value, testCase.broadcast, messageBroadcasted)
		}
	}
}

func newMockBroadcastEventHub(mockBroadcastEvent func(eventName string, data interface{}) error) ws.EventHub {

	return &mocks.MockEventHub{MockBroadcastEvent: mockBroadcastEvent}
}
