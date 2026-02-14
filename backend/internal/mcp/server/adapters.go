package server

import (
	"node-herder/internal/mcp/resolver"
	"node-herder/models/devices"
)

type deviceInfoAdapter struct {
	store DeviceStore
}

func (a *deviceInfoAdapter) AllDeviceInfo() ([]resolver.DeviceInfo, error) {
	state, err := a.store.LoadHubState()
	if err != nil {
		return nil, err
	}
	result := make([]resolver.DeviceInfo, len(state.Devices))
	for i, d := range state.Devices {
		result[i] = resolver.DeviceInfo{
			ID:   d.Id,
			Name: d.FriendlyName,
		}
	}
	return result, nil
}

type deviceLookupAdapter struct {
	store DeviceStore
}

func (a *deviceLookupAdapter) FindDeviceByIds(ids []string) ([]*devices.Device, error) {
	state, err := a.store.LoadHubState()
	if err != nil {
		return nil, err
	}

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	var result []*devices.Device
	for _, d := range state.Devices {
		if idSet[d.Id] {
			result = append(result, d)
		}
	}
	return result, nil
}
