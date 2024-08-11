package settings

type Repository interface {
	Save(config *AppConfig) error
	Load() (*AppConfig, error)
	FindOrAddDeviceConfigIfNotExists(id string) (*DeviceConfig, error)
	SaveDeviceConfig(deviceConfig *DeviceConfig) error
	Close() error
}
