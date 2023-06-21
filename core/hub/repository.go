package hub

import "sort"

type Device struct {
	Name    string      `json:"name"`
	Payload interface{} `json:"payload"`
}

func newDevice(name string, payload interface{}) *Device {
	return &Device{Name: name, Payload: payload}
}

type Repository interface {
	AddDevice(deviceName string, payload interface{})
	ListAllDevices() []*Device
}

type MemoryRepository struct {
	store map[string]*Device
}

func NewMemoryRepository() Repository {
	return &MemoryRepository{store: map[string]*Device{}}
}

func (s *MemoryRepository) AddDevice(deviceName string, payload interface{}) {

	if _, ok := s.store[deviceName]; !ok {
		s.store[deviceName] = newDevice(deviceName, payload)
	}
	s.store[deviceName].Payload = payload
}

func (s *MemoryRepository) ListAllDevices() []*Device {

	// sort before returning values
	keys := make([]string, 0, len(s.store))
	for k := range s.store {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	devices := make([]*Device, 0, len(s.store))
	for _, key := range keys {
		devices = append(devices, s.store[key])
	}
	return devices
}
