package services

import (
	"context"
	"node-herder/models/automations"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/settings"
	"node-herder/utils"
	"time"
)

const (
	deviceAvailabilityTimeoutOverride = 3600
	lastSeenKey                       = "last_seen"
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
	for name, value := range payload {
		expose, ok := d.device.GetExpose(name)
		if !ok {
			continue
		}

		// skip if debounced, we need that to initialise the first debouncer timer
		// so that we can debounce the next updates
		if d.debouncerService.DebounceExpose(name, expose.Category) {
			continue
		}

		d.device.Exposes[name].Data.SetValue(value)
	}

	d.device.LastSeen = getLastSeen(payload)

	d.device.Availability = devices.OnlineAvailability
	utils.LogInfof("device [%s] %s is online", d.device.Id, d.device.FriendlyName)

	d.attempToEmitMeasurementUpdate(payload)
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

	var updatePackage = devices.NewUpdatePackage(d.device.Id)
	for name, newValue := range payload {

		expose, ok := d.device.GetExpose(name)
		if !ok {
			continue
		}

		if d.debouncerService.DebounceExpose(name, expose.Category) {
			continue
		}

		if utils.ComparePayloadValues(expose.Data.Value(), newValue) {
			continue
		}

		updatePackage.Data[name] = newValue
	}

	// if we are here even with no expose changes, it still means that the device is online
	if d.device.Availability == devices.OfflineAvailability {
		d.device.Availability = devices.OnlineAvailability

		// TODO: handle this below better
		// -updatePackage contains Availability only if we have a change on Device Availability. else its ommited.
		// thats because we use updatePackage for either measurement data or device availability change
		// -if everything is debounced then we wont emit anything. so need to think about that
		updatePackage.Availability = devices.OnlineAvailability // we handle it manually for now.

		utils.LogInfof("device [%s] %s is online", d.device.Id, d.device.FriendlyName)
		d.resetAvailabilityTimer()
	}

	d.device.LastSeen = getLastSeen(payload) // we need that.

	if updatePackage.HasData() {
		updatePackage.LastSeen = d.device.LastSeen

		// update device with expose changes
		for expose, value := range updatePackage.Data {
			d.device.Exposes[expose].Data.SetValue(value)
		}

		d.attempToEmitMeasurementUpdate(updatePackage.Data)

		// this will update device in store and emit ws event to connected clients
		d.events.OnDeviceUpdated(d.device, updatePackage)
	}
}

func (d *DeviceLifetimeService) attempToEmitMeasurementUpdate(payload map[string]interface{}) {

	// collect measurement data only if below conditions are enabled

	// TODO: BUG!
	// bug here if device is not from bridge then id will be auto geerated and wont find if automation is enabld
	if !d.configCache.IsMetricsEnabled(d.device.Id) && !d.automationQueries.IsAutomationEnabled(d.device.Id) {
		return
	}

	// send measurement updates to metrics store
	data := map[string]interface{}{}
	for name, value := range payload {
		if expose, ok := d.device.Exposes[name]; ok && expose.Category == bridge.MeasurementCategory {
			data[name] = value
		}
	}

	if len(data) > 0 {
		// this will attempt to run automation (if enabled) and store to metrics store (if enabled)
		d.events.OnDeviceMeasurementsUpdated(d.device, data)
	}
}

func (s *DeviceLifetimeService) startAvailabilityMonitoring(timeoutInSecs int, onChangeCallback func(p *devices.UpdatePackage)) {

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
				if diff.Seconds() >= float64(timeoutInSecs) {

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
