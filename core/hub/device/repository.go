package device

import (
	"errors"
	"sort"
)

type Repository interface {
	Store(deviceName string, payload *NodePayload)
	ListAllDevices() []*NodePayload
	FindDevice(deviceName string) (*NodePayload, error)
}

type MemoryNodeRepository struct {
	store map[string]*NodePayload
}

func NewMemoryNodeRepository() Repository {
	return &MemoryNodeRepository{store: map[string]*NodePayload{}}
}

func (s *MemoryNodeRepository) Store(deviceName string, payload *NodePayload) {
	s.store[deviceName] = payload
}

func (s *MemoryNodeRepository) FindDevice(deviceName string) (*NodePayload, error) {

	if val, ok := s.store[deviceName]; ok {
		return val, nil
	}

	return nil, errors.New("device not found")

}

func (s *MemoryNodeRepository) ListAllDevices() []*NodePayload {

	// sort before returning values
	keys := make([]string, 0, len(s.store))
	for k := range s.store {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	devices := make([]*NodePayload, 0, len(s.store))
	for _, key := range keys {
		device := s.store[key]
		devices = append(devices, device)
	}
	return devices
}
