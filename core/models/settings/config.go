package settings

import (
	"node-herder/utils"
	"time"
)

type DeviceConfig struct {
	Id             string                         `json:"id"`
	Disabled       bool                           `json:"disabled"`
	MetricsEnabled bool                           `json:"history"`
	RateLimit      *utils.TimeInterval            `json:"rateLimit"`
	Debounce       map[string]*utils.TimeInterval `json:"debounce"`
}

func (d *DeviceConfig) RateLimitDuration() time.Duration {
	if d.RateLimit == nil {
		return 0
	}
	return d.RateLimit.Duration()
}

func NewDeviceConfig(id string) *DeviceConfig {
	return &DeviceConfig{
		Id:             id,
		Disabled:       false,
		MetricsEnabled: false,
		RateLimit:      utils.IntervalFromSeconds(60), // default to 60 seconds
		Debounce:       map[string]*utils.TimeInterval{},
	}
}

type HistoryConfig struct {
	SleepTimeout *utils.TimeInterval `json:"sleeptTmeout"`
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
	Devices        map[string]*DeviceConfig   `json:"devices"`
	History        *HistoryConfig             `json:"history"`
	Logger         *LoggerConfig              `json:"logger"`
	DashboardGroup map[string]*DashboardGroup `json:"dashboardGroup"`
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
		Devices:        map[string]*DeviceConfig{},
		History:        DefaultHistoryConfig(),
		Logger:         DefaultLoggingConfig(),
		DashboardGroup: map[string]*DashboardGroup{},
	}
}

type AppConfig struct {
	Hub    *HubConfig    `json:"hub"`
	Bridge *BridgeConfig `json:"bridge"`
}

func (s *AppConfig) AddDeviceConfig(cfg *DeviceConfig) {
	s.Hub.Devices[cfg.Id] = cfg
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		Hub: &HubConfig{
			Devices:        map[string]*DeviceConfig{},
			History:        DefaultHistoryConfig(),
			Logger:         DefaultLoggingConfig(),
			DashboardGroup: map[string]*DashboardGroup{},
		},
		Bridge: DefaultBridgeConfig(),
	}
}
