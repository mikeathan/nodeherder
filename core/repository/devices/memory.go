package repository

import (
	"errors"
	"node-herder/models/devices"
	"sort"
	"sync"
)

type MemoryDeviceRepo struct {
	storeV2 map[string]*devices.DeviceV2
	mutex   sync.RWMutex
}

func NewMemoryDeviceRepo() devices.Repository {
	return &MemoryDeviceRepo{
		storeV2: map[string]*devices.DeviceV2{},
		mutex:   sync.RWMutex{},
	}
}

func (s *MemoryDeviceRepo) StoreV2(key string, device *devices.DeviceV2) {

	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.storeV2[key] = device
}

func (s *MemoryDeviceRepo) FindDeviceV2(id string) (*devices.DeviceV2, error) {

	defer s.mutex.RUnlock()

	s.mutex.RLock()
	if val, ok := s.storeV2[id]; ok {
		return val, nil
	}

	return nil, errors.New("device not found")
}

func (s *MemoryDeviceRepo) ListAllDevicesV2() []*devices.DeviceV2 {

	// sort before returning values
	s.mutex.RLock()
	defer s.mutex.RUnlock()

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

	return devices
}
