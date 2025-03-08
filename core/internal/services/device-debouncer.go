package services

import (
	"node-herder/models/settings"
	"node-herder/utils"
	"sync"
	"time"
)

type DeviceDebouncer struct {
	clock       utils.Clock
	configCache *settings.DeviceConfigCache
	id          string
	debounceMap map[string]time.Time
	mutex       sync.RWMutex
}

func NewDeviceDebouncer(deviceId string, configCache *settings.DeviceConfigCache, clock utils.Clock) *DeviceDebouncer {

	return &DeviceDebouncer{
		id:          deviceId,
		clock:       clock,
		configCache: configCache,
		debounceMap: map[string]time.Time{},
		mutex:       sync.RWMutex{},
	}
}

func (d *DeviceDebouncer) DebounceExpose(exposeName string) bool {
	duration, ok := d.configCache.GetDebounce(d.id, exposeName)
	if !ok {
		// no debounce time set, so don't debounce
		return false
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

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
