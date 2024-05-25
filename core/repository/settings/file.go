package repository

import (
	"encoding/json"
	"fmt"
	"node-herder/models/settings"
	"sync"

	"github.com/boltdb/bolt"
)

const baseFilename = "settings.db"
const settingsBucketName = "settings"
const settingsKeyName = "device_settings"

type FileSettingsRepo struct {
	mutex sync.RWMutex
	db    *bolt.DB
}

func NewFileSettingsRepo() (settings.Repository, error) {
	return NewFileSettingsRepoFromFile(baseFilename)
}

func NewFileSettingsRepoFromFile(filename string) (settings.Repository, error) {

	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &FileSettingsRepo{
		db:    db,
		mutex: sync.RWMutex{},
	}, nil
}

func (s *FileSettingsRepo) Close() error {
	err := s.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *FileSettingsRepo) Save(value *settings.AppConfig) error {
	err := s.db.Update(func(tx *bolt.Tx) error {

		bucket, err := tx.CreateBucketIfNotExists([]byte(settingsBucketName))
		if err != nil {
			return err
		}

		buf, err := json.Marshal(value)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(settingsKeyName), buf)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (s *FileSettingsRepo) FindDeviceConfig(id string) (*settings.DeviceConfig, error) {

	config, err := s.Load()
	if err != nil {
		return nil, err
	}

	if val, ok := config.Devices[id]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("device config for id %v not found", id)
}

func (s *FileSettingsRepo) SaveDeviceConfig(deviceConfig *settings.DeviceConfig) error {

	config, err := s.Load()
	if err != nil {
		return err
	}

	config.Devices[deviceConfig.Id] = deviceConfig
	return s.Save(config)
}

func (s *FileSettingsRepo) Load() (*settings.AppConfig, error) {

	var settings *settings.AppConfig
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(settingsBucketName))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}

		buffer := bucket.Get([]byte(settingsKeyName))
		if buffer == nil {
			return fmt.Errorf("key %v not found", settings)
		}

		err := json.Unmarshal(buffer, &settings)
		if err != nil {
			return err
		}

		return nil
	})

	return settings, err
}
