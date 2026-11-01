package hub

import (
	"errors"
	"time"

	"node-herder/internal/metrics/domain"
	"node-herder/models/devices"
	"node-herder/models/hub"
	"node-herder/models/settings"
	"node-herder/store"
)

// publicDevices is the only device projection passed to unauthenticated context
// and MCP consumers. The underlying store remains authoritative for signed-in users.
type publicDevices struct{ source store.AppStore }

func (p publicDevices) LoadHubState() (*hub.HubState, error) {
	state, err := p.source.LoadHubState()
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, errors.New("hub state unavailable")
	}
	visible := make([]*devices.Device, 0, len(state.Devices))
	for _, device := range state.Devices {
		if device != nil && device.Provider == "" {
			visible = append(visible, device)
		}
	}
	return hub.NewHubState(state.Config, visible), nil
}

func (p publicDevices) RegisterIsDirtyCallback(callback store.AppStoreDirtyFlagCallback) {
	p.source.RegisterIsDirtyCallback(callback)
}
func (p publicDevices) AppConfig() *settings.AppConfigCache { return p.source.AppConfig() }

func (p publicDevices) FindDeviceByIds(ids []string) ([]*devices.Device, error) {
	found, err := p.source.FindDeviceByIds(ids)
	if err != nil {
		return nil, err
	}
	visible := make([]*devices.Device, 0, len(found))
	for _, device := range found {
		if device != nil && device.Provider == "" {
			visible = append(visible, device)
		}
	}
	return visible, nil
}

func (p publicDevices) QueryDevice(id string, from, to time.Time, filters []domain.MetricFilter, collectors map[string]domain.ExposeResult) (*domain.DeviceMetricsResult, error) {
	device, err := p.source.FindDeviceById(id)
	if err != nil || device == nil || device.Provider != "" {
		return nil, errors.New("device unavailable")
	}
	return p.source.QueryDevice(id, from, to, filters, collectors)
}
