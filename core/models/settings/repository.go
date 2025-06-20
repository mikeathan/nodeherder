package settings

type Repository interface {
	SaveAppConfig(config *AppConfig) error
	Load() (*AppConfig, error)
	LoadBridgeConfig() (*BridgeConfig, error)
	SaveBridgeConfig(bridgeConfig *BridgeConfig) error
	SaveHubConfig(hubConfig *HubConfig) error
	LoadOrDefaultDeviceConfig(id string) (*DeviceConfig, error)
	SaveDeviceConfig(deviceConfig *DeviceConfig) error
	Close() error
}
