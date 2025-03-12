package services

import (
	"fmt"
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/utils"
)

type DeviceProcessor struct {
	registrar           *HubRegisterService
	deviceServices      map[string]*DeviceLifetimeService
	hub                 *Hub // needs fixing
	eventHub            ws.EventHub
	availabilityTimeout int
}

func NewDeviceProcessor(registrar *HubRegisterService, hub *Hub, eventHub ws.EventHub, timeout int) *DeviceProcessor {
	return &DeviceProcessor{
		registrar:           registrar,
		deviceServices:      make(map[string]*DeviceLifetimeService),
		hub:                 hub,
		eventHub:            eventHub,
		availabilityTimeout: timeout,
	}
}

func (dm *DeviceProcessor) CreateOrUpdateDevice(friendlyName, connType string, dataMap map[string]interface{}) (*devices.Device, error) {
	device, err := dm.registrar.LookupByName(friendlyName)
	if err != nil {
		return nil, fmt.Errorf("error looking up device: %w", err)
	}

	if device == nil {
		return dm.createNewDevice(friendlyName, connType, dataMap)
	}

	return dm.updateExistingDevice(device, dataMap)
}

func (dm *DeviceProcessor) createNewDevice(friendlyName, connType string, dataMap map[string]interface{}) (*devices.Device, error) {
	device, err := dm.registrar.CreateNewDevice(friendlyName, connType, dataMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create new device: %w", err)
	}

	if err := dm.createDeviceService(device); err != nil {
		return nil, err
	}

	dm.eventHub.Broadcast(ws.DeviceAdded, device)
	dm.hub.deviceAdded(device, dataMap)

	return device, nil
}

func (dm *DeviceProcessor) createDeviceService(device *devices.Device) error {
	appConfig := dm.hub.store.AppConfig()
	debouncer := NewDeviceDebouncer(device.Id, appConfig.GetDeviceConfigCache(device.Id), utils.NewRealClock())

	s := NewDeviceLifetimeService(device, debouncer)
	s.Monitor(dm.availabilityTimeout, func(p any) {
		dm.eventHub.Broadcast(ws.DeviceUpdated, p)
	})

	dm.deviceServices[device.Id] = s
	return nil
}

func (dm *DeviceProcessor) updateExistingDevice(device *devices.Device, dataMap map[string]interface{}) (*devices.Device, error) {
	deviceService, ok := dm.deviceServices[device.Id]
	if !ok {
		return nil, fmt.Errorf("device service not found for device ID: %s", device.Id)
	}

	updatedData := deviceService.Update(dataMap)
	if updatedData.IsEmpty() {
		return device, nil
	}

	dm.eventHub.Broadcast(ws.DeviceUpdated, updatedData)
	dm.hub.deviceUpdated(device, updatedData.Data)
	return device, nil
}
