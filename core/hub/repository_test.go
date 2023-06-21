package hub_test

import (
	"node-herder/hub"
	"testing"
)

const data1 = "{'battery':100,'humidity':59.8,'last_seen':'2023-05-31T19:02:28+01:00','linkquality':51,'temperature':18.4,'voltage':3000}"

const data2 = "{'battery':95,'humidity':60.8,'last_seen':'2023-05-31T19:05:28+01:00','linkquality':50,'temperature':22.1,'voltage':3000}"

func TestRepositoryCanAddOneDevice(t *testing.T) {

	repo := hub.NewMemoryRepository()
	repo.Store("TH1", data1)

	devices := repo.ListAllDevices()

	if len(devices) == 0 {
		t.Errorf("empty device list")
	}
	if len(devices) > 1 {
		t.Errorf("contains invalid mismatch")
	}
	if devices[0].Payload != data1 {
		t.Errorf("device payload mismatch ")
	}
	if devices[0].Name != "TH1" {
		t.Errorf("device name mismatch ")
	}
}

func TestRepositoryCanAddMultipleDevices(t *testing.T) {

	repo := hub.NewMemoryRepository()
	repo.Store("TH1", data1)
	repo.Store("TH2", data2)
	devices := repo.ListAllDevices()

	if len(devices) == 0 {
		t.Errorf("empty device list")
	}
	if len(devices) > 2 {
		t.Errorf("contains invalid devices")
	}
	if devices[0].Payload != data1 {
		t.Errorf("device 1 payload mismatc ")
	}
	if devices[1].Payload != data2 {
		t.Errorf("device 2 payload mismatch ")
	}
	if devices[0].Name != "TH1" {
		t.Errorf("device 1 name mismatch ")
	}
	if devices[1].Name != "TH2" {
		t.Errorf("device 2 name mismatch ")
	}
}

func TestRepositoryCanUpdateExistingDevice(t *testing.T) {

	repo := hub.NewMemoryRepository()
	repo.Store("TH1", data1)
	repo.Store("TH1", data2)
	devices := repo.ListAllDevices()

	if len(devices) == 0 {
		t.Errorf("empty device list")
	}
	if len(devices) > 1 {
		t.Errorf("contains invalid devices")
	}
	if devices[0].Payload != data2 {
		t.Errorf("device payload mismatch")
	}

	if devices[0].Name != "TH1" {
		t.Errorf("device name mismatc ")
	}
}
