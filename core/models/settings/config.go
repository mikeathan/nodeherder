package settings

import (
	"node-herder/utils"
	"time"
)

type BridgeConfig struct {
	MaxTimeAllowed *utils.TimeInterval `json:"maxTimeAllowed"`
	PermitJoin     bool                `json:"permitJoin"`
}

type DeviceConfig struct {
	Id             string `json:"id"`
	Disabled       bool   `json:"disabled"`
	MetricsEnabled bool   `json:"history"`
	RateLimit      int    `json:"rateLimit"`
}

func (d *DeviceConfig) RateLimitDuration() time.Duration {
	return time.Duration(d.RateLimit) * time.Millisecond
}

func NewDeviceConfig(id string) *DeviceConfig {
	return &DeviceConfig{
		Id:             id,
		Disabled:       false,
		MetricsEnabled: false,
		RateLimit:      int(time.Minute.Milliseconds()), // default to 1 minute
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
		MaxTimeAllowed: utils.IntervalFromSeconds(120),
		PermitJoin:     false,
	}
}

type AppConfig struct {
	Devices map[string]*DeviceConfig `json:"devices"`
	History *HistoryConfig           `json:"history"`
	Logger  *LoggerConfig            `json:"logger"`
	Bridge  *BridgeConfig            `json:"bridge"`
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		Devices: map[string]*DeviceConfig{},
		History: DefaultHistoryConfig(),
		Logger:  DefaultLoggingConfig(),
		Bridge:  DefaultBridgeConfig(),
	}
}

func (s *AppConfig) Add(cfg *DeviceConfig) {
	s.Devices[cfg.Id] = cfg
}
