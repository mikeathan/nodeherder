package services

import (
	"node-herder/models/settings"
	"node-herder/store"
	"node-herder/utils"
)

type DeviceDebouncer struct {
	clock     utils.Clock
	store     store.AppStore
	configMap map[string]*settings.DeviceConfig
}

func NewDeviceDebouncer(store store.AppStore, clock utils.Clock) *DeviceDebouncer {
	app, err := store.LoadAppConfig()

	if err != nil {
		// DO STH
	}

	return &DeviceDebouncer{
		clock:     clock,
		store:     store,
		configMap: app.Hub.Devices,
	}
}

func (d *DeviceDebouncer) Debounce(friendlyName string, payload map[string]interface{}, processFunc func() error) error {
	deviceId := d.store.ResolveFriendlyName(friendlyName)
	if cfg, err := d.store.LoadDeviceConfig(deviceId); err == nil {
		now := d.clock.Now()

		// if cfg, ok := d.configMap[id]; ok {

		// 	if now.Sub(d.lastEvent) < d.debounceTime {
		// 		// Event within debounce window, ignore
		// 		return
		// 	}
		// }
		d.lastEvent = d.clock.Now()
	}

	return processFunc()

	if eventTime.Sub(d.lastEvent) < d.debounceTime {
		// Event within debounce window, ignore
		return
	}

	d.lastEvent = eventTime
	processFunc()
}
