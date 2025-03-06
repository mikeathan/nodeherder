package settings

import "time"

type Cache[T any] interface {
	Get(key string) (T, bool)
	Set(value T)
	Delete(key string)
}

// DeviceConfigCache

type DeviceConfigCache struct {
	deviceConfigs  map[string]*DeviceConfig
	exposeDebounce map[string]time.Duration
}

func NewDeviceConfigCache(store Repository) (*DeviceConfigCache, error) {
	appconfig, err := store.Load()
	if err != nil {
		return nil, err
	}

	deviceConfigs := make(map[string]*DeviceConfig)
	for _, dev := range appconfig.Hub.Devices {
		deviceConfigs[dev.Id] = dev
	}

	exposeDebounce := map[string]time.Duration{}
	for _, deviceConfig := range deviceConfigs {
		for expose, debounce := range deviceConfig.Debounce {
			exposeDebounce[expose] = debounce.Duration()
		}
	}

	return &DeviceConfigCache{
		deviceConfigs:  deviceConfigs,
		exposeDebounce: exposeDebounce,
	}, nil
}

func (d *DeviceConfigCache) Get(id string) (*DeviceConfig, bool) {
	if deviceConfig, ok := d.deviceConfigs[id]; ok {
		return deviceConfig, true
	}

	return nil, false
}

func (d *DeviceConfigCache) Set(deviceConfig *DeviceConfig) {
	d.deviceConfigs[deviceConfig.Id] = deviceConfig

	for expose, debounce := range deviceConfig.Debounce {
		d.exposeDebounce[expose] = debounce.Duration()
	}
}

func (d *DeviceConfigCache) Delete(name string) {
	// not implemented
}

func (d *DeviceConfigCache) GetExposeDebounce(expose string) (time.Duration, bool) {
	debounce, ok := d.exposeDebounce[expose]
	return debounce, ok
}

// AppConfigCache

type AppConfigCache struct {
	deviceCache Cache[*DeviceConfig]
	store       Repository
}

func NewAppConfigCache(store Repository) (*AppConfigCache, error) {

	deviceCache, err := NewDeviceConfigCache(store)
	if err != nil {
		return nil, err
	}
	return &AppConfigCache{
		deviceCache: deviceCache,
		store:       store,
	}, nil
}

func (s *AppConfigCache) LoadAppConfig() (*AppConfig, error) {
	return s.store.Load()
}

func (s *AppConfigCache) LoadBridgeConfig() (*BridgeConfig, error) {
	config, err := s.store.LoadBridgeConfig()
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (s *AppConfigCache) SaveBridgePermitJoin(enabled bool) error {
	config, err := s.store.LoadBridgeConfig()
	if err != nil {
		return err
	}

	config.PermitJoin = enabled

	return s.store.SaveBridgeConfig(config)
}

func (s *AppConfigCache) SaveLoggerConfig(loggerConfig *LoggerConfig) (*AppConfig, error) {
	config, err := s.LoadAppConfig()
	if err != nil {
		return nil, err
	}

	config.Hub.Logger = loggerConfig

	err = s.store.SaveAppConfig(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (s *AppConfigCache) SaveHistoryConfig(historyConfig *HistoryConfig) (*AppConfig, error) {
	config, err := s.LoadAppConfig()
	if err != nil {
		return nil, err
	}

	config.Hub.History = historyConfig

	err = s.store.SaveAppConfig(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (d *AppConfigCache) SetDeviceConfig(deviceConfig *DeviceConfig) error {

	// store in cache first
	d.deviceCache.Set(deviceConfig)

	// store in db
	return d.store.SaveDeviceConfig(deviceConfig)
}

func (d *AppConfigCache) GetDeviceConfig(id string) (*DeviceConfig, error) {

	// check if device config is cached
	if deviceConfig, ok := d.deviceCache.Get(id); ok {
		return deviceConfig, nil
	}

	// load from db
	return d.store.FindOrAddDeviceConfigIfNotExists(id)
}
