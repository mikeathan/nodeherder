package services

import (
	"node-herder/models/devices"
	"node-herder/utils"
	"time"
)

const (
	deviceAvailabilityTimeoutOverride = 3600
	lastSeenKey                       = "last_seen"
)

type DeviceService struct {
	device           *devices.Device
	debouncerService *DeviceDebouncer

	availabilityTicker *time.Ticker
	availablityDone    chan bool
}

func NewDeviceService(device *devices.Device, debouncerService *DeviceDebouncer) *DeviceService {

	s := &DeviceService{
		device:           device,
		debouncerService: debouncerService,
		availablityDone:  make(chan bool),
	}

	return s
}

func (d *DeviceService) Update(payload map[string]interface{}) *devices.UpdatePackage {

	var updatePackage = devices.NewUpdatePackage(d.device.Id)
	for name, newValue := range payload {

		// TODO:
		// debounce needs to happen here for each expose hat has debounce value

		if expose, ok := d.device.GetExpose(name); ok && expose.Data != newValue {

			if len(updatePackage.Data) != 0 {
				updatePackage.Data[name] = newValue
			} else if expose.Category == devices.MeasurementCategory {
				updatePackage.Data[name] = newValue
			}

			// update device expose with updated data
			d.device.Exposes[name].Data = newValue
		}
	}

	// defer device.mutex.Unlock()
	// device.mutex.Lock()

	if updatePackage.HasData() {
		updatePackage.LastSeen = getLastSeen(payload)
	}

	if d.device.Availability == devices.OfflineAvailability {
		d.device.Availability = devices.OnlineAvailability

		// TODO: handle this below better
		// updatePackage contains Availability only if we have a change on Device Availability. else its ommited.
		// thats because we use updatePackage for either measurement data or device availability change
		updatePackage.Availability = devices.OnlineAvailability // we handle it manually for now.

		utils.LogInfof("device [%s] %s is online", d.device.Id, d.device.FriendlyName)
		d.device.ResetAvailabilityTimer()
	}

	d.device.LastSeen = getLastSeen(payload) // we need that.
	return updatePackage
}

func (s *DeviceService) Monitor(timeoutInSecs int, onChangeCallback func(p interface{})) {

	s.availabilityTicker = time.NewTicker(1 * time.Second)
	s.availablityDone = make(chan bool)

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

func (s *DeviceService) Dispose() {
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
