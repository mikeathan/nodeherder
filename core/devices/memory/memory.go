package memory

import (
	"errors"
	"node-herder/devices"
	"sort"
	"sync"
)

type MemoryNodeRepository struct {
	store map[string]*devices.Device
	mutex sync.RWMutex
}

func NewMemoryNodeRepository() devices.Repository {
	return &MemoryNodeRepository{store: map[string]*devices.Device{}, mutex: sync.RWMutex{}}
}

func (s *MemoryNodeRepository) Store(deviceName string, Device *devices.Device) {
	s.mutex.Lock()
	s.store[deviceName] = Device
	s.mutex.Unlock()
}

func (s *MemoryNodeRepository) FindDevice(deviceName string) (*devices.Device, error) {
	defer s.mutex.RUnlock()

	s.mutex.RLock()
	if val, ok := s.store[deviceName]; ok {
		return val, nil
	}

	return nil, errors.New("device not found")

}

func (s *MemoryNodeRepository) ListAllDevices() []*devices.Device {

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
