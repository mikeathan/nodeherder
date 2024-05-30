package utils_test

import (
	"fmt"
	"io/ioutil"
	"math"
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/models/settings"
	"node-herder/repository"
	"node-herder/store"
	"os"
	"reflect"
	"testing"
	"time"
)

func CreateStore() store.AppStore {
	repo := repository.NewMemoryDeviceRepo()
	metricsRepo := mocks.NopMetricsRepo{}
	settingsRepo := mocks.NopSettingsrepo{}
	store, _ := store.NewAppStore(repo, &metricsRepo, &settingsRepo)
	return store
}

func CreateStoreFromDeviceRepo(repo devices.Repository) store.AppStore {
	metricsRepo := mocks.NopMetricsRepo{}
	settingsRepo := mocks.NopSettingsrepo{}
	store, _ := store.NewAppStore(repo, &metricsRepo, &settingsRepo)
	return store
}

func CreateStoreFromRepos(repo devices.Repository, metrics metrics.Repository, settings settings.Repository) store.AppStore {
	store, _ := store.NewAppStore(repo, metrics, settings)
	return store
}

func ValidateDevice(t *testing.T, dev1 *devices.Device, dev2 *devices.Device) {

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

		if !equalityCheck(expose.Data, inputExpose.Data) {
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

func ValidateBridge(t *testing.T, dev1 *devices.BridgeInfo, dev2 *devices.BridgeInfo) {

	if dev1.IeeeAddress != dev2.IeeeAddress {
		t.Fatalf("device IeeeAddress mismatch")
	}
	if dev1.FriendlyName != dev2.FriendlyName {
		t.Fatalf("device FriendlyName mismatch")
	}
	if dev1.DateCode != dev2.DateCode {
		t.Fatalf("device DateCode mismatch")
	}

	if dev1.Manufacturer != dev2.Manufacturer {
		t.Fatalf("device Manufacturer mismatch")
	}

	if dev1.Type != dev2.Type {
		t.Fatalf("device Type mismatch")
	}
	if dev1.SoftwareBuildID != dev2.SoftwareBuildID {
		t.Fatalf("device SoftwareBuildID mismatch")
	}
	if dev1.PowerSource != dev2.PowerSource {
		t.Fatalf("device PowerSource mismatch")
	}
	if dev1.Definition.Description != dev2.Definition.Description {
		t.Fatalf("device Definition.Description mismatch")
	}
	if dev1.Definition.Model != dev2.Definition.Model {
		t.Fatalf("device Definition.Model mismatch")
	}
	if dev1.Definition.Vendor != dev2.Definition.Vendor {
		t.Fatalf("device Definition.Vendor mismatch")
	}
	if dev1.Definition.SupportsOta != dev2.Definition.SupportsOta {
		t.Fatalf("device Definition.SupportsOta mismatch")
	}

	if !reflect.DeepEqual(dev1.Definition.Options, dev2.Definition.Options) {
		t.Fatalf("device Definition.Options mismatch")
	}

	for eidx, expose := range dev1.Definition.Exposes {
		expose2 := dev2.Definition.Exposes[eidx]
		if expose.Name != expose2.Name {
			t.Fatalf("unexpected expose.Name value")
		}

		if expose.Description != expose2.Description {
			t.Fatalf("unexpected expose.Description value")
		}

		if expose.Property != expose2.Property {
			t.Fatalf("unexpected expose.Property value")
		}
		if expose.Type != expose2.Type {
			t.Fatalf("unexpected expose.Type value")
		}
		if expose.Unit != expose2.Unit {
			t.Fatalf("unexpected expose.Unit value")
		}
		if expose.Access != expose2.Access {
			t.Fatalf("unexpected expose.Access value")
		}
		if expose.ValueMax != expose2.ValueMax {
			t.Fatalf("unexpected expose.ValueMax value")
		}
		if expose.ValueMin != expose2.ValueMin {
			t.Fatalf("unexpected expose.ValueMin value")
		}
		if expose.ValueOn != expose2.ValueOn {
			t.Fatalf("unexpected expose.ValueOn value")
		}
		if expose.ValueOff != expose2.ValueOff {
			t.Fatalf("unexpected expose.ValueOff value")
		}

		if !reflect.DeepEqual(expose.Values, expose2.Values) {
			t.Fatalf("unexpected expose.Values value")
		}
	}
}

func equalityCheck(a interface{}, b interface{}) bool {

	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	switch va.Kind() {
	case reflect.Float32, reflect.Float64:

		fl1 := math.Float32bits(float32(va.Float()))
		fl2 := math.Float32bits(float32(vb.Float()))

		return fl1 == fl2
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return va.Int() == vb.Int()
	default:
		return a == b
	}
}

func CreateDevice(deviceId string, friendlyName string, property string, data any, min float64, max float64) *devices.Device {
	device1 := &devices.Device{}
	device1.Id = deviceId
	device1.FriendlyName = friendlyName
	device1.ConnectionType = "mqtt"
	device1.Description = fmt.Sprintf("Test device %s description", deviceId)
	device1.PowerSource = "mains"
	device1.Properties = map[string]any{}
	device1.Properties["last_seen"] = time.Now().Format(time.RFC3339)
	device1.Properties["link_quality"] = 45.0
	device1.Exposes = make(map[string]*devices.Entity)

	ent1 := &devices.Entity{}
	ent1.Description = fmt.Sprintf("%s readings", property)
	ent1.Name = property
	ent1.Unit = "test"
	ent1.Data = data

	device1.Exposes[property] = ent1
	device1.Exposes[property].Attributes = make(map[string]any)
	device1.Exposes[property].Attributes["min"] = min
	device1.Exposes[property].Attributes["max"] = max
	return device1
}

func CreateBridgeInfoList(deviceList []*devices.Device) []*devices.BridgeInfo {

	bridgeInfoList := []*devices.BridgeInfo{}
	for _, dev := range deviceList {
		bridge := &devices.BridgeInfo{}
		bridge.IeeeAddress = dev.Id
		bridge.Definition.Description = dev.Description
		bridge.FriendlyName = dev.FriendlyName
		bridge.Type = "EndDevice"
		bridge.PowerSource = dev.PowerSource
		bridge.Disabled = false
		bridge.InterviewCompleted = true

		for _, expose := range dev.Exposes {
			e := devices.BridgeExpose{}
			e.Name = expose.Name
			e.Property = expose.Type
			e.Type = expose.Type
			e.Unit = expose.Unit
			e.ValueMin = expose.Attributes["min"]
			e.ValueMax = expose.Attributes["max"]
			e.Description = expose.Description

			f := devices.BridgeInfoFeature{}
			f.Name = expose.Name
			f.Property = expose.Type
			f.Type = expose.Type
			f.Unit = expose.Unit
			f.ValueMin = expose.Attributes["min"]
			f.ValueMax = expose.Attributes["max"]
			f.Description = expose.Description
			e.Features = append(e.Features, f)

			bridge.Definition.Exposes = append(bridge.Definition.Exposes, e)
		}

		bridgeInfoList = append(bridgeInfoList, bridge)
	}

	return bridgeInfoList
}

func Tempfile() string {
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
