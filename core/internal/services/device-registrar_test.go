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

	measureExposes, err := devices.FindAllExposesByCategory(data, devices.MeasurementCategory)
	if err != nil {
		t.Fatal("Error loading measurementDevices:", err)
		return
	}

	storeDevices, _ := store.AllDevices()

	for _, device := range storeDevices {
		for _, expose := range device.Exposes {
			found := false
			for _, bridgeExpose := range measureExposes {

				if expose.Name == bridgeExpose.Name {
					if expose.Category != devices.DiagnosticCategory {
						t.Errorf("Error measurement device mismatch want: %s got: %s", bridgeExpose.Name, expose.Name)
					}

					found = true
				}
			}

			if !found {
				t.Errorf("Error measurement expose %s not found in bridge", expose.Name)
			}

		}
	}

	// for _, d := range storeDevices {
	// 	for k, v := range d.Exposes {
	// 	}
	// }
}
