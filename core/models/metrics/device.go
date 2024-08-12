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

type DeviceMetricsResult struct {
	DeviceId string         `json:"deviceId"`
	Exposes  []ExposeResult `json:"exposes"`
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

func ToTimeRangeExposeResults(r ExposeResult) *ExposeTimeRangeMetricsResult {
	if e, ok := r.(*ExposeTimeRangeMetricsResult); ok {
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

type ExposeTimeRangeMetricsResult struct {
	Name string            `json:"name"`
	Type string            `json:"type"`
	From int64             `json:"from"`
	To   int64             `json:"to"`
	Data []*TimeRangeValue `json:"data"`
}

func (e *ExposeTimeRangeMetricsResult) GetType() string {
	return e.Type
}

type BinaryValue struct {
	X string   `json:"x"`
	Y [2]int64 `json:"y"`
}

type EnumValue struct {
	X string   `json:"x"`
	Y [2]int64 `json:"y"`
}

type TimeRangeValue struct {
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

func NewExposeBinaryMetricResult(name string, from time.Time, to time.Time) *ExposeTimeRangeMetricsResult {
	return &ExposeTimeRangeMetricsResult{
		Name: name,
		Type: "binary",
		From: from.UnixMilli(),
		To:   to.UnixMilli(),
		Data: []*TimeRangeValue{},
	}
}

func NewExposeEnumMetricResult(name string, from time.Time, to time.Time) *ExposeTimeRangeMetricsResult {
	return &ExposeTimeRangeMetricsResult{
		Name: name,
		Type: "enum",
		From: from.UnixMilli(),
		To:   to.UnixMilli(),
		Data: []*TimeRangeValue{},
	}
}
func (e *ExposeTimeRangeMetricsResult) Add(value string, from time.Time, to time.Time) {
	e.Data = append(e.Data, &TimeRangeValue{
		X: value,
		Y: [2]int64{from.UnixMilli(), to.UnixMilli()},
	})
}

func (c *DeviceMetricsResult) MarshalJSON() ([]byte, error) {

	res := struct {
		DeviceId string         `json:"deviceId"`
		Exposes  []ExposeResult `json:"exposes"`
	}{
		DeviceId: c.DeviceId,
		Exposes:  c.Exposes,
	}

	return json.Marshal(res)
}

func (c *DeviceMetricsResult) UnmarshalJSON(data []byte) error {

	var tmpJson map[string]interface{}
	err := json.Unmarshal(data, &tmpJson)
	if err != nil {
		return err
	}
	for key, value := range tmpJson {
		switch key {
		case "deviceId":
			c.DeviceId = value.(string)
		case "exposes":
			c.Exposes = []ExposeResult{}
			for _, expose := range value.([]interface{}) {
				exposeResult := expose.(map[string]interface{})
				if exposeType, ok := exposeResult["type"]; ok {
					switch exposeType {
					case "numeric":
						exposeBytes, err := json.Marshal(exposeResult)
						if err != nil {
							return err
						}
						var n *ExposeNumericMetricsResult = &ExposeNumericMetricsResult{}
						err = json.Unmarshal(exposeBytes, &n)
						if err != nil {
							return err
						}
						c.Exposes = append(c.Exposes, n)
					case "binary":
						exposeBytes, err := json.Marshal(exposeResult)
						if err != nil {
							return err
						}
						var b *ExposeTimeRangeMetricsResult = &ExposeTimeRangeMetricsResult{}
						err = json.Unmarshal(exposeBytes, &b)
						if err != nil {
							return err
						}
						c.Exposes = append(c.Exposes, b)
					case "enum":
						exposeBytes, err := json.Marshal(exposeResult)
						if err != nil {
							return err
						}
						var b *ExposeTimeRangeMetricsResult = &ExposeTimeRangeMetricsResult{}
						err = json.Unmarshal(exposeBytes, &b)
						if err != nil {
							return err
						}
						c.Exposes = append(c.Exposes, b)
					default:
						return fmt.Errorf("unknown expose type: %s", exposeType)
					}
				}
			}
		}
	}

	return err
}

func NewDeviceMetricsResult(deviceid string) *DeviceMetricsResult {
	return &DeviceMetricsResult{DeviceId: deviceid, Exposes: []ExposeResult{}}
}

func (e *DeviceMetricsResult) Add(event ExposeResult) {
	e.Exposes = append(e.Exposes, event)
}
