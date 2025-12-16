package domain

import (
	"encoding/json"
	"fmt"
	"node-herder/utils"
	"time"
)

// View Range rwaw metrics
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
	Size() int
	Collect(timestamp time.Time, data []byte) error
	Flush()
}

type ExposeMetricsResult struct {
	Type string `json:"type"`
}

func (e *ExposeMetricsResult) GetType() string {
	return e.Type
}

func (e *ExposeMetricsResult) ProcessData(timestamp time.Time, data []byte) error {
	return nil
}

func (e *ExposeMetricsResult) Finish() {
}

func (e *ExposeMetricsResult) Size() int {
	return 0
}

type ExposeNumericMetricsResult struct {
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	From       int64           `json:"from"`
	To         int64           `json:"to"`
	Data       []*NumericValue `json:"data"`
	Aggregator AggregationType
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

func ToExposeBinaryEventsResults(r ExposeResult) *ExposeBinaryEventsResult {
	if e, ok := r.(*ExposeBinaryEventsResult); ok {
		return e
	}
	return nil
}

func (e *ExposeNumericMetricsResult) GetType() string {
	return "numeric"
}

func (e *ExposeNumericMetricsResult) Size() int {
	return len(e.Data)
}

type timeRangeStartValue struct {
	Value     string
	Timestamp time.Time
}

type BinaryEvent struct {
	Timestamp int64 `json:"timestamp"`
	Value     bool  `json:"value"`
}

type ExposeBinaryEventsResult struct {
	Name string         `json:"name"`
	Type string         `json:"type"`
	From int64          `json:"from"`
	To   int64          `json:"to"`
	Data []*BinaryEvent `json:"data"`
}

// ExposeBinaryEventsResult
func NewExposeBinaryEventsResult(name string, from, to time.Time) *ExposeBinaryEventsResult {
	return &ExposeBinaryEventsResult{
		Name: name,
		Type: "binary",
		From: from.UnixMilli(),
		To:   to.UnixMilli(),
		Data: []*BinaryEvent{},
	}
}

func ToExposeBinaryEventsResult(r ExposeResult) *ExposeBinaryEventsResult {
	if e, ok := r.(*ExposeBinaryEventsResult); ok {
		return e
	}
	return nil
}

func (e *ExposeBinaryEventsResult) GetType() string {
	return e.Type
}

func (e *ExposeBinaryEventsResult) Size() int {
	return len(e.Data)
}

func (e *ExposeBinaryEventsResult) Collect(timestamp time.Time, raw []byte) error {
	var valueStr string
	if err := utils.ByteArrayToAny(raw, &valueStr); err != nil {
		return err
	}

	val, ok := utils.ParseBool(valueStr)
	if !ok {
		return fmt.Errorf("failed to parse bool value: %v", valueStr)
	}

	e.Data = append(e.Data, &BinaryEvent{
		Timestamp: timestamp.UnixMilli(),
		Value:     val,
	})

	return nil
}

func (e *ExposeBinaryEventsResult) Flush() {
	// do nothing
}

func (c *ExposeBinaryEventsResult) MarshalJSON() ([]byte, error) {
	res := struct {
		Name string         `json:"name"`
		Type string         `json:"type"`
		From int64          `json:"from"`
		To   int64          `json:"to"`
		Data []*BinaryEvent `json:"data"`
	}{
		Name: c.Name,
		Type: c.Type,
		From: c.From,
		To:   c.To,
		Data: c.Data,
	}

	return json.Marshal(res)
}

// ExposeTimeRangeMetricsResult
type ExposeTimeRangeMetricsResult struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	From      int64             `json:"from"`
	To        int64             `json:"to"`
	Data      []*TimeRangeValue `json:"data"`
	prevValue *timeRangeStartValue
}

func (e *ExposeTimeRangeMetricsResult) GetType() string {
	return e.Type
}

func (e *ExposeTimeRangeMetricsResult) Size() int {
	return len(e.Data)
}

type TimeRangeValue struct {
	X string   `json:"x"`
	Y [2]int64 `json:"y"`
}

type NumericValue struct {
	X int64   `json:"x"`
	Y float32 `json:"y"`
}

func NewExposeResult(name string, dataType string, aggregator AggregationType, from time.Time, to time.Time) (ExposeResult, error) {
	var result ExposeResult
	switch dataType {
	case "numeric":
		result = NewExposeNumericMetricResult(name, aggregator, from, to)
	case "binary":
		result = NewExposeBinaryMetricResult(name, from, to)
	case "enum":
		result = NewExposeTimeRangeMetricResult(name, dataType, from, to)
	default:
		return nil, fmt.Errorf("expose type %v not supported", dataType)
	}

	return result, nil
}

func NewExposeNumericMetricResult(name string, aggregator AggregationType, from time.Time, to time.Time) *ExposeNumericMetricsResult {
	return &ExposeNumericMetricsResult{
		Name:       name,
		Type:       "numeric",
		Aggregator: aggregator,
		From:       from.UnixMilli(),
		To:         to.UnixMilli(),
		Data:       []*NumericValue{},
	}
}

func (e *ExposeNumericMetricsResult) Collect(timestamp time.Time, data []byte) error {
	var value float32
	err := utils.ByteArrayToAny(data, &value)
	if err != nil {
		return err
	}

	truncated := utils.TruncateFloat32(value, 1)
	e.Add(truncated, timestamp)
	return nil
}

func (e *ExposeNumericMetricsResult) Flush() {
	// do nothing
}

func (e *ExposeNumericMetricsResult) Add(value float32, timestamp time.Time) {
	e.Data = append(e.Data, &NumericValue{
		X: timestamp.UnixMilli(),
		Y: value,
	})
}
func (c *ExposeNumericMetricsResult) MarshalJSON() ([]byte, error) {
	res := struct {
		Name string          `json:"name"`
		Type string          `json:"type"`
		From int64           `json:"from"`
		To   int64           `json:"to"`
		Data []*NumericValue `json:"data"`
	}{
		Name: c.Name,
		Type: c.Type,
		From: c.From,
		To:   c.To,
		Data: c.Data,
	}

	return json.Marshal(res)
}

func NewExposeTimeRangeMetricResult(name string, dataType string, from time.Time, to time.Time) ExposeResult {
	return &ExposeTimeRangeMetricsResult{
		Name: name,
		Type: dataType,
		From: from.UnixMilli(),
		To:   to.UnixMilli(),
		Data: []*TimeRangeValue{},
	}
}

func NewExposeBinaryMetricResult(name string, from time.Time, to time.Time) ExposeResult {
	return &ExposeBinaryEventsResult{
		Name: name,
		Type: "binary",
		From: from.UnixMilli(),
		To:   to.UnixMilli(),
		Data: []*BinaryEvent{},
	}
}

func NewExposeEnumMetricResult(name string, from time.Time, to time.Time) ExposeResult {
	return &ExposeTimeRangeMetricsResult{
		Name: name,
		Type: "enum",
		From: from.UnixMilli(),
		To:   to.UnixMilli(),
		Data: []*TimeRangeValue{},
	}
}

func (e *ExposeTimeRangeMetricsResult) Collect(timestamp time.Time, data []byte) error {
	var value string
	err := utils.ByteArrayToAny(data, &value)
	if err != nil {
		return err
	}
	if e.prevValue == nil {
		e.prevValue = &timeRangeStartValue{value, timestamp}
		return nil
	}
	if e.prevValue.Value != value {

		e.Add(e.prevValue.Value, e.prevValue.Timestamp, timestamp)
		e.prevValue = &timeRangeStartValue{value, timestamp}
	}

	return nil
}

func (e *ExposeTimeRangeMetricsResult) Flush() {
	if len(e.Data)%2 != 0 {
		e.Add(e.prevValue.Value, e.prevValue.Timestamp, time.UnixMilli(e.To))
	}
}

func (e *ExposeTimeRangeMetricsResult) Add(value string, from time.Time, to time.Time) {
	e.Data = append(e.Data, &TimeRangeValue{
		X: value,
		Y: [2]int64{from.UnixMilli(), to.UnixMilli()},
	})
}

func (c *ExposeTimeRangeMetricsResult) MarshalJSON() ([]byte, error) {
	res := struct {
		Name string            `json:"name"`
		Type string            `json:"type"`
		From int64             `json:"from"`
		To   int64             `json:"to"`
		Data []*TimeRangeValue `json:"data"`
	}{
		Name: c.Name,
		Type: c.Type,
		From: c.From,
		To:   c.To,
		Data: c.Data,
	}

	return json.Marshal(res)
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
						var b *ExposeBinaryEventsResult = &ExposeBinaryEventsResult{}
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
