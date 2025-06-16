package services_test

import (
	"fmt"
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/repository"
	utils_test "node-herder/testing"
	"os"
	"path/filepath"
	"testing"
	"time"
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
	registrar.RegisterBridge(bridgeInfoes)

	storeDevices, _ := store.AllDevices()

	// assert measurement devices
	measureExposes, err := devices.FindAllExposesByCategory(data, bridge.MeasurementCategory)
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
						assetExpose(bridgeExpose, expose, bridge.MeasurementCategory, t)
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
	diagnosticExposes, err := devices.FindAllExposesByCategory(data, bridge.DiagnosticCategory)
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
						assetExpose(bridgeExpose, expose, bridge.DiagnosticCategory, t)
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
	configExposes, err := devices.FindAllExposesByCategory(data, bridge.ConfigCategory)
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
						assetExpose(bridgeExpose, expose, bridge.ConfigCategory, t)
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

func TestCreateNewDevice(t *testing.T) {

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
	registrar.RegisterBridge(bridgeInfoes)

	deviceName := "Living room light"

	lastSeen := time.Now().Format(time.RFC3339)
	payload := map[string]interface{}{}
	payload["brightness"] = 120.1
	payload["color_temp"] = 100.1
	payload["state"] = "on"
	payload["last_seen"] = lastSeen
	payload["battery"] = 100
	newDevice, err := registrar.CreateNewDevice(deviceName, "mqtt", payload)

	if err != nil {
		t.Errorf("Error creating new device: %s", err)
	}

	assertDevicePayload(newDevice, deviceName, payload, t)
}

func TestDefaultDebounceforDiagnosticExposes(t *testing.T) {

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

	store, cleanup, err := utils_test.CreateFileStore()
	if err != nil {
		t.Fatalf("CreateFileStore failed. err %v ", err)
	}
	defer cleanup()
	eventHub := &mocks.MockEventHub{}

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(bridgeInfoes)

	ds, err := store.AllDevices()
	if err != nil {
		t.Errorf("Error loading devices: %s", err)
	}

	if len(ds) == 0 {
		t.Errorf("Error not found any devices")
	}

	appConfig := store.AppConfig()

	for _, device := range ds {
		deviceConfig, err := appConfig.GetDeviceConfig(device.Id)
		if err != nil {
			t.Errorf("Error getting device config: %s", err)
		}
		for _, expose := range device.Exposes {
			if expose.Category == bridge.DiagnosticCategory {
				d, ok := deviceConfig.DebounceOverrides[expose.Name]
				if !ok {
					t.Errorf("Error diagnostic expose %s debounce is 0", expose.Name)
				}
				if d.Value != 300 {
					t.Errorf("Error diagnostic expose %s debounce is not 300. got %v", expose.Name, d.Value)
				}
				if d.Unit != "seconds" {
					t.Errorf("Error diagnostic expose %s debounce unit is not seconds", d.Unit)
				}
			}
		}
	}

}

func assertDeviceUpdatePackage(device *devices.Device, updatePackage *devices.UpdatePackage, t *testing.T) {

	if device.Id != updatePackage.Id {
		t.Errorf("Error device name mismatch want: %s got: %s", updatePackage.Id, device.Id)
	}
	if device.LastSeen != updatePackage.LastSeen {
		t.Errorf("Error device last seen mismatch want: %s got: %s", updatePackage.LastSeen, device.LastSeen)
	}

	for name, value := range updatePackage.Data {
		if expose, ok := device.Exposes[name]; ok {

			if expose.Data != value {
				t.Errorf("Error device expose value mismatch want: %v got: %v", expose.Data, value)
			}
		}
	}
}

func assertDevicePayload(newDevice *devices.Device, deviceName string, payload map[string]interface{}, t *testing.T) {
	if newDevice.FriendlyName != deviceName {
		t.Errorf("Error device name mismatch want: %s got: %s", deviceName, newDevice.FriendlyName)
	}

	if newDevice.LastSeen != payload["last_seen"] {
		t.Errorf("Error device last seen mismatch want: %s got: %s", payload["last_seen"], newDevice.LastSeen)
	}

	if newDevice.Availability != devices.OnlineAvailability {
		t.Errorf("Error device availability mismatch want: %v got: %v", devices.OnlineAvailability, newDevice.Availability)
	}

	if newDevice.ConnectionType != "mqtt" {
		t.Errorf("Error device connection type mismatch want: %s got: %s", "mqtt", newDevice.ConnectionType)
	}

	if newDevice.PowerSource != "battery" {
		t.Errorf("Error device power source mismatch want: %v got: %v", "battery", newDevice.PowerSource)
	}

	if newDevice.Exposes["brightness"].Data != 120.1 {
		t.Errorf("Error device brightness mismatch want: %v got: %v", 120.1, newDevice.Exposes["brightness"].Data)
	}

	if newDevice.Exposes["color_temp"].Data != 100.1 {
		t.Errorf("Error device color temp mismatch want: %v got: %v", 100.1, newDevice.Exposes["color_temp"].Data)
	}

	if newDevice.Exposes["state"].Data != "on" {
		t.Errorf("Error device state mismatch want: %v got: %v", "on", newDevice.Exposes["state"].Data)
	}
}
func assetExpose(bridgeExpose devices.BridgeExpose, expose *devices.Entity, category bridge.ExposeCategory, t *testing.T) {
	if expose.Category != category {
		t.Errorf("Error %s device mismatch want: %s got: %s", category, bridgeExpose.Name, expose.Name)
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
	if expose.Type == bridge.BinaryDataType {

		if expose.Values["on"] != bridgeExpose.ValueOn {
			t.Errorf("Error %s device value mismatch want: %v got: %v", bridgeExpose.Name, bridgeExpose.ValueOn, expose.Values["on"])
		}
		if expose.Values["off"] != bridgeExpose.ValueOff {
			t.Errorf("Error %s device value mismatch want: %v got: %v", bridgeExpose.Name, bridgeExpose.ValueOff, expose.Values["off"])
		}

		if bridgeExpose.ValueToggle != "" {
			if expose.Values["toggle"] != bridgeExpose.ValueToggle {
				t.Errorf("Error %s device value mismatch want: %v got: %v", bridgeExpose.Name, bridgeExpose.ValueToggle, expose.Values["toggle"])
			}
		}

		if bridgeExpose.ValueToggle != "" && expose.Values["toggle"] != bridgeExpose.ValueToggle {
			t.Errorf("Error %s device value mismatch want: %v got: %v", bridgeExpose.Name, bridgeExpose.ValueToggle, expose.Values["toggle"])
		}
	}

	if expose.Type == bridge.EnumDataType {
		for id, item := range bridgeExpose.Values {
			if expose.Values[fmt.Sprintf("%d", id)] != item {
				t.Errorf("Error %s device value mismatch want: %v got: %v", bridgeExpose.Name, item, expose.Values[fmt.Sprintf("%d", id)])
			}
		}
	}

	if expose.Type == bridge.NumericDataType {
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
