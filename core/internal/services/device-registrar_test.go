package services_test

import (
	"fmt"
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/repository"
	utils_test "node-herder/testing"
	"os"
	"path/filepath"
	"testing"
)

func TestRegisterBridge(t *testing.T) {

	bridgeInfoFile := filepath.Join("../../../docs", "device_bridge.json")
	data, err := os.ReadFile(bridgeInfoFile)
	if err != nil {
		t.Fatal("Error reading file:", err)
		return
	}
	bridgeInfoes, err := devices.LoadBridgeDevices(data)
	if err != nil {
		t.Fatal("Error parsing bridge info data:", err)
		return
	}
	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(bridgeInfoes, 30000)

	storeDevices, _ := store.AllDevices()

	// assert measurement devices
	measureExposes, err := devices.FindAllExposesByCategory(data, devices.MeasurementCategory)
	if err != nil {
		t.Fatal("Error loading MeasurementCategory devices:", err)
		return
	}

	if len(measureExposes) == 0 {
		t.Errorf("Error not found any measurement devices")
	}

	for deviceridgeId, deviceBridgeExposes := range measureExposes {
		for _, bridgeExpose := range deviceBridgeExposes {
			found := false
			for _, device := range storeDevices {
				for _, expose := range device.Exposes {
					if deviceridgeId == device.Id && expose.Name == bridgeExpose.Name {
						assetExpose(bridgeExpose, expose, devices.MeasurementCategory, t)
						assertDataType(bridgeExpose, expose, t)

						found = true
					}
				}
			}

			if !found {
				t.Errorf("Error measurement bridge expose %s not found in store", bridgeExpose.Name)
			}
		}
	}

	// assert diagnostic devices
	diagnosticExposes, err := devices.FindAllExposesByCategory(data, devices.DiagnosticCategory)
	if err != nil {
		t.Fatal("Error loading DiagnosticCategory devices:", err)
		return
	}

	if len(diagnosticExposes) == 0 {
		t.Errorf("Error not found any diagnostic devices")
	}

	for deviceBridgeId, deviceBridgeExposes := range diagnosticExposes {
		for _, bridgeExpose := range deviceBridgeExposes {
			found := false
			for _, device := range storeDevices {
				for _, expose := range device.Exposes {
					if deviceBridgeId == device.Id && expose.Name == bridgeExpose.Name {
						assetExpose(bridgeExpose, expose, devices.DiagnosticCategory, t)
						assertDataType(bridgeExpose, expose, t)


						found = true
					}
				}
			}

			if !found {
				t.Errorf("Error diagnostic bridge expose %s not found in store", bridgeExpose.Name)
			}
		}
	}

	// // assert config devices
	configExposes, err := devices.FindAllExposesByCategory(data, devices.ConfigCategory)
	if err != nil {
		t.Fatal("Error loading ConfigCategory devices:", err)
		return
	}

	if len(configExposes) == 0 {
		t.Errorf("Error not found any config devices")
	}

	for deviceBridgeId, deviceBridgeExposes := range configExposes {
		for _, bridgeExpose := range deviceBridgeExposes {
			found := false
			for _, device := range storeDevices {
				for _, expose := range device.Exposes {
					if deviceBridgeId == device.Id && expose.Name == bridgeExpose.Name {
						assetExpose(bridgeExpose, expose, devices.ConfigCategory, t)
						assertDataType(bridgeExpose, expose, t)


						found = true
					}
				}
			}

			if !found {
				t.Errorf("Error config bridge expose %s not found in store", bridgeExpose.Name)
			}
		}
	}
}

func assetExpose(bridgeExpose devices.BridgeExpose, expose *devices.Entity, category devices.ExposeCategory, t *testing.T) {
	if expose.Category != category {
		t.Errorf("Error %s device mismatch want: %s got: %s", category, bridgeExpose.Name, expose.Name)
	}

	if expose.AccessMode != bridgeExpose.Access {
		t.Errorf("Error %s device access mismatch want: %v got: %v", category, bridgeExpose.Access, expose.AccessMode)
	}

	if expose.Type != bridgeExpose.Type {
		t.Errorf("Error %s device type mismatch want: %v got: %v", category, bridgeExpose.Type, expose.Type)
	}

	if expose.Unit != bridgeExpose.Unit {
		t.Errorf("Error %s device unit mismatch want: %v got: %v", category, bridgeExpose.Unit, expose.Unit)
	}
	if expose.Description != bridgeExpose.Description {
		t.Errorf("Error %s device description mismatch want: %v got: %v", category, bridgeExpose.Description, expose.Description)
	}

}

func assertDataType(bridgeExpose devices.BridgeExpose, expose *devices.Entity, t *testing.T) {
	if expose.Type == devices.BinaryDataType {

		if expose.Values["on"] != bridgeExpose.ValueOn {
			t.Errorf("Error %s device value mismatch want: %v got: %v", bridgeExpose.Name, bridgeExpose.ValueOn, expose.Values["on"])
		}
		if expose.Values["off"] != bridgeExpose.ValueOff {
			t.Errorf("Error %s device value mismatch want: %v got: %v", bridgeExpose.Name, bridgeExpose.ValueOff, expose.Values["off"])
		}
		if bridgeExpose.ValueToggle != "" && expose.Values["toggle"] != bridgeExpose.ValueToggle {
			t.Errorf("Error %s device value mismatch want: %v got: %v", bridgeExpose.Name, bridgeExpose.ValueToggle, expose.Values["toggle"])
		}
	}
	if expose.Type == devices.EnumDataType {
		for id, item := range bridgeExpose.Values {
			if expose.Values[fmt.Sprintf("%d", id)] != item {
				t.Errorf("Error %s device value mismatch want: %v got: %v", bridgeExpose.Name, item, expose.Values[fmt.Sprintf("%d", id)])
			}
		}
	}

	if expose.Type == devices.NumericDataType {
		if bridgeExpose.ValueMin != nil && expose.Attributes["min"] != bridgeExpose.ValueMin {
			t.Errorf("Error %s device value mismatch want: %v got: %v", bridgeExpose.Name, bridgeExpose.ValueMin, expose.Attributes["min"])
		}
		if bridgeExpose.ValueMax != nil && expose.Attributes["max"] != bridgeExpose.ValueMax {
			t.Errorf("Error %s device value mismatch want: %v got: %v", bridgeExpose.Name, bridgeExpose.ValueMax, expose.Attributes["max"])
		}

		if len(bridgeExpose.Presets) != 0 {
			for _, preset := range bridgeExpose.Presets {
				if expose.Values[preset.Name] != preset.Value {
					t.Errorf("Error %s device value mismatch want: %v got: %v", bridgeExpose.Name, preset.Value, expose.Values[preset.Name])
				}
			}
		}
	}
}
