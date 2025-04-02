package services

import (
	"fmt"
	"node-herder/models/devices"
	"node-herder/store"
	"node-herder/utils"
	"sync"
)

type DeviceProcessor struct {
	registrar      *HubRegisterService
	deviceServices map[string]*DeviceLifetimeService
	store          store.AppStore
	events         *devices.DeviceRequestEvents
	mutex          *sync.RWMutex
}

func NewDeviceProcessor(registrar *HubRegisterService, store store.AppStore, events *devices.DeviceRequestEvents) *DeviceProcessor {
	return &DeviceProcessor{
		registrar:      registrar,
		deviceServices: make(map[string]*DeviceLifetimeService),
		store:          store,
		events:         events,
		mutex:          &sync.RWMutex{},
	}
}

func (dm *DeviceProcessor) CreateOrUpdateDevice(friendlyName, connType string, dataMap map[string]interface{}) error {
	device, _ := dm.registrar.LookupByName(friendlyName)
	if device == nil {
		return dm.createNewDevice(friendlyName, connType, dataMap)
	}

	dm.updateExistingDevice(device, dataMap)
	return nil
}

func (dm *DeviceProcessor) createNewDevice(friendlyName, connType string, dataMap map[string]interface{}) error {
	device, err := dm.registrar.CreateNewDevice(friendlyName, connType, dataMap)
	if err != nil {
		return fmt.Errorf("failed to create new device: %w", err)
	}

	dm.createDeviceService(device, dataMap)

	return nil
}

func (dm *DeviceProcessor) createDeviceService(device *devices.Device, dataMap map[string]interface{}) *DeviceLifetimeService {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	appConfig := dm.store.AppConfig()
	ls := NewDeviceLifetimeService(device, dm.events, appConfig.GetDeviceConfigCache(), utils.NewRealClock())
	ls.Start(dataMap)

	dm.deviceServices[device.Id] = ls

	return ls
}

func (dm *DeviceProcessor) getDeviceLifetime(device *devices.Device) (*DeviceLifetimeService, bool) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	ls, ok := dm.deviceServices[device.Id]
	return ls, ok

}
func (dm *DeviceProcessor) updateExistingDevice(device *devices.Device, dataMap map[string]interface{}) {

	if lf, ok := dm.getDeviceLifetime(device); ok {
		lf.Update(dataMap)
		return
	}

	// we are here because device is registered via bridge
	// but we dont have a device lifetime service created yet
	lf := dm.createDeviceService(device, dataMap)
	lf.Update(dataMap)
}
