package repository_test

import (
	"math"
	"node-herder/models/devices"
	repository "node-herder/repository/devices"
	"reflect"
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
	device, _ := devices.CreateNewDevice("1", name, "mqtt", nil, createMockPayload(name, 50, 60.1, 23.5, 120.0))

	repo.Store(name, device)
	res, err := repo.FindDevice(name)

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

func validateDevice(t *testing.T, dev1 *devices.Device, dev2 *devices.Device) {

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

		if equalityCheck(expose.Data, inputExpose.Data) == false {
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

func compareStructs(s1, s2 interface{}) bool {
	v1 := reflect.ValueOf(s1).Elem()
	v2 := reflect.ValueOf(s2).Elem()

	for i := 0; i < v1.NumField(); i++ {
		field1 := v1.Field(i)
		field2 := v2.Field(i)
		if field1.Kind() != field2.Kind() || !reflect.DeepEqual(field1.Interface(), field2.Interface()) {
			return false
		}
	}
	return true
}

func validateBridge(t *testing.T, dev1 *devices.BridgeInfo, dev2 *devices.BridgeInfo) {
	if !compareStructs(dev1, dev2) {
		t.Fatalf("bridge mismatch")
	}
	// if dev1.IeeeAddress != dev2.IeeeAddress {
	// 	t.Fatalf("device Id mismatch")
	// }
	// if dev1.FriendlyName != dev2.FriendlyName {
	// 	t.Fatalf("device FriendlyName mismatch")
	// }
	// if dev1.DateCode != dev2.DateCode {
	// 	t.Fatalf("device Description mismatch")
	// }

	// if dev1.Manufacturer != dev2.Manufacturer {
	// 	t.Fatalf("device PowerSource mismatch")
	// }

	// if dev1.Type != dev2.Type {
	// 	t.Fatalf("device PowerSource mismatch")
	// }
	// if dev1.SoftwareBuildID != dev2.SoftwareBuildID {
	// 	t.Fatalf("device PowerSource mismatch")
	// }
	// if dev1.PowerSource != dev2.PowerSource {
	// 	t.Fatalf("device PowerSource mismatch")
	// }
	// if dev1.Definition.Description != dev2.Definition.Description {
	// 	t.Fatalf("device Definition.Description mismatch")
	// }
	// if dev1.Definition.Model != dev2.Definition.Model {
	// 	t.Fatalf("device Definition.Model mismatch")
	// }
	// if dev1.Definition.Vendor != dev2.Definition.Vendor {
	// 	t.Fatalf("device Definition.Vendor mismatch")
	// }
	// if dev1.Definition.SupportsOta != dev2.Definition.SupportsOta {
	// 	t.Fatalf("device Definition.SupportsOta mismatch")
	// }

	// for eidx, expose := range dev1.Definition.Exposes {
	// 	expose2:=dev2.Definition.Exposes[eidx]
	// 	if expose.Name != expose2.Name {
	// 		t.Fatalf("unexpected expose.Name value")
	// 	}

	// 	if expose.Name != expose2.Name {
	// 		t.Fatalf("unexpected expose.Name value")
	// 	}

	// }
	// for eidx, expose := range dev1.Exposes {
	// 	inputExpose := dev2.Exposes[eidx]

	// 	if expose.Name != inputExpose.Name {
	// 		t.Fatalf("unexpected expose.Name value")
	// 	}
	// 	if expose.Description != inputExpose.Description {
	// 		t.Fatalf("unexpected expose.Description value")
	// 	}

	// 	if equalityCheck(expose.Data, inputExpose.Data) == false {
	// 		t.Fatalf("unexpected expose.Data value")
	// 	}
	// 	if expose.Unit != inputExpose.Unit {
	// 		t.Fatalf("unexpected expose.Unit value")
	// 	}
	// 	for pidx, property := range expose.Properties {
	// 		inputproperty := inputExpose.Properties[pidx]
	// 		if property != inputproperty {
	// 			t.Fatalf("unexpected property value")
	// 		}

	// 	}
	// }
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
func TestRepositoryCanAddMultipleDevices(t *testing.T) {

	repo := repository.NewMemoryDeviceRepo()
	dev1Name := "device 1"
	device1, _ := devices.CreateNewDevice("1", dev1Name, "mqtt", nil, createMockPayload(dev1Name, 50, 60.1, 23.5, 120.0))

	dev2Name := "device 2"
	device2, _ := devices.CreateNewDevice("2", dev2Name, "mqtt", nil, createMockPayload(dev2Name, 90, 34.7, 36.2, 56.0))

	repo.Store(dev1Name, device1)
	repo.Store(dev2Name, device2)
	devices, _ := repo.AllDevices()

	if len(devices) == 0 {
		t.Fatalf("empty device list")
	}
	if len(devices) > 2 {
		t.Fatalf("contains invalid devices")
	}
	validateDevice(t, devices[0], device1)
	validateDevice(t, devices[1], device2)

	if devices[0].FriendlyName != dev1Name {
		t.Fatalf("device 1 name mismatch")
	}
	if devices[1].FriendlyName != dev2Name {
		t.Fatalf("device 2 name mismatch")
	}
}

func TestRepositoryCanUpdateExistingDevice(t *testing.T) {

	repo := repository.NewMemoryDeviceRepo()
	dev1Name := "device 1"
	device1, _ := devices.CreateNewDevice("1", dev1Name, "mqtt", nil, createMockPayload(dev1Name, 50, 60.1, 23.5, 120.0))

	device1b, _ := devices.CreateNewDevice("1", dev1Name, "mqtt", nil, createMockPayload(dev1Name, 90, 34.7, 36.2, 56.0))
	repo.Store(dev1Name, device1)
	repo.Store(dev1Name, device1b)
	devices, _ := repo.AllDevices()

	if len(devices) == 0 {
		t.Fatalf("empty device list")
	}
	if len(devices) > 1 {
		t.Fatalf("contains invalid devices")
	}
	validateDevice(t, devices[0], device1b)
}
