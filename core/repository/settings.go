package repository

import (
	"encoding/json"
	"errors"
	"node-herder/models/settings"
	"node-herder/utils/storage"
)

const settingsBaseFilename = "settings.db"
const settingsBucketName = "settings"
const settingsKeyName = "device_settings"
const bridgeSettingsKeyName = "bridge_settings"

type FileSettingsRepo struct {
	persistantStorage storage.KeyValueDatabase
	memoryStorage     *MemoryRepo[*settings.BridgeConfig]
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
		memoryStorage:     NewMemoryRepo[*settings.BridgeConfig](),
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
	config, err := s.memoryStorage.Find(bridgeSettingsKeyName)
	if err != nil {
		// if not found, create default config
		config := settings.DefaultBridgeConfig()
		return config, nil
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

func (s *FileSettingsRepo) LoadOrDefaultDeviceConfig(id string) (*settings.DeviceConfig, error) {

	appConfig, err := s.Load()
	if err != nil {
		return nil, err
	}

	// return override if exists
	cfg := appConfig.Hub.Devices.Overrides[id]
	if cfg != nil {
		return cfg, nil

	}

	// return default device config
	return appConfig.Hub.Devices.Defaults, nil
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

	config.Hub.Devices.AddOverride(deviceConfig)
	return s.SaveAppConfig(config)
}

func (s *FileSettingsRepo) DeleteDeviceConfig(id string) error {
	config, err := s.Load()
	if err != nil {
		return err
	}

	if ok := config.Hub.Devices.DeleteOverride(id); !ok {
		return errors.New("device config not found")
	}

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

	bridgeCfg, err := s.LoadBridgeConfig()
	if err == nil {
		settings.Bridge = bridgeCfg
	}

	return settings, err
}
