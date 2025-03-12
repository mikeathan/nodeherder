package services

import (
	"fmt"
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/store"
	"node-herder/utils"
)

type DeviceProcessor struct {
	registrar           *HubRegisterService
	deviceServices      map[string]*DeviceLifetimeService
	store               store.AppStore
	events              *devices.DeviceRequestEvents
	eventHub            ws.EventHub
	availabilityTimeout int
}

func NewDeviceProcessor(registrar *HubRegisterService, store store.AppStore, events *devices.DeviceRequestEvents, eventHub ws.EventHub, timeout int) *DeviceProcessor {
	return &DeviceProcessor{
		registrar:           registrar,
		deviceServices:      make(map[string]*DeviceLifetimeService),
		eventHub:            eventHub,
		availabilityTimeout: timeout,
		store:               store,
		events:              events,
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

	// dm.eventHub.Broadcast(ws.DeviceAdded, device)
	// dm.hub.deviceAdded(device, dataMap)

	return nil
}

func (dm *DeviceProcessor) createDeviceService(device *devices.Device, dataMap map[string]interface{}) {
	appConfig := dm.store.AppConfig()
	debouncer := NewDeviceDebouncer(device.Id, appConfig.GetDeviceConfigCache(device.Id), utils.NewRealClock())

	s := NewDeviceLifetimeService(device, dm.events, debouncer)
	s.Start(dataMap)

	dm.deviceServices[device.Id] = s
}

func (dm *DeviceProcessor) updateExistingDevice(device *devices.Device, dataMap map[string]interface{}) error {
	deviceService, ok := dm.deviceServices[device.Id]
	if !ok {
		return fmt.Errorf("device service not found for device ID: %s", device.Id)
	}

	deviceService.Update(dataMap)
	// if updatedData.HasData() {
	// 	return device, nil
	// }

	// dm.eventHub.Broadcast(ws.DeviceUpdated, updatedData)
	// dm.hub.deviceUpdated(device, updatedData.Data)
	return nil
}

func (dm *DeviceProcessor) OnNewDevice(action func(device *devices.Device, dataMap map[string]interface{})) {

}
func (dm *DeviceProcessor) OnDeviceUpdated(action func(device *devices.Device, dataMap map[string]interface{})) {

}
