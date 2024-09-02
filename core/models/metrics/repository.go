package metrics

import (
	"node-herder/models/devices"
	"time"
)

type Repository interface {
	Store(id string, data map[string]any) error
	ViewDeviceTimeRange(device *devices.Device, from time.Time, to time.Time) (*DeviceMetricsResult, error)
	Close() error
}
