package services

import (
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/utils"
)

type devicePayload struct {
	status string
	data   interface{}
}

func newDeviceAddedPayload(data interface{}) *devicePayload {
	return &devicePayload{status: ws.DeviceAdded, data: data}
}

type DeviceRegistrar struct {
	bridgeInfoList            []*devices.BridgeInfo
	idMapper                  map[string]string
	repo                      devices.Repository
	eventHub                  ws.EventHub
	deviceAvailabilityTimeout int
}

func NewHubRegisterService(repo devices.Repository, hub ws.EventHub, deviceAvailabilityTimeout int) *DeviceRegistrar {
	return &DeviceRegistrar{repo: repo, eventHub: hub, idMapper: make(map[string]string), deviceAvailabilityTimeout: deviceAvailabilityTimeout}
}

func (s *DeviceRegistrar) Register(friendlyName string, device *devices.Device) {
	id := s.ResolveId(friendlyName)

	s.repo.Store(id, device)

	s.idMapper[friendlyName] = id // store id in mapper for easy access
}

func (s *DeviceRegistrar) LookupByName(name string) (*devices.Device, error) {

	id := s.ResolveId(name)

	return s.repo.FindDevice(id)
}

func (s *DeviceRegistrar) LookupById(id string) (*devices.Device, error) {

	return s.repo.FindDevice(id)
}

func (s *DeviceRegistrar) CreateNewDevice(friendlyName string, connType string, data map[string]interface{}) (*devices.Device, error) {

	id := s.ResolveId(friendlyName)
	bridgeInfo := s.FindBridgeInfo(id)

	device, err := devices.CreateNewDevice(id, friendlyName, connType, bridgeInfo, data)
	if err != nil {
		return nil, err
	}

	device.Monitor(s.deviceAvailabilityTimeout, func(p interface{}) {
		s.eventHub.Broadcast(ws.DeviceUpdated, p)
	})

	return device, nil
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

func (a *DeviceRegistrar) FindBridgeInfo(id string) *devices.BridgeInfo {
	for _, device := range a.bridgeInfoList {
		if device.IeeeAddress == id {
			return device
		}
	}

	return nil
}

func (s *DeviceRegistrar) RegisterBridge(bridgeInfoList []*devices.BridgeInfo, deviceAvailabilityTimeoutOverride int) {

	s.bridgeInfoList = bridgeInfoList

	s.configureIdMapper(bridgeInfoList)

	for _, bridgeInfo := range bridgeInfoList {
		if !bridgeInfo.IsActive() {
			continue
		}

		d, err := s.repo.FindDevice(bridgeInfo.IeeeAddress)
		if err != nil {
			// not found in repo, new it here
			d = devices.NewDevice(bridgeInfo.IeeeAddress)
			d.ConnectionType = "mqtt"

			// load exposes
			for _, expose := range bridgeInfo.Definition.Exposes {
				entity, err := devices.CreateFromExpose(expose)
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
					entity, err := devices.CreateFromFeature(feature)
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

func (a *DeviceRegistrar) ResolveId(name string) string {

	// check if id already exists
	if id, ok := a.idMapper[name]; ok {
		return id
	}

	return utils.HashName(name)
}
