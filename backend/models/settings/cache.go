package settings

import (
	"fmt"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/utils"
	"sync"
	"time"
)

type DeviceConfigUpdateListener func(cfg *DeviceConfig)

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

func (d *DeviceDebouncer) DebounceExpose(expose *devices.Entity) bool {
	if expose == nil {
		return false
	}

	if expose.Type == "binary" || expose.Type == "enum" {
		return false
	}

	duration, ok := d.configCache.GetDebounce(d.id, expose.Name, expose.Category)
	if !ok {
		// no debounce time set, so don't debounce
		return false
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()

	now := d.clock.Now()
	if lastEvent, ok := d.debounceMap[expose.Name]; ok {
		if now.Sub(lastEvent) < duration {
			return true
		}
	}

	// update the last event time
	d.debounceMap[expose.Name] = now
	return false
}

type DeviceConfigCache struct {
	appConfig *AppConfig
	mutex     *sync.RWMutex
	store     Repository
}

func NewDeviceConfigCache(store Repository, appconfig *AppConfig, mutex *sync.RWMutex) *DeviceConfigCache {
	return &DeviceConfigCache{
		store:     store,
		appConfig: appconfig,
		mutex:     mutex,
	}
}

func (d *DeviceConfigCache) Size() int {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return len(d.appConfig.Hub.Devices.Overrides)
}

func (d *DeviceConfigCache) IsMetricsEnabled(deviceId string) bool {
	config, err := d.Get(deviceId)
	if err != nil {
		return false
	}
	return config.MetricsEnabled
}

func (d *DeviceConfigCache) IsDeviceDisabled(deviceId string) bool {
	config, err := d.Get(deviceId)
	if err != nil {
		return false
	}
	return config.Disabled
}

func (d *DeviceConfigCache) Get(id string) (*DeviceConfig, error) {

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if deviceConfig, ok := d.appConfig.Hub.Devices.Overrides[id]; ok {
		return deviceConfig, nil
	}

	// Not in overrides, return default derived from AppConfig
	defaultConfig := NewDeviceConfigFrom(d.appConfig.Hub.Devices.Defaults)
	defaultConfig.Id = id
	return defaultConfig, nil
}

func (d *DeviceConfigCache) Set(deviceConfig *DeviceConfig) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.appConfig.Hub.Devices.AddOverride(deviceConfig)
	return d.store.SaveDeviceConfig(deviceConfig)
}

func (d *DeviceConfigCache) Delete(id string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	err := d.store.DeleteDeviceConfig(id)
	if err != nil {
		return err
	}

	d.appConfig.Hub.Devices.DeleteOverride(id)
	return nil
}

func (d *DeviceConfigCache) DeleteDebounce(id string, exposeName string) (bool, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	config, exists := d.appConfig.Hub.Devices.Overrides[id]
	if !exists {
		return false, nil
	}
	delete(config.DebounceOverrides, exposeName)
	if err := d.store.SaveDeviceConfig(config); err != nil {
		return false, err
	}
	return true, nil
}

func (d *DeviceConfigCache) SetDebounce(id string, exposeName string, timeInterval *utils.TimeInterval) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	config, exists := d.appConfig.Hub.Devices.Overrides[id]
	if !exists {
		config = NewDeviceConfig(id)
		d.appConfig.Hub.Devices.AddOverride(config)
	}
	config.DebounceOverrides[exposeName] = timeInterval
	return d.store.SaveDeviceConfig(config)
}

func (d *DeviceConfigCache) GetDebounce(id string, exposeName string, category bridge.ExposeCategory) (time.Duration, bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	config, exists := d.appConfig.Hub.Devices.Overrides[id]
	if !exists {
		return 0, false
	}

	if interval, ok := config.DebounceOverrides[exposeName]; ok {
		return interval.Duration(), true
	}

	if defaultInterval, ok := d.appConfig.Hub.Devices.Defaults.DefaultDebounceByCategory[category]; ok {
		return defaultInterval.Duration(), true
	}

	return 0, false
}

type AppConfigCache struct {
	deviceCache           *DeviceConfigCache
	store                 Repository
	tasks                 []Task
	configUpdateListeners []DeviceConfigUpdateListener
	appConfig             *AppConfig
	mutex                 sync.RWMutex
}

func NewAppConfigCache(store Repository, task []Task) (*AppConfigCache, error) {

	config, err := store.Load()
	if err != nil {
		return nil, err
	}

	cache := &AppConfigCache{
		store:                 store,
		tasks:                 task,
		configUpdateListeners: []DeviceConfigUpdateListener{},
		appConfig:             config,
	}
	cache.deviceCache = NewDeviceConfigCache(store, config, &cache.mutex)

	cache.startTasks(config)
	return cache, nil
}

func (s *AppConfigCache) RegisterDeviceConfigUpdateListener(listener DeviceConfigUpdateListener) {
	s.configUpdateListeners = append(s.configUpdateListeners, listener)
}

func (s *AppConfigCache) GetDeviceConfigCache() *DeviceConfigCache {
	return s.deviceCache
}

func (s *AppConfigCache) LoadAppConfig() (*AppConfig, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.appConfig, nil
}

func (s *AppConfigCache) LoadBridgeConfig() (*BridgeConfig, error) {
	config, err := s.store.LoadBridgeConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (s *AppConfigCache) SaveBridgePermitJoin(enabled bool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Clone or update in place? AppConfig is a pointer, but we want to ensure consistency.
	// Since we lock, we can update in place and then save.
	s.appConfig.Bridge.PermitJoin = enabled

	// We also need to save the specific bridge config if there is a separate method?
	// The repo has SaveBridgeConfig. And SaveAppConfig implies saving everything?
	// Repository structure separates bridge config?
	// Looking at Repo methods: LoadBridgeConfig, SaveBridgeConfig.
	// But appConfig.Bridge comes from LoadBridgeConfig.
	// So we should update appConfig.Bridge AND save via repo.

	return s.store.SaveBridgeConfig(s.appConfig.Bridge)
}

func (s *AppConfigCache) SaveLoggerConfig(loggerConfig *LoggerConfig) (*AppConfig, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.appConfig.Hub.Logger = loggerConfig

	err := s.store.SaveAppConfig(s.appConfig)
	if err != nil {
		return nil, err
	}

	s.reloadTasks(s.appConfig)

	return s.appConfig, nil
}

func (s *AppConfigCache) LoadLoggerConfig() (*LoggerConfig, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.appConfig.Hub.Logger, nil
}

func (s *AppConfigCache) SaveHistoryConfig(historyConfig *HistoryConfig) (*AppConfig, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.appConfig.Hub.History = historyConfig

	err := s.store.SaveAppConfig(s.appConfig)
	if err != nil {
		return nil, err
	}

	s.reloadTasks(s.appConfig)

	return s.appConfig, nil
}

func (s *AppConfigCache) SaveMCPConfig(mcpConfig *MCPConfig) (*AppConfig, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.appConfig.Hub.MCP = mcpConfig

	err := s.store.SaveAppConfig(s.appConfig)
	if err != nil {
		return nil, err
	}

	s.reloadTasks(s.appConfig)

	return s.appConfig, nil
}

func (s *AppConfigCache) RenameDashboardGroup(oldName string, newName string) (*DashboardGroup, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	group, ok := s.appConfig.Hub.DashboardGroups[oldName]
	if !ok {
		return nil, fmt.Errorf("dashboard group %q not found", oldName)
	}

	delete(s.appConfig.Hub.DashboardGroups, oldName)
	group.Name = newName
	s.appConfig.Hub.DashboardGroups[newName] = group

	err := s.store.SaveAppConfig(s.appConfig)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (s *AppConfigCache) SaveDashboardGroup(exposeGroup *DashboardGroup) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.appConfig.Hub.DashboardGroups[exposeGroup.Name] = exposeGroup

	err := s.store.SaveAppConfig(s.appConfig)
	if err != nil {
		return err
	}

	return nil
}

func (s *AppConfigCache) DeleteDashboardGroup(name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.appConfig.Hub.DashboardGroups, name)

	err := s.store.SaveAppConfig(s.appConfig)
	if err != nil {
		return err
	}

	return nil
}

func (s *AppConfigCache) ImportDashboardGroups(groups map[string]*DashboardGroup) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// delete all
	for name, _ := range s.appConfig.Hub.DashboardGroups {
		if _, ok := groups[name]; !ok {
			delete(s.appConfig.Hub.DashboardGroups, name)
		}
	}

	// import new groups
	for name, group := range groups {
		s.appConfig.Hub.DashboardGroups[name] = group
	}

	err := s.store.SaveAppConfig(s.appConfig)
	if err != nil {
		return err
	}

	return nil
}

func (d *AppConfigCache) SetDeviceConfigOverrides(deviceConfig *DeviceConfig) error {
	// Simple delegation now
	err := d.deviceCache.Set(deviceConfig)
	if err != nil {
		return err
	}
	d.setDirty(deviceConfig)
	return nil
}

func (d *AppConfigCache) DeleteDeviceConfigOverrides(id string) error {
	// Simple delegation now
	return d.deviceCache.Delete(id)
}

func (s *AppConfigCache) setDirty(cfg *DeviceConfig) {
	for _, listener := range s.configUpdateListeners {
		listener(cfg)
	}
}

func (s *AppConfigCache) SetDeviceConfigDefaults(deviceDefaults *DeviceConfig) error {

	// We need to update appConfig too?
	// SetDeviceConfigDefaults loads, saves, and updates device cache.
	// We should update s.appConfig.Hub.Devices.Defaults as well.
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.appConfig.Hub.Devices.Defaults = deviceDefaults
	err := s.store.SaveAppConfig(s.appConfig)
	if err != nil {
		return err
	}

	s.setDirty(deviceDefaults)
	return nil
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
