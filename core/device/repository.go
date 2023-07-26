package device

import (
	"errors"
	"sort"
	"sync"
)

type MemoryNodeRepository struct {
	store map[string]*Payload
	mutex sync.RWMutex
}

func NewMemoryNodeRepository() Repository {
	return &MemoryNodeRepository{store: map[string]*Payload{}, mutex: sync.RWMutex{}}
}

func (s *MemoryNodeRepository) Store(deviceName string, payload *Payload) {
	s.mutex.Lock()
	s.store[deviceName] = payload
	s.mutex.Unlock()
}

func (s *MemoryNodeRepository) FindDevice(deviceName string) (*Payload, error) {
	defer s.mutex.RUnlock()

	s.mutex.RLock()
	if val, ok := s.store[deviceName]; ok {
		return val, nil
	}

	return nil, errors.New("device not found")

}

func (s *MemoryNodeRepository) ListAllDevices() []*Payload {

	// sort before returning values
	s.mutex.RLock()
	keys := make([]string, 0, len(s.store))
	for k := range s.store {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	devices := make([]*Payload, 0, len(s.store))
	for _, key := range keys {
		device := s.store[key]
		devices = append(devices, device)
	}
	s.mutex.RUnlock()

	return devices
}
