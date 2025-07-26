package services

import (
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
	device           *devices.Device
	debouncerService *settings.DeviceDebouncer

	availabilityTicker *time.Ticker
	availablityDone    chan bool
	events             *devices.DeviceRequestEvents
	configCache        *settings.DeviceConfigCache
	automationQueries  automations.AutomationQuerier
}

func NewDeviceLifetimeService(device *devices.Device, events *devices.DeviceRequestEvents, configCache *settings.DeviceConfigCache, automationQueries automations.AutomationQuerier, clock utils.Clock) *DeviceLifetimeService {

	return &DeviceLifetimeService{
		configCache:       configCache,
		debouncerService:  settings.NewDeviceDebouncer(device.Id, configCache, clock),
		device:            device,
		events:            events,
		automationQueries: automationQueries,
		availablityDone:   make(chan bool, 1),
	}
}

func (d *DeviceLifetimeService) Start(payload map[string]interface{}) {

	d.startAvailabilityMonitoring(d.events.AvailabilityTimeout, func(p *devices.UpdatePackage) {
		d.events.OnDeviceAvailabilityChanged(p)
	})

	d.events.OnNewDevice(d.device, payload)
}

func (d *DeviceLifetimeService) Update(payload map[string]interface{}) {

	TODO
	//  WIP, 
	config,_:=d.configCache.Get(d.device.Id)

	if config.Disabled{
		return
	}
	/// 


	var updatePackage = devices.NewUpdatePackage(d.device.Id)
	for name, newValue := range payload {

		expose, ok := d.device.GetExpose(name)
		if !ok {
			continue
		}

		if d.debouncerService.DebounceExpose(name, expose.Category) {
			continue
		}

		if utils.ComparePayloadValues(expose.Data, newValue) {
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
			d.device.Exposes[expose].Data = value
		}

		// collect measurement data only if below conditions are enabled
		if d.configCache.IsMetricsEnabled(d.device.Id) ||
			d.automationQueries.IsAutomationEnabled(d.device.Id) {

			// send measurement updates to metrics store
			measumementUpdateData := map[string]any{}
			for name, value := range updatePackage.Data {
				if e, ok := d.device.Exposes[name]; ok && e.Category == bridge.MeasurementCategory {
					measumementUpdateData[name] = value
				}
			}

			if len(measumementUpdateData) > 0 {
				// this will attempt to run automation (if enabled) and store to metrics store (if enabled)
				d.events.OnDeviceMeasurementsUpdated(d.device, measumementUpdateData)
			}
		}

		// this will update device in store and emit ws event to connected clients
		d.events.OnDeviceUpdated(d.device, updatePackage)
	}
}

func (s *DeviceLifetimeService) startAvailabilityMonitoring(timeoutInSecs int, onChangeCallback func(p *devices.UpdatePackage)) {

	s.availabilityTicker = time.NewTicker(1 * time.Second)

	go func() {
		defer close(s.availablityDone)

		for {
			select {
			case <-s.availablityDone:

				s.device.SetAvailable(false)
				utils.LogInfof("device %s availability timer killed", s.device.Id)

				// todo: move it in one place
				if onChangeCallback != nil {
					p := devices.NewUpdatePackage(s.device.Id)
					p.Availability = devices.OfflineAvailability
					onChangeCallback(p)
				}

				return

			case <-s.availabilityTicker.C:

				if !s.device.IsAvailable() {
					return
				}

				lastSeen, err := s.device.LastSeenTime()
				if err != nil {
					utils.LogErrorf("device %s failed to parse time %s", s.device.Id, err.Error())

					s.Dispose()
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

func (s *DeviceLifetimeService) Dispose() {
	s.availablityDone <- true
	s.availabilityTicker.Stop()

	utils.LogDebugf("device %s disposed", s.device.Id)
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
