package repository

import (
	"errors"
	"node-herder/models/devices"
	"node-herder/utils"
	"sort"
	"sync"

	"github.com/boltdb/bolt"
)

const baseFilename = "devices.db"

type FileDeviceRepo struct {
	mutex sync.RWMutex
	db    *bolt.DB
}

func NewFileDeviceRepo() (devices.Repository, error) {
	return NewFileDeviceRepoFromFile(baseFilename)
}

func NewFileDeviceRepoFromFile(filename string) (devices.Repository, error) {

	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		utils.LogError(err)
		return nil, err
	}
	return &FileDeviceRepo{
		db:    db,
		mutex: sync.RWMutex{},
	}, nil
}

func (s *FileDeviceRepo) StoreBridge(key string, brigeInfo []*devices.BridgeInfo) {

}

func (s *FileDeviceRepo) Store(key string, device *devices.Device) {

	defer s.mutex.Unlock()
	s.mutex.Lock()
	s.store[key] = device
}

func (s *FileDeviceRepo) FindDevice(id string) (*devices.Device, error) {

	defer s.mutex.RUnlock()

	s.mutex.RLock()
	if val, ok := s.store[id]; ok {
		return val, nil
	}

	return nil, errors.New("device not found")
}

func (s *FileDeviceRepo) FindDevices(ids []string) []*devices.Device {
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

func (s *FileDeviceRepo) AllDevices() []*devices.Device {

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
