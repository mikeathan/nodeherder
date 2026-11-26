package domain

import (
	"node-herder/models/devices"
	"time"
)

type CachedEntry struct {
	ExposeName string
	Timestamp  time.Time
	Value      []byte
}

type Repository interface {
	Store(id string, data map[string]any) error

	ViewDeviceTimeRange(device *devices.Device, from time.Time, to time.Time) (*DeviceMetricsResult, error)

	QueryDevice(deviceID string, from, to time.Time, filters []MetricFilter, collectors map[string]ExposeResult) (*DeviceMetricsResult, error)

	Prune(expireAt time.Duration) error

	// DeleteBefore removes stored readings strictly before cutoff. The explicit cutoff keeps
	// a caller from clearing recent data through a relative window by accident.
	DeleteBefore(cutoff time.Time) error

	Close() error
}
