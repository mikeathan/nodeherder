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

type ExposeMetricsResult interface {
	GetType() string
}

type ExposeNumericMetricsResult struct {
	Name string          `json:"name"`
	Type string          `json:"type"`
	From int64           `json:"from"`
	To   int64           `json:"to"`
	Data []*NumericValue `json:"data"`
}

func ToBinaryExposeResults(r ExposeMetricsResult) *ExposeBinaryMetricsResult {
	if e, ok := r.(*ExposeBinaryMetricsResult); ok {
		return e
	}
	return nil
}

func ToNumericExposeResults(r ExposeMetricsResult) *ExposeNumericMetricsResult {
	if e, ok := r.(*ExposeNumericMetricsResult); ok {
		return e
	}
	return nil
}

func (e *ExposeNumericMetricsResult) GetType() string {
	return "numeric"
}

type ExposeBinaryMetricsResult struct {
	Name string         `json:"name"`
	Type string         `json:"type"`
	From int64          `json:"from"`
	To   int64          `json:"to"`
	Data []*BinaryValue `json:"data"`
}

func (e *ExposeBinaryMetricsResult) GetType() string {
	return "binary"
}

type BinaryValue struct {
	X string   `json:"x"`
	Y [2]int64 `json:"y"`
}

type NumericValue struct {
	X int64   `json:"x"`
	Y float32 `json:"y"`
}

func NewExposeNumericMetricResult(name string, from time.Time, to time.Time) *ExposeNumericMetricsResult {
	return &ExposeNumericMetricsResult{
		Name: name,
		Type: "numeric",
		From: from.UnixMilli(),
		To:   to.UnixMilli(),
		Data: []*NumericValue{},
	}
}

func (e *ExposeNumericMetricsResult) Add(value float32, timestamp time.Time) {
	e.Data = append(e.Data, &NumericValue{
		X: timestamp.UnixMilli(),
		Y: value,
	})
}

func NewExposeBinaryMetricResult(name string, from time.Time, to time.Time) *ExposeBinaryMetricsResult {
	return &ExposeBinaryMetricsResult{
		Name: name,
		Type: "binary",
		From: from.UnixMilli(),
		To:   to.UnixMilli(),
		Data: []*BinaryValue{},
	}
}

func (e *ExposeBinaryMetricsResult) Add(value string, from time.Time, to time.Time) {
	e.Data = append(e.Data, &BinaryValue{
		X: value,
		Y: [2]int64{from.UnixMilli(), to.UnixMilli()},
	})
}

type DeviceMetricsResult struct {
	DeviceId string `json:"deviceId"`
	Expose   []ExposeMetricsResult
}

func NewDeviceMetricsResult(deviceid string) *DeviceMetricsResult {
	return &DeviceMetricsResult{DeviceId: deviceid, Expose: []ExposeMetricsResult{}}
}

func (e *DeviceMetricsResult) Add(event ExposeMetricsResult) {
	e.Expose = append(e.Expose, event)
}
