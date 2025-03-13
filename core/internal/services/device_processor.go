package services

import (
	"fmt"
	"node-herder/models/devices"
	"node-herder/models/settings"
	"node-herder/store"
	"node-herder/utils"
)

type DeviceProcessor struct {
	registrar      *HubRegisterService
	deviceServices map[string]*DeviceLifetimeService
	store          store.AppStore
	events         *devices.DeviceRequestEvents
}

func NewDeviceProcessor(registrar *HubRegisterService, store store.AppStore, events *devices.DeviceRequestEvents) *DeviceProcessor {
	return &DeviceProcessor{
		registrar:      registrar,
		deviceServices: make(map[string]*DeviceLifetimeService),
		store:          store,
		events:         events,
	}
}

func (dm *DeviceProcessor) CreateOrUpdateDevice(friendlyName, connType string, dataMap map[string]interface{}) error {
	device, err := dm.registrar.LookupByName(friendlyName)
	if err != nil {
		return fmt.Errorf("error looking up device: %w", err)
	}

	if device == nil {
		return dm.createNewDevice(friendlyName, connType, dataMap)
	}

	return dm.updateExistingDevice(device, dataMap)
}

func (dm *DeviceProcessor) createNewDevice(friendlyName, connType string, dataMap map[string]interface{}) error {
	device, err := dm.registrar.CreateNewDevice(friendlyName, connType, dataMap)
	if err != nil {
		return fmt.Errorf("failed to create new device: %w", err)
	}

	dm.createDeviceService(device, dataMap)
	return nil
}

func (dm *DeviceProcessor) createDeviceService(device *devices.Device, dataMap map[string]interface{}) {
	appConfig := dm.store.AppConfig()
	debouncer := settings.NewDeviceDebouncer(device.Id, appConfig.GetDeviceConfigCache(device.Id), utils.NewRealClock())

	ls := NewDeviceLifetimeService(device, dm.events, debouncer)
	ls.Start(dataMap)

	dm.deviceServices[device.Id] = ls
}

func (dm *DeviceProcessor) updateExistingDevice(device *devices.Device, dataMap map[string]interface{}) error {
	deviceService, ok := dm.deviceServices[device.Id]
	if !ok {
		return fmt.Errorf("device service not found for device ID: %s", device.Id)
	}

	deviceService.Update(dataMap)
	return nil
}
