package device

import "time"

type Repository interface {
	Store(deviceName string, payload *Payload)
	ListAllDevices() []*Payload
	FindDevice(deviceName string) (*Payload, error)
}

type Payload struct {
	Id                 string         `json:"id"`
	ConnectionType     string         `json:"conn"`
	PowerSource        string         `json:"power_source"`
	Sensors            map[string]any `json:"sensors"`
	Stats              map[string]any `json:"stats"`
	availabilityTicker time.Ticker
}

func NewPayload() *Payload {
	return &Payload{
		Sensors:            map[string]any{},
		Stats:              map[string]any{},
		availabilityTicker: time.Ticker{},
	}
}
