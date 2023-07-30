package repository_test

import (
	"encoding/json"
	"node-herder/models/devices"
	repository "node-herder/repository/devices"
	"testing"
)

const device1Payload = `{"id":"device 1","conn":"mqtt","power_source":"battery","sensors":{"humidity":92.49999999999999,"temperature":19.000000000000004},"stats":{"availability":"online","last_seen":"2023-07-20T19:48:35+01:00","linkquality":47,"battery":98}}`
const device1Payload2 = `{"id":"device 1","conn":"mqtt","power_source":"battery","sensors":{"humidity":92.49999999999999,"temperature":19.000000000000004},"stats":{"availability":"online","last_seen":"2023-07-20T19:48:35+01:00","linkquality":47,"battery":98}}`
const device2Payload = `{"id":"device 2","conn":"http","power_source":"mains","sensors":{"humidity":45,"temperature":14},"stats":{"availability":"online","last_seen":"2023-07-12T10:10:35+01:00","linkquality":120}}`

func createDevice(payload string) *devices.Device {
	var device *devices.Device
	err := json.Unmarshal([]byte(payload), &device)
	if err != nil {
		panic(err.Error())
	}
	return device
}

func TestRepositoryCanAddOneDevice(t *testing.T) {

	repo := repository.NewMemoryDeviceRepo()
	repo.Store("device 1", createDevice(device1Payload))
	device, err := repo.FindDevice("device 1")

	if err != nil {
		t.Fatalf(err.Error())
	}

	connType := device.ConnectionType
	if connType != "mqtt" {
		t.Fatalf("device 1 ConectionType mismatch  got %s want %s", connType, "mqtt")
	}

	if device.Id != "device 1" {
		t.Fatalf("device name mismatch ")
	}
}

func toJson(payload interface{}) string {
	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(data)
}

func TestRepositoryCanAddMultipleDevices(t *testing.T) {

	repo := repository.NewMemoryDeviceRepo()
	repo.Store("device 1", createDevice(device1Payload))
	repo.Store("device 2", createDevice(device2Payload))
	devices := repo.ListAllDevices()

	if len(devices) == 0 {
		t.Fatalf("empty device list")
	}
	if len(devices) > 2 {
		t.Fatalf("contains invalid devices")
	}
	// gotPayload := toJson(devices[0].Payload)
	// if gotPayload != data1 {
	// 	t.Fatalf("device 1 payload mismatc ")
	// }

	// gotPayload2 := toJson(devices[1].Payload)
	// if gotPayload2 != data2 {
	// 	t.Fatalf("device 2 payload mismatch ")
	// }

	if devices[0].Id != "device 1" {
		t.Fatalf("device 1 name mismatch ")
	}
	if devices[1].Id != "device 2" {
		t.Fatalf("device 2 name mismatch ")
	}
}

func TestRepositoryCanUpdateExistingDevice(t *testing.T) {

	repo := repository.NewMemoryDeviceRepo()
	repo.Store("device 1", createDevice(device1Payload))
	repo.Store("device 1", createDevice(device1Payload2))
	devices := repo.ListAllDevices()

	if len(devices) == 0 {
		t.Fatalf("empty device list")
	}
	if len(devices) > 1 {
		t.Fatalf("contains invalid devices")
	}
	// gotPayload := toJson(devices[0].Payload)
	// if gotPayload != data2 {
	// 	t.Fatalf("device payload mismatch")
	// }

	if devices[0].Id != "device 1" {
		t.Fatalf("device name mismatc ")
	}
}
