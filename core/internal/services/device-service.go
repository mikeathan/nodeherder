package services

import (
	"node-herder/internal/ws"
	"node-herder/models/devices"
	"node-herder/utils"
)

func (s *) RegisterBridge(bridgeInfoList []*devices.BridgeInfo) {

	// remove items from idMapper, that use to have a bridge info but dont exist in current bridge info list
	// but cant clean idmapper because it contains non bridge infor items
	s.bridgeInfoList = bridgeInfoList

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

func BuildFromBridge(hub ws.EventHub, repo devices.Repository, bridgeDevices []*devices.BridgeInfo, deviceAvailabilityTimeoutOverride int) []*devices.DeviceV2 {

	allDevices := []*devices.DeviceV2{}
	for _, bridgeInfo := range bridgeDevices {
		if !bridgeInfo.IsActive() {
			continue
		}

		d, err := repo.FindDeviceV2ById(bridgeInfo.IeeeAddress)
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
				hub.Broadcast(ws.DevicePropertiesUpdated, p)
			})
		}

		d.Description = bridgeInfo.Definition.Description
		d.FriendlyName = bridgeInfo.FriendlyName
		d.PowerSource = bridgeInfo.PowerSource

		allDevices = append(allDevices, d)
	}

	return allDevices

}
