package services

import (
	"node-herder/models/settings"
	"node-herder/utils"
	"time"
)

type ExposeDebouncer struct {
	lastEvent    time.Time
	debounceTime time.Duration
}

type DeviceDebouncer struct {
	clock utils.Clock

	configCache *settings.DeviceConfigCache
	debounce    *DeviceDebouncer
	id          string
	debounceMap map[string]time.Time
}

func NewDeviceDebouncer(deviceId string, configCache *settings.DeviceConfigCache, clock utils.Clock) *DeviceDebouncer {

	return &DeviceDebouncer{
		id:          deviceId,
		clock:       clock,
		configCache: configCache,
		debounceMap: map[string]time.Time{},
	}
}

func (d *DeviceDebouncer) DebounceExpose(exposeName string) bool {
	duration, ok := d.configCache.GetDebounce(d.id, exposeName)
	if !ok {
		// no debounce time set, so don't debounce
		return false
	}

	now := d.clock.Now()
	if lastEvent, ok := d.debounceMap[exposeName]; ok {
		if now.Sub(lastEvent) < duration {
			return true
		}
	}

	// update the last event time
	d.debounceMap[exposeName] = now
	return false
}
