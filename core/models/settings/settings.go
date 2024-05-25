package settings

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

type DeviceConfig struct {
	Id       string `json:"id"`
	Disabled bool   `json:"disabled"`
	History  bool   `json:"history"`
}
