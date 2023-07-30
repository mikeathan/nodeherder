package repository_test

import (
	"node-herder/models/devices"
	repository "node-herder/repository/devices"
	"testing"
	"time"
)

func createMockPayload(id string, battery int, humidity float32, temperature float32, linkquality float32) map[string]interface{} {

	return map[string]interface{}{
		"battery":     battery,
		"id":          id,
		"humidity":    humidity,
		"last_seen":   time.Now().Format(time.RFC3339),
		"linkquality": linkquality,
		"temperature": 17.1,
	}
}

func TestRepositoryCanAddOneDevice(t *testing.T) {

	repo := repository.NewMemoryDeviceRepo()
	id := "device 1"
	device := devices.CreateNewDevice(id, createMockPayload(id, 50, 60.1, 23.5, 120.0))

	repo.Store(id, device)
	res, err := repo.FindDevice(id)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if device != res {
		t.Fatalf("device result mismatch: got %v want %v", res, device)
	}

	if device.Id != id {
		t.Fatalf("device name mismatch")
	}
}

func TestRepositoryCanAddMultipleDevices(t *testing.T) {

	repo := repository.NewMemoryDeviceRepo()
	devId1 := "device 1"
	device1 := devices.CreateNewDevice(devId1, createMockPayload(devId1, 50, 60.1, 23.5, 120.0))

	devId2 := "device 2"
	device2 := devices.CreateNewDevice(devId2, createMockPayload(devId2, 90, 34.7, 36.2, 56.0))

	repo.Store(devId1, device1)
	repo.Store(devId2, device2)
	devices := repo.ListAllDevices()

	if len(devices) == 0 {
		t.Fatalf("empty device list")
	}
	if len(devices) > 2 {
		t.Fatalf("contains invalid devices")
	}
	if devices[0] != device1 {
		t.Fatalf("device 1 payload mismatc")
	}

	if devices[1] != device2 {
		t.Fatalf("device 2 payload mismatch")
	}

	if devices[0].Id != devId1 {
		t.Fatalf("device 1 name mismatch")
	}
	if devices[1].Id != devId2 {
		t.Fatalf("device 2 name mismatch")
	}
}

func TestRepositoryCanUpdateExistingDevice(t *testing.T) {

	repo := repository.NewMemoryDeviceRepo()
	devId1 := "device 1"
	device1 := devices.CreateNewDevice(devId1, createMockPayload(devId1, 50, 60.1, 23.5, 120.0))

	device1b := devices.CreateNewDevice(devId1, createMockPayload(devId1, 90, 34.7, 36.2, 56.0))
	repo.Store(devId1, device1)
	repo.Store(devId1, device1b)
	devices := repo.ListAllDevices()

	if len(devices) == 0 {
		t.Fatalf("empty device list")
	}
	if len(devices) > 1 {
		t.Fatalf("contains invalid devices")
	}

	if devices[0] != device1b {
		t.Fatalf("device 1 payload mismatch")
	}

	if devices[0].Id != devId1 {
		t.Fatalf("device name mismatch")
	}
}
