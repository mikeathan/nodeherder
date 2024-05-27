package settings

type DeviceConfig struct {
	Id       string `json:"id"`
	Disabled bool   `json:"disabled"`
	History  bool   `json:"history"`
}

type Repository interface {
	Save(config *AppConfig) error
	Load() (*AppConfig, error)
	FindDeviceConfig(id string) (*DeviceConfig, error)
	SaveDeviceConfig(deviceConfig *DeviceConfig) error
	Close() error
}
