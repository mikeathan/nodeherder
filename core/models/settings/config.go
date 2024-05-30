package settings

import "time"

type DeviceConfig struct {
	Id             string `json:"id"`
	Disabled       bool   `json:"disabled"`
	MetricsEnabled bool   `json:"history"`
	RateLimit      int    `json:"rateLimit"`
}

func NewDeviceConfig(id string) *DeviceConfig {
	return &DeviceConfig{
		Id:             id,
		Disabled:       false,
		MetricsEnabled: false,
		RateLimit:      int(time.Minute.Milliseconds()), // default to 1 minute
	}
}

type AppConfig struct {
	Devices map[string]*DeviceConfig `json:"devices"`
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		Devices: map[string]*DeviceConfig{},
	}
}

func (s *AppConfig) Add(cfg *DeviceConfig) {
	s.Devices[cfg.Id] = cfg
}
