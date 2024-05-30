package metrics

import (
	"node-herder/models/devices"
	"time"
)

type ExposeMetricsResult struct {
	Name      string       `json:"name"`
	Type      string       `json:"type"`
	Timestamp []*time.Time `json:"timestamp"`
	Values    []any        `json:"values"`
}

func NewExposeMetricsResult(name string, dataType string) *ExposeMetricsResult {
	return &ExposeMetricsResult{
		Name:      name,
		Type:      dataType,
		Timestamp: []*time.Time{},
		Values:    []any{},
	}
}

func (e *ExposeMetricsResult) Add(value any, timestamp *time.Time) {
	e.Values = append(e.Values, value)
	e.Timestamp = append(e.Timestamp, timestamp)
}

type DeviceMetricsResult struct {
	DeviceId string `json:"deviceId"`
	Expose   []*ExposeMetricsResult
}

func NewDeviceMetricsResult(deviceid string) *DeviceMetricsResult {
	return &DeviceMetricsResult{DeviceId: deviceid, Expose: []*ExposeMetricsResult{}}
}

func (e *DeviceMetricsResult) Add(event *ExposeMetricsResult) {
	e.Expose = append(e.Expose, event)
}

type Repository interface {
	Store(device *devices.Device) error
	ViewDeviceTimeRange(device *devices.Device, from time.Time, to time.Time) (*DeviceMetricsResult, error)
	ViewExposeTimeRange(device *devices.Device, exposeName string, from time.Time, to time.Time) (*DeviceMetricsResult, error)
	Close() error
}
