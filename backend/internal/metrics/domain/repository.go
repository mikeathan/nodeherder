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

	Close() error
}
