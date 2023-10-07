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
	store map[string]*devices.Device

	storeV2        map[string]*devices.DeviceV2
	mutex          sync.RWMutex
	bridgeInfoList []*devices.BridgeInfo
	idMapper       map[string]string
}

func NewMemoryDeviceRepo() devices.Repository {
	return &MemoryDeviceRepo{
		store:          map[string]*devices.Device{},
		storeV2:        map[string]*devices.DeviceV2{},
		mutex:          sync.RWMutex{},
		bridgeInfoList: []*devices.BridgeInfo{},
		idMapper:       map[string]string{},
	}
}

func (s *MemoryDeviceRepo) Store(deviceName string, Device *devices.Device) {
	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.store[deviceName] = Device
}

func (s *MemoryDeviceRepo) FindDevice(name string) (*devices.Device, error) {

	defer s.mutex.RUnlock()

	s.mutex.RLock()
	if val, ok := s.store[name]; ok {
		return val, nil
	}

	return nil, errors.New("device not found")
}

func (s *MemoryDeviceRepo) StoreV2(friendlyName string, device *devices.DeviceV2) {

	id := s.ResolveId(friendlyName)

	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.storeV2[id] = device

	s.idMapper[friendlyName] = id // store id in mapper for easy access
}

func (s *MemoryDeviceRepo) FindDeviceV2(friendlyName string) (*devices.DeviceV2, error) {

	id := s.ResolveId(friendlyName)

	defer s.mutex.RUnlock()

	s.mutex.RLock()
	if val, ok := s.storeV2[id]; ok {
		return val, nil
	}

	return nil, errors.New("device not found")
}

func (s *MemoryDeviceRepo) ListAllDevices() []*devices.Device {

	// sort before returning values
	s.mutex.RLock()
	keys := make([]string, 0, len(s.store))
	for k := range s.store {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	devices := make([]*devices.Device, 0, len(s.store))
	for _, key := range keys {
		device := s.store[key]
		devices = append(devices, device)
	}
	s.mutex.RUnlock()

	return devices
}
func (s *MemoryDeviceRepo) ListAllDevicesV2() []*devices.DeviceV2 {

	// sort before returning values
	s.mutex.RLock()
	keys := make([]string, 0, len(s.store))
	for k := range s.store {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	devices := make([]*devices.DeviceV2, 0, len(s.store))
	for _, key := range keys {
		device := s.storeV2[key]
		devices = append(devices, device)
	}
	s.mutex.RUnlock()

	return devices
}
func (s *MemoryDeviceRepo) RegisterBridge(bridgeInfoList []*devices.BridgeInfo) {

	s.bridgeInfoList = bridgeInfoList

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
