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
// TODO:

// device/
//   ├── device_repository.go        // Interface definition for DeviceRepository
//   └── repository/
//       ├── file_repository.go        // File-based implementation of DeviceRepository
//       └── memory_repository.go      // Memory-based implementation of DeviceRepository
