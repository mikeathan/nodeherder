package settings

type Repository interface {
	SaveAppConfig(config *AppConfig) error
	Load() (*AppConfig, error)
	LoadBridgeConfig() (*HubConfig, error)
	SaveBridgeConfig(bridgeConfig *BridgeConfig) error
	FindOrAddDeviceConfigIfNotExists(id string) (*DeviceConfig, error)
	SaveDeviceConfig(deviceConfig *DeviceConfig) error
	Close() error
}
