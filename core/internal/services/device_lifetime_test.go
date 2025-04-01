package services_test

import (
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/devices"
	"node-herder/models/settings"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"sync"
	"testing"
	"time"
)

func TestDeviceLifetimeService_Start(t *testing.T) {
	wg := sync.WaitGroup{}
	device := utils_test.CreateDevice("x01234", "testDevice", "brightness", 124, 0.0, 255.0)
	device.Availability = devices.OnlineAvailability

	events := &devices.DeviceRequestEvents{
		OnNewDevice: func(d *devices.Device, p map[string]interface{}) {
			if d != device {
				t.Errorf("OnNewDevice device = %v, want %v", d, device)
				return
			}
			if d.Id != "x01234" {
				t.Errorf("OnNewDevice id = %v, want %v", d.Id, "x01234")
			}

			if p["brightness"] != 10.2 {
				t.Errorf("OnNewDevice payload = %v, want %v", p["brightness"], 10.2)
			}

			wg.Done()
		},
		AvailabilityTimeout: 1,
		OnDeviceAvailabilityChanged: func(p *devices.UpdatePackage) {
			if p == nil {
				t.Errorf("OnDeviceAvailabilityChanged payload is nil")
				return
			}
			if p.Availability != devices.OfflineAvailability {
				t.Errorf("OnDeviceAvailabilityChanged payload.Availability = %v, want %v", p.Availability, devices.OfflineAvailability)
			}
			wg.Done()
		}}

	app := settings.NewAppConfig()
	cache := settings.NewDeviceConfigCache(app)
	service := services.NewDeviceLifetimeService(device, events, cache, utils.NewRealClock())
	payload := map[string]interface{}{"brightness": 10.2}

	wg.Add(2)
	service.Start(payload)

	// wait until it becomes offline and assert it
	wg.Wait()

}

func TestDeviceLifetimeService_UpdateWithNewData(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	device := utils_test.CreateLightDevice("x01234", "testDevice", "brightness", 124.2)
	device.Availability = devices.OfflineAvailability

	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			if d != device {
				t.Errorf("OnDeviceUpdated device = %v, want %v", d, device)
			}
			if p.Data["brightness"] != 35.4 {
				t.Errorf("OnDeviceUpdated payload.Exposes[brightness].Data = %v, want %v", p.Data["brightness"], 35.4)
			}
			if p.Availability != devices.OnlineAvailability {
				t.Errorf("OnDeviceUpdated availability = %v, want %v", p.Availability, devices.OnlineAvailability)
			}
			wg.Done()
		},
	}

	// needs refacoring !!!!!!
	app := settings.NewAppConfig()
	d1 := settings.NewDeviceConfig("x01234")
	app.AddDeviceConfig(d1)
	cache := settings.NewDeviceConfigCache(app)
	service := services.NewDeviceLifetimeService(device, events, cache, utils.NewRealClock())
	payload := map[string]interface{}{"brightness": 35.4, "last_seen": "2023-01-01T00:00:00Z"}

	service.Update(payload)

	wg.Wait()
	if device.Exposes["brightness"].Data != 35.4 {
		t.Errorf("Device value not updated")
	}

	if device.Availability != devices.OnlineAvailability {
		t.Errorf("Device availability not updated")
	}
	if device.LastSeen != "2023-01-01T00:00:00Z" {
		t.Errorf("Device last seen not updated")
	}
}

func TestDeviceLifetimeService_UpdateWithSameData(t *testing.T) {

	device := utils_test.CreateLightDevice("x01234", "testDevice", "brightness", 124.2)
	device.Availability = devices.OfflineAvailability

	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			t.Error("OnDeviceUpdated should not be called")
		},
	}

	// needs refacoring !!!!!!
	app := settings.NewAppConfig()
	d1 := settings.NewDeviceConfig("x01234")
	app.AddDeviceConfig(d1)
	cache := settings.NewDeviceConfigCache(app)

	service := services.NewDeviceLifetimeService(device, events, cache, utils.NewRealClock())
	payload := map[string]interface{}{"brightness": 124.2, "last_seen": "2023-01-01T00:00:00Z"}

	service.Update(payload)
	time.Sleep(1 * time.Second)
	if device.Exposes["brightness"].Data != 124.2 {
		t.Errorf("Device value not updated")
	}

	if device.Availability != devices.OnlineAvailability {
		t.Errorf("Device availability not updated")
	}
	if device.LastSeen != "2023-01-01T00:00:00Z" {
		t.Errorf("Device last seen not updated")
	}
}

func TestDeviceLifetimeService_UpdateWithDebouncer(t *testing.T) {

	device := utils_test.CreateLightDevice("x01234", "testDevice", "brightness", 124.1)
	device.Availability = devices.OfflineAvailability

	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			//t.Error("OnDeviceUpdated should not be called")
		},
	}
	now := time.Now()
	mockClock := mocks.NewMockClock(func() time.Time {
		return now
	})
	appConfig := settings.NewAppConfig()
	d1 := settings.NewDeviceConfig("x01234")
	d1.Debounce = map[string]*utils.TimeInterval{
		"brightness": utils.IntervalFromSeconds(3),
	}
	appConfig.AddDeviceConfig(d1)
	cache := settings.NewDeviceConfigCache(appConfig)

	service := services.NewDeviceLifetimeService(device, events, cache, mockClock)

	testCases := []struct {
		payload      map[string]interface{}
		timestamp    time.Time
		shouldUpdate bool
	}{
		{
			payload:      map[string]interface{}{"brightness": 124.2},
			timestamp:    createTimestamp(11, 05, 10),
			shouldUpdate: true,
		},
		{
			payload:      map[string]interface{}{"brightness": 124.5},
			timestamp:    createTimestamp(11, 05, 11),
			shouldUpdate: false,
		},
		{
			payload:      map[string]interface{}{"brightness": 124.6},
			timestamp:    createTimestamp(11, 05, 14),
			shouldUpdate: true,
		},
		{
			payload:      map[string]interface{}{"brightness": 124.7},
			timestamp:    createTimestamp(11, 05, 17),
			shouldUpdate: true,
		},
		{
			payload:      map[string]interface{}{"brightness": 124.8},
			timestamp:    createTimestamp(11, 05, 18),
			shouldUpdate: false,
		},
		{
			payload:      map[string]interface{}{"brightness": 124.9},
			timestamp:    createTimestamp(11, 05, 19),
			shouldUpdate: false,
		},
	}

	for _, tc := range testCases {

		currValue := device.Exposes["brightness"].Data
		mockClock.SetMockTime(tc.timestamp)

		service.Update(tc.payload)
		time.Sleep(50 * time.Millisecond)

		if tc.shouldUpdate {

			if device.Exposes["brightness"].Data != tc.payload["brightness"] {
				t.Errorf("Device value not updated. want %v, got %v", tc.payload["brightness"], device.Exposes["brightness"].Data)
			}
		} else {

			if device.Exposes["brightness"].Data != currValue {
				t.Errorf("Error: Device value updated. want %v, got %v", currValue, device.Exposes["brightness"].Data)
			}
		}
	}

}

func TestDeviceLifetimeService_Availability(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	device := &devices.Device{Id: "testDevice", Availability: devices.OnlineAvailability}

	events := &devices.DeviceRequestEvents{
		AvailabilityTimeout: 10,
		OnNewDevice: func(d *devices.Device, p map[string]interface{}) {
			if d.Availability != devices.OnlineAvailability {
				t.Errorf("OnNewDevice availability = %v, want %v", d.Availability, devices.OfflineAvailability)
				return
			}
			wg.Done()

		},
		OnDeviceAvailabilityChanged: func(updatePackage *devices.UpdatePackage) {
			if updatePackage.Availability != devices.OfflineAvailability {
				t.Errorf("OnDeviceAvailabilityChanged availability = %v, want %v", updatePackage.Availability, devices.OfflineAvailability)
				return
			}
			wg.Done()

		},
	}

	app := settings.NewAppConfig()
	cache := settings.NewDeviceConfigCache(app)
	service := services.NewDeviceLifetimeService(device, events, cache, mocks.NewMockClock(func() time.Time { return time.Now() }))
	payload := map[string]interface{}{"test": "data"}

	// make last seen 11 seconds ago as our availability timeout is 10 seconds
	device.LastSeen = time.Now().Add(-11 * time.Second).Format(time.RFC3339)

	service.Start(payload)
	wg.Wait()

	if device.Availability != devices.OfflineAvailability {
		t.Errorf("Device availability not changed to offline")
	}
}


todo add more test cases here
func TestDeviceLifetimeService_MetricsAvailability(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	device := utils_test.CreateLightDevice("x01234", "testDevice", "brigthness", 124.1)
	device.Availability = devices.OfflineAvailability

	events := &devices.DeviceRequestEvents{
		AvailabilityTimeout: 30000,
		
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			if d != device {
				t.Errorf("OnDeviceUpdated device = %v, want %v", d, device)
			}
			wg.Done()
		},
		OnDeviceMetricsAvailable: func(d *devices.Device, p map[string]interface{}) {

			if p["brigthness"] != 34.5 {
				t.Errorf("OnDeviceMetricsAvailable value = %v, want %v", p["brigthness"], 34.5)
			}
			wg.Done()
		},
	}

	app := settings.NewAppConfig()
	d1 := settings.NewDeviceConfig("x01234")
	d1.MetricsEnabled = true
	app.AddDeviceConfig(d1)
	cache := settings.NewDeviceConfigCache(app)
	service := services.NewDeviceLifetimeService(device, events, cache, mocks.NewMockClock(func() time.Time { return time.Now() }))
	payload := map[string]interface{}{"test": "data", "brigthness": 34.5}

	service.Update(payload)
	wg.Wait()
}

func createTimestamp(hour, minute, second int) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, now.Location())
}
