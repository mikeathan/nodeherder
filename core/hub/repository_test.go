package hub_test

import (
	"encoding/json"
	"node-herder/hub"
	"testing"
)

const data1 = `{"battery":100,"humidity":60.8,"last_seen":"2023-05-31T19:05:28+01:00","linkquality":50,"temperature":22.1,"voltage":3000}`
const data2 = `{"battery":100,"humidity":61.2,"last_seen":"2023-06-31T19:05:28+01:00","linkquality":34,"temperature":16.6,"voltage":2999}`

func TestRepositoryCanAddOneDevice(t *testing.T) {

	repo := hub.NewMemoryRepository()
	repo.StoreJson("TH1", []byte(data1))
	devices := repo.ListAllDevices()

	if len(devices) == 0 {
		t.Fatalf("empty device list")
	}
	if len(devices) > 1 {
		t.Fatalf("contains invalid mismatch")
	}

	gotPayload := toJson(devices[0].Payload)
	if gotPayload != data1 {
		t.Fatalf("device payload mismatch want %s got %s", data1, gotPayload)
	}
	if devices[0].Name != "TH1" {
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

	repo := hub.NewMemoryRepository()
	repo.StoreJson("TH1", []byte(data1))
	repo.StoreJson("TH2", []byte(data2))
	devices := repo.ListAllDevices()

	if len(devices) == 0 {
		t.Fatalf("empty device list")
	}
	if len(devices) > 2 {
		t.Fatalf("contains invalid devices")
	}
	gotPayload := toJson(devices[0].Payload)
	if gotPayload != data1 {
		t.Fatalf("device 1 payload mismatc ")
	}

	gotPayload2 := toJson(devices[1].Payload)
	if gotPayload2 != data2 {
		t.Fatalf("device 2 payload mismatch ")
	}

	if devices[0].Name != "TH1" {
		t.Fatalf("device 1 name mismatch ")
	}
	if devices[1].Name != "TH2" {
		t.Fatalf("device 2 name mismatch ")
	}
}

func TestRepositoryCanUpdateExistingDevice(t *testing.T) {

	repo := hub.NewMemoryRepository()
	repo.StoreJson("TH1", []byte(data1))
	repo.StoreJson("TH1", []byte(data2))
	devices := repo.ListAllDevices()

	if len(devices) == 0 {
		t.Fatalf("empty device list")
	}
	if len(devices) > 1 {
		t.Fatalf("contains invalid devices")
	}
	gotPayload := toJson(devices[0].Payload)
	if gotPayload != data2 {
		t.Fatalf("device payload mismatch")
	}

	if devices[0].Name != "TH1" {
		t.Fatalf("device name mismatc ")
	}
}
