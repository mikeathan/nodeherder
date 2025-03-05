package services

import (
	"node-herder/store"
	"node-herder/utils"
	"time"
)

type ExposeDebouncer struct {
	lastEvent    time.Time
	debounceTime time.Duration
}

 we will have a deviceconfigurationProcessor or handler

 that will keep the debouncer Map per expose
 and will be updated when we update device config
 will be passed here and we wont need to do many ifs, just one
type DeviceDebouncer struct {
	clock            utils.Clock
	store            store.AppStore
	id               string
	debounceMap      map[string]time.Time

}

func NewDeviceDebouncer(deviceId string, store store.AppStore, clock utils.Clock) *DeviceDebouncer {

	return &DeviceDebouncer{
		id:          deviceId,
		clock:       clock,
		store:       store,
		debounceMap: map[string]time.Time{},
	}
}

func (d *DeviceDebouncer) DebounceExpose(exposeName string) bool {

	// 3 ifs not sure is good !!!!!
	if cfg, err := d.store.LoadDeviceConfig(d.id); err == nil {
		if debounce, ok := cfg.Debounce[exposeName]; ok {

			now := d.clock.Now()

			if lastEvent, ok := d.debounceMap[exposeName]; ok {
				if now.Sub(lastEvent) < debounce.Duration() {
					return true
				}
			}
			d.debounceMap[exposeName] = now
		}
	}
	return false
}
