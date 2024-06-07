package repository

import (
	"encoding/json"
	"fmt"
	"node-herder/models/settings"
	"node-herder/utils"
	"sync"

	"github.com/boltdb/bolt"
)

const settingsBaseFilename = "settings.db"
const settingsBucketName = "settings"
const settingsKeyName = "device_settings"

type FileSettingsRepo struct {
	mutex sync.RWMutex
	db    *bolt.DB
}

func NewFileSettingsRepo() (settings.Repository, error) {
	return NewFileSettingsRepoFromFile(settingsBaseFilename)
}

func NewFileSettingsRepoFromFile(filename string) (settings.Repository, error) {

	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		return nil, err
	}
	repo := &FileSettingsRepo{
		db:    db,
		mutex: sync.RWMutex{},
	}
	err = repo.init()
	if err != nil {
		utils.LogError(err)
		return nil, err
	}

	return repo, nil
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

func (s *FileSettingsRepo) init() error {
	tx, err := s.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.CreateBucketIfNotExists([]byte(settingsBucketName)); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *FileSettingsRepo) Load() (*settings.AppConfig, error) {

	settings := settings.NewAppConfig()
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(settingsBucketName))
		if bucket == nil {
			return bolt.ErrBucketNotFound
		}

		buffer := bucket.Get([]byte(settingsKeyName))
		if buffer == nil {
			return nil
			//return fmt.Errorf("key %v not found", settings)
		}

		err := json.Unmarshal(buffer, &settings)
		if err != nil {
			return err
		}

		return nil
	})

	return settings, err
}
