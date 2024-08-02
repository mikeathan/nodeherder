package metrics

import (
	"encoding/json"
	"fmt"
	"time"
)

type LoadDeviceMetricsRequest struct {
	Id     string `json:"id"`
	Expose string `json:"expose,omitempty"`
	From   int64  `json:"from"`
	To     int64  `json:"to"`
}

type ExposeResult interface {
	GetType() string
}

type ExposeMetricsResult struct {
	Type string `json:"type"`
}

func (e *ExposeMetricsResult) GetType() string {
	return e.Type
}

type ExposeNumericMetricsResult struct {
	Name string          `json:"name"`
	Type string          `json:"type"`
	From int64           `json:"from"`
	To   int64           `json:"to"`
	Data []*NumericValue `json:"data"`
}

func ToBinaryExposeResults(r ExposeResult) *ExposeBinaryMetricsResult {
	if e, ok := r.(*ExposeBinaryMetricsResult); ok {
		return e
	}
	return nil
}

func ToNumericExposeResults(r ExposeResult) *ExposeNumericMetricsResult {
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


INEED UNMARSHALLER FOR deviceMetricsobject
func (c *ExposeMetricsResult) UnmarshalJSON(data []byte) error {
	var s *ExposeMetricsResult = &ExposeMetricsResult{}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s.GetType() == "binary" {
		var b *ExposeBinaryMetricsResult = &ExposeBinaryMetricsResult{}
		//b:=ToBinaryExposeResults(s)
		return json.Unmarshal(data, &b)
	}
	if s.GetType() == "numeric" {
		var n *ExposeNumericMetricsResult = &ExposeNumericMetricsResult{}
		//b:=ToBinaryExposeResults(s)
		return json.Unmarshal(data, &n)
	}

	return nil
}
func (c *ExposeMetricsResult) MarshalJSON() ([]byte, error) {
	switch c.GetType() {
	case "binary":
		b := ToBinaryExposeResults(c)
		return json.Marshal(b)
		// return json.Marshal(struct {
		// 	Name string         `json:"name"`
		// 	Type string         `json:"type"`
		// 	From int64          `json:"from"`
		// 	To   int64          `json:"to"`
		// 	Data []*BinaryValue `json:"data"`
		// }{
		// 	Name: b.Name,
		// 	Type: c.GetType(),
		// 	From: b.From,
		// 	To:   b.To,
		// 	Data: b.Data,
		// })

	case "numeric":
		n := ToNumericExposeResults(c)
		return json.Marshal(n)
		// return json.Marshal(struct {
		// 	Name string          `json:"name"`
		// 	Type string          `json:"type"`
		// 	From int64           `json:"from"`
		// 	To   int64           `json:"to"`
		// 	Data []*NumericValue `json:"data"`
		// }{
		// 	Name: n.Name,
		// 	Type: c.GetType(),
		// 	From: n.From,
		// 	To:   n.To,
		// 	Data: n.Data,
		// })

	default:
		return nil, fmt.Errorf("unknown customer type")
	}
}

type DeviceMetricsResult struct {
	DeviceId string `json:"deviceId"`
	Expose   []ExposeResult
}

func NewDeviceMetricsResult(deviceid string) *DeviceMetricsResult {
	return &DeviceMetricsResult{DeviceId: deviceid, Expose: []ExposeResult{}}
}

func (e *DeviceMetricsResult) Add(event ExposeResult) {
	e.Expose = append(e.Expose, event)
}
