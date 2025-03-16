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
	device, _ := dm.registrar.LookupByName(friendlyName)
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

	return dm.createDeviceService(device, dataMap)
}

func (dm *DeviceProcessor) createDeviceService(device *devices.Device, dataMap map[string]interface{}) error {
	appConfig := dm.store.AppConfig()
	debouncer := settings.NewDeviceDebouncer(device.Id, appConfig.GetDeviceConfigCache(device.Id), utils.NewRealClock())

	ls := NewDeviceLifetimeService(device, dm.events, debouncer)
	ls.Start(dataMap)

	if err := dm.registrar.Register(device.FriendlyName, device); err != nil {
		return err
	}

	dm.deviceServices[device.Id] = ls

	return nil
}

func (dm *DeviceProcessor) updateExistingDevice(device *devices.Device, dataMap map[string]interface{}) error {
	if deviceService, ok := dm.deviceServices[device.Id]; ok {
		deviceService.Update(dataMap)
		if err := dm.registrar.Register(device.FriendlyName, device); err != nil {
			return err
		}
		return nil
	}
	todo  check code path. do we need to registerd device if service not exists but device is registered via bridge?

	// we are here because device is registered via bridge
	// but we dont have a device lifetime service created yet
	err := dm.createDeviceService(device, dataMap)
	if err != nil {
		return err
	}

	dm.deviceServices[device.Id].Update(dataMap)
	return nil
}
