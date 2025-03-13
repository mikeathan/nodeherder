package services_test

import (
	"fmt"
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
	device := utils_test.CreateDevice("x01234", "testDevice", "brigthness", 124, 0.0, 255.0)
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

			if p["brigthness"] != 10.2 {
				t.Errorf("OnNewDevice payload = %v, want %v", p["brigthness"], 10.2)
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

	debouncer := utils_test.CreateDebouncer("x01234")

	service := services.NewDeviceLifetimeService(device, events, debouncer)
	payload := map[string]interface{}{"brigthness": 10.2}

	wg.Add(2)
	service.Start(payload)

	// wait until it becomes offline and assert it
	wg.Wait()

}

func TestDeviceLifetimeService_UpdateWithNewData(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	device := utils_test.CreateLightDevice("x01234", "testDevice", "brigthness", 124.2)
	device.Availability = devices.OfflineAvailability

	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			if d != device {
				t.Errorf("OnDeviceUpdated device = %v, want %v", d, device)
			}
			if p.Data["brigthness"] != 35.4 {
				t.Errorf("OnDeviceUpdated payload.Exposes[brigthness].Data = %v, want %v", p.Data["brigthness"], 35.4)
			}
			if p.Availability != devices.OnlineAvailability {
				t.Errorf("OnDeviceUpdated availability = %v, want %v", p.Availability, devices.OnlineAvailability)
			}
			wg.Done()
		},
	}

	debouncer := utils_test.CreateDebouncer("x01234")
	service := services.NewDeviceLifetimeService(device, events, debouncer)
	payload := map[string]interface{}{"brigthness": 35.4, "last_seen": "2023-01-01T00:00:00Z"}

	service.Update(payload)

	wg.Wait()
	if device.Exposes["brigthness"].Data != 35.4 {
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

	device := utils_test.CreateLightDevice("x01234", "testDevice", "brigthness", 124.2)
	device.Availability = devices.OfflineAvailability

	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			t.Error("OnDeviceUpdated should not be called")
		},
	}

	debouncer := utils_test.CreateDebouncer("x01234")
	service := services.NewDeviceLifetimeService(device, events, debouncer)
	payload := map[string]interface{}{"brigthness": 124.2, "last_seen": "2023-01-01T00:00:00Z"}

	service.Update(payload)
	time.Sleep(1 * time.Second)
	if device.Exposes["brigthness"].Data != 124.2 {
		t.Errorf("Device value not updated")
	}

	if device.Availability != devices.OnlineAvailability {
		t.Errorf("Device availability not updated")
	}
	if device.LastSeen != "2023-01-01T00:00:00Z" {
		t.Errorf("Device last seen not updated")
	}
}

to fix

func TestDeviceLifetimeService_UpdateWithDebouncer(t *testing.T) {

	device := utils_test.CreateLightDevice("x01234", "testDevice", "brigthness", 124.2)
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
		"brigthness": utils.IntervalFromSeconds(2),
	}
	appConfig.AddDeviceConfig(d1)

	debouncer := utils_test.CreateDebouncerFromAppConfig("x01234", appConfig, mockClock)
	service := services.NewDeviceLifetimeService(device, events, debouncer)

	testCases := []struct {
		payload      map[string]interface{}
		shouldUpdate bool
	}{
		{
			payload:      map[string]interface{}{"brigthness": 124.2, "last_seen": createLastSeen(11, 05, 10)},
			shouldUpdate: true,
		},
		{
			payload:      map[string]interface{}{"brigthness": 124.5, "last_seen": createLastSeen(11, 05, 11)},
			shouldUpdate: false,
		},
		{
			payload:      map[string]interface{}{"brigthness": 124.6, "last_seen": createLastSeen(11, 05, 13)},
			shouldUpdate: true,
		},
	}

	for idx, tc := range testCases {

		currValue := device.Exposes["brigthness"].Data
		lastSeen := tc.payload["last_seen"].(string)
		nextTimestamp, _ := time.Parse(time.RFC3339, lastSeen)
		mockClock.SetMockTime(nextTimestamp)

		service.Update(tc.payload)
		time.Sleep(50 * time.Millisecond)

		if tc.shouldUpdate {
			fmt.Printf("shouldUpdate=true idx %v last seen %v \n",idx, device.LastSeen)

			if device.Exposes["brigthness"].Data != tc.payload["brigthness"] {
				t.Errorf("Device value not updated. want %v, got %v", tc.payload["brigthness"], device.Exposes["brigthness"].Data)
			}
		} else {
			fmt.Printf("shouldUpdate=false idx %v last seen %v \n",idx, device.LastSeen)

			if device.Exposes["brigthness"].Data != currValue {
				t.Errorf("Device value not updated. want %v, got %v", currValue, device.Exposes["brigthness"].Data)
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
	debouncer := settings.NewDeviceDebouncer("testDevice", cache, mocks.NewMockClock(func() time.Time { return time.Now() }))

	service := services.NewDeviceLifetimeService(device, events, debouncer)
	payload := map[string]interface{}{"test": "data"}

	// make last seen 11 seconds ago as our availability timeout is 10 seconds
	device.LastSeen = time.Now().Add(-11 * time.Second).Format(time.RFC3339)

	service.Start(payload)
	wg.Wait()

	service.Dispose()

	if device.Availability != devices.OfflineAvailability {
		t.Errorf("Device availability not changed to offline")
	}
}

func createLastSeen(hour, minute, second int) string {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, now.Location()).Format(time.RFC3339)
}
