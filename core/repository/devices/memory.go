package repository

import (
	"errors"
	"node-herder/models/devices"
	"sort"
	"sync"
)

type MemoryDeviceRepo struct {
	store map[string]*devices.Device
	mutex sync.RWMutex
}

func NewMemoryDeviceRepo() devices.Repository {
	return &MemoryDeviceRepo{
		store: map[string]*devices.Device{},
		mutex: sync.RWMutex{},
	}
}

func (s *MemoryDeviceRepo) StoreBridge(brigeInfo []*devices.BridgeInfo) error {
	return nil
}
func (s *MemoryDeviceRepo) FindBridgeInfo(ids []string) ([]*devices.BridgeInfo, error) {
	return nil, nil
}

func (s *MemoryDeviceRepo) Store(key string, device *devices.Device) error {

	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.store[key] = device

	return nil
}

func (s *MemoryDeviceRepo) FindDevice(id string) (*devices.Device, error) {

	defer s.mutex.RUnlock()

	s.mutex.RLock()
	if val, ok := s.store[id]; ok {
		return val, nil
	}

	return nil, errors.New("device not found")
}

func (s *MemoryDeviceRepo) FindDevices(ids []string) ([]*devices.Device, error) {
	ds := []*devices.Device{}

	defer s.mutex.RUnlock()

	s.mutex.RLock()
	for _, id := range ids {
		if val, ok := s.store[id]; ok {
			ds = append(ds, val)
		}
	}

	return ds, nil
}
func (s *MemoryDeviceRepo) AllDevices() ([]*devices.Device, error) {

	// sort before returning values
	s.mutex.RLock()
	defer s.mutex.RUnlock()

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

	return devices, nil
}
