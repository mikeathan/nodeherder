package metrics

import (
	"node-herder/models/devices"
	"time"
)

type Repository interface {
	Store(device *devices.Device) error
	ViewDeviceTimeRange(device *devices.Device, from time.Time, to time.Time) (*DeviceMetricsResult, error)
	ViewExposeTimeRange(device *devices.Device, exposeName string, from time.Time, to time.Time) (*DeviceMetricsResult, error)
	Close() error
}
