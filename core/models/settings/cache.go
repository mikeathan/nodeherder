package settings

import (
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

func (d *DeviceDebouncer) DebounceExpose(exposeName string) bool {
	duration, ok := d.configCache.GetDebounce(d.id, exposeName)
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

// cache
type Cache[T any] interface {
	Get(key string) (T, bool)
	Set(value T)
	Size() int
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
	store           Repository
}

func NewDeviceConfigCache(store Repository, appconfig *AppConfig) *DeviceConfigCache {
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
		store:           store,
		devicesConfigs:  deviceConfigs,
		devicesDebounce: deviceDebounce,
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
	config, err := d.store.FindOrAddDeviceConfigIfNotExists(id)
	if err != nil {
		return nil, err
	}

	// store in cache
	d.devicesConfigs[id] = config
	return config, nil
}

func (d *DeviceConfigCache) Set(deviceConfig *DeviceConfig) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.devicesConfigs[deviceConfig.Id] = deviceConfig

	for expose, debounce := range deviceConfig.Debounce {
		if _, ok := d.devicesDebounce[deviceConfig.Id]; !ok {
			d.devicesDebounce[deviceConfig.Id] = NewDeviceDebounce(deviceConfig)
			continue
		}

		d.devicesDebounce[deviceConfig.Id].SetDebounce(expose, debounce.Duration())
	}

	// store in db
	return d.store.SaveDeviceConfig(deviceConfig)
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

func (d *DeviceConfigCache) SetDebounce(id string, exposeName string, timeInterval *utils.TimeInterval) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if deviceDebounce, ok := d.devicesDebounce[id]; ok {
		deviceDebounce.SetDebounce(exposeName, timeInterval.Duration())
	} else {

		deviceDebounce = NewDeviceDebounce(NewDeviceConfig(id))
		deviceDebounce.SetDebounce(exposeName, timeInterval.Duration())
		d.devicesDebounce[id] = deviceDebounce
	}
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

func (s *AppConfigCache) SaveExposeGroup(name string, exposeGroup *ExposeGroup) error {
	config, err := s.LoadAppConfig()
	if err != nil {
		return err
	}

	config.Hub.Groups[name] = exposeGroup

	err = s.store.SaveAppConfig(config)
	if err != nil {
		return err
	}

	return nil
}

func (s *AppConfigCache) DeleteExposeGroup(name string) error {
	config, err := s.LoadAppConfig()
	if err != nil {
		return err
	}

	delete(config.Hub.Groups, name)

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
