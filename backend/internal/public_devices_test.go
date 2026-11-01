package hub

import (
	"encoding/json"
	"testing"
	"time"

	"node-herder/internal/metrics/domain"
	"node-herder/models/devices"
	"node-herder/models/hub"
	"node-herder/store"
)

type publicProjectionStore struct {
	store.AppStore
	items   []*devices.Device
	queries int
}

func (s *publicProjectionStore) LoadHubState() (*hub.HubState, error) {
	return hub.NewHubState(nil, s.items), nil
}
func (s *publicProjectionStore) FindDeviceByIds(_ []string) ([]*devices.Device, error) {
	return s.items, nil
}
func (s *publicProjectionStore) FindDeviceById(id string) (*devices.Device, error) {
	for _, d := range s.items {
		if d.Id == id {
			return d, nil
		}
	}
	return nil, nil
}
func (s *publicProjectionStore) QueryDevice(_ string, _, _ time.Time, _ []domain.MetricFilter, _ map[string]domain.ExposeResult) (*domain.DeviceMetricsResult, error) {
	s.queries++
	return &domain.DeviceMetricsResult{}, nil
}

func TestPublicDevicesExcludesProviderDevices(t *testing.T) {
	legacy := devices.NewDevice("legacy")
	private := devices.NewDevice("private")
	private.Provider = "esphome"
	source := &publicProjectionStore{items: []*devices.Device{legacy, private}}
	projection := publicDevices{source: source}
	state, err := projection.LoadHubState()
	if err != nil || len(state.Devices) != 1 || state.Devices[0].Id != "legacy" {
		t.Fatalf("state: %+v %v", state, err)
	}
	found, err := projection.FindDeviceByIds([]string{"legacy", "private"})
	if err != nil || len(found) != 1 || found[0].Id != "legacy" {
		t.Fatalf("found: %+v %v", found, err)
	}
	if _, err := projection.QueryDevice("private", time.Time{}, time.Time{}, nil, nil); err == nil {
		t.Fatal("private metrics query accepted")
	}
	if source.queries != 0 {
		t.Fatal("private metrics reached storage")
	}
	if _, err := projection.QueryDevice("legacy", time.Time{}, time.Time{}, nil, nil); err != nil || source.queries != 1 {
		t.Fatal("legacy metrics unavailable")
	}
	data, err := json.Marshal(private)
	if err != nil {
		t.Fatal(err)
	}
	var restored devices.Device
	if err := json.Unmarshal(data, &restored); err != nil || restored.Provider != "esphome" {
		t.Fatalf("provider was not persisted: %v", err)
	}
	if len(source.items) != 2 {
		t.Fatal("public projection mutated source")
	}
}
