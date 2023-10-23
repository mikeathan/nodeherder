package services

import (
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/utils"
	"strings"
)

type DeviceRegistrar struct {
	bridgeInfoList            []*devices.BridgeInfo
	idMapper                  map[string]string
	repo                      devices.Repository
	eventHub                  ws.EventHub
	deviceAvailabilityTimeout int
}

func NewHubRegisterService(repo devices.Repository, hub ws.EventHub, deviceAvailabilityTimeout int) *DeviceRegistrar {
	return &DeviceRegistrar{repo: repo, eventHub: hub, deviceAvailabilityTimeout: deviceAvailabilityTimeout}
}

func (s *DeviceRegistrar) Register(friendlyName string, device *devices.DeviceV2) {
	id := s.ResolveId(friendlyName)

	s.repo.StoreV2(id, device)

	s.idMapper[friendlyName] = id // store id in mapper for easy access

}

func (s *DeviceRegistrar) LookupByName(name string) (*devices.DeviceV2, error) {

	id := s.ResolveId(name)

	return s.repo.FindDeviceV2ById(id)
}

func (s *DeviceRegistrar) LookupById(id string) (*devices.DeviceV2, error) {

	return s.repo.FindDeviceV2ById(id)
}

func (s *DeviceRegistrar) configureIdMapper(bridgeInfoList []*devices.BridgeInfo) {

	// remove items from idMapper, that use to have a bridge info but dont exist in current bridge info list
	// but cant clean idmapper because it contains non bridge infor items

	// clean up
	for name, id := range s.idMapper {
		var found = false
		for _, device := range s.bridgeInfoList {

			if device.IeeeAddress == id {
				found = true
				break
			}
		}

		if !found {
			delete(s.idMapper, name)
		}
	}

	// setup
	for _, device := range s.bridgeInfoList {
		s.idMapper[device.FriendlyName] = device.IeeeAddress
	}
}

func (s *DeviceRegistrar) RegisterBridge(bridgeInfoList []*devices.BridgeInfo, deviceAvailabilityTimeoutOverride int) []*devices.DeviceV2 {

	s.bridgeInfoList = bridgeInfoList

	s.configureIdMapper(bridgeInfoList)

	allDevices := []*devices.DeviceV2{}
	for _, bridgeInfo := range bridgeInfoList {
		if !bridgeInfo.IsActive() {
			continue
		}

		d, err := s.repo.FindDeviceV2ById(bridgeInfo.IeeeAddress)
		if err != nil {
			// not found in repo, new it here
			d = devices.NewDeviceV2(bridgeInfo.IeeeAddress)
			d.ConnectionType = "mqtt"

			var entity *devices.Entity

			for _, expose := range bridgeInfo.Definition.Exposes {
				entity, err = devices.CreateFromExpose(expose)
				if err != nil {
					utils.LogDebugf("failed loading %s error %s", bridgeInfo.FriendlyName, err.Error())
					continue
				}

				if entity == nil {
					continue
				}

				// it shoud be coming from database
				// since it doesnt we dont have below info
				d.Exposes[entity.Name] = entity
				d.Properties["availability"] = "offline"
			}

			d.Monitor(deviceAvailabilityTimeoutOverride, func(p interface{}) {
				s.eventHub.Broadcast(ws.DevicePropertiesUpdated, p)
			})
		}

		d.Description = bridgeInfo.Definition.Description
		d.FriendlyName = bridgeInfo.FriendlyName
		d.PowerSource = bridgeInfo.PowerSource

		allDevices = append(allDevices, d)
	}

	return allDevices

}

func (a *DeviceRegistrar) ResolveId(name string) string {

	// check if id already exists
	if id, ok := a.idMapper[name]; ok {
		return id
	}

	// create new id from name
	newId := strings.ReplaceAll(name, " ", "_")
	return utils.Hash(newId)
}
