package api

import (
	"node-herder/models/bridge"
	"time"
)

type AggregationType string

const DeviceContextVersion = "1"

const (
	AggNone  AggregationType = ""
	AggLast  AggregationType = "last"
	AggMin   AggregationType = "min"
	AggMax   AggregationType = "max"
	AggAvg   AggregationType = "avg"
	AggCount AggregationType = "count"
)

type DeviceContextResponse struct {
	Version     string           `json:"version"`
	GeneratedAt time.Time        `json:"generatedAt"`
	Devices     []*DeviceContext `json:"devices"`
}

type DeviceContext struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Exposes     []*ExposeInfo `json:"exposes"`
}

type ExposeInfo struct {
	Name string                `json:"name"`
	Type bridge.ExposeDataType `json:"type"`
	//AccessMode   bridge.ExposeAccessMode `json:"accessMode"` // TODO: implement after fixing the bridge access flatting as now is wrong
	Unit         string            `json:"unit,omitempty"`
	Values       []string          `json:"values,omitempty"`
	ValueOn      any               `json:"valueOn,omitempty"`
	ValueOff     any               `json:"valueOff,omitempty"`
	ValueToggle  any               `json:"valueToggle,omitempty"`
	Aggregations []AggregationType `json:"aggregations"`
}
