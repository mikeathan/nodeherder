package store

import (
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/metrics"
	"node-herder/models/settings"
	"node-herder/repository"
	"node-herder/utils"
	"sync"
	"time"
)

type rateLimiter struct {
	mutex     sync.Mutex
	lastWrite time.Time
	appConfig *settings.AppConfig
	store     map[string]time.Time // In-memory store for device IDs and last write times
}

func NewRateLimiter(appConfig *settings.AppConfig) *rateLimiter {
	return &rateLimiter{
		mutex:     sync.Mutex{},
		lastWrite: time.Time{},
		appConfig: appConfig,
		store:     map[string]time.Time{},
	}
}

func (rl *rateLimiter) AllowWrite(id string, rateLimit time.Duration) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	currentTime := time.Now()

	// Check if device exists in the in-memory store
	if lastWrite, ok := rl.store[id]; ok {
		if currentTime.Sub(lastWrite) < rateLimit {
			return false // Rate limit exceeded
		}
	}

	// Update lastWrite time and store in map
	rl.lastWrite = currentTime
	rl.store[id] = currentTime

	return true
}

type AppStore interface {
	StoreDevice(friendlyName string, device *devices.Device) error
	RemoveDeviceById(id string) error

	FindDeviceByFriendlyName(friendlyName string) (*devices.Device, error)
	FindDeviceById(id string) (*devices.Device, error)
	FindDeviceByIds(ids []string) ([]*devices.Device, error)
	AllDevices() ([]*devices.Device, error)

	AppConfig() *settings.AppConfigCache

	StoreBridgeInfoList(bridgeInfoList []*devices.BridgeInfo) error
	FindBridgeInfoByFriendlyName(friendlyName string) (*devices.BridgeInfo, error)
	FindBridgeInfoById(id string) (*devices.BridgeInfo, error)

	StoreMetrics(friendlyName string, data map[string]any) error
	ViewMetrics(device *devices.Device, from time.Time, to time.Time) (*metrics.DeviceMetricsResult, error)
	ResolveFriendlyName(friendlyName string) string
}

type appStore struct {
	metrics        metrics.Repository
	devices        devices.Repository
	config         *settings.AppConfigCache
	deviceIdMapper *repository.DeviceIdMapper
	rateLimiter    *rateLimiter
}

func NewAppStore(devices devices.Repository, metrics metrics.Repository, config *settings.AppConfigCache) (AppStore, error) {

	appconfig, err := config.LoadAppConfig()
	if err != nil {
		return nil, err
	}

	deviceConfigs := make(map[string]*settings.DeviceConfig)
	for _, dev := range appconfig.Hub.Devices.Config {
		deviceConfigs[dev.Id] = dev
	}

	app := &appStore{
		metrics:        metrics,
		devices:        devices,
		config:         config,
		deviceIdMapper: repository.NewDeviceIdMapper(devices),
		rateLimiter:    NewRateLimiter(nil),
	}

	return app, nil
}

func (s *appStore) ViewMetrics(device *devices.Device, from time.Time, to time.Time) (*metrics.DeviceMetricsResult, error) {
	return s.metrics.ViewDeviceTimeRange(device, from, to)
}

func (s *appStore) AppConfig() *settings.AppConfigCache {
	return s.config
}

func (s *appStore) StoreMetrics(friendlyName string, data map[string]any) error {
	id := s.ResolveFriendlyName(friendlyName)

	config, err := s.config.GetDeviceConfig(id)
	if err != nil {
		return nil
	}

	if config.MetricsEnabled &&
		s.rateLimiter.AllowWrite(id, config.RateLimitDuration()) {

		err := s.metrics.Store(id, data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *appStore) RemoveDeviceById(id string) error {
	return s.devices.Remove(id)
}

func (s *appStore) StoreDevice(friendlyName string, device *devices.Device) error {
	id := s.ResolveFriendlyName(friendlyName)

	isNew, err := s.devices.Store(id, device)
	if err != nil {
		return err
	}

	if isNew {
		err := s.initialiseDeviceConfig(device)
		if err != nil {
			return err
		}
	}

	s.deviceIdMapper.UpdateId(friendlyName, id)
	return nil
}

func (s *appStore) initialiseDeviceConfig(device *devices.Device) error {

	// make sure new device has a configuration if added for first time
	deviceConfig, err := s.config.GetDeviceConfig(device.Id)
	if err != nil {
		return err
	}
	// for diagnostic entities, set default debounce to 5 min
	for _, entity := range device.Exposes {
		if entity.Category == bridge.DiagnosticCategory {

			if _, ok := deviceConfig.Debounce[entity.Name]; !ok {
				deviceConfig.Debounce[entity.Name] = utils.IntervalFromSeconds(300)
				s.config.SetDeviceConfig(deviceConfig)
			}
		}
	}
	return nil
}

func (s *appStore) StoreBridgeInfoList(bridgeInfoList []*devices.BridgeInfo) error {
	err := s.devices.StoreBridge(bridgeInfoList)
	if err != nil {
		return err
	}

	s.deviceIdMapper.Configure()
	return nil
}

func (s *appStore) FindDeviceByFriendlyName(friendlyName string) (*devices.Device, error) {

	id := s.ResolveFriendlyName(friendlyName)

	return s.FindDeviceById(id)
}

func (s *appStore) FindDeviceById(id string) (*devices.Device, error) {
	return s.devices.FindDevice(id)
}

func (s *appStore) FindBridgeInfoByFriendlyName(friendlyName string) (*devices.BridgeInfo, error) {

	id := s.ResolveFriendlyName(friendlyName)
	return s.FindBridgeInfoById(id)
}

func (s *appStore) FindBridgeInfoById(id string) (*devices.BridgeInfo, error) {
	return s.devices.FindBridgeInfo(id)
}

func (s *appStore) FindDeviceByIds(ids []string) ([]*devices.Device, error) {
	return s.devices.FindDevices(ids)
}

func (s *appStore) AllDevices() ([]*devices.Device, error) {
	return s.devices.AllDevices()
}

func (a *appStore) ResolveFriendlyName(friendlyName string) string {
	return a.deviceIdMapper.ResolveFriendlyName(friendlyName)
}
