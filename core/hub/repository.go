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
	StoreJson(deviceName string, payload []byte) error
	StoreObject(deviceName string, payload interface{}) error
	ListAllDevices() []*models.Device
}

type MemoryRepository struct {
	store map[string]*models.Device
}

func NewMemoryRepository() Repository {
	return &MemoryRepository{store: map[string]*models.Device{}}
}

func (s *MemoryRepository) StoreJson(deviceName string, payload []byte) error {
	var data map[string]interface{}
	err := json.Unmarshal(payload, &data)
	if err != nil {
		return errors.New("invalid device data")
	}

	s.storePayload(deviceName, data)
	return nil
}

func (s *MemoryRepository) StoreObject(deviceName string, payload interface{}) error {
	if data, ok := payload.(map[string]interface{}); ok {
		s.storePayload(deviceName, data)
		return nil
	}
	return errors.New("invalid device data")
}

func (s *MemoryRepository) storePayload(deviceName string, payload map[string]interface{}) {

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
		device := s.store[key]
		devices = append(devices, device)
	}
	return devices
}
