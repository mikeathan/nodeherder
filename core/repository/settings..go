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
const bridgeSettingsKeyName = "bridge_settings"

type FileSettingsRepo struct {
	persistantStorage storage.KeyValueDatabase
	memoryStorage     *MemoryRepo[settings.BridgeConfig]
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
		persistantStorage: kvdb,
		memoryStorage:     NewMemoryRepo[settings.BridgeConfig](),
	}, nil
}

func (s *FileSettingsRepo) Close() error {
	s.memoryStorage.Close()

	err := s.persistantStorage.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *FileSettingsRepo) SaveAppConfig(value *settings.AppConfig) error {

	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.persistantStorage.Set([]byte(settingsKeyName), buf)
}


func (s *FileSettingsRepo) LoadBridgeConfig() (*settings.BridgeConfig, error) {
		buffer, err := s.persistantStorage.Get([]byte(bridgeSettingsKeyName))
	if err != nil {
		return nil, err
	}
	config := settings.DefaultBridgeConfig()
	if buffer != nil {
		err = json.Unmarshal(buffer, &config)
		if err != nil {
			return nil, err
		}
	}
	return config, err
}

func (s *FileSettingsRepo) SaveHubConfig(hubConfig *settings.HubConfig) error {

	config, err := s.Load()
	if err != nil {
		return err
	}

	return s.SaveAppConfig(config)
}

func (s *FileSettingsRepo) SaveBridgeConfig(bridgeConfig *settings.BridgeConfig) error {
	_, err := s.memoryStorage.Store(bridgeSettingsKeyName, bridgeConfig)
	return err
}

func (s *FileSettingsRepo) FindOrAddDeviceConfigIfNotExists(id string) (*settings.DeviceConfig, error) {

	config, err := s.Load()
	if err != nil {
		return nil, err
	}

	if val, ok := config.Hub.Devices[id]; ok {
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
func (s *FileSettingsRepo) SaveHistoryConfig(historyConfig *settings.HistoryConfig) error {
	config, err := s.Load()
	if err != nil {
		return err
	}

	config.Hub.History = historyConfig
	return s.SaveAppConfig(config)
}
func (s *FileSettingsRepo) SaveLoggerConfig(loggerConfig *settings.LoggerConfig) error {
	config, err := s.Load()
	if err != nil {
		return err
	}

	config.Hub.Logger = loggerConfig
	return s.SaveAppConfig(config)
}
func (s *FileSettingsRepo) SaveDeviceConfig(deviceConfig *settings.DeviceConfig) error {

	config, err := s.Load()
	if err != nil {
		return err
	}

	config.Hub.Devices[deviceConfig.Id] = deviceConfig
	return s.SaveAppConfig(config)
}

func (s *FileSettingsRepo) Load() (*settings.AppConfig, error) {

	buffer, err := s.persistantStorage.Get([]byte(settingsKeyName))
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
