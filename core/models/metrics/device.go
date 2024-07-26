package metrics

import (
	"time"
)

type LoadDeviceMetricsRequest struct {
	Id     string `json:"id"`
	Expose string `json:"expose,omitempty"`
	From   int64  `json:"from"`
	To     int64  `json:"to"`
}

type ExposeMetricsResult struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	From       int64   `json:"from"`
	To         int64   `json:"to"`
	Timestamps []int64 `json:"timestamps"`
	Values     []any   `json:"values"`
}

func NewExposeMetricsResult(name string, from time.Time, to time.Time, dataType string) *ExposeMetricsResult {
	return &ExposeMetricsResult{
		Name:       name,
		Type:       dataType,
		From:       from.UnixMilli(),
		To:         to.UnixMilli(),
		Timestamps: []int64{},
		Values:     []any{},
	}
}

func (e *ExposeMetricsResult) Add(value any, timestamp time.Time) {
	e.Values = append(e.Values, value)
	e.Timestamps = append(e.Timestamps, timestamp.UnixMilli())
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
