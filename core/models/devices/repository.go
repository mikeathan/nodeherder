package devices

type Repository interface {
	AllDevices() ([]*Device, error)
	Store(key string, device *Device) error
	FindDevice(key string) (*Device, error)
	FindDevices(ids []string) ([]*Device, error)
	StoreBridge(brigeInfo []*BridgeInfo) error
	AllBridgeInfo() ([]*BridgeInfo, error)
	FindBridgeInfo(key string) (*BridgeInfo, error)
	Close() error
}
