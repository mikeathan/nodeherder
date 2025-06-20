package settings

import (
	"node-herder/models/bridge"
	"node-herder/utils"
	"sync"
	"time"
)

// Debouncer
type DeviceDebouncer struct {
	clock       utils.Clock
	configCache *DeviceConfigCache
	id          string
	debounceMap map[string]time.Time
	mutex       sync.RWMutex
}

func NewDeviceDebouncer(deviceId string, configCache *DeviceConfigCache, clock utils.Clock) *DeviceDebouncer {
	return &DeviceDebouncer{
		id:          deviceId,
		clock:       clock,
		configCache: configCache,
		debounceMap: map[string]time.Time{},
		mutex:       sync.RWMutex{},
	}
}

func (d *DeviceDebouncer) DebounceExpose(exposeName string, category bridge.ExposeCategory) bool {
	duration, ok := d.configCache.GetDebounce(d.id, exposeName, category)
	if !ok {
		// no debounce time set, so don't debounce
		return false
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()

	now := d.clock.Now()
	if lastEvent, ok := d.debounceMap[exposeName]; ok {
		if now.Sub(lastEvent) < duration {
			return true
		}
	}

	// update the last event time
	d.debounceMap[exposeName] = now
	return false
}

type DeviceConfigCache struct {
	devicesConfigs            map[string]*DeviceConfig
	defaultDebounceByCategory map[bridge.ExposeCategory]*utils.TimeInterval
	mutex                     sync.RWMutex
	store                     Repository
}

func NewDeviceConfigCache(store Repository, appconfig *AppConfig) *DeviceConfigCache {
	deviceConfigs := make(map[string]*DeviceConfig)
	for _, dev := range appconfig.Hub.Devices.Overrides {
		deviceConfigs[dev.Id] = dev
	}

	return &DeviceConfigCache{
		store:                     store,
		devicesConfigs:            deviceConfigs,
		mutex:                     sync.RWMutex{},
		defaultDebounceByCategory: appconfig.Hub.Devices.Default.DefaultDebounceByCategory,
	}
}

func (d *DeviceConfigCache) Size() int {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return len(d.devicesConfigs)
}

func (d *DeviceConfigCache) IsMetricsEnabled(deviceId string) bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	config, err := d.Get(deviceId)
	if err != nil {
		return false
	}
	return config.MetricsEnabled
}

func (d *DeviceConfigCache) Get(id string) (*DeviceConfig, error) {

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if deviceConfig, ok := d.devicesConfigs[id]; ok {
		return deviceConfig, nil
	}

	// load from db
	config, err := d.store.LoadOrDefaultDeviceConfig(id)
	if err != nil {
		return nil, err
	}

	// store in cache
	d.devicesConfigs[id] = config
	d.mutex.Unlock()
	return config, nil
}

func (d *DeviceConfigCache) Set(deviceConfig *DeviceConfig) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.devicesConfigs[deviceConfig.Id] = deviceConfig
	return d.store.SaveDeviceConfig(deviceConfig)
}

func (d *DeviceConfigCache) DeleteDebounce(id string, exposeName string) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	config, exists := d.devicesConfigs[id]
	if !exists {
		return false
	}
	delete(config.DebounceOverrides, exposeName)
	return true
}

func (d *DeviceConfigCache) SetDebounce(id string, exposeName string, timeInterval *utils.TimeInterval) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	config, exists := d.devicesConfigs[id]
	if !exists {
		config = NewDeviceConfig(id)
		d.devicesConfigs[id] = config
	}
	config.DebounceOverrides[exposeName] = timeInterval
}

func (d *DeviceConfigCache) GetDebounce(id string, exposeName string, category bridge.ExposeCategory) (time.Duration, bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	config, exists := d.devicesConfigs[id]
	if !exists {
		return 0, false
	}

	if interval, ok := config.DebounceOverrides[exposeName]; ok {
		return interval.Duration(), true
	}

	if defaultInterval, ok := d.defaultDebounceByCategory[category]; ok {
		return defaultInterval.Duration(), true
	}

	return 0, false
}

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
		deviceCache: NewDeviceConfigCache(store, config),
		store:       store,
		tasks:       task,
	}

	cache.startTasks(config)
	return cache, nil
}

func (s *AppConfigCache) GetDeviceConfigCache() *DeviceConfigCache {
	return s.deviceCache
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

func (s *AppConfigCache) LoadLoggerConfig() (*LoggerConfig, error) {
	config, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	return config.Hub.Logger, nil
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

func (s *AppConfigCache) SaveDashboardGroup(exposeGroup *DashboardGroup) error {
	config, err := s.LoadAppConfig()
	if err != nil {
		return err
	}

	config.Hub.DashboardGroups[exposeGroup.Name] = exposeGroup

	err = s.store.SaveAppConfig(config)
	if err != nil {
		return err
	}

	return nil
}

func (s *AppConfigCache) DeleteDashboardGroup(name string) error {
	config, err := s.LoadAppConfig()
	if err != nil {
		return err
	}

	delete(config.Hub.DashboardGroups, name)

	err = s.store.SaveAppConfig(config)
	if err != nil {
		return err
	}

	return nil
}

func (d *AppConfigCache) SetDeviceConfig(deviceConfig *DeviceConfig) error {
	return d.deviceCache.Set(deviceConfig)
}

func (d *AppConfigCache) GetDeviceConfig(id string) (*DeviceConfig, error) {
	return d.deviceCache.Get(id)
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
