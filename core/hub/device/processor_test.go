package device_test

import (
	"encoding/json"
	"fmt"
	"node-herder/hub/device"
	"testing"
	"time"
)

const device1BatterySource = `{"battery":98, "humidity":71.2, "last_seen":"2023-07-31T19:05:28+01:00", "linkquality":36.1,"temperature":17.1,"voltage":2999}`
const device2 = `{"battery":98, "humidity":71.2,  "linkquality":36.1,"temperature":17.1,"voltage":2999}`
const device3NoLastSeen = `{"battery":98, "humidity":71.2,  "linkquality":36.1,"temperature":17.1,"voltage":2999}`

func TestProcessorAddsNewDevice(t *testing.T) {

	repo := device.NewMemoryNodeRepository()
	id := "device1"
	powerSource := "battery"
	var payload = createPayload(device1BatterySource)

	p := device.NewPayloadProcessor(repo)
	addDevice(p, id, payload)

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

	if device.PowerSource != powerSource {
		t.Fatalf("want %s got %s", powerSource, device.PowerSource)
	}
}

func createPayload(data string) interface{} {
	var payload interface{}
	err := json.Unmarshal([]byte(device1BatterySource), &payload)
	if err != nil {
		panic(err.Error())
	}
	return payload
}
func addDevice(p device.Processor, id string, payload interface{}) {
	err := p.Process(id, payload)
	if err != nil {
		panic(err.Error())
	}
}

func TestProcessorUpdatesExistingDevice(t *testing.T) {

	repo := device.NewMemoryNodeRepository()

	var payload1 = createPayload(device1BatterySource)
	var payload2 = createPayload(device2)

	p := device.NewPayloadProcessor(repo)
	addDevice(p, "device1", payload1)
	addDevice(p, "device2", payload2)
	addDevice(p, "device2", payload1)

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

	repo := device.NewMemoryNodeRepository()
	id := "device1"
	var payload interface{}
	err := json.Unmarshal([]byte(device3NoLastSeen), &payload)
	if err != nil {
		t.Fatalf(err.Error())
	}

	p := device.NewPayloadProcessor(repo)
	err = p.Process(id, payload)
	if err != nil {
		t.Fatalf(err.Error())
	}

	device, err := repo.FindDevice(id)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if device.Id != id {
		t.Fatalf("want %s got %s", id, device.Id)
	}

	if device.Device["last_seen"] == nil {
		t.Fatalf("want %s got %s", "last_seen", "nil")
	}
	want := time.Now().Format(time.RFC3339)
	if device.Device["last_seen"] != want {
		t.Fatalf("want %s got %s", want, device.Device["last_seen"])
	}
}

func TestProcessorLastSeen(t *testing.T) {
	fmt.Println(time.Now().Format(time.RFC3339))
}
