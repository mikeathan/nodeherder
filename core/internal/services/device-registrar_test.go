package services_test

import (
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

	//wg := &sync.WaitGroup{}
	//mqtt := &mocks.MockMqttClient{}

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

	for _, bridgeExpose := range measureExposes {
		found := false
		for _, device := range storeDevices {
			for _, expose := range device.Exposes {
				
				need to match the device that bridgeExpose is coming from so we can assert all data

				if expose.Name == bridgeExpose.Name {
					if expose.Category != devices.MeasurementCategory {
						t.Errorf("Error measurement device mismatch want: %s got: %s", bridgeExpose.Name, expose.Name)
					}

					if expose.AccessMode != bridgeExpose.Access {
						t.Errorf("Error measurement device access mismatch want: %v got: %v", bridgeExpose.Access, expose.AccessMode)
					}

					if expose.Type != bridgeExpose.Type {
						t.Errorf("Error measurement device type mismatch want: %v got: %v", bridgeExpose.Type, expose.Type)
					}

					if expose.Unit != bridgeExpose.Unit {
						t.Errorf("Error measurement device unit mismatch want: %v got: %v", bridgeExpose.Unit, expose.Unit)
					}
					if expose.Description != bridgeExpose.Description {
						t.Errorf("Error measurement device description mismatch want: %v got: %v", bridgeExpose.Description, expose.Description)
					}

					found = true
				}
			}
		}

		if !found {
			t.Errorf("Error measurement bridge expose %s not found in store", bridgeExpose.Name)
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

	for _, bridgeExpose := range diagnosticExposes {
		found := false
		for _, device := range storeDevices {
			for _, expose := range device.Exposes {
				if expose.Name == bridgeExpose.Name {
					if expose.Category != devices.DiagnosticCategory {
						t.Errorf("Error diagnostic device mismatch want: %s got: %s", bridgeExpose.Name, expose.Name)
					}

					found = true
				}
			}
		}

		if !found {
			t.Errorf("Error diagnostic bridge expose %s not found in store", bridgeExpose.Name)
		}
	}

	// assert config devices
	configExposes, err := devices.FindAllExposesByCategory(data, devices.ConfigCategory)
	if err != nil {
		t.Fatal("Error loading ConfigCategory devices:", err)
		return
	}

	if len(configExposes) == 0 {
		t.Errorf("Error not found any config devices")
	}

	for _, bridgeExpose := range configExposes {
		found := false
		for _, device := range storeDevices {
			for _, expose := range device.Exposes {
				if expose.Name == bridgeExpose.Name {
					if expose.Category != devices.ConfigCategory {
						t.Errorf("Error config device mismatch want: %s got: %s", bridgeExpose.Name, expose.Name)
					}

					found = true
				}
			}
		}

		if !found {
			t.Errorf("Error config bridge expose %s not found in store", bridgeExpose.Name)
		}
	}
}
