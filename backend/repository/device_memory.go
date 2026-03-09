package repository

import (
	"errors"
	"fmt"
	"node-herder/models/devices"
	"sort"
	"sync"
)

type MemoryDeviceRepo struct {
	store          map[string]*devices.Device
	mutex          sync.RWMutex
	bridgeInfoList []*devices.BridgeInfo
	updated        map[string]bool
}

func NewMemoryDeviceRepo() devices.Repository {
	return &MemoryDeviceRepo{
		store:          map[string]*devices.Device{},
		mutex:          sync.RWMutex{},
		bridgeInfoList: []*devices.BridgeInfo{},
		updated:        map[string]bool{},
	}
}

func (s *MemoryDeviceRepo) Close() error {
	return nil
}

func (s *MemoryDeviceRepo) StoreBridge(brigeInfo []*devices.BridgeInfo) error {
	s.bridgeInfoList = brigeInfo
	return nil
}

func (s *MemoryDeviceRepo) AllBridgeInfo() ([]*devices.BridgeInfo, error) {
	return s.bridgeInfoList, nil
}

func (s *MemoryDeviceRepo) FindBridgeInfo(key string) (*devices.BridgeInfo, error) {

	for _, device := range s.bridgeInfoList {
		if device.IeeeAddress == key {
			return device, nil
		}
	}
	return nil, fmt.Errorf("bridge id %v not found", key)
}

func (s *MemoryDeviceRepo) Store(key string, device *devices.Device) (bool, error) {

	defer s.mutex.Unlock()
	s.mutex.Lock()

	ok := s.updated[key]
	s.store[key] = device

	s.updated[key] = true
	return !ok, nil
}

func (s *MemoryDeviceRepo) Remove(key string) error {

	defer s.mutex.Unlock()
	s.mutex.Lock()

	delete(s.store, key)
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
	defer s.mutex.RUnlock()

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

	return devices, nil
}
