package services_test

import (
	"errors"
	"node-herder/models/settings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeviceProcessor_CreateOrUpdateDevice_NewDevice(t *testing.T) {
	mockRegistrar := new(MockHubRegisterService)
	mockStore := new(MockAppStore)
	mockEvents := new(MockDeviceRequestEvents)

	dataMap := map[string]interface{}{"key": "value"}
	appConfig := settings.NewAppConfig()
	mockStore.On("AppConfig").Return(appConfig)

	mockRegistrar.On("LookupByName", "testDevice").Return((*devices.Device)(nil), nil)
	newDevice := &devices.Device{Id: "testID", FriendlyName: "testDevice"}
	mockRegistrar.On("CreateNewDevice", "testDevice", "wifi", dataMap).Return(newDevice, nil)
	mockEvents.On("Subscribe", newDevice).Return(make(chan devices.DeviceRequest))

	processor := NewDeviceProcessor(mockRegistrar, mockStore, mockEvents)
	err := processor.CreateOrUpdateDevice("testDevice", "wifi", dataMap)

	assert.NoError(t, err)
	assert.Contains(t, processor.deviceServices, "testID")
	mockRegistrar.AssertExpectations(t)
	mockStore.AssertExpectations(t)
	mockEvents.AssertExpectations(t)
}

func TestDeviceProcessor_CreateOrUpdateDevice_ExistingDevice(t *testing.T) {
	mockRegistrar := new(MockHubRegisterService)
	mockStore := new(MockAppStore)
	mockEvents := new(MockDeviceRequestEvents)

	dataMap := map[string]interface{}{"key": "value"}
	existingDevice := &devices.Device{Id: "existingID", FriendlyName: "existingDevice"}
	mockRegistrar.On("LookupByName", "existingDevice").Return(existingDevice, nil)

	appConfig := settings.NewAppConfig()
	mockStore.On("AppConfig").Return(appConfig)

	mockEvents.On("Subscribe", existingDevice).Return(make(chan devices.DeviceRequest))

	processor := NewDeviceProcessor(mockRegistrar, mockStore, mockEvents)
	processor.createDeviceService(existingDevice, dataMap)

	updatedDataMap := map[string]interface{}{"newKey": "newValue"}
	err := processor.CreateOrUpdateDevice("existingDevice", "wifi", updatedDataMap)

	assert.NoError(t, err)
	assert.Contains(t, processor.deviceServices, "existingID")
	mockRegistrar.AssertExpectations(t)
	mockStore.AssertExpectations(t)
	mockEvents.AssertExpectations(t)
}

func TestDeviceProcessor_CreateOrUpdateDevice_LookupError(t *testing.T) {
	mockRegistrar := new(MockHubRegisterService)
	mockStore := new(MockAppStore)
	mockEvents := new(MockDeviceRequestEvents)

	mockRegistrar.On("LookupByName", "testDevice").Return((*devices.Device)(nil), errors.New("lookup error"))

	processor := NewDeviceProcessor(mockRegistrar, mockStore, mockEvents)
	err := processor.CreateOrUpdateDevice("testDevice", "wifi", map[string]interface{}{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error looking up device")
	mockRegistrar.AssertExpectations(t)
}

func TestDeviceProcessor_CreateOrUpdateDevice_CreateNewError(t *testing.T) {
	mockRegistrar := new(MockHubRegisterService)
	mockStore := new(MockAppStore)
	mockEvents := new(MockDeviceRequestEvents)

	mockRegistrar.On("LookupByName", "testDevice").Return((*devices.Device)(nil), nil)
	mockRegistrar.On("CreateNewDevice", "testDevice", "wifi", map[string]interface{}{}).Return((*devices.Device)(nil), errors.New("create error"))

	processor := NewDeviceProcessor(mockRegistrar, mockStore, mockEvents)
	err := processor.CreateOrUpdateDevice("testDevice", "wifi", map[string]interface{}{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create new device")
	mockRegistrar.AssertExpectations(t)
}

func TestDeviceProcessor_CreateOrUpdateDevice_DeviceServiceNotFound(t *testing.T) {
	mockRegistrar := new(MockHubRegisterService)
	mockStore := new(MockAppStore)
	mockEvents := new(MockDeviceRequestEvents)

	existingDevice := &devices.Device{Id: "existingID", FriendlyName: "existingDevice"}
	mockRegistrar.On("LookupByName", "existingDevice").Return(existingDevice, nil)

	processor := NewDeviceProcessor(mockRegistrar, mockStore, mockEvents)
	err := processor.CreateOrUpdateDevice("existingDevice", "wifi", map[string]interface{}{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "device service not found for device ID")
	mockRegistrar.AssertExpectations(t)
}

func TestDeviceProcessor_createDeviceService(t *testing.T) {
	mockRegistrar := new(MockHubRegisterService)
	mockStore := new(MockAppStore)
	mockEvents := new(MockDeviceRequestEvents)
	device := &devices.Device{Id: "testDeviceID", FriendlyName: "testDevice"}
	appConfig := settings.NewAppConfig()
	mockStore.On("AppConfig").Return(appConfig)
	mockEvents.On("Subscribe", device).Return(make(chan devices.DeviceRequest))
	processor := NewDeviceProcessor(mockRegistrar, mockStore, mockEvents)
	processor.createDeviceService(device, map[string]interface{}{})
	assert.Contains(t, processor.deviceServices, "testDeviceID")
}

func TestDeviceProcessor_updateExistingDevice(t *testing.T) {
	mockRegistrar := new(MockHubRegisterService)
	mockStore := new(MockAppStore)
	mockEvents := new(MockDeviceRequestEvents)
	device := &devices.Device{Id: "testDeviceID", FriendlyName: "testDevice"}
	appConfig := settings.NewAppConfig()
	mockStore.On("AppConfig").Return(appConfig)
	mockEvents.On("Subscribe", device).Return(make(chan devices.DeviceRequest))
	processor := NewDeviceProcessor(mockRegistrar, mockStore, mockEvents)
	processor.createDeviceService(device, map[string]interface{}{})

	err := processor.updateExistingDevice(device, map[string]interface{}{})
	assert.NoError(t, err)
}
