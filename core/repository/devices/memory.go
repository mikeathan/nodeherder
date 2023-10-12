package repository

import (
	"errors"
	"node-herder/models/devices"
	"node-herder/utils"
	"sort"
	"strings"
	"sync"
)

type MemoryDeviceRepo struct {
	storeV2        map[string]*devices.DeviceV2
	mutex          sync.RWMutex
	bridgeInfoList []*devices.BridgeInfo
	idMapper       map[string]string
}

func NewMemoryDeviceRepo() devices.Repository {
	return &MemoryDeviceRepo{
		storeV2:        map[string]*devices.DeviceV2{},
		mutex:          sync.RWMutex{},
		bridgeInfoList: []*devices.BridgeInfo{},
		idMapper:       map[string]string{},
	}
}

func (s *MemoryDeviceRepo) StoreV2(friendlyName string, device *devices.DeviceV2) {

	id := s.ResolveId(friendlyName)

	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.storeV2[id] = device

	s.idMapper[friendlyName] = id // store id in mapper for easy access
}

func (s *MemoryDeviceRepo) FindDeviceV2ById(id string) (*devices.DeviceV2, error) {
	defer s.mutex.RUnlock()

	s.mutex.RLock()
	if val, ok := s.storeV2[id]; ok {
		return val, nil
	}

	return nil, errors.New("device not found")
}

func (s *MemoryDeviceRepo) FindDeviceV2(friendlyName string) (*devices.DeviceV2, error) {

	id := s.ResolveId(friendlyName)

	return s.FindDeviceV2ById(id)
}

func (s *MemoryDeviceRepo) ListAllDevicesV2() []*devices.DeviceV2 {

	// sort before returning values
	s.mutex.RLock()
	keys := make([]string, 0, len(s.storeV2))
	for k := range s.storeV2 {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	devices := make([]*devices.DeviceV2, 0, len(s.storeV2))
	for _, key := range keys {
		device := s.storeV2[key]
		devices = append(devices, device)
	}
	s.mutex.RUnlock()

	return devices
}
func (s *MemoryDeviceRepo) RegisterBridge(bridgeInfoList []*devices.BridgeInfo) {

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

func (a *MemoryDeviceRepo) FindBridgeInfo(id string) *devices.BridgeInfo {
	for _, device := range a.bridgeInfoList {
		if device.IeeeAddress == id {
			return device
		}
	}
	return nil
}

func (a *MemoryDeviceRepo) ResolveId(name string) string {

	// check if id already exists
	if id, ok := a.idMapper[name]; ok {
		return id
	}

	// create new id from name
	newId := strings.ReplaceAll(name, " ", "_")
	return utils.Hash(newId)
}
