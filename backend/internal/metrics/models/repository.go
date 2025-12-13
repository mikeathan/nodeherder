package metrics

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

	Query(query MetricsQuery) (*[]MetricsQueryResult, error)
	
	Prune(expireAt time.Duration) error

	Close() error
}
