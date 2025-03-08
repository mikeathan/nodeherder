package settings

import (
	"node-herder/utils"
	"sync"
	"time"
)

type Cache[T any] interface {
	Get(key string) (T, bool)
	Set(value T)
	Delete(key string)
}

// DeviceConfigCache

type DeviceDebounce struct {
	exposeDebounce map[string]time.Duration
}

func NewDeviceDebounce(config *DeviceConfig) *DeviceDebounce {

	exposeDebounce := map[string]time.Duration{}
	for expose, debounce := range config.Debounce {
		exposeDebounce[expose] = debounce.Duration()
	}
	return &DeviceDebounce{
		exposeDebounce: exposeDebounce,
	}
}

func (d *DeviceDebounce) GetDebounce(expose string) (time.Duration, bool) {
	debounce, ok := d.exposeDebounce[expose]
	return debounce, ok
}

func (d *DeviceDebounce) SetDebounce(expose string, debounce time.Duration) {
	d.exposeDebounce[expose] = debounce
}

type DeviceConfigCache struct {
	devicesConfigs  map[string]*DeviceConfig
	devicesDebounce map[string]*DeviceDebounce
	mutex           sync.RWMutex
}

func NewDeviceConfigCache(appconfig *AppConfig) *DeviceConfigCache {
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

	deviceDebounce := map[string]*DeviceDebounce{}
	for _, deviceConfig := range deviceConfigs {
		deviceDebounce[deviceConfig.Id] = NewDeviceDebounce(deviceConfig)
	}

	return &DeviceConfigCache{
		devicesConfigs:  deviceConfigs,
		devicesDebounce: deviceDebounce,
	}
}

func (d *DeviceConfigCache) Get(id string) (*DeviceConfig, bool) {

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if deviceConfig, ok := d.devicesConfigs[id]; ok {
		return deviceConfig, true
	}

	return nil, false
}

func (d *DeviceConfigCache) Set(deviceConfig *DeviceConfig) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.devicesConfigs[deviceConfig.Id] = deviceConfig

	for expose, debounce := range deviceConfig.Debounce {
		d.devicesDebounce[deviceConfig.Id].SetDebounce(expose, debounce.Duration())
	}
}

func (d *DeviceConfigCache) Delete(id string) {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	delete(d.devicesConfigs, id)
	delete(d.devicesDebounce, id)
}

func (d *DeviceConfigCache) DeleteDebounce(id string, exposeName string) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if deviceDebounce, ok := d.devicesDebounce[id]; ok {
		delete(deviceDebounce.exposeDebounce, exposeName)
		return true
	}
	return false
}

func (d *DeviceConfigCache) GetDebounce(id string, exposeName string) (time.Duration, bool) {

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	deviceDebounce, ok := d.devicesDebounce[id]
	if !ok {
		return 0, false
	}

	return deviceDebounce.GetDebounce(exposeName)
}

// AppConfigCache

type AppConfigCache struct {
	deviceCache *DeviceConfigCache
	store       Repository
	tasks       []Task
}

func NewAppConfigCache(store Repository, task []Task) (*AppConfigCache, error) {

	config, err := store.Load()
	if err != nil {
		return nil, err
	}

	cache := &AppConfigCache{
		deviceCache: NewDeviceConfigCache(config),
		store:       store,
		tasks:       task,
	}

	cache.startTasks(config)
	return cache, nil
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

	s.reloadTasks(config)

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

	s.reloadTasks(config)

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

func (s *AppConfigCache) startTasks(config *AppConfig) {

	for _, task := range s.tasks {
		err := task.Start(config)
		if err != nil {
			utils.LogErrorf("Error starting task: %v\n", err)
		}
	}
}

func (s *AppConfigCache) reloadTasks(config *AppConfig) {

	for _, task := range s.tasks {
		err := task.Stop()
		if err != nil {
			utils.LogErrorf("Error stopping task: %v\n", err)
			continue
		}

		err = task.Start(config)
		if err != nil {
			utils.LogErrorf("Error starting task: %v\n", err)
		}
	}
}
