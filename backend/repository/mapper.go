package repository

import (
	"node-herder/models/devices"
	"node-herder/utils"
)

type DeviceIdMapper struct {
	deviceRepo     devices.Repository
	deviceIdMapper map[string]string
}

func NewDeviceIdMapper(deviceRepo devices.Repository) *DeviceIdMapper {
	return &DeviceIdMapper{
		deviceRepo:     deviceRepo,
		deviceIdMapper: map[string]string{},
	}
}

func (s *DeviceIdMapper) Configure() {
	bridgeInfoList, err := s.deviceRepo.AllBridgeInfo()
	if err != nil {
		utils.LogErrorf("Error loading bridgeInfoList %s", err.Error())
		return
	}
	// clean up
	for name, id := range s.deviceIdMapper {
		var found = false
		for _, device := range bridgeInfoList {
			if device.IeeeAddress == id && device.FriendlyName == name {
				found = true
				break
			}
		}

		if !found {
			utils.LogDebugf("Removing %s from IdMapper", name)
			delete(s.deviceIdMapper, name)
		}
	}

	// setup
	for _, device := range bridgeInfoList {
		s.deviceIdMapper[device.FriendlyName] = device.IeeeAddress
	}
}

func (a *DeviceIdMapper) ResolveFriendlyName(friendlyName string) string {

	// check if id already exists
	if id, ok := a.deviceIdMapper[friendlyName]; ok {
		return id
	}

	return utils.HashName(friendlyName)
}

func (s *DeviceIdMapper) UpdateId(friendlyName string, id string) {
	s.deviceIdMapper[friendlyName] = id
}
