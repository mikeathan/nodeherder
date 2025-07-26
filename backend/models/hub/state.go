package hub

import (
	"node-herder/models/devices"
	"node-herder/models/settings"
)

type HubState struct {
	Config  *settings.AppConfig `json:"config"`
	Devices []*devices.Device   `json:"devices"`
}

func NewHubState(config *settings.AppConfig, devices []*devices.Device) *HubState {
	return &HubState{
		Config:  config,
		Devices: devices,
	}
}
