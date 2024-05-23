package repository

import (
	"encoding/json"
	"errors"
	"node-herder/models/devices"
	"node-herder/utils"
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

func (s *FileDeviceRepo) StoreBridge(brigeInfo []*devices.BridgeInfo) error {
	err := s.db.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucketIfNotExists([]byte("bridge"))
		if err != nil {
			return err
		}

		buf, err := json.Marshal(brigeInfo)
		if err != nil {
			return err
		}

		return bucket.Put([]byte("deviceInfo"), buf)
	})

	return err
}

func (s *FileDeviceRepo) Close() {
	err := s.db.Close()
	if err != nil {
		utils.LogError(err)
	}
}

func (s *FileDeviceRepo) Store(key string, device *devices.Device) error {

	defer s.mutex.Unlock()
	s.mutex.Lock()

	err := s.db.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucketIfNotExists([]byte("devices"))
		if err != nil {
			return err
		}

		buf, err := json.Marshal(device)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(key), buf)

	})

	return err
}

func (s *FileDeviceRepo) FindBridgeInfo(ids []string) ([]*devices.BridgeInfo, error) {

	var bridgeInfo []*devices.BridgeInfo
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("bridge"))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}

		buffer := bucket.Get([]byte("deviceInfo"))
		if buffer == nil {
			return errors.New("key not found")
		}

		err := json.Unmarshal(buffer, &bridgeInfo)
		if err != nil {
			return err
		}

		return nil
	})

	return bridgeInfo, err
}

func (s *FileDeviceRepo) FindDevice(id string) (*devices.Device, error) {

	defer s.mutex.RUnlock()

	s.mutex.RLock()
	return s.findDevice(id)
}

func (s *FileDeviceRepo) FindDevices(ids []string) ([]*devices.Device, error) {

	defer s.mutex.RUnlock()

	s.mutex.RLock()
	return s.FindDevices(ids)
}

func (s *FileDeviceRepo) findAllDevices() ([]*devices.Device, error) {
	var deviceList []*devices.Device
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("devices"))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}

		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var device *devices.Device
			err := json.Unmarshal(v, &device)
			if err != nil {
				return err
			}
			deviceList = append(deviceList, device)

		}

		return nil
	})

	return deviceList, err
}

func (s *FileDeviceRepo) findDevices(keys []string) ([]*devices.Device, error) {
	var deviceList []*devices.Device
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("devices"))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}

		for _, key := range keys {
			buffer := bucket.Get([]byte(key))
			if buffer != nil {
				return errors.New("key not found")
			}

			var device *devices.Device
			err := json.Unmarshal(buffer, &device)
			if err != nil {
				return err
			}

			deviceList = append(deviceList, device)
		}

		return nil
	})

	return deviceList, err
}

func (s *FileDeviceRepo) findDevice(key string) (*devices.Device, error) {
	var device *devices.Device
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("devices"))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}

		buffer := bucket.Get([]byte(key))
		if buffer == nil {
			return errors.New("key not found")
		}

		err := json.Unmarshal(buffer, &device)
		if err != nil {
			return err
		}

		return nil
	})

	return device, err
}

func (s *FileDeviceRepo) AllDevices() ([]*devices.Device, error) {

	// sort before returning values
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.findAllDevices()
}
