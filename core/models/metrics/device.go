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

type NumericValue struct {
	X int64   `json:"x"`
	Y float32 `json:"y"`
}

type ExposeNumericMetricResult struct {
	Name string          `json:"name"`
	Type string          `json:"type"`
	From int64           `json:"from"`
	To   int64           `json:"to"`
	Data []*NumericValue `json:"data"`
}

func NewExposeNumericMetricResult(name string, from time.Time, to time.Time) *ExposeNumericMetricResult {
	return &ExposeNumericMetricResult{
		Name: name,
		Type: "numeric",
		From: from.UnixMilli(),
		To:   to.UnixMilli(),
		Data: []*NumericValue{},
	}
}

func (e *ExposeNumericMetricResult) Add(value float32, timestamp time.Time) {
	e.Data = append(e.Data, &NumericValue{
		X: timestamp.UnixMilli(),
		Y: value,
	})
}

type BinaryValue struct {
	X string   `json:"x"`
	Y [2]int64 `json:"y"`
}

type ExposeBinaryMetricResult struct {
	Name string         `json:"name"`
	Type string         `json:"type"`
	From int64          `json:"from"`
	To   int64          `json:"to"`
	Data []*BinaryValue `json:"data"`
}

func NewExposeBinaryMetricResult(name string, from time.Time, to time.Time) *ExposeBinaryMetricResult {
	return &ExposeBinaryMetricResult{
		Name: name,
		Type: "binary",
		From: from.UnixMilli(),
		To:   to.UnixMilli(),
		Data: []*BinaryValue{},
	}
}

func (e *ExposeBinaryMetricResult) Add(value string, from time.Time, to time.Time) {
	e.Data = append(e.Data, &BinaryValue{
		X: value,
		Y: [2]int64{from.UnixMilli(), to.UnixMilli()},
	})
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
