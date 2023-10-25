package repository

import (
	"errors"
	"node-herder/models/devices"
	"sort"
	"sync"
)

type MemoryDeviceRepo struct {
	store map[string]*devices.DeviceV2
	mutex sync.RWMutex
}

func NewMemoryDeviceRepo() devices.Repository {
	return &MemoryDeviceRepo{
		store: map[string]*devices.DeviceV2{},
		mutex: sync.RWMutex{},
	}
}

func (s *MemoryDeviceRepo) StoreV2(key string, device *devices.DeviceV2) {

	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.store[key] = device
}

func (s *MemoryDeviceRepo) FindDeviceV2(id string) (*devices.DeviceV2, error) {

	defer s.mutex.RUnlock()

	s.mutex.RLock()
	if val, ok := s.store[id]; ok {
		return val, nil
	}

	return nil, errors.New("device not found")
}

func (s *MemoryDeviceRepo) ListAllDevicesV2() []*devices.DeviceV2 {

	// sort before returning values
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	keys := make([]string, 0, len(s.store))
	for k := range s.store {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	devices := make([]*devices.DeviceV2, 0, len(s.store))
	for _, key := range keys {
		device := s.store[key]
		devices = append(devices, device)
	}

	return devices
}
