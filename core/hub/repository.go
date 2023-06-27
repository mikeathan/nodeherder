package hub

import (
	"encoding/json"
	"errors"
	"node-herder/models"
	"sort"
)

func newDevice(name string, payload interface{}) *models.Device {
	return &models.Device{Name: name, Payload: payload}
}

type Repository interface {
	Store(deviceName string, payload interface{})
	ListAllDevices() []*models.Device
}

func ToJson(data any) ([]byte, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, errors.New("failed to marshal payload")
	}

	return bytes, nil
}

type MemoryRepository struct {
	store map[string]*models.Device
}

func NewMemoryRepository() Repository {
	return &MemoryRepository{store: map[string]*models.Device{}}
}

func (s *MemoryRepository) Store(deviceName string, payload interface{}) {

	if _, ok := s.store[deviceName]; !ok {
		s.store[deviceName] = newDevice(deviceName, payload)
	}
	s.store[deviceName].Payload = payload
}

func (s *MemoryRepository) ListAllDevices() []*models.Device {

	// sort before returning values
	keys := make([]string, 0, len(s.store))
	for k := range s.store {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	devices := make([]*models.Device, 0, len(s.store))
	for _, key := range keys {
		devices = append(devices, s.store[key])
	}
	return devices
}
