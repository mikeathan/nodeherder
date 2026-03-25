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

const (
	deviceAvailabilityTimeoutOverride = 3600
	lastSeenKey                       = "last_seen"
	defaultAutomationCooldown         = 20 * time.Millisecond
)

type DeviceLifetimeService struct {
	device             *devices.Device
	debouncerService   *settings.DeviceDebouncer
	stopped            bool
	availabilityTicker *time.Ticker
	events             *devices.DeviceRequestEvents
	configCache        *settings.DeviceConfigCache
	automationQueries  automations.AutomationQuerier
	availabilityCtx    context.Context
	availabilityCancel context.CancelFunc

	// automation cooldown: prevents rapid re-triggering from feedback loops
	lastAutomationTime time.Time
	automationCooldown time.Duration
	cooldownMu         sync.Mutex
}

func NewDeviceLifetimeService(device *devices.Device, events *devices.DeviceRequestEvents, configCache *settings.DeviceConfigCache, automationQueries automations.AutomationQuerier, clock utils.Clock) *DeviceLifetimeService {

	return &DeviceLifetimeService{
		configCache:        configCache,
		debouncerService:   settings.NewDeviceDebouncer(device.Id, configCache, clock),
		device:             device,
		events:             events,
		stopped:            false,
		automationQueries:  automationQueries,
		availabilityCtx:    context.Background(),
		availabilityCancel: func() {},
		automationCooldown: defaultAutomationCooldown,
	}
}

func (d *DeviceLifetimeService) Seed(payload map[string]interface{}) {

	if d.configCache.IsDeviceDisabled(d.device.Id) {
		d.stopped = true
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

	d.device.LastSeen = getLastSeen(payload)

	d.device.Availability = devices.OnlineAvailability
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

	if cfg.Disabled == d.stopped {
		return
	}

	if cfg.Disabled {
		d.stopped = true
		d.stopAvailabilityMonitoring()
		return
	}

	d.stopped = false
	d.startAvailabilityMonitoring(d.events.AvailabilityTimeout, func(p *devices.UpdatePackage) {
		d.events.OnDeviceAvailabilityChanged(p)
	})

}

func (d *DeviceLifetimeService) Update(payload map[string]interface{}) {
	if d.stopped {
		return
	}

	var updatePackage *devices.UpdatePackage
	var hasChanges bool
	var delta map[string]interface{}

	for name, newValue := range payload {

		expose, ok := d.device.GetExpose(name)
		if !ok {
			utils.LogTracef("device %s: expose %s not found in device model, skipping", d.device.Id, name)
			continue
		}

		// Deduplicate stateful properties to prevent automation feedback loops (echoes)
		// and database spam from noisy sensors. Event values bypass this to ensure
		// repeated physical triggers (like consecutive button clicks) are always processed.
		if !expose.IsEventValue(newValue) && utils.ComparePayloadValues(expose.Data.Value(), newValue) {
			utils.LogTracef("device %s: expose %s value unchanged, skipping", d.device.Id, name)
			continue
		}

		// Lazy initialize delta and updatePackage only when a real change is detected
		if delta == nil {
			delta = make(map[string]interface{})
		}
		// TODO: needs refactoring to introduce a Filter-First approach
		if updatePackage == nil {
			updatePackage = devices.NewUpdatePackage(d.device.Id)
		}

		// always update in-memory state so reads (e.g. automation step
		// calculations) see the latest value regardless of debounce
		d.device.Exposes[name].Data.SetValue(newValue)
		hasChanges = true
		delta[name] = newValue

		if d.debouncerService.DebounceExpose(expose) {
			utils.LogTracef("device %s: expose %s is debounced, skipping", d.device.Id, name)
			continue
		}

		utils.LogTracef("device %s: expose %s scheduled for update", d.device.Id, name)
		updatePackage.Data[name] = newValue
	}

	// if we are here even with no expose changes, it still means that the device is online
	if d.device.Availability == devices.OfflineAvailability {
		d.device.Availability = devices.OnlineAvailability

		// Lazy initialize updatePackage if it hasn't been created yet
		if updatePackage == nil {
			updatePackage = devices.NewUpdatePackage(d.device.Id)
		}

		// TODO: handle this below better
		// -updatePackage contains Availability only if we have a change on Device Availability. else its ommited.
		// thats because we use updatePackage for either measurement data or device availability change
		// -if everything is debounced then we wont emit anything. so need to think about that
		updatePackage.Availability = devices.OnlineAvailability // we handle it manually for now.

		utils.LogInfof("device [%s] %s is online", d.device.Id, d.device.FriendlyName)
		d.resetAvailabilityTimer()
	}

	d.device.LastSeen = getLastSeen(payload) // we need that.

	// automation: only trigger when relevant exposes have changed
	if hasChanges {
		d.attemptToTriggerAutomation(delta)
	}

	// storage + UI: only emit for non-debounced changes or availability changes
	if updatePackage != nil && (updatePackage.HasData() || updatePackage.Availability != "") {
		updatePackage.LastSeen = d.device.LastSeen

		if updatePackage.HasData() {
			d.attemptToStoreMetrics(updatePackage.Data)
		}

		// this will update device in store and emit ws event to connected clients
		d.events.OnDeviceUpdated(d.device, updatePackage)
	}
}

// attemptToTriggerAutomation fires the automation engine if automation is
// enabled for this device. Only triggers when relevant exposes have changed.
func (d *DeviceLifetimeService) attemptToTriggerAutomation(delta map[string]interface{}) {
	if !d.automationQueries.IsAutomationEnabled(d.device.Id) {
		utils.LogDebugf("device %s: automation disabled, skipping", d.device.Id)
		return
	}

	if d.events.OnDeviceAutomationTriggered != nil {
		utils.LogDebugf("device %s: triggering automation for changed exposes", d.device.Id)
		d.events.OnDeviceAutomationTriggered(d.device, delta)
	}
}

// attemptToStoreMetrics sends debounced measurement data to the metrics store.
// Only measurement-category exposes are included.
func (d *DeviceLifetimeService) attemptToStoreMetrics(payload map[string]interface{}) {

	if !d.configCache.IsMetricsEnabled(d.device.Id) {
		utils.LogDebugf("device %s: metrics disabled, skipping storage", d.device.Id)
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

	if s.availabilityTicker != nil {
		utils.LogDebugf("device %s availability monitor already running", s.device.Id)
		return
	}

	s.availabilityTicker = time.NewTicker(1 * time.Second)
	s.availabilityCtx, s.availabilityCancel = context.WithCancel(context.Background())

	go func() {

		for {
			select {
			case <-s.availabilityCtx.Done():

				s.availabilityTicker.Stop()
				s.availabilityTicker = nil
				//s.device.SetAvailable(false)
				utils.LogDebugf("device %s availability ticker cancelled", s.device.Id)

				return

			case <-s.availabilityTicker.C:

				if !s.device.IsAvailable() {
					return
				}

				lastSeen, err := s.device.LastSeenTime()
				if err != nil {
					lastSeen = time.Now()
					utils.LogErrorf("device %s failed to parse time %s. fallback to now()", s.device.Id, err.Error())
				}

				now := time.Now()
				diff := now.Sub(lastSeen)
				if diff >= timeoutDuration {

					s.device.SetAvailable(false)
					utils.LogInfof("device %s is offine", s.device.Id)

					// todo: move it in one place
					if onChangeCallback != nil {
						p := devices.NewUpdatePackage(s.device.Id)
						p.Availability = devices.OfflineAvailability
						onChangeCallback(p)
					}

					s.availabilityTicker.Stop()
				}

			}
		}
	}()
}
func (d *DeviceLifetimeService) resetAvailabilityTimer() {
	if d.availabilityTicker != nil {
		d.availabilityTicker.Reset(1 * time.Second)
	} else {
		utils.LogDebugf("device %s availability ticker not initialized, cannot reset", d.device.Id)
	}
}

func (s *DeviceLifetimeService) stopAvailabilityMonitoring() {

	if s.availabilityCancel != nil {
		s.availabilityCancel()
		s.availabilityCancel = nil
	}
	utils.LogDebugf("device %s monitoring stopped", s.device.Id)
}

func getLastSeen(data map[string]interface{}) string {
	if val, ok := data[lastSeenKey]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	utils.LogDebug("lastSeen not in payload, using current time.")
	return getCurrentTime()
}

func getCurrentTime() string {
	return time.Now().Format(time.RFC3339)
}
