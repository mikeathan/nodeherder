package repository

import (
	"errors"
	"node-herder/models/devices"
	"sort"
	"sync"
)

type MetricsRepo struct {
	store map[string]*devices.Device
	mutex sync.RWMutex
}

func NewMetricsRepo() devices.Repository {
	return &MetricsRepo{
		store: map[string]*devices.Device{},
		mutex: sync.RWMutex{},
	}
}

func (s *MetricsRepo) Store(key string, device *devices.Device) {

	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.store[key] = device
}

func (s *MetricsRepo) FindDevice(id string) (*devices.Device, error) {

	defer s.mutex.RUnlock()

	s.mutex.RLock()
	if val, ok := s.store[id]; ok {
		return val, nil
	}

	return nil, errors.New("device not found")
}

func (s *MetricsRepo) FindDevices(ids []string) []*devices.Device {
	ds := []*devices.Device{}

	defer s.mutex.RUnlock()

	s.mutex.RLock()
	for _, id := range ids {
		if val, ok := s.store[id]; ok {
			ds = append(ds, val)
		}
	}

	return ds
}
func (s *MetricsRepo) AllDevices() []*devices.Device {

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

	return devices
}
