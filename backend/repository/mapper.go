package repository

import (
	"node-herder/models/devices"
	"node-herder/utils"
	"sync"
)

type DeviceIdMapper struct {
	mutex          sync.RWMutex
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
	registered, err := s.deviceRepo.AllDevices()
	if err != nil {
		utils.LogErrorf("Error loading devices for IdMapper: %s", err.Error())
		return
	}
	// clean up
	s.mutex.Lock()
	defer s.mutex.Unlock()
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
	// Provider identities survive bridge refresh and process restart; Zigbee
	// mappings take precedence if an invalid historical name collision exists.
	for _, device := range registered {
		if device.Provider != "" && device.FriendlyName != "" {
			if _, exists := s.deviceIdMapper[device.FriendlyName]; !exists {
				s.deviceIdMapper[device.FriendlyName] = device.Id
			}
		}
	}
}

func (a *DeviceIdMapper) ResolveFriendlyName(friendlyName string) string {

	// check if id already exists
	a.mutex.RLock()
	id, ok := a.deviceIdMapper[friendlyName]
	a.mutex.RUnlock()
	if ok {
		return id
	}

	hashed := utils.HashName(friendlyName)
	utils.LogDebugf("DeviceIdMapper.ResolveFriendlyName: no IEEE mapping for %q, using hash=%s", friendlyName, hashed)
	return hashed
}

func (s *DeviceIdMapper) UpdateId(friendlyName string, id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.deviceIdMapper[friendlyName] = id
}

// UpdateProviderId replaces stale names for a provider's stable ID after rename.
func (s *DeviceIdMapper) UpdateProviderId(friendlyName string, id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for name, existing := range s.deviceIdMapper {
		if existing == id && name != friendlyName {
			delete(s.deviceIdMapper, name)
		}
	}
	s.deviceIdMapper[friendlyName] = id
}
