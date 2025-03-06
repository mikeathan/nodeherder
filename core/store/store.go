package store

import (
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
	//UpdateDevice(friendlyName string, device *devices.Device) error

	FindDeviceByFriendlyName(friendlyName string) (*devices.Device, error)
	FindDeviceById(id string) (*devices.Device, error)
	FindDeviceByIds(ids []string) ([]*devices.Device, error)
	AllDevices() ([]*devices.Device, error)

	AppConfig() *settings.AppConfigCache
	// LoadAppConfig() (*settings.AppConfig, error)
	// LoadDeviceConfig(id string) (*settings.DeviceConfig, error)
	// SaveDeviceConfig(deviceconfig *settings.DeviceConfig) error
	// SaveHistoryConfig(historyConfig *settings.HistoryConfig) error
	// SaveLoggerConfig(loggerConfig *settings.LoggerConfig) error
	// SaveBridgePermitJoin(enabled bool) error
	// LoadBridgeConfig() (*settings.BridgeConfig, error)

	StoreBridgeInfoList(bridgeInfoList []*devices.BridgeInfo) error
	FindBridgeInfoByFriendlyName(friendlyName string) (*devices.BridgeInfo, error)
	FindBridgeInfoById(id string) (*devices.BridgeInfo, error)

	StoreMetrics(friendlyName string, data map[string]any) error
	ViewMetrics(device *devices.Device, from time.Time, to time.Time) (*metrics.DeviceMetricsResult, error)
	ResolveFriendlyName(friendlyName string) string

	AddTask(task Task)
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
	for _, dev := range appconfig.Hub.Devices {
		deviceConfigs[dev.Id] = dev
	}

	app := &appStore{
		metrics:        metrics,
		devices:        devices,
		config:         config,
		deviceIdMapper: repository.NewDeviceIdMapper(devices),
		rateLimiter:    NewRateLimiter(nil),
		tasks:          tasks,
	}

	app.startTasks(appconfig)
	return app, nil
}

func (s *appStore) AddTask(task Task) {
	s.tasks = append(s.tasks, task)
}

func (s *appStore) ViewMetrics(device *devices.Device, from time.Time, to time.Time) (*metrics.DeviceMetricsResult, error) {
	return s.metrics.ViewDeviceTimeRange(device, from, to)
}

func (s *appStore) AppConfig() *settings.AppConfigCache {
	return s.config
}

we need to do the task reload  inthe cache now or somewhere 
// func (s *appStore) LoadAppConfig() (*settings.AppConfig, error) {
// 	return s.config.LoadAppConfig()
// }

// func (s *appStore) SaveHistoryConfig(historyConfig *settings.HistoryConfig) error {

// 	app, err := s.config.SaveHistoryConfig(historyConfig)
// 	if err != nil {
// 		return err
// 	}

// 	s.reloadTasks(app)

// 	return nil
// }

// func (s *appStore) SaveLoggerConfig(loggerConfig *settings.LoggerConfig) error {
// 	app, err := s.config.SaveLoggerConfig(loggerConfig)
// 	if err != nil {
// 		return err
// 	}

// 	s.reloadTasks(app)

// 	return nil
// }

// func (s *appStore) LoadBridgeConfig() (*settings.BridgeConfig, error) {
// 	config, err := s.config.LoadBridgeConfig()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return config, nil
// }

// func (s *appStore) SaveBridgePermitJoin(enabled bool) error {
// 	return s.config.SaveBridgePermitJoin(enabled)
// }

// func (s *appStore) SaveDeviceConfig(deviceconfig *settings.DeviceConfig) error {
// 	return s.config.SetDeviceConfig(deviceconfig)
// }

// func (s *appStore) LoadDeviceConfig(id string) (*settings.DeviceConfig, error) {
// 	return s.config.GetDeviceConfig(id)
// }

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
		// make sure new device has a configuration if added for first time
		_, err := s.config.GetDeviceConfig(id)
		if err != nil {
			return err
		}
	}

	s.deviceIdMapper.UpdateId(friendlyName, id)
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
