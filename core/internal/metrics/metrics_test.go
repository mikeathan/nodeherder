package metrics_test

import (
	"fmt"
	"node-herder/internal/metrics"
	"node-herder/models/devices"
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {

	repo, err := metrics.NewMetricsRepo()
	if err != nil {
		t.Errorf("failed to initialise metrics repo", err.Error())
	}
	for i := 0; i < 2; i++ {
		dev := createMockDevice(fmt.Sprintf("x000%v", i), fmt.Sprintf("device %v", i), time.Now().Add(-(time.Second * 2)), i*2)
		repo.Store(dev)
	}
}

func createMockDevice(id string, name string, timestamp time.Time, data any) *devices.Device {
	device1 := &devices.Device{}
	device1.Id = id
	device1.FriendlyName = name
	device1.ConnectionType = "mqtt"
	device1.Description = fmt.Sprintf("Test device %s description", id)
	device1.PowerSource = "mains"
	device1.Properties = map[string]any{}
	device1.Properties["last_seen"] = timestamp
	device1.Properties["link_quality"] = 45.0
	device1.Exposes = make(map[string]*devices.Entity)

	property := fmt.Sprintf("property_device_%v", id)

	ent1 := &devices.Entity{}
	ent1.Description = fmt.Sprintf("%s readings", property)
	ent1.Name = property
	ent1.Unit = "test"
	ent1.Data = data

	device1.Exposes[property] = ent1
	device1.Exposes[property].Attributes = make(map[string]any)
	device1.Exposes[property].Attributes["min"] = 0.0
	device1.Exposes[property].Attributes["max"] = 255.0
	return device1
}
