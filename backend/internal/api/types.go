package api

import (
	"node-herder/models/bridge"
	"time"
)

type AggregationType string

const (
	AggNone  AggregationType = ""
	AggLast  AggregationType = "last"
	AggMin   AggregationType = "min"
	AggMax   AggregationType = "max"
	AggAvg   AggregationType = "avg"
	AggCount AggregationType = "count"
)

type DeviceContextResponse struct {
	Version     string          `json:"version"`
	GeneratedAt time.Time       `json:"generatedAt"`
	Devices     []DeviceContext `json:"devices"`
}

type DeviceContext struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Room    string       `json:"room,omitempty"`
	Type    string       `json:"type"`
	Exposes []ExposeInfo `json:"exposes"`
}

type ExposeInfo struct {
	Name         string                `json:"name"`
	Type         bridge.ExposeDataType `json:"type"`
	Unit         string                `json:"unit,omitempty"`
	Values       []string              `json:"values,omitempty"`
	Aggregations []AggregationType     `json:"aggregations"`
}
