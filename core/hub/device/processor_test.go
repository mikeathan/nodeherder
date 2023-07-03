package device_test

import (
	"encoding/json"
	"node-herder/hub/device"
	"testing"
)

const device1BatterySource = `{"battery":98, "humidity":71.2, "last_seen":"2023-07-31T19:05:28+01:00", "linkquality":36.1,"temperature":17.1,"voltage":2999}`
const device3NoLastSeen = `{"battery":98, "humidity":71.2,  "linkquality":36.1,"temperature":17.1,"voltage":2999}`

func TestProcessorDeviceBatterySource(t *testing.T) {

	repo := device.NewMemoryNodeRepository()
	id := "device1"
	powerSource := "battery"
	var payload interface{}
	err := json.Unmarshal([]byte(device1BatterySource), &payload)
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

	if device.PowerSource != powerSource {
		t.Fatalf("want %s got %s", powerSource, device.PowerSource)
	}
}

func TestProcessorDeviceNoLastSeen(t *testing.T) {

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

	// TODO: fix time format: we want
	// 2023-07-31T19:05:28+01:00
	// 2023-07-03 20:50:53.136466802 +0100 BST m=+0.000381630"

}
