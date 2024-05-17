package devices

import (
	"time"
)

// this is going to be DeviceRepository
type Repository interface {
	AllDevices() []*Device
	Store(key string, device *Device)
	FindDevice(key string) (*Device, error)
	FindDevices(ids []string) []*Device
}

type MetricsRepository interface {
	Store(device *Device) error
	ViewDeviceTimeRange(device *Device, from time.Time, to time.Time) (*DeviceMetricsResult, error)
	ViewExposeTimeRange(deviceId string, exposeName string, from time.Time, to time.Time) (*DeviceMetricsResult, error)
	Close()
}
