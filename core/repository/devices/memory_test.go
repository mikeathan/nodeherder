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
	name := "device 1"
	device, _ := devices.CreateNewDeviceV2(repo, name, "mqtt", createMockPayload(name, 50, 60.1, 23.5, 120.0))

	repo.StoreV2(name, device)
	res, err := repo.FindDeviceV2(name)

	if err != nil {
		t.Fatalf(err.Error())
	}

	if device != res {
		t.Fatalf("device result mismatch: got %v want %v", res, device)
	}

	if device.FriendlyName != name {
		t.Fatalf("device name mismatch")
	}
}

func validateDevice(t *testing.T, dev1 *devices.DeviceV2, dev2 *devices.DeviceV2) {

	if dev1.Id != dev2.Id {
		t.Fatalf("device Id mismatch")
	}
	if dev1.FriendlyName != dev2.FriendlyName {
		t.Fatalf("device FriendlyName mismatch")
	}
	if dev1.Description != dev2.Description {
		t.Fatalf("device Description mismatch")
	}
	if dev1.ConnectionType != dev2.ConnectionType {
		t.Fatalf("device ConnectionType mismatch")
	}
	if dev1.PowerSource != dev2.PowerSource {
		t.Fatalf("device PowerSource mismatch")
	}

	for eidx, expose := range dev1.Exposes {
		inputExpose := dev2.Exposes[eidx]

		if expose.Name != inputExpose.Name {
			t.Fatalf("unexpected expose.Name value")
		}
		if expose.Description != inputExpose.Description {
			t.Fatalf("unexpected expose.Description value")
		}
		if expose.Data != inputExpose.Data {
			t.Fatalf("unexpected expose.Data value")
		}
		if expose.Unit != inputExpose.Unit {
			t.Fatalf("unexpected expose.Unit value")
		}
		for pidx, property := range expose.Properties {
			inputproperty := inputExpose.Properties[pidx]
			if property != inputproperty {
				t.Fatalf("unexpected property value")
			}

		}
	}

}

func TestRepositoryCanAddMultipleDevices(t *testing.T) {

	repo := repository.NewMemoryDeviceRepo()
	dev1Name := "device 1"
	device1, _ := devices.CreateNewDeviceV2(repo, dev1Name, "mqtt", createMockPayload(dev1Name, 50, 60.1, 23.5, 120.0))

	dev2Name := "device 2"
	device2, _ := devices.CreateNewDeviceV2(repo, dev2Name, "mqtt", createMockPayload(dev2Name, 90, 34.7, 36.2, 56.0))

	repo.StoreV2(dev1Name, device1)
	repo.StoreV2(dev2Name, device2)
	devices := repo.ListAllDevicesV2()

	if len(devices) == 0 {
		t.Fatalf("empty device list")
	}
	if len(devices) > 2 {
		t.Fatalf("contains invalid devices")
	}
	validateDevice(t, devices[1], device1)
	validateDevice(t, devices[0], device2)

	if devices[1].Id != dev1Name {
		t.Fatalf("device 1 name mismatch")
	}
	if devices[0].Id != dev2Name {
		t.Fatalf("device 2 name mismatch")
	}
}

func TestRepositoryCanUpdateExistingDevice(t *testing.T) {

	repo := repository.NewMemoryDeviceRepo()
	dev1Name := "device 1"
	device1, _ := devices.CreateNewDeviceV2(repo, dev1Name, "mqtt", createMockPayload(dev1Name, 50, 60.1, 23.5, 120.0))

	device1b, _ := devices.CreateNewDeviceV2(repo, dev1Name, "mqtt", createMockPayload(dev1Name, 90, 34.7, 36.2, 56.0))
	repo.StoreV2(dev1Name, device1)
	repo.StoreV2(dev1Name, device1b)
	devices := repo.ListAllDevicesV2()

	if len(devices) == 0 {
		t.Fatalf("empty device list")
	}
	if len(devices) > 1 {
		t.Fatalf("contains invalid devices")
	}

	if devices[0] != device1b {
		t.Fatalf("device 1 payload mismatch")
	}

	if devices[0].Id != dev1Name {
		t.Fatalf("device name mismatch")
	}
}
