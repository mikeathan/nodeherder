package models

import (
	"node-herder/internal/metrics/query"
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

	Query(query query.MetricsQueryRequest) (*[]query.MetricsQueryResponse, error)

	Prune(expireAt time.Duration) error

	Close() error
}
