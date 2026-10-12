package services

import (
	"context"
	"node-herder/models/automations"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/settings"
	"node-herder/utils"
	"sync"
	"time"
)

const lastSeenKey = "last_seen"

// DeviceLifetimeService owns one device's state. Seed and Update for a device
// must not run concurrently: the hub runs them on that device's lane, one
// message at a time in arrival order, so the automations they trigger read the
// state written by that message (spec 005, FR-01).
type DeviceLifetimeService struct {
	device            *devices.Device
	debouncerService  *settings.DeviceDebouncer
	stopped           bool
	events            *devices.DeviceRequestEvents
	configCache       *settings.DeviceConfigCache
	automationQueries automations.AutomationQuerier
	monitorMu         sync.Mutex
	monitor           *availabilityMonitor
}

type availabilityMonitor struct {
	ticker *time.Ticker
	ctx    context.Context
	cancel context.CancelFunc
}

func NewDeviceLifetimeService(device *devices.Device, events *devices.DeviceRequestEvents, configCache *settings.DeviceConfigCache, automationQueries automations.AutomationQuerier, clock utils.Clock) *DeviceLifetimeService {

	return &DeviceLifetimeService{
		configCache:       configCache,
		debouncerService:  settings.NewDeviceDebouncer(device.Id, configCache, clock),
		device:            device,
		events:            events,
		stopped:           false,
		automationQueries: automationQueries,
	}
}

func (d *DeviceLifetimeService) Seed(payload map[string]interface{}) {

	if d.configCache.IsDeviceDisabled(d.device.Id) {
		d.monitorMu.Lock()
		d.stopped = true
		d.stopAvailabilityMonitoringLocked()
		d.monitorMu.Unlock()
		return
	}

	d.startAvailabilityMonitoring(d.events.AvailabilityTimeout, func(p *devices.UpdatePackage) {
		d.events.OnDeviceAvailabilityChanged(p)
	})

	// update device with initial data
	var hasChanges bool
	for name, value := range payload {
		expose, ok := d.device.GetExpose(name)
		if !ok {
			continue
		}

		// skip if debounced, we need that to initialise the first debouncer timer
		// so that we can debounce the next updates
		if d.debouncerService.DebounceExpose(expose) {
			continue
		}

		d.device.Exposes[name].Data.SetValue(value)
		hasChanges = true
	}

	d.device.SetLastSeen(getLastSeen(payload))

	d.device.SetAvailability(devices.OnlineAvailability)
	utils.LogInfof("device [%s] %s is online", d.device.Id, d.device.FriendlyName)

	// Seed fires automation directly, no cooldown is required
	if hasChanges && d.automationQueries.IsAutomationEnabled(d.device.Id) {
		utils.LogDebugf("device %s: triggering automation for seed event", d.device.Id)
		d.events.OnDeviceAutomationTriggered(d.device, payload)
	}
	d.attemptToStoreMetrics(payload)
	d.events.OnNewDevice(d.device)
}

func (d *DeviceLifetimeService) OnConfigUpdated(cfg *settings.DeviceConfig) {
	d.monitorMu.Lock()
	defer d.monitorMu.Unlock()
	if cfg.Disabled == d.stopped {
		return
	}

	if cfg.Disabled {
		d.stopped = true
		d.stopAvailabilityMonitoringLocked()
		return
	}

	d.stopped = false
	d.startAvailabilityMonitoringLocked(d.events.AvailabilityTimeout, func(p *devices.UpdatePackage) {
		d.events.OnDeviceAvailabilityChanged(p)
	})

}

func (d *DeviceLifetimeService) Update(payload map[string]interface{}) {
	d.monitorMu.Lock()
	stopped := d.stopped
	d.monitorMu.Unlock()
	if stopped {
		return
	}

	live, delta, sampled := d.collectUpdate(payload)
	d.restoreAvailability(live)
	live.LastSeen = getLastSeen(payload)
	d.device.SetLastSeen(live.LastSeen)

	// Accepted state is visible to automation before any output callback.
	if len(delta) > 0 {
		d.attemptToTriggerAutomation(delta)
	}
	if len(sampled) > 0 {
		d.attemptToStoreMetrics(sampled)
	}
	if live.HasData() || live.Availability != "" {
		d.events.OnDeviceUpdated(d.device, live)
	}
}

// collectUpdate separates device evidence, changed automation inputs and sampled
// metrics. Writable same-value reports can confirm a command without retriggering
// automation or consuming the metrics sampling window.
func (d *DeviceLifetimeService) collectUpdate(payload map[string]interface{}) (*devices.UpdatePackage, map[string]interface{}, map[string]interface{}) {
	live := devices.NewUpdatePackage(d.device.Id)
	delta := make(map[string]interface{})
	sampled := make(map[string]interface{})
	for name, value := range payload {
		expose, ok := d.device.GetExpose(name)
		if !ok {
			continue
		}
		changed := acceptExposeReport(expose, value)
		if changed {
			delta[name] = value
			if !d.debouncerService.DebounceExpose(expose) {
				sampled[name] = value
			}
		}
		if changed || isWritableExpose(expose) {
			live.Data[name] = value
		}
	}
	return live, delta, sampled
}

func acceptExposeReport(expose *devices.Entity, value interface{}) bool {
	if !expose.IsEventValue(value) && utils.ComparePayloadValues(expose.Data.Value(), value) {
		return false
	}
	expose.Data.SetValue(value)
	return true
}

func isWritableExpose(expose *devices.Entity) bool {
	return expose.AccessMode == bridge.WriteAccessMode || expose.AccessMode == bridge.ReadWriteAccessMode
}

func (d *DeviceLifetimeService) restoreAvailability(live *devices.UpdatePackage) {
	if d.device.GetAvailability() != devices.OfflineAvailability {
		return
	}
	d.device.SetAvailability(devices.OnlineAvailability)
	live.Availability = devices.OnlineAvailability
	utils.LogInfof("device [%s] %s is online", d.device.Id, d.device.FriendlyName)
	d.resetAvailabilityTimer()
}

// attemptToTriggerAutomation fires the automation engine if automation is
// enabled for this device. Only triggers when relevant exposes have changed.
func (d *DeviceLifetimeService) attemptToTriggerAutomation(delta map[string]interface{}) {
	if !d.automationQueries.IsAutomationEnabled(d.device.Id) {
		utils.LogTracef("device %s: automation disabled, skipping", d.device.Id)
		return
	}

	if d.events.OnDeviceAutomationTriggered != nil {
		utils.LogTracef("device %s: dispatching changed exposes to automation evaluation", d.device.Id)
		d.events.OnDeviceAutomationTriggered(d.device, delta)
	}
}

// attemptToStoreMetrics sends debounced measurement data to the metrics store.
// Only measurement-category exposes are included.
func (d *DeviceLifetimeService) attemptToStoreMetrics(payload map[string]interface{}) {

	if !d.configCache.IsMetricsEnabled(d.device.Id) {
		utils.LogTracef("device %s: metrics disabled, skipping storage", d.device.Id)
		return
	}

	// filter to measurement exposes only
	data := map[string]interface{}{}
	for name, value := range payload {
		if expose, ok := d.device.Exposes[name]; ok && expose.Category == bridge.MeasurementCategory {
			data[name] = normalizeMeasurementValue(expose, value)
		}
	}

	if len(data) > 0 {
		d.events.OnDeviceMeasurementsUpdated(d.device, data)
	}
}

func normalizeMeasurementValue(expose *devices.Entity, value any) any {

	if expose == nil {
		return value
	}

	if expose.Type != bridge.BinaryDataType {
		return value
	}

	switch v := value.(type) {
	case bool:
		return v
	case string:
		if boolVal, ok := utils.ConvertToBool(v); ok {
			return boolVal
		}
	case []byte:
		if boolVal, ok := utils.ConvertToBool(string(v)); ok {
			return boolVal
		}
	}

	return value
}

func (s *DeviceLifetimeService) startAvailabilityMonitoring(timeoutDuration time.Duration, onChangeCallback func(p *devices.UpdatePackage)) {
	s.monitorMu.Lock()
	defer s.monitorMu.Unlock()
	if !s.stopped {
		s.startAvailabilityMonitoringLocked(timeoutDuration, onChangeCallback)
	}
}

func (s *DeviceLifetimeService) startAvailabilityMonitoringLocked(timeoutDuration time.Duration, onChangeCallback func(p *devices.UpdatePackage)) {
	if s.monitor != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	monitor := &availabilityMonitor{ticker: time.NewTicker(time.Second), ctx: ctx, cancel: cancel}
	s.monitor = monitor

	go func() {
		defer func() {
			monitor.ticker.Stop()
			monitor.cancel()
			s.monitorMu.Lock()
			if s.monitor == monitor {
				s.monitor = nil
			}
			s.monitorMu.Unlock()
		}()

		for {
			select {
			case <-monitor.ctx.Done():
				utils.LogDebugf("device %s availability ticker cancelled", s.device.Id)

				return

			case <-monitor.ticker.C:

				if !s.device.IsAvailable() {
					s.monitorMu.Lock()
					if s.monitor == monitor && !s.device.IsAvailable() {
						monitor.ticker.Stop()
					}
					s.monitorMu.Unlock()
					continue // retain the worker so an online update can reset it
				}

				lastSeen, err := s.device.LastSeenTime()
				if err != nil {
					lastSeen = time.Now()
					utils.LogErrorf("device %s failed to parse time %s. fallback to now()", s.device.Id, err.Error())
				}

				now := time.Now()
				diff := now.Sub(lastSeen)
				if diff >= timeoutDuration {
					s.monitorMu.Lock()
					if s.monitor != monitor || s.stopped || monitor.ctx.Err() != nil {
						s.monitorMu.Unlock()
						return
					}
					s.device.SetAvailable(false)
					monitor.ticker.Stop()
					s.monitorMu.Unlock()
					utils.LogInfof("device %s is offine", s.device.Id)

					// todo: move it in one place
					if onChangeCallback != nil {
						p := devices.NewUpdatePackage(s.device.Id)
						p.Availability = devices.OfflineAvailability
						onChangeCallback(p)
					}

				}

			}
		}
	}()
}
func (d *DeviceLifetimeService) resetAvailabilityTimer() {
	d.monitorMu.Lock()
	defer d.monitorMu.Unlock()
	if d.stopped {
		return
	}
	if d.monitor != nil {
		d.monitor.ticker.Reset(time.Second)
		return
	}
	d.startAvailabilityMonitoringLocked(d.events.AvailabilityTimeout, d.events.OnDeviceAvailabilityChanged)
}

func (s *DeviceLifetimeService) stopAvailabilityMonitoringLocked() {
	if s.monitor != nil {
		s.monitor.cancel()
		s.monitor.ticker.Stop()
		s.monitor = nil
	}
}

func getLastSeen(data map[string]interface{}) string {
	if val, ok := data[lastSeenKey]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	utils.LogTrace("lastSeen not in payload, using current time.")
	return getCurrentTime()
}

func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}
