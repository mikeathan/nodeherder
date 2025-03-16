package services_test

import (
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/devices"
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
	registrar := services.NewHubRegisterService(store, eventHub, 30000)

	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			t.Errorf("Error: OnDeviceUpdated called for new device")
		},
		OnNewDevice: func(d *devices.Device, p map[string]interface{}) {
			// if err := registrar.Register(d.FriendlyName, d); err != nil {
			// 	t.Errorf("Error registering device add: %s", err)
			// }
			wg.Done()
		},
		OnDeviceAvailabilityChanged: func(p *devices.UpdatePackage) {
		},
		AvailabilityTimeout: 1,
	}
	deviceName := "Living room light"

	lastSeen := time.Now().Format(time.RFC3339)
	payload := map[string]interface{}{}
	payload["brightness"] = 120.1
	payload["color_temp"] = 100.1
	payload["state"] = "on"
	payload["last_seen"] = lastSeen
	payload["battery"] = 100
	//appConfig := settings.NewAppConfig()

	processor := services.NewDeviceProcessor(registrar, store, events)

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
	if d.Exposes["brightness"].Data != 120.1 {
		t.Errorf("Device Expose brightness mismatch want: %f got: %f", 120.1, d.Exposes["brightness"].Data)
	}
	if d.Exposes["color_temp"].Data != 100.1 {
		t.Errorf("Device Expose color_temp mismatch want: %f got: %f", 100.1, d.Exposes["color_temp"].Data)
	}
	if d.Exposes["state"].Data != "on" {
		t.Errorf("Device Expose state mismatch want: %s got: %s", "on", d.Exposes["state"].Data)
	}

}

func TestDeviceProcessor_CreateOrUpdateDevice_ExistingDevice(t *testing.T) {
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
	registrar := services.NewHubRegisterService(store, eventHub, 30000)
	registrar.RegisterBridge(bridgeInfoes)

	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			// if err := registrar.Register(d.FriendlyName, d); err != nil {
			// 	t.Errorf("Error registering device update: %s", err)
			// }
			wg.Done()
		},
		OnNewDevice: func(d *devices.Device, p map[string]interface{}) {
			// if err := registrar.Register(d.FriendlyName, d); err != nil {
			// 	t.Errorf("Error registering device add: %s", err)
			// }
			wg.Done()
		},
		OnDeviceAvailabilityChanged: func(p *devices.UpdatePackage) {
		},
		AvailabilityTimeout: 1,
	}
	deviceName := "Living room light"

	

	processor := services.NewDeviceProcessor(registrar, store, events)

	

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
	if d.Exposes["brightness"].Data != 120.1 {
		t.Errorf("Device Expose brightness mismatch want: %f got: %f", 120.1, d.Exposes["brightness"].Data)
	}
	if d.Exposes["color_temp"].Data != 100.1 {
		t.Errorf("Device Expose color_temp mismatch want: %f got: %f", 100.1, d.Exposes["color_temp"].Data)
	}
	if d.Exposes["state"].Data != "on" {
		t.Errorf("Device Expose state mismatch want: %s got: %s", "on", d.Exposes["state"].Data)
	}

}

// func TestDeviceProcessor_CreateOrUpdateDevice_LookupError(t *testing.T) {
// 	mockRegistrar := new(MockHubRegisterService)
// 	mockStore := new(MockAppStore)
// 	mockEvents := new(MockDeviceRequestEvents)

// 	mockRegistrar.On("LookupByName", "testDevice").Return((*devices.Device)(nil), errors.New("lookup error"))

// 	processor := NewDeviceProcessor(mockRegistrar, mockStore, mockEvents)
// 	err := processor.CreateOrUpdateDevice("testDevice", "wifi", map[string]interface{}{})

// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "error looking up device")
// 	mockRegistrar.AssertExpectations(t)
// }

// func TestDeviceProcessor_CreateOrUpdateDevice_CreateNewError(t *testing.T) {
// 	mockRegistrar := new(MockHubRegisterService)
// 	mockStore := new(MockAppStore)
// 	mockEvents := new(MockDeviceRequestEvents)

// 	mockRegistrar.On("LookupByName", "testDevice").Return((*devices.Device)(nil), nil)
// 	mockRegistrar.On("CreateNewDevice", "testDevice", "wifi", map[string]interface{}{}).Return((*devices.Device)(nil), errors.New("create error"))

// 	processor := NewDeviceProcessor(mockRegistrar, mockStore, mockEvents)
// 	err := processor.CreateOrUpdateDevice("testDevice", "wifi", map[string]interface{}{})

// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "failed to create new device")
// 	mockRegistrar.AssertExpectations(t)
// }

// func TestDeviceProcessor_CreateOrUpdateDevice_DeviceServiceNotFound(t *testing.T) {
// 	mockRegistrar := new(MockHubRegisterService)
// 	mockStore := new(MockAppStore)
// 	mockEvents := new(MockDeviceRequestEvents)

// 	existingDevice := &devices.Device{Id: "existingID", FriendlyName: "existingDevice"}
// 	mockRegistrar.On("LookupByName", "existingDevice").Return(existingDevice, nil)

// 	processor := NewDeviceProcessor(mockRegistrar, mockStore, mockEvents)
// 	err := processor.CreateOrUpdateDevice("existingDevice", "wifi", map[string]interface{}{})

// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "device service not found for device ID")
// 	mockRegistrar.AssertExpectations(t)
// }

// func TestDeviceProcessor_createDeviceService(t *testing.T) {
// 	mockRegistrar := new(MockHubRegisterService)
// 	mockStore := new(MockAppStore)
// 	mockEvents := new(MockDeviceRequestEvents)
// 	device := &devices.Device{Id: "testDeviceID", FriendlyName: "testDevice"}
// 	appConfig := settings.NewAppConfig()
// 	mockStore.On("AppConfig").Return(appConfig)
// 	mockEvents.On("Subscribe", device).Return(make(chan devices.DeviceRequest))
// 	processor := NewDeviceProcessor(mockRegistrar, mockStore, mockEvents)
// 	processor.createDeviceService(device, map[string]interface{}{})
// 	assert.Contains(t, processor.deviceServices, "testDeviceID")
// }

// func TestDeviceProcessor_updateExistingDevice(t *testing.T) {
// 	mockRegistrar := new(MockHubRegisterService)
// 	mockStore := new(MockAppStore)
// 	mockEvents := new(MockDeviceRequestEvents)
// 	device := &devices.Device{Id: "testDeviceID", FriendlyName: "testDevice"}
// 	appConfig := settings.NewAppConfig()
// 	mockStore.On("AppConfig").Return(appConfig)
// 	mockEvents.On("Subscribe", device).Return(make(chan devices.DeviceRequest))
// 	processor := NewDeviceProcessor(mockRegistrar, mockStore, mockEvents)
// 	processor.createDeviceService(device, map[string]interface{}{})

// 	err := processor.updateExistingDevice(device, map[string]interface{}{})
// 	assert.NoError(t, err)
// }
