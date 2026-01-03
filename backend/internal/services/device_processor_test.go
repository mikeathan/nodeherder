package services_test

import (
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/models/settings"
	"node-herder/repository"
	utils_test "node-herder/testing"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestDeviceProcessor_CreateOrUpdateDevice_NewDevice(t *testing.T) {

	wg := sync.WaitGroup{}
	wg.Add(1)
	repo := repository.NewMemoryDeviceRepo()
	store := utils_test.CreateStoreFromDeviceRepo(repo)
	eventHub := &mocks.MockEventHub{}
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()
	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			t.Errorf("Error: OnDeviceUpdated called for new device")
		},
		OnNewDevice: func(d *devices.Device) {
			if err := store.StoreDevice(d.FriendlyName, d); err != nil {
				t.Errorf("Error store device add: %s", err)
			}
			wg.Done()
		},
		OnDeviceAvailabilityChanged: func(p *devices.UpdatePackage) {
		},
		AvailabilityTimeout: 1,
	}
	deviceName := "Attic room Light"

	lastSeen := time.Now().Format(time.RFC3339)
	payload := map[string]interface{}{}
	payload["brightness"] = 120.1
	payload["color_temp"] = 100.1
	payload["state"] = "on"
	payload["last_seen"] = lastSeen
	payload["battery"] = 100

	processor := services.NewDeviceProcessorBuilder().
		WithRegistrar(registrar).
		WithStore(store).
		WithEvents(events).
		WithAutomationQuerier(deviceQuerier).
		Build()

	err := processor.CreateOrUpdateDevice(deviceName, "wifi", payload)

	if err != nil {
		t.Errorf("Error creating new device: %s", err)
	}
	wg.Wait()
	d, err := store.FindDeviceByFriendlyName(deviceName)
	if err != nil {
		t.Errorf("Error finding device: %s", err)
	}

	if d.FriendlyName != deviceName {
		t.Errorf("Device FriendlyName mismatch want: %s got: %s", deviceName, d.FriendlyName)
	}
	if d.LastSeen != lastSeen {
		t.Errorf("Device LastSeen mismatch want: %s got: %s", lastSeen, d.LastSeen)
	}
	if d.ConnectionType != "wifi" {
		t.Errorf("Device ConnectionType mismatch want: %s got: %s", "wifi", d.ConnectionType)
	}
	if d.PowerSource != "battery" {
		t.Errorf("Device PowerSource mismatch want: %s got: %s", "battery", d.PowerSource)
	}
	if d.Exposes["brightness"].Data.Value() != 120.1 {
		t.Errorf("Device Expose brightness mismatch want: %f got: %f", 120.1, d.Exposes["brightness"].Data.Value())
	}
	if d.Exposes["color_temp"].Data.Value() != 100.1 {
		t.Errorf("Device Expose color_temp mismatch want: %f got: %f", 100.1, d.Exposes["color_temp"].Data.Value())
	}
	if d.Exposes["state"].Data.Value() != "on" {
		t.Errorf("Device Expose state mismatch want: %s got: %s", "on", d.Exposes["state"].Data.Value())
	}
}

func TestDeviceProcessor_CreateOrUpdateDevice_ExistingDevice(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

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
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(bridgeInfoes)

	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			if err := store.StoreDevice(d.FriendlyName, d); err != nil {
				t.Errorf("Error store device update: %s", err)
			}
			wg.Done()
		},
		OnNewDevice: func(d *devices.Device) {
			if err := store.StoreDevice(d.FriendlyName, d); err != nil {
				t.Errorf("Error store device add: %s", err)
			}
			wg.Done()
		},
		OnDeviceAvailabilityChanged: func(p *devices.UpdatePackage) {
		},
		OnDeviceMeasurementsUpdated: func(d *devices.Device, p map[string]interface{}) {

		},
		AvailabilityTimeout: 1,
	}
	deviceName := "Attic room Light"

	processor := services.NewDeviceProcessorBuilder().
		WithRegistrar(registrar).
		WithStore(store).
		WithEvents(events).
		WithAutomationQuerier(deviceQuerier).
		Build()

	lastSeen := time.Now().Format(time.RFC3339)
	updatePayload := map[string]interface{}{}
	updatePayload["brightness"] = 10.1
	updatePayload["color_temp"] = 120.1
	updatePayload["state"] = "false"
	updatePayload["last_seen"] = lastSeen
	updatePayload["battery"] = 100
	err = processor.CreateOrUpdateDevice(deviceName, "wifi", updatePayload)

	wg.Wait()
	if err != nil {
		t.Errorf("Error updating device: %s", err)
	}
	d, err := store.FindDeviceByFriendlyName(deviceName)
	if err != nil {
		t.Errorf("Error finding device: %s", err)
	}

	if d.FriendlyName != deviceName {
		t.Errorf("Device FriendlyName mismatch want: %s got: %s", deviceName, d.FriendlyName)
	}
	if d.LastSeen != lastSeen {
		t.Errorf("Device LastSeen mismatch want: %s got: %s", lastSeen, d.LastSeen)
	}
	if d.ConnectionType != "mqtt" {
		t.Errorf("Device ConnectionType mismatch want: %s got: %s", "mqtt", d.ConnectionType)
	}
	if d.PowerSource != "mains (single phase)" {
		t.Errorf("Device PowerSource mismatch want: %s got: %s", "mains (single phase)", d.PowerSource)
	}
	if d.Exposes["brightness"].Data.Value() != 10.1 {
		t.Errorf("Device Expose brightness mismatch want: %f got: %f", 10.1, d.Exposes["brightness"].Data.Value())
	}
	if d.Exposes["color_temp"].Data.Value() != 120.1 {
		t.Errorf("Device Expose color_temp mismatch want: %f got: %f", 120.1, d.Exposes["color_temp"].Data.Value())
	}
	if d.Exposes["state"].Data.Value() != "false" {
		t.Errorf("Device Expose state mismatch want: %s got: %s", "false", d.Exposes["state"].Data.Value())
	}
}

func TestOnDeviceConfigUpdated_WithDeviceOverride_ShouldDisableDevice(t *testing.T) {

	wg := sync.WaitGroup{}
	wg.Add(2)

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
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(bridgeInfoes)

	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			t.Errorf("Error	should 	not call OnDeviceUpdated for disabled device")
		},
		OnNewDevice: func(d *devices.Device) {
			if err := store.StoreDevice(d.FriendlyName, d); err != nil {
				t.Errorf("Error store device add: %s", err)
			}
			wg.Done()
		},
		OnDeviceAvailabilityChanged: func(p *devices.UpdatePackage) {
		},
		OnDeviceMeasurementsUpdated: func(d *devices.Device, p map[string]interface{}) {

		},
		AvailabilityTimeout: 1,
	}

	processor := services.NewDeviceProcessorBuilder().
		WithRegistrar(registrar).
		WithStore(store).
		WithEvents(events).
		WithAutomationQuerier(deviceQuerier).
		Build()

	//  Send payload 1
	// "friendly_name": "Attic room Light",
	// "ieee_address": "0x70ac08fffefafeca",
	deviceName := "Attic room Light"
	lastSeen := time.Now().Format(time.RFC3339)
	updatePayload := map[string]interface{}{}
	updatePayload["brightness"] = 10.1
	updatePayload["color_temp"] = 120.1
	updatePayload["state"] = "false"
	updatePayload["last_seen"] = lastSeen
	updatePayload["battery"] = 100
	processor.CreateOrUpdateDevice(deviceName, "wifi", updatePayload)

	//  Send payload 2
	// "friendly_name": "Living room presence sensor",
	// "ieee_address": "0xa4c13894070052fc",
	deviceName2 := "Living room presence sensor"
	lastSeen2 := time.Now().Format(time.RFC3339)
	updatePayload2 := map[string]interface{}{}
	updatePayload2["presence"] = true
	updatePayload2["target_distance"] = 102.1
	updatePayload2["last_seen"] = lastSeen2
	processor.CreateOrUpdateDevice(deviceName2, "mqtt", updatePayload2)

	cfg := settings.NewDeviceConfig("0xa4c13894070052fc")
	cfg.Disabled = true

	processor.OnDeviceConfigUpdated(cfg)

	time.Sleep(200 * time.Millisecond)
}

func TestOnDeviceConfigUpdated_WithDeviceDefaults_ShouldDisableAllDevices(t *testing.T) {

	//wg := sync.WaitGroup{}

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
		t.Fatal("Error creating file store:", err)
		return
	}
	defer cleanup()

	eventHub := &mocks.MockEventHub{}
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()

	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(bridgeInfoes)

	newDeviceIndex := 0
	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			t.Errorf("Error	should 	not call OnDeviceUpdated for disabled device")
		},
		OnNewDevice: func(d *devices.Device) {
			newDeviceIndex++
			if newDeviceIndex > 2 {
				t.Errorf("Error	should 	not call OnNewDevice for disabled device")
			}
		},
		OnDeviceAvailabilityChanged: func(p *devices.UpdatePackage) {
		},
		OnDeviceMeasurementsUpdated: func(d *devices.Device, p map[string]interface{}) {

		},
		AvailabilityTimeout: 1,
	}

	processor := services.NewDeviceProcessorBuilder().
		WithRegistrar(registrar).
		WithStore(store).
		WithEvents(events).
		WithAutomationQuerier(deviceQuerier).
		Build()

	// send two new devices before setting defaults

	//  Send payload 1
	// "friendly_name": "Attic room Light",
	// "ieee_address": "0x70ac08fffefafeca",
	deviceName := "Attic room Light"
	lastSeen := time.Now().Format(time.RFC3339)
	updatePayload := map[string]interface{}{}
	updatePayload["brightness"] = 10.1
	updatePayload["color_temp"] = 120.1
	updatePayload["state"] = "false"
	updatePayload["last_seen"] = lastSeen
	updatePayload["battery"] = 100
	processor.CreateOrUpdateDevice(deviceName, "wifi", updatePayload)

	//  Send payload 2
	// "friendly_name": "Living room presence sensor",
	// "ieee_address": "0xa4c13894070052fc",
	deviceName2 := "Living room presence sensor"
	lastSeen2 := time.Now().Format(time.RFC3339)
	updatePayload2 := map[string]interface{}{}
	updatePayload2["presence"] = true
	updatePayload2["target_distance"] = 102.1
	updatePayload2["last_seen"] = lastSeen2
	processor.CreateOrUpdateDevice(deviceName2, "mqtt", updatePayload2)

	cfg := store.AppConfig()
	cfg.RegisterDeviceConfigUpdateListener(func(cfg *settings.DeviceConfig) {
		processor.OnDeviceConfigUpdated(cfg)
	})
	// create defaults and set devices disabled
	defaults := settings.DefaultDeviceConfig()
	defaults.Disabled = true
	cfg.SetDeviceConfigDefaults(defaults)

	// send again, this update should be ignored
	lastSeen2 = time.Now().Format(time.RFC3339)
	updatePayload2["presence"] = false
	updatePayload2["target_distance"] = 12.1
	updatePayload2["last_seen"] = lastSeen2
	processor.CreateOrUpdateDevice(deviceName2, "mqtt", updatePayload2)

	// send new device and it should be ignored

	//  Send payload 3
	// "friendly_name": "Attic alarm",
	// "ieee_address": "0xa4c1389b273366c3",
	deviceName3 := "Attic alarm"
	lastSeen3 := time.Now().Format(time.RFC3339)
	updatePayload3 := map[string]interface{}{}
	updatePayload3["alarm"] = true
	updatePayload3["last_seen"] = lastSeen3
	processor.CreateOrUpdateDevice(deviceName3, "mqtt", updatePayload3)
}
