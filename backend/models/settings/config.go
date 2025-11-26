package settings

import (
	"node-herder/models/bridge"
	"node-herder/utils"
	"time"
)

type DeviceSettings struct {
	Defaults  *DeviceConfig            `json:"defaults"`
	Overrides map[string]*DeviceConfig `json:"overrides"`
}

func (d *DeviceSettings) GetEffectiveConfig(deviceId string) *DeviceConfig {
	if override, exists := d.Overrides[deviceId]; exists {
		return override
	}
	return d.Defaults
}

func (d *DeviceSettings) GetDebounceForEntity(deviceId, entityName string, entityCategory bridge.ExposeCategory) *utils.TimeInterval {
	// First check if device has entity specific override
	if deviceOverride, exists := d.Overrides[deviceId]; exists {
		if debounce, exists := deviceOverride.DebounceOverrides[entityName]; exists {
			return debounce
		}
	}

	//  Fall back to category default from base config
	if debounce, exists := d.Defaults.DefaultDebounceByCategory[entityCategory]; exists {
		return debounce
	}

	return nil
}

func (d *DeviceSettings) IsEntityDebounced(deviceId, entityName string, entityCategory bridge.ExposeCategory) bool {
	return d.GetDebounceForEntity(deviceId, entityName, entityCategory) != nil
}

func (d *DeviceSettings) AddOverride(deviceConfig *DeviceConfig) {
	if deviceConfig.Id == "" {
		return
	}

	if deviceConfig.DebounceOverrides == nil {
		deviceConfig.DebounceOverrides = make(map[string]*utils.TimeInterval)
	}

	d.Overrides[deviceConfig.Id] = deviceConfig
}

func (d *DeviceSettings) DeleteOverride(deviceId string) bool {
	if _, ok := d.Overrides[deviceId]; !ok {
		return false
	}

	delete(d.Overrides, deviceId)
	return true
}

func NewDeviceSettings() *DeviceSettings {
	return &DeviceSettings{
		Defaults:  DefaultDeviceConfig(),
		Overrides: map[string]*DeviceConfig{},
	}
}

func DefaultDeviceConfig() *DeviceConfig {
	return &DeviceConfig{
		Disabled:                  false,
		MetricsEnabled:            false,
		RateLimit:                 utils.IntervalFromSeconds(60),
		DefaultDebounceByCategory: map[bridge.ExposeCategory]*utils.TimeInterval{bridge.DiagnosticCategory: utils.IntervalFromSeconds(300)},
		DebounceOverrides:         nil,
	}
}

type DeviceConfig struct {
	Id                        string                                        `json:"id,omitempty"`
	Disabled                  bool                                          `json:"disabled"`
	MetricsEnabled            bool                                          `json:"metricsEnabled"`
	RateLimit                 *utils.TimeInterval                           `json:"rateLimit"`
	DefaultDebounceByCategory map[bridge.ExposeCategory]*utils.TimeInterval `json:"defaultDebounceByCategory,omitempty"`
	DebounceOverrides         map[string]*utils.TimeInterval                `json:"debounceOverrides,omitempty"`
}

func (d *DeviceConfig) RateLimitDuration() time.Duration {
	if d.RateLimit == nil {
		return 0
	}
	return d.RateLimit.Duration()
}

func NewDeviceConfig(id string) *DeviceConfig {
	return &DeviceConfig{
		Id:                        id,
		Disabled:                  false,
		MetricsEnabled:            false,
		RateLimit:                 utils.IntervalFromSeconds(60), // default to 60 seconds
		DebounceOverrides:         map[string]*utils.TimeInterval{},
		DefaultDebounceByCategory: nil,
	}
}

func NewDeviceConfigFrom(config *DeviceConfig) *DeviceConfig {
	return &DeviceConfig{
		Id:                        config.Id,
		Disabled:                  config.Disabled,
		MetricsEnabled:            config.MetricsEnabled,
		RateLimit:                 config.RateLimit,
		DebounceOverrides:         config.DebounceOverrides,
		DefaultDebounceByCategory: config.DefaultDebounceByCategory,
	}
}

type HistoryConfig struct {
	SleepTimeout *utils.TimeInterval `json:"sleepTimeout"`
	ExpireAt     *utils.TimeInterval `json:"expireAt"`
}

func NewHistoryConfig(sleepTimeout *utils.TimeInterval, expireAt *utils.TimeInterval) *HistoryConfig {
	return &HistoryConfig{
		SleepTimeout: sleepTimeout,
		ExpireAt:     expireAt,
	}
}

func DefaultHistoryConfig() *HistoryConfig {
	return &HistoryConfig{
		SleepTimeout: utils.IntervalFromHours(12),
		ExpireAt:     utils.IntervalFromDays(10),
	}
}

type LoggerConfig struct {
	EnableRemoteLogger bool `json:"enableRemoteLogger"`
}

func NewLoggerConfig(enableRemoteLogger bool) *LoggerConfig {
	return &LoggerConfig{
		EnableRemoteLogger: enableRemoteLogger,
	}
}

func DefaultLoggingConfig() *LoggerConfig {
	return &LoggerConfig{
		EnableRemoteLogger: false,
	}
}

func DefaultBridgeConfig() *BridgeConfig {
	return &BridgeConfig{
		TimeExpireAt: utils.IntervalFromSeconds(120),
		PermitJoin:   false,
	}
}

type HubConfig struct {
	Devices         *DeviceSettings            `json:"devices"`
	History         *HistoryConfig             `json:"history"`
	Logger          *LoggerConfig              `json:"logger"`
	DashboardGroups map[string]*DashboardGroup `json:"dashboardGroups"`
}

type BridgeConfig struct {
	TimeExpireAt *utils.TimeInterval `json:"maxTimeAllowed"`
	PermitJoin   bool                `json:"permitJoin"`
}

func NewBridgeConfig() *BridgeConfig {
	return &BridgeConfig{
		TimeExpireAt: utils.IntervalFromSeconds(120),
		PermitJoin:   false,
	}
}

func NewHubConfig() *HubConfig {
	return &HubConfig{
		Devices:         NewDeviceSettings(),
		History:         DefaultHistoryConfig(),
		Logger:          DefaultLoggingConfig(),
		DashboardGroups: map[string]*DashboardGroup{},
	}
}

type AppConfig struct {
	Hub    *HubConfig    `json:"hub"`
	Bridge *BridgeConfig `json:"bridge"`
}

func (s *AppConfig) AddDeviceConfig(cfg *DeviceConfig) {
	s.Hub.Devices.AddOverride(cfg)
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		Hub: &HubConfig{
			Devices:         NewDeviceSettings(),
			History:         DefaultHistoryConfig(),
			Logger:          DefaultLoggingConfig(),
			DashboardGroups: map[string]*DashboardGroup{},
		},
		Bridge: DefaultBridgeConfig(),
	}
}
