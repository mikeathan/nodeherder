package services

import (
	"fmt"
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/store"
	"node-herder/utils"
)

type DeviceRegistrar interface {
	LookupByName(name string) (*devices.Device, error)
	LookupById(id string) (*devices.Device, error)
	RetrieveEntityData(id string, property string) (any, error)
	CreateNewDevice(friendlyName string, connType string, data map[string]interface{}) (*devices.Device, error)
	FindBridgeInfo(id string) (*devices.BridgeInfo, error)
	RegisterBridge(bridgeInfoList []*devices.BridgeInfo, deviceAvailabilityTimeoutOverride int)
}

type HubRegisterService struct {
	idMapper                  map[string]string
	store                     store.AppStore
	eventHub                  ws.EventHub
	deviceAvailabilityTimeout int
}

func NewHubRegisterService(store store.AppStore, hub ws.EventHub, deviceAvailabilityTimeout int) *HubRegisterService {
	return &HubRegisterService{store: store, eventHub: hub, idMapper: make(map[string]string), deviceAvailabilityTimeout: deviceAvailabilityTimeout}
}

func (s *HubRegisterService) Register(friendlyName string, device *devices.Device) {

	// TODO:
	// have some config to check if device is allowed to publish metrics
	// and we store metrics here if so.

	id := s.ResolveId(friendlyName)

	s.store.StoreDevice(device)
	//	s.store.Devices().Store(id, device)

	s.idMapper[friendlyName] = id // store id in mapper for easy access
}

func (s *HubRegisterService) RetrieveEntityData(id string, property string) (any, error) {
	device, err := s.LookupById(id)
	if err != nil {
		return nil, err
	}
	if entity, ok := device.Exposes[property]; ok {
		return entity.Data, nil
	}

	return nil, fmt.Errorf("entity %s not found", property)
}

func (s *HubRegisterService) LookupByName(name string) (*devices.Device, error) {

	id := s.ResolveId(name)

	return s.store.Devices().FindDevice(id)
}

func (s *HubRegisterService) LookupById(id string) (*devices.Device, error) {

	return s.store.Devices().FindDevice(id)
}

func (s *HubRegisterService) CreateNewDevice(friendlyName string, connType string, data map[string]interface{}) (*devices.Device, error) {

	id := s.ResolveId(friendlyName)
	bridgeInfo, err := s.store.Devices().FindBridgeInfo(id)
	if err != nil {
		utils.LogInfof("BrideInfo not found for device id:%v friendlyName:%v", id, friendlyName)
	}

	device, err := devices.CreateNewDevice(id, friendlyName, connType, bridgeInfo, data)
	if err != nil {
		return nil, err
	}

	device.Monitor(s.deviceAvailabilityTimeout, func(p interface{}) {
		s.eventHub.Broadcast(ws.DeviceUpdated, p)
	})

	return device, nil
}

func (s *HubRegisterService) configureIdMapper() {
	bridgeInfoList, err := s.store.Devices().AllBridgeInfo()
	if err != nil {
		utils.LogErrorf("Error loading bridgeInfoList %s", err.Error())
		return
	}

	// clean up
	for name, id := range s.idMapper {
		var found = false
		for _, device := range bridgeInfoList {
			if device.IeeeAddress == id && device.FriendlyName == name {
				found = true
				break
			}
		}

		if !found {
			utils.LogDebugf("Removing %s from IdMapper", name)
			delete(s.idMapper, name)
		}
	}

	// setup
	for _, device := range bridgeInfoList {
		s.idMapper[device.FriendlyName] = device.IeeeAddress
	}
}

func (a *HubRegisterService) FindBridgeInfo(id string) (*devices.BridgeInfo, error) {
	return a.store.Devices().FindBridgeInfo(id)
}

func (s *HubRegisterService) RegisterBridge(bridgeInfoList []*devices.BridgeInfo, deviceAvailabilityTimeoutOverride int) {

	err := s.store.Devices().StoreBridge(bridgeInfoList)
	if err != nil {
		utils.LogErrorf("store bridgeinfo failed: %s", err.Error())
		return
	}

	s.configureIdMapper()

	for _, bridgeInfo := range bridgeInfoList {
		if !bridgeInfo.IsActive() {
			continue
		}

		d, err := s.store.Devices().FindDevice(bridgeInfo.IeeeAddress)
		if err != nil {
			// not found in repo, new it here
			d = devices.NewDevice(bridgeInfo.IeeeAddress)
			d.ConnectionType = "mqtt"

			// load exposes
			for _, expose := range bridgeInfo.Definition.Exposes {
				entity, err := devices.CreateEntityFromExpose(expose, nil)
				if err != nil {
					utils.LogDebugf("expose failed loading %s error %s", bridgeInfo.FriendlyName, err.Error())
					continue
				}
				d.Exposes[entity.Name] = entity
				d.Properties["availability"] = "offline"
			}

			// load features
			for _, expose := range bridgeInfo.Definition.Exposes {
				for _, feature := range expose.Features {
					entity, err := devices.CreateEntityFromFeature(feature, nil)
					if err != nil {
						utils.LogDebugf("feature failed loading %s error %s", bridgeInfo.FriendlyName, err.Error())
						continue
					}

					d.Exposes[entity.Name] = entity
					d.Properties["availability"] = "offline"
				}
			}

			if len(d.Exposes) == 0 {
				utils.LogInfof("device %s failed loading. error contains no valid exposed data", bridgeInfo.FriendlyName)
				continue
			}

			d.Monitor(deviceAvailabilityTimeoutOverride, func(p interface{}) {
				s.eventHub.Broadcast(ws.DeviceUpdated, p)
			})
		}

		d.Description = bridgeInfo.Definition.Description
		d.FriendlyName = bridgeInfo.FriendlyName
		d.PowerSource = bridgeInfo.PowerSource

		s.Register(d.FriendlyName, d)
	}
}

func (a *HubRegisterService) ResolveId(name string) string {

	// check if id already exists
	if id, ok := a.idMapper[name]; ok {
		return id
	}

	return utils.HashName(name)
}
