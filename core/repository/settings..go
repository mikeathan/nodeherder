package repository

import (
	"encoding/json"
	"fmt"
	"node-herder/models/settings"
	"node-herder/utils/storage"
)

const settingsBaseFilename = "settings.db"
const settingsBucketName = "settings"
const settingsKeyName = "device_settings"

type FileSettingsRepo struct {
	kvdb storage.KeyValueDatabase
}

func NewFileSettingsRepo() (settings.Repository, error) {
	return NewFileSettingsRepoFromFile(settingsBaseFilename)
}

func NewFileSettingsRepoFromFile(filename string) (settings.Repository, error) {

	kvdb, err := storage.NewBoltKeyValueDatabase(filename, settingsBucketName)
	if err != nil {
		return nil, err
	}

	return &FileSettingsRepo{
		kvdb: kvdb,
	}, nil
}

func (s *FileSettingsRepo) Close() error {
	err := s.kvdb.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *FileSettingsRepo) Save(value *settings.AppConfig) error {

	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.kvdb.Set([]byte(settingsKeyName), buf)
}

func (s *FileSettingsRepo) FindOrAddDeviceConfigIfNotExists(id string) (*settings.DeviceConfig, error) {

	config, err := s.Load()
	if err != nil {
		return nil, err
	}

	if val, ok := config.Devices[id]; ok {
		return val, nil
	}

	// if device config not found, create one with default values
	cfg := settings.NewDeviceConfig(id)
	err = s.SaveDeviceConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise new device config for id %v", id)
	}
	return cfg, nil
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

	buffer, err := s.kvdb.Get([]byte(settingsBucketName))
	if err != nil {
		return nil, err
	}

	settings := settings.NewAppConfig()

	if buffer != nil {
		err = json.Unmarshal(buffer, &settings)
		if err != nil {
			return nil, err
		}

	}
	return settings, err
}
