package settings

type Repository interface {
	Save(config *AppConfig) error
	Load() (*AppConfig, error)
	FindDeviceConfig(id string) (*DeviceConfig, error)
	SaveDeviceConfig(deviceConfig *DeviceConfig) error
	Close() error
}
