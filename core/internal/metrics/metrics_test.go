package metrics_test

import (
	"fmt"
	"io/ioutil"
	"node-herder/internal/metrics"
	"node-herder/models/devices"
	"os"
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {

	tempfile := tempfile()
	defer os.Remove(tempfile)

	repo, err := metrics.NewMetricsRepoFromFile(tempfile)
	if err != nil {
		t.Error("failed to initialise metrics repo", err.Error())
	}
	var devices map[string]*devices.Device = make(map[string]*devices.Device)
	for i := 0; i < 2; i++ {
		dev := createMockDevice(fmt.Sprintf("x000%v", i), fmt.Sprintf("device %v", i), time.Now().Add(-(time.Second * 10)), i+1*2)
		err = repo.Store(dev)
		if err != nil {
			t.Error("failed to store metrics ", err.Error())
		}
		dev = createMockDevice(fmt.Sprintf("x000%v", i), fmt.Sprintf("device %v", i), time.Now().Add(-(time.Minute * 20)), i+1*5)
		err = repo.Store(dev)
		if err != nil {
			t.Error("failed to store metrics ", err.Error())
		}

		dev = createMockDevice(fmt.Sprintf("x000%v", i), fmt.Sprintf("device %v", i), time.Now().Add(-(time.Second * 100)), i+1*15)
		err = repo.Store(dev)
		if err != nil {
			t.Error("failed to store metrics ", err.Error())
		}

		dev = createMockDevice(fmt.Sprintf("x000%v", i), fmt.Sprintf("device %v", i), time.Now().Add(-(time.Hour * 1)), i+1*20)
		err = repo.Store(dev)
		if err != nil {
			t.Error("failed to store metrics ", err.Error())
		}
		devices[dev.Id] = dev
	}

	id1 := fmt.Sprintf("x000%v", 0)
	dev1 := devices[id1]
	err = repo.ViewTimeRange(dev1, time.Now().Add(-(time.Minute * 30)), time.Now())
	if err != nil {
		t.Error("failed to query metrics: ", err.Error())
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
	device1.Properties["last_seen"] = timestamp.Format(time.RFC3339)
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

func tempfile() string {
	f, err := ioutil.TempFile("", "bolt-")
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}
	return f.Name()
}
