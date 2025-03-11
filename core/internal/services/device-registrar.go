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
	StoreMetrics(friendlyName string, data map[string]interface{}) error
}

type HubRegisterService struct {
	deviceIdMapper            map[string]string
	store                     store.AppStore
	eventHub                  ws.EventHub
	deviceAvailabilityTimeout int
}

func NewHubRegisterService(store store.AppStore, hub ws.EventHub, deviceAvailabilityTimeout int) *HubRegisterService {
	return &HubRegisterService{store: store, eventHub: hub, deviceIdMapper: make(map[string]string), deviceAvailabilityTimeout: deviceAvailabilityTimeout}
}

func (s *HubRegisterService) Register(friendlyName string, device *devices.Device) error {
	return s.store.StoreDevice(friendlyName, device)
}

func (s *HubRegisterService) StoreMetrics(friendlyName string, data map[string]interface{}) error {
	return s.store.StoreMetrics(friendlyName, data)
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

func (s *HubRegisterService) LookupByName(friendlyName string) (*devices.Device, error) {

	return s.store.FindDeviceByFriendlyName(friendlyName)
}

func (s *HubRegisterService) LookupById(id string) (*devices.Device, error) {
	return s.store.FindDeviceById(id)
}

func (s *HubRegisterService) RemoveDevice(id string) error {
	return s.store.RemoveDeviceById(id)
}

func (s *HubRegisterService) CreateNewDevice(friendlyName string, connType string, data map[string]interface{}) (*devices.Device, error) {

	bridgeInfo, err := s.store.FindBridgeInfoByFriendlyName(friendlyName)
	if err != nil {
		utils.LogInfof("BrideInfo not found friendlyName: %v", friendlyName)
	}

	id := s.store.ResolveFriendlyName(friendlyName)
	device, err := devices.CreateNewDevice(id, friendlyName, connType, bridgeInfo, data)
	if err != nil {
		return nil, err
	}

	// s.Monitor(s.deviceAvailabilityTimeout, func(p interface{}) {
	// 	s.eventHub.Broadcast(ws.DeviceUpdated, p)
	// })

	return device, nil
}

func (a *HubRegisterService) FindBridgeInfo(id string) (*devices.BridgeInfo, error) {
	return a.store.FindBridgeInfoById(id)
}

func (s *HubRegisterService) RegisterBridge(bridgeInfoList []*devices.BridgeInfo, deviceAvailabilityTimeoutOverride int) {

	err := s.store.StoreBridgeInfoList(bridgeInfoList)
	if err != nil {
		utils.LogErrorf("store bridgeInfoList failed: %s", err.Error())
		return
	}

	for _, bridgeInfo := range bridgeInfoList {
		if !bridgeInfo.IsActive() {
			continue
		}

		d, err := s.store.FindDeviceById(bridgeInfo.IeeeAddress)
		if err != nil {
			// not found in repo, new it here
			d = devices.NewDevice(bridgeInfo.IeeeAddress)
			d.ConnectionType = "mqtt"

			// load exposes if any
			// load features if any
			// both using a single constructor function
			// then categorise to correct type expose or feature (possibly rename) using AccessMode

			// load exposes
			for _, expose := range bridgeInfo.Definition.Exposes {
				entity, err := devices.CreateEntityFromExpose(expose, nil)
				if err != nil {
					utils.LogDebugf("expose failed loading %s error %s", bridgeInfo.FriendlyName, err.Error())
					continue
				}
				d.Exposes[entity.Name] = entity
				d.Availability = devices.OfflineAvailability
			}

			// load features
			for _, expose := range bridgeInfo.Definition.Exposes {

				for _, feature := range expose.Features {
					entity, err := devices.CreateEntityFromExpose(feature, nil)
					if err != nil {
						utils.LogDebugf("feature failed loading %s error %s", bridgeInfo.FriendlyName, err.Error())
						continue
					}

					d.Exposes[entity.Name] = entity
					d.Availability = devices.OfflineAvailability
				}

			}

			if len(d.Exposes) == 0 {
				utils.LogInfof("device %s failed loading. error contains no valid exposed data", bridgeInfo.FriendlyName)
				continue
			}

			// d.Monitor(deviceAvailabilityTimeoutOverride, func(p interface{}) {
			// 	s.eventHub.Broadcast(ws.DeviceUpdated, p)
			// })
		}

		d.Description = bridgeInfo.Definition.Description
		d.FriendlyName = bridgeInfo.FriendlyName
		d.PowerSource = bridgeInfo.PowerSource

		// register device
		s.store.StoreDevice(d.FriendlyName, d)
	}
}
