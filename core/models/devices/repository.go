package devices

import (
	"time"
)

// this is going to be DeviceRepository
type Repository interface {
	AllDevices() ([]*Device, error)
	Store(key string, device *Device) error
	FindDevice(key string) (*Device, error)
	FindDevices(ids []string) ([]*Device, error)
	StoreBridge(brigeInfo []*BridgeInfo) error
	FindBridgeInfo(ids []string) ([]*BridgeInfo, error)
}

type MetricsRepository interface {
	Store(device *Device) error
	ViewDeviceTimeRange(device *Device, from time.Time, to time.Time) (*DeviceMetricsResult, error)
	ViewExposeTimeRange(device *Device, exposeName string, from time.Time, to time.Time) (*DeviceMetricsResult, error)
	Close()
}
