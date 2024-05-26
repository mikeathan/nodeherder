package repository

import (
	"encoding/json"
	"fmt"
	"node-herder/models/devices"
	"sync"

	"github.com/boltdb/bolt"
)

const deviceBaseFilename = "devices.db"
const devicesBucketName = "devices"
const bridgeBucketName = "bridge"
const bridgeKeyName = "bridgeInfo"

type FileDeviceRepo struct {
	mutex sync.RWMutex
	db    *bolt.DB
}

func NewFileDeviceRepo() (devices.Repository, error) {
	return NewFileDeviceRepoFromFile(deviceBaseFilename)
}

func NewFileDeviceRepoFromFile(filename string) (devices.Repository, error) {

	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &FileDeviceRepo{
		db:    db,
		mutex: sync.RWMutex{},
	}, nil
}

func (s *FileDeviceRepo) StoreBridge(brigeInfo []*devices.BridgeInfo) error {
	err := s.db.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucketIfNotExists([]byte(bridgeBucketName))
		if err != nil {
			return err
		}

		// save all bridge info in one key so we dont have to update/remove
		// devices when changed
		buf, err := json.Marshal(brigeInfo)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(bridgeKeyName), buf)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (s *FileDeviceRepo) Close() error {
	err := s.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *FileDeviceRepo) Store(key string, device *devices.Device) error {

	defer s.mutex.Unlock()
	s.mutex.Lock()

	err := s.db.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucketIfNotExists([]byte(devicesBucketName))
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
func (s *FileDeviceRepo) AllBridgeInfo() ([]*devices.BridgeInfo, error) {
	var bridgeInfo []*devices.BridgeInfo
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bridgeBucketName))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}

		buffer := bucket.Get([]byte(bridgeKeyName))
		if buffer == nil {
			return fmt.Errorf("key %v not found", bridgeKeyName)
		}

		err := json.Unmarshal(buffer, &bridgeInfo)
		if err != nil {
			return err
		}

		return nil
	})

	return bridgeInfo, err
}

func (s *FileDeviceRepo) FindBridgeInfo(id string) (*devices.BridgeInfo, error) {

	var bridgeInfo []*devices.BridgeInfo
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bridgeBucketName))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}

		buffer := bucket.Get([]byte(bridgeKeyName))
		if buffer == nil {
			return fmt.Errorf("key %v not found", bridgeKeyName)
		}

		err := json.Unmarshal(buffer, &bridgeInfo)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// just loop over here to find the bridge id, we do that so we dont keep state to
	// update/delete or insert new bridges. we might want to do that eventually

	for _, bridge := range bridgeInfo {
		if bridge.IeeeAddress == id {
			return bridge, nil
		}
	}
	return nil, fmt.Errorf("bridge %v not found", id)
}

func (s *FileDeviceRepo) FindDevice(id string) (*devices.Device, error) {

	defer s.mutex.RUnlock()

	s.mutex.RLock()
	return s.findDevice(id)
}

func (s *FileDeviceRepo) FindDevices(ids []string) ([]*devices.Device, error) {

	defer s.mutex.RUnlock()

	s.mutex.RLock()
	return s.findDevices(ids)
}

func (s *FileDeviceRepo) findAllDevices() ([]*devices.Device, error) {
	var deviceList []*devices.Device
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(devicesBucketName))
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
		bucket := tx.Bucket([]byte(devicesBucketName))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}

		for _, key := range keys {
			buffer := bucket.Get([]byte(key))
			if buffer == nil {
				return fmt.Errorf("key %v not found", key)
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
		bucket := tx.Bucket([]byte(devicesBucketName))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}

		buffer := bucket.Get([]byte(key))
		if buffer == nil {
			return fmt.Errorf("key %v not found", key)
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
