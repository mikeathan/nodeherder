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
	"time"
)

func TestDeviceProcessor_CreateOrUpdateDevice_NewDevice(t *testing.T) {

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
	//registrar.RegisterBridge(bridgeInfoes)
	fmt.Println(bridgeInfoes)
	//device := utils_test.CreateLightDevice("x01234", "testDevice", "brigthness", 124.2)

	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
		},
		OnNewDevice: func(d *devices.Device, p map[string]interface{}) {
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

	err = processor.CreateOrUpdateDevice(deviceName, "wifi", payload)

	if err != nil {
		t.Errorf("Error creating new device: %s", err)
	}
	time.Sleep(1 * time.Second)
	d, err := store.FindDeviceByFriendlyName(deviceName)
	if err != nil {
		t.Errorf("Error finding device: %s", err)
	}
	
	if d.FriendlyName != deviceName {
		t.Errorf("Device FriendlyName mismatch want: %s got: %s", deviceName, d.FriendlyName)
	}

}

// func TestDeviceProcessor_CreateOrUpdateDevice_ExistingDevice(t *testing.T) {
// 	mockRegistrar := new(MockHubRegisterService)
// 	mockStore := new(MockAppStore)
// 	mockEvents := new(MockDeviceRequestEvents)

// 	dataMap := map[string]interface{}{"key": "value"}
// 	existingDevice := &devices.Device{Id: "existingID", FriendlyName: "existingDevice"}
// 	mockRegistrar.On("LookupByName", "existingDevice").Return(existingDevice, nil)

// 	appConfig := settings.NewAppConfig()
// 	mockStore.On("AppConfig").Return(appConfig)

// 	mockEvents.On("Subscribe", existingDevice).Return(make(chan devices.DeviceRequest))

// 	processor := NewDeviceProcessor(mockRegistrar, mockStore, mockEvents)
// 	processor.createDeviceService(existingDevice, dataMap)

// 	updatedDataMap := map[string]interface{}{"newKey": "newValue"}
// 	err := processor.CreateOrUpdateDevice("existingDevice", "wifi", updatedDataMap)

// 	assert.NoError(t, err)
// 	assert.Contains(t, processor.deviceServices, "existingID")
// 	mockRegistrar.AssertExpectations(t)
// 	mockStore.AssertExpectations(t)
// 	mockEvents.AssertExpectations(t)
// }

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
