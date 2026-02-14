package services_test

import (
	"fmt"
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/settings"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"sync"
	"testing"
	"time"
)

func TestDeviceLifetimeService_Seed(t *testing.T) {
	wg := sync.WaitGroup{}
	device := utils_test.CreateDevice("x01234", "testDevice", "brightness", 124, 0.0, 255.0)
	device.Availability = devices.OnlineAvailability

	events := &devices.DeviceRequestEvents{
		OnNewDevice: func(d *devices.Device) {
			if d != device {
				t.Errorf("OnNewDevice device = %v, want %v", d, device)
				return
			}
			if d.Id != "x01234" {
				t.Errorf("OnNewDevice id = %v, want %v", d.Id, "x01234")
			}

			if d.Exposes["brightness"].Data.Value() != 10.2 {
				t.Errorf("OnNewDevice payload = %v, want %v", d.Exposes["brightness"].Data.Value(), 10.2)
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
	repo := mocks.NopSettingsrepo{}
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()

	cache := settings.NewDeviceConfigCache(&repo, app, &sync.RWMutex{})
	service := services.NewDeviceLifetimeService(device, events, cache, deviceQuerier, utils.NewRealClock())
	payload := map[string]interface{}{"brightness": 10.2}

	wg.Add(2)
	service.Seed(payload)

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
		}, OnDeviceMeasurementsUpdated: func(d *devices.Device, p map[string]interface{}) {

		},
	}

	// needs refacoring !!!!!!
	app := settings.NewAppConfig()
	d1 := settings.NewDeviceConfig("x01234")
	app.AddDeviceConfig(d1)
	repo := mocks.NopSettingsrepo{}
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()

	cache := settings.NewDeviceConfigCache(&repo, app, &sync.RWMutex{})
	service := services.NewDeviceLifetimeService(device, events, cache, deviceQuerier, utils.NewRealClock())
	payload := map[string]interface{}{"brightness": 35.4, "last_seen": "2023-01-01T00:00:00Z"}

	service.Update(payload)

	wg.Wait()
	if device.Exposes["brightness"].Data.Value() != 35.4 {
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
	repo := mocks.NopSettingsrepo{}
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()

	cache := settings.NewDeviceConfigCache(&repo, app, &sync.RWMutex{})

	service := services.NewDeviceLifetimeService(device, events, cache, deviceQuerier, utils.NewRealClock())
	payload := map[string]interface{}{"brightness": 124.2, "last_seen": "2023-01-01T00:00:00Z"}

	service.Update(payload)
	time.Sleep(1 * time.Second)
	if device.Exposes["brightness"].Data.Value() != 124.2 {
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
		OnDeviceMeasurementsUpdated: func(d *devices.Device, p map[string]interface{}) {

		},
	}
	now := time.Now()
	mockClock := mocks.NewMockClock(func() time.Time {
		return now
	})
	app := settings.NewAppConfig()
	d1 := settings.NewDeviceConfig("x01234")
	d1.DebounceOverrides = map[string]*utils.TimeInterval{
		"brightness": utils.IntervalFromSeconds(3),
	}
	app.AddDeviceConfig(d1)
	repo := mocks.NopSettingsrepo{}
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()

	cache := settings.NewDeviceConfigCache(&repo, app, &sync.RWMutex{})
	service := services.NewDeviceLifetimeService(device, events, cache, deviceQuerier, mockClock)

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

		currValue := device.Exposes["brightness"].Data.Value()
		mockClock.SetMockTime(tc.timestamp)

		service.Update(tc.payload)
		time.Sleep(50 * time.Millisecond)

		if tc.shouldUpdate {

			if device.Exposes["brightness"].Data.Value() != tc.payload["brightness"] {
				t.Errorf("Device value not updated. want %v, got %v", tc.payload["brightness"], device.Exposes["brightness"].Data.Value())
			}
		} else {

			if device.Exposes["brightness"].Data.Value() != currValue {
				t.Errorf("Error: Device value updated. want %v, got %v", currValue, device.Exposes["brightness"].Data.Value())
			}
		}
	}
}

func TestDeviceLifetimeService_ShouldChangeAvailability_ToOffline(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	device := &devices.Device{Id: "testDevice", Availability: devices.OnlineAvailability}

	events := &devices.DeviceRequestEvents{
		AvailabilityTimeout: 10,
		OnNewDevice: func(d *devices.Device) {
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
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()
	repo := mocks.NopSettingsrepo{}
	cache := settings.NewDeviceConfigCache(&repo, app, &sync.RWMutex{})
	service := services.NewDeviceLifetimeService(device, events, cache, deviceQuerier, mocks.NewMockClock(func() time.Time { return time.Now() }))
	payload := map[string]interface{}{"test": "data"}

	// make last_seen 11 seconds ago as our availability timeout is 10 seconds
	device.LastSeen = time.Now().Add(-11 * time.Second).Format(time.RFC3339)

	service.Seed(payload)
	wg.Wait()

	if device.Availability != devices.OfflineAvailability {
		t.Errorf("Device availability not changed to offline")
	}
}

func TestDeviceLifetimeService_MetricsAvailabilityWithMetricsEnabled(t *testing.T) {
	wg := sync.WaitGroup{}

	brightness := utils_test.CreateEntity("brightness", "number", 12.5)
	brightness.Category = bridge.MeasurementCategory
	battery := utils_test.CreateEntity("battery", "number", 80)
	battery.Category = bridge.DiagnosticCategory
	color_temp := utils_test.CreateEntity("color_temp", "number", 156)
	color_temp.Category = bridge.MeasurementCategory
	linkquality := utils_test.CreateEntity("linkquality", "number", 112)
	linkquality.Category = bridge.DiagnosticCategory

	device := utils_test.CreateDeviceWithExposes("x01234", "testDevice", []*devices.Entity{brightness, battery, color_temp, linkquality})
	device.Availability = devices.OfflineAvailability

	events := &devices.DeviceRequestEvents{
		AvailabilityTimeout: 30000,

		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			if d != device {
				t.Errorf("OnDeviceUpdated device = %v, want %v", d, device)
			}
			wg.Done()
		},
		OnDeviceMeasurementsUpdated: func(d *devices.Device, p map[string]interface{}) {

			_, ok1 := p["brightness"]
			_, ok2 := p["color_temp"]

			if !ok1 && !ok2 {
				t.Errorf("OnDeviceMetricsAvailable metrics = %v, want brightness and/or color_temp", p)
			}
			wg.Done()
		},
	}

	app := settings.NewAppConfig()
	d1 := settings.NewDeviceConfig("x01234")
	d1.MetricsEnabled = true
	d1.DebounceOverrides["battery"] = utils.IntervalFromMilliseconds(5000)
	d1.DebounceOverrides["linkquality"] = utils.IntervalFromMilliseconds(5000)

	app.AddDeviceConfig(d1)
	repo := mocks.NopSettingsrepo{}
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()

	cache := settings.NewDeviceConfigCache(&repo, app, &sync.RWMutex{})
	service := services.NewDeviceLifetimeService(device, events, cache, deviceQuerier, mocks.NewMockClock(func() time.Time { return time.Now() }))

	testCases := []struct {
		payload         map[string]interface{}
		expectedUpdates int
	}{
		{
			payload:         map[string]interface{}{"battery": 34, "brightness": 34.5},
			expectedUpdates: 2,
		},
		{
			payload:         map[string]interface{}{"battery": 23, "brightness": 120},
			expectedUpdates: 2,
		},
		{
			payload:         map[string]interface{}{"battery": 12, "brightness": 120}, // debounced and no changes
			expectedUpdates: 0,
		},
		{
			payload:         map[string]interface{}{"battery": 34, "linkquality": 12.5}, // 1st debounced and 2nd change
			expectedUpdates: 1,
		},
		{
			payload:         map[string]interface{}{"battery": 34, "linkquality": 12.5, "color_temp": 106.5}, //1 st changed, 2nd debounced and 3rd changed
			expectedUpdates: 2,
		},
		{
			payload:         map[string]interface{}{"battery": 34, "linkquality": 12.5, "color_temp": 12.5, "brightness": 12},
			expectedUpdates: 2,
		},
	}

	for id, tc := range testCases {

		payload := tc.payload

		if id == 2 {
			fmt.Println("")
		}
		if tc.expectedUpdates > 0 {
			wg.Add(tc.expectedUpdates)
		}
		service.Update(payload)

		if tc.expectedUpdates > 0 {
			wg.Wait()
		}
	}
}

func TestDeviceLifetimeService_NormalizesBinaryMeasurementValues(t *testing.T) {

	device := utils_test.CreateDoorSensorDevice("x0binary", "door sensor", false)

	var measurements map[string]interface{}
	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {},
		OnDeviceMeasurementsUpdated: func(d *devices.Device, p map[string]interface{}) {
			measurements = p
		},
	}

	app := settings.NewAppConfig()
	cfg := settings.NewDeviceConfig(device.Id)
	cfg.MetricsEnabled = true
	app.AddDeviceConfig(cfg)

	repo := mocks.NopSettingsrepo{}
	cache := settings.NewDeviceConfigCache(&repo, app, &sync.RWMutex{})
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()

	service := services.NewDeviceLifetimeService(device, events, cache, deviceQuerier, utils.NewRealClock())

	payload := map[string]interface{}{"contact": "open"}
	service.Update(payload)

	if measurements == nil {
		t.Fatalf("expected measurements to be emitted")
	}

	val, ok := measurements["contact"].(bool)
	if !ok {
		t.Fatalf("expected binary measurement to be bool, got %T", measurements["contact"])
	}

	if !val {
		t.Fatalf("expected contact measurement to be true")
	}
}

func TestDeviceLifetimeService_MetricsAvailabilityWithAutomationEnabled(t *testing.T) {
	wg := sync.WaitGroup{}

	brightness := utils_test.CreateEntity("brightness", "number", 12.5)
	brightness.Category = bridge.MeasurementCategory
	battery := utils_test.CreateEntity("battery", "number", 80)
	battery.Category = bridge.DiagnosticCategory
	color_temp := utils_test.CreateEntity("color_temp", "number", 156)
	color_temp.Category = bridge.MeasurementCategory
	linkquality := utils_test.CreateEntity("linkquality", "number", 112)
	linkquality.Category = bridge.DiagnosticCategory

	device := utils_test.CreateDeviceWithExposes("x01234", "testDevice", []*devices.Entity{brightness, battery, color_temp, linkquality})
	device.Availability = devices.OfflineAvailability

	events := &devices.DeviceRequestEvents{
		AvailabilityTimeout: 30000,

		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			if d != device {
				t.Errorf("OnDeviceUpdated device = %v, want %v", d, device)
			}
			wg.Done()
		},
		OnDeviceMeasurementsUpdated: func(d *devices.Device, p map[string]interface{}) {

			_, ok1 := p["brightness"]
			_, ok2 := p["color_temp"]

			if !ok1 && !ok2 {
				t.Errorf("OnDeviceMetricsAvailable metrics = %v, want brightness and/or color_temp", p)
			}
			wg.Done()
		},
	}

	app := settings.NewAppConfig()
	d1 := settings.NewDeviceConfig("x01234")
	d1.DebounceOverrides["battery"] = utils.IntervalFromMilliseconds(5000)
	d1.DebounceOverrides["linkquality"] = utils.IntervalFromMilliseconds(5000)

	app.AddDeviceConfig(d1)
	repo := mocks.NopSettingsrepo{}

	// enable automation for device so we can collect measurement data changes
	deviceQuerier := mocks.NewMockAutomationDeviceQuerierWithValues(map[string]bool{"x01234": true})
	cache := settings.NewDeviceConfigCache(&repo, app, &sync.RWMutex{})
	service := services.NewDeviceLifetimeService(device, events, cache, deviceQuerier, mocks.NewMockClock(func() time.Time { return time.Now() }))

	testCases := []struct {
		payload         map[string]interface{}
		expectedUpdates int
	}{
		{
			payload:         map[string]interface{}{"battery": 34, "brightness": 34.5},
			expectedUpdates: 2,
		},
		{
			payload:         map[string]interface{}{"battery": 23, "brightness": 120},
			expectedUpdates: 2,
		},
		{
			payload:         map[string]interface{}{"battery": 12, "brightness": 120}, // debounced and no changes
			expectedUpdates: 0,
		},
		{
			payload:         map[string]interface{}{"battery": 34, "linkquality": 12.5}, // 1st debounced and 2nd change
			expectedUpdates: 1,
		},
		{
			payload:         map[string]interface{}{"battery": 34, "linkquality": 12.5, "color_temp": 106.5}, //1 st changed, 2nd debounced and 3rd changed
			expectedUpdates: 2,
		},
		{
			payload:         map[string]interface{}{"battery": 34, "linkquality": 12.5, "color_temp": 12.5, "brightness": 12},
			expectedUpdates: 2,
		},
	}

	for _, tc := range testCases {

		payload := tc.payload

		if tc.expectedUpdates > 0 {
			wg.Add(tc.expectedUpdates)
		}
		service.Update(payload)

		if tc.expectedUpdates > 0 {
			wg.Wait()
		}
	}
}

func TestOnConfigUpdated_ShouldDisableDevice(t *testing.T) {
	wg := sync.WaitGroup{}

	device := utils_test.CreateDevice("x01234", "testDevice", "brightness", 124, 0.0, 255.0)
	device.Availability = devices.OnlineAvailability

	events := &devices.DeviceRequestEvents{
		OnNewDevice: func(d *devices.Device) {
			if d != device {
				t.Errorf("OnNewDevice device = %v, want %v", d, device)
				return
			}
			wg.Done()
		},
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			// we should only get here before the device config disabled is set
			wg.Done()

		},
		OnDeviceMeasurementsUpdated: func(d *devices.Device, p map[string]interface{}) {
			t.Errorf("OnDeviceMeasurementsUpdated should not be called")
		},
		OnDeviceAvailabilityChanged: func(p *devices.UpdatePackage) {
		},
	}

	app := settings.NewAppConfig()
	repo := mocks.NopSettingsrepo{}
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()

	cache := settings.NewDeviceConfigCache(&repo, app, &sync.RWMutex{})
	service := services.NewDeviceLifetimeService(device, events, cache, deviceQuerier, utils.NewRealClock())
	payload := map[string]interface{}{"brightness": 10.2}

	// one event for device added and one for device updated
	wg.Add(2)
	service.Seed(payload)

	time.Sleep(200 * time.Millisecond)
	service.Update(map[string]interface{}{"brightness": 12.5})
	wg.Wait()

	// we expect no events after disabled is true
	newConfig := settings.NewDeviceConfig("x01234")
	newConfig.Disabled = true
	service.OnConfigUpdated(newConfig)

	service.Update(map[string]interface{}{"brightness": 20.5})
	time.Sleep(200 * time.Millisecond)
}

func TestOnConfigUpdated_ShouldDisableDevice_OnStartUp(t *testing.T) {
	wg := sync.WaitGroup{}

	device := utils_test.CreateDevice("x01234", "testDevice", "brightness", 124, 0.0, 255.0)
	device.Availability = devices.OnlineAvailability

	events := &devices.DeviceRequestEvents{
		OnNewDevice: func(d *devices.Device) {
			if d != device {
				t.Errorf("OnNewDevice device = %v, want %v", d, device)
				return
			}
			wg.Done()
		},
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			// we should only get here before the device config disabled is set
			wg.Done()

		},
		OnDeviceMeasurementsUpdated: func(d *devices.Device, p map[string]interface{}) {
			t.Errorf("OnDeviceMeasurementsUpdated should not be called")
		},
		OnDeviceAvailabilityChanged: func(p *devices.UpdatePackage) {
		},
	}

	app := settings.NewAppConfig()

	newConfig := settings.NewDeviceConfig("x01234")
	newConfig.Disabled = true

	app.AddDeviceConfig(newConfig)
	repo := mocks.NopSettingsrepo{}
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()

	cache := settings.NewDeviceConfigCache(&repo, app, &sync.RWMutex{})
	service := services.NewDeviceLifetimeService(device, events, cache, deviceQuerier, utils.NewRealClock())
	payload := map[string]interface{}{"brightness": 10.2}

	// one event for device added and one for device updated
	service.Seed(payload)

	time.Sleep(200 * time.Millisecond)
	service.Update(map[string]interface{}{"brightness": 12.5})

	wg.Add(1)

	newConfig.Disabled = false
	service.OnConfigUpdated(newConfig)

	time.Sleep(200 * time.Millisecond)
	service.Update(map[string]interface{}{"brightness": 20.5})

	wg.Wait()
}

func createTimestamp(hour, minute, second int) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, now.Location())
}
