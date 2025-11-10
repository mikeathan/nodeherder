package repository

import (
	"encoding/json"
	"fmt"
	"node-herder/models/devices"
	"node-herder/utils"
	"path/filepath"
	"sync"

	"github.com/boltdb/bolt"
)

const deviceBaseFilename = "devices.db"
const devicesBucketName = "devices"
const bridgeBucketName = "bridge"
const bridgeKeyName = "bridgeInfo"

// TODO: use kv database
type FileDeviceRepo struct {
	mutex   sync.RWMutex
	db      *bolt.DB
	updated map[string]bool
}

func NewFileDeviceRepo() (devices.Repository, error) {
	return NewFileDeviceRepoFromFile(filepath.Join("data", deviceBaseFilename))
}

func NewFileDeviceRepoFromFile(filename string) (devices.Repository, error) {

	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		return nil, err
	}
	repo := &FileDeviceRepo{
		mutex:   sync.RWMutex{},
		db:      db,
		updated: map[string]bool{},
	}
	err = repo.init()
	if err != nil {
		utils.LogError(err)
		return nil, err
	}

	return repo, nil
}

func (s *FileDeviceRepo) Remove(key string) error {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.db.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(devicesBucketName))
		if b == nil {
			return bolt.ErrBucketNotFound
		}

		err := b.Delete([]byte(key))
		if err != nil {
			return fmt.Errorf("delete key: %w", err)
		}
		return nil
	})
}

func (s *FileDeviceRepo) init() error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.CreateBucketIfNotExists([]byte(devicesBucketName)); err != nil {
		return err
	}
	if _, err := tx.CreateBucketIfNotExists([]byte(bridgeBucketName)); err != nil {
		return err
	}

	return tx.Commit()
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

func (s *FileDeviceRepo) Store(key string, device *devices.Device) (bool, error) {

	s.mutex.Lock()
	defer s.mutex.Unlock()

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

	ok := s.updated[key]
	if err == nil {
		s.updated[key] = true
	}
	return !ok, err
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

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.findDevice(id)
}

func (s *FileDeviceRepo) FindDevices(ids []string) ([]*devices.Device, error) {

	s.mutex.RLock()
	defer s.mutex.RUnlock()

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
