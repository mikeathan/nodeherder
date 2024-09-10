package settings

import "time"

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
	SleepTimeout time.Duration `json:"sleeptimeout"`
	ExpireAt     time.Duration `json:"expireat"`
}

func NewHistoryConfig(sleepTimeout time.Duration, expireAt time.Duration) *HistoryConfig {
	return &HistoryConfig{
		SleepTimeout: sleepTimeout,
		ExpireAt:     expireAt,
	}
}
func DefaultHistoryConfig() *HistoryConfig {
	return &HistoryConfig{
		SleepTimeout: time.Hour * 12,      // 12 hours timeout
		ExpireAt:     time.Hour * 24 * 10, // 10 days expiration
	}
}

type AppConfig struct {
	Devices map[string]*DeviceConfig `json:"devices"`
	History *HistoryConfig           `json:"history"`
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		Devices: map[string]*DeviceConfig{},
		History: DefaultHistoryConfig(),
	}
}

func (s *AppConfig) Add(cfg *DeviceConfig) {
	s.Devices[cfg.Id] = cfg
}
