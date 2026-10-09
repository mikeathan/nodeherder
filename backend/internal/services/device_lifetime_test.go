package services_test

import (
	"maps"
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/settings"
	utils_test "node-herder/testing"
	"node-herder/utils"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestEntityValueAutomationReadSemantics(t *testing.T) {
	for _, tc := range []struct {
		name, property    string
		dataType          bridge.ExposeDataType
		initial           any
		values            []any
		triggers, updates int
	}{
		{"fresh state before debounce", "brightness", bridge.NumericDataType, float64(0), []any{float64(1), float64(2), float64(2)}, 2, 1},
		{"identical physical events", "action", bridge.EnumDataType, "single", []any{"single", "single"}, 2, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			device := devices.NewDevice(tc.name)
			device.SetAvailability(devices.OnlineAvailability)
			entity := devices.NewEntity(tc.property)
			entity.Type = tc.dataType
			entity.Data.SetValue(tc.initial)
			device.Exposes[tc.property] = entity
			app := settings.NewAppConfig()
			cfg := settings.NewDeviceConfig(device.Id)
			cfg.DebounceOverrides[tc.property] = utils.IntervalFromSeconds(10)
			app.AddDeviceConfig(cfg)
			cache := settings.NewDeviceConfigCache(&mocks.NopSettingsrepo{}, app, &sync.RWMutex{})
			triggers, updates := 0, 0
			var expected any
			events := &devices.DeviceRequestEvents{
				OnDeviceUpdated: func(*devices.Device, *devices.UpdatePackage) { updates++ },
				OnDeviceAutomationTriggered: func(d *devices.Device, payload map[string]interface{}) {
					triggers++
					if d.Exposes[tc.property].Data.Value() != expected || payload[tc.property] != expected {
						t.Error("automation did not receive fresh state/event")
					}
					// A callback may read/write the entity; no model lock may surround it.
					d.Exposes[tc.property].Data.SetValue(expected)
				},
			}
			clock := mocks.NewMockClock(func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) })
			service := services.NewDeviceLifetimeService(device, events, cache,
				mocks.NewMockAutomationDeviceQuerierWithValues(map[string]bool{device.Id: true}), clock)
			for _, value := range tc.values {
				expected = value
				service.Update(map[string]interface{}{tc.property: value})
			}
			if triggers != tc.triggers || updates != tc.updates {
				t.Fatalf("triggers/updates = %d/%d, want %d/%d", triggers, updates, tc.triggers, tc.updates)
			}
		})
	}
}

func TestAvailabilityMonitorConcurrentConfiguration(t *testing.T) {
	device := devices.NewDevice("monitor-concurrent")
	device.SetAvailability(devices.OnlineAvailability)
	device.SetLastSeen(time.Now().Format(time.RFC3339))
	cache := settings.NewDeviceConfigCache(&mocks.NopSettingsrepo{}, settings.NewAppConfig(), &sync.RWMutex{})
	service := services.NewDeviceLifetimeService(device, &devices.DeviceRequestEvents{
		AvailabilityTimeout:         time.Hour,
		OnDeviceAvailabilityChanged: func(*devices.UpdatePackage) { t.Error("unexpected expiry") },
	}, cache, mocks.NewMockAutomationDeviceQuerier(), utils.NewRealClock())
	enabled, disabled := settings.NewDeviceConfig(device.Id), settings.NewDeviceConfig(device.Id)
	disabled.Disabled = true
	t.Cleanup(func() { service.OnConfigUpdated(disabled) })
	start := make(chan struct{})
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for i := 0; i < 100; i++ {
				service.OnConfigUpdated(disabled)
				service.OnConfigUpdated(enabled)
			}
		}()
	}
	close(start)
	wg.Wait()
	service.OnConfigUpdated(disabled)
	service.OnConfigUpdated(disabled)
}

func TestAvailabilityMonitorCallbackCanDisableAndReenable(t *testing.T) {
	device := devices.NewDevice("monitor-reentry")
	device.SetAvailability(devices.OnlineAvailability)
	old := time.Now().Add(-time.Hour).Format(time.RFC3339)
	device.SetLastSeen(old)
	cache := settings.NewDeviceConfigCache(&mocks.NopSettingsrepo{}, settings.NewAppConfig(), &sync.RWMutex{})
	enabled, disabled := settings.NewDeviceConfig(device.Id), settings.NewDeviceConfig(device.Id)
	disabled.Disabled = true
	var service *services.DeviceLifetimeService
	var callbacks atomic.Int32
	finished := make(chan struct{}, 1)
	events := &devices.DeviceRequestEvents{AvailabilityTimeout: time.Second}
	events.OnDeviceAvailabilityChanged = func(p *devices.UpdatePackage) {
		if p.Availability != devices.OfflineAvailability {
			t.Errorf("unexpected availability: %v", p.Availability)
		}
		service.OnConfigUpdated(disabled) // must not wait for this callback
		switch callbacks.Add(1) {
		case 1:
			device.SetAvailability(devices.OnlineAvailability)
			device.SetLastSeen(old)
			service.OnConfigUpdated(enabled)
		case 2:
			finished <- struct{}{}
		default:
			t.Error("unexpected duplicate availability callback")
		}
	}
	service = services.NewDeviceLifetimeService(device, events, cache,
		mocks.NewMockAutomationDeviceQuerier(), utils.NewRealClock())
	t.Cleanup(func() { service.OnConfigUpdated(disabled) })
	service.OnConfigUpdated(disabled)
	service.OnConfigUpdated(enabled)
	select {
	case <-finished:
	case <-time.After(5 * time.Second):
		t.Fatal("monitor callback/re-enable blocked or replacement monitor was lost")
	}
	if callbacks.Load() != 2 {
		t.Fatalf("callbacks = %d, want 2", callbacks.Load())
	}
}

func TestAvailabilityMonitorResumesAfterOnlineUpdate(t *testing.T) {
	device := devices.NewDevice("monitor-reset")
	device.SetAvailability(devices.OnlineAvailability)
	old := time.Now().Add(-time.Hour).Format(time.RFC3339)
	device.SetLastSeen(old)
	cache := settings.NewDeviceConfigCache(&mocks.NopSettingsrepo{}, settings.NewAppConfig(), &sync.RWMutex{})
	changed := make(chan struct{}, 2)
	events := &devices.DeviceRequestEvents{
		AvailabilityTimeout: time.Second,
		OnDeviceAvailabilityChanged: func(p *devices.UpdatePackage) {
			if p.Availability != devices.OfflineAvailability {
				t.Errorf("unexpected availability: %v", p.Availability)
			}
			changed <- struct{}{}
		},
		OnDeviceUpdated: func(*devices.Device, *devices.UpdatePackage) {},
	}
	service := services.NewDeviceLifetimeService(device, events, cache,
		mocks.NewMockAutomationDeviceQuerier(), utils.NewRealClock())
	disabled := settings.NewDeviceConfig(device.Id)
	disabled.Disabled = true
	t.Cleanup(func() { service.OnConfigUpdated(disabled) })
	service.OnConfigUpdated(disabled)
	service.OnConfigUpdated(settings.NewDeviceConfig(device.Id))
	awaitOffline := func() {
		t.Helper()
		select {
		case <-changed:
		case <-time.After(5 * time.Second):
			t.Fatal("availability monitor did not expire")
		}
	}
	awaitOffline()
	service.Update(map[string]interface{}{"last_seen": old})
	if !device.IsAvailable() {
		t.Fatal("update did not bring device online")
	}
	awaitOffline() // Reset must resume the same monitor after its ticker stopped.
}

func TestDeviceLifetimeServiceLastSeenConcurrentRead(t *testing.T) {
	device := utils_test.CreateLightDevice("last-seen", "timestamp test", "brightness", -1.0)
	device.SetAvailability(devices.OnlineAvailability)
	initial := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	device.SetLastSeen(initial.Format(time.RFC3339))
	app := settings.NewAppConfig()
	app.Hub.Devices.Defaults.DefaultDebounceByCategory = nil
	repo := &mocks.NopSettingsrepo{}
	cache := settings.NewDeviceConfigCache(repo, app, &sync.RWMutex{})
	var expected string
	callbacks := 0
	events := &devices.DeviceRequestEvents{OnDeviceUpdated: func(_ *devices.Device, p *devices.UpdatePackage) {
		callbacks++
		if p.LastSeen != expected {
			t.Errorf("payload timestamp = %q, want %q", p.LastSeen, expected)
		}
		// Read from a callback too: the setter must not retain the device lock.
		if _, err := device.LastSeenTime(); err != nil {
			t.Error(err)
		}
	}}
	service := services.NewDeviceLifetimeService(device, events, cache,
		mocks.NewMockAutomationDeviceQuerier(), utils.NewRealClock())
	ready, stop, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	go func() {
		defer close(done)
		close(ready)
		for {
			select {
			case <-stop:
				return
			default:
				if _, err := device.LastSeenTime(); err != nil {
					t.Error(err)
					return
				}
			}
		}
	}()
	<-ready
	for i := 0; i < 200; i++ {
		expected = initial.Add(time.Duration(i) * time.Second).Format(time.RFC3339)
		service.Update(map[string]interface{}{"brightness": float64(i), "last_seen": expected})
	}
	close(stop)
	<-done
	if callbacks != 200 {
		t.Fatalf("updates emitted = %d, want 200", callbacks)
	}
	actual, err := device.LastSeenTime()
	if err != nil || actual.Format(time.RFC3339) != expected {
		t.Fatalf("final LastSeen = %v (%v), want %s", actual, err, expected)
	}
}

func TestDeviceLifetimeService_Seed(t *testing.T) {
	device := utils_test.CreateDevice("x01234", "testDevice", "brightness", 124, 0.0, 255.0)
	var added int
	events := &devices.DeviceRequestEvents{
		AvailabilityTimeout: time.Hour,
		OnNewDevice: func(d *devices.Device) {
			added++
			if d != device || d.Id != "x01234" {
				t.Errorf("OnNewDevice returned unexpected device: %v", d)
			}
			if got := d.Exposes["brightness"].Data.Value(); got != 10.2 {
				t.Errorf("seed value = %v, want 10.2", got)
			}
			if got := d.GetAvailability(); got != devices.OnlineAvailability {
				t.Errorf("seed availability = %v, want online", got)
			}
		},
	}
	cache := settings.NewDeviceConfigCache(&mocks.NopSettingsrepo{}, settings.NewAppConfig(), &sync.RWMutex{})
	service := services.NewDeviceLifetimeService(device, events, cache,
		mocks.NewMockAutomationDeviceQuerier(), utils.NewRealClock())
	stopLifetimeAfterTest(t, service, device.Id)
	service.Seed(map[string]interface{}{"brightness": 10.2})
	if added != 1 {
		t.Fatalf("OnNewDevice calls = %d, want 1", added)
	}
}

// Stop through the existing configuration API; lifecycle joining belongs to T014.
func stopLifetimeAfterTest(t *testing.T, service *services.DeviceLifetimeService, id string) {
	t.Helper()
	t.Cleanup(func() {
		cfg := settings.NewDeviceConfig(id)
		cfg.Disabled = true
		service.OnConfigUpdated(cfg)
	})
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
	t.Cleanup(func() {
		disabled := settings.NewDeviceConfig(device.Id)
		disabled.Disabled = true
		service.OnConfigUpdated(disabled)
	})

	service.Update(payload)

	wg.Wait()
	if device.Exposes["brightness"].Data.Value() != 35.4 {
		t.Errorf("Device value not updated")
	}

	if device.GetAvailability() != devices.OnlineAvailability {
		t.Errorf("Device availability not updated")
	}
	if device.LastSeen != "2023-01-01T00:00:00Z" {
		t.Errorf("Device last seen not updated")
	}
}

func TestDeviceLifetimeService_UpdateWithSameData(t *testing.T) {

	device := utils_test.CreateLightDevice("x01234", "testDevice", "brightness", 124.2)
	device.Availability = devices.OnlineAvailability

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
	if device.Exposes["brightness"].Data.Value() != 124.2 {
		t.Errorf("Device value not updated")
	}

	if device.GetAvailability() != devices.OnlineAvailability {
		t.Errorf("Device availability not updated")
	}
	if device.LastSeen != "2023-01-01T00:00:00Z" {
		t.Errorf("Device last seen not updated")
	}
}

func TestDeviceLifetimeService_UpdateWithDebouncer(t *testing.T) {

	device := utils_test.CreateLightDevice("x01234", "testDevice", "brightness", 124.1)
	device.Availability = devices.OnlineAvailability

	var updates int
	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			updates++
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
		updates = 0
		mockClock.SetMockTime(tc.timestamp)

		service.Update(tc.payload)

		if device.Exposes["brightness"].Data.Value() != tc.payload["brightness"] {
			t.Errorf("Device value not updated (should always be fresh). want %v, got %v", tc.payload["brightness"], device.Exposes["brightness"].Data.Value())
		}

		expectedUpdates := 0
		if tc.shouldUpdate {
			expectedUpdates = 1
		}
		if updates != expectedUpdates {
			t.Errorf("at time %v, expected %d updates, got %d", tc.timestamp, expectedUpdates, updates)
		}
	}
}

func TestDeviceLifetimeService_ShouldChangeAvailability_ToOffline(t *testing.T) {
	device := &devices.Device{Id: "testDevice", Availability: devices.OnlineAvailability}
	offline := make(chan *devices.UpdatePackage, 1)
	added := 0
	events := &devices.DeviceRequestEvents{
		AvailabilityTimeout: 10 * time.Second,
		OnNewDevice: func(d *devices.Device) {
			added++
			if got := d.GetAvailability(); got != devices.OnlineAvailability {
				t.Errorf("seed availability = %v, want online", got)
			}
		},
		OnDeviceAvailabilityChanged: func(p *devices.UpdatePackage) { offline <- p },
	}
	cache := settings.NewDeviceConfigCache(&mocks.NopSettingsrepo{}, settings.NewAppConfig(), &sync.RWMutex{})
	service := services.NewDeviceLifetimeService(device, events, cache,
		mocks.NewMockAutomationDeviceQuerier(), utils.NewRealClock())
	stopLifetimeAfterTest(t, service, device.Id)
	// Seed reads last_seen from its payload; assigning the device field first is
	// ineffective because Seed overwrites it. Keep the real one-second monitor tick.
	service.Seed(map[string]interface{}{
		"last_seen": time.Now().Add(-11 * time.Second).Format(time.RFC3339),
	})
	if added != 1 {
		t.Fatalf("OnNewDevice calls = %d, want 1", added)
	}
	select {
	case p := <-offline:
		if p.Id != device.Id || p.Availability != devices.OfflineAvailability {
			t.Fatalf("offline update = %+v", p)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("availability monitor did not report offline")
	}
	if got := device.GetAvailability(); got != devices.OfflineAvailability {
		t.Errorf("availability = %v, want offline", got)
	}
}

func TestDeviceLifetimeService_MetricsAvailabilityWithMetricsEnabled(t *testing.T) {
	var updatesCount int
	var measurementsCount int

	brightness := utils_test.CreateEntity("brightness", "number", 12.5)
	brightness.Category = bridge.MeasurementCategory
	battery := utils_test.CreateEntity("battery", "number", 80)
	battery.Category = bridge.DiagnosticCategory
	color_temp := utils_test.CreateEntity("color_temp", "number", 156)
	color_temp.Category = bridge.MeasurementCategory
	linkquality := utils_test.CreateEntity("linkquality", "number", 112)
	linkquality.Category = bridge.DiagnosticCategory

	device := utils_test.CreateDeviceWithExposes("x01234", "testDevice", []*devices.Entity{brightness, battery, color_temp, linkquality})
	device.Availability = devices.OnlineAvailability

	events := &devices.DeviceRequestEvents{
		AvailabilityTimeout: 30000,

		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			if d != device {
				t.Errorf("OnDeviceUpdated device = %v, want %v", d, device)
			}
			updatesCount++
		},
		OnDeviceMeasurementsUpdated: func(d *devices.Device, p map[string]interface{}) {

			_, ok1 := p["brightness"]
			_, ok2 := p["color_temp"]

			if !ok1 && !ok2 {
				t.Errorf("OnDeviceMetricsAvailable metrics = %v, want brightness and/or color_temp", p)
			}
			measurementsCount++
		},
	}

	now := time.Now()
	mockClock := mocks.NewMockClock(func() time.Time {
		return now
	})

	app := settings.NewAppConfig()
	d1 := settings.NewDeviceConfig("x01234")
	d1.MetricsEnabled = true
	d1.DebounceOverrides["battery"] = utils.IntervalFromMilliseconds(5000)
	d1.DebounceOverrides["linkquality"] = utils.IntervalFromMilliseconds(5000)
	d1.DebounceOverrides["brightness"] = utils.IntervalFromMilliseconds(10)
	d1.DebounceOverrides["color_temp"] = utils.IntervalFromMilliseconds(10)

	app.AddDeviceConfig(d1)
	repo := mocks.NopSettingsrepo{}
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()

	cache := settings.NewDeviceConfigCache(&repo, app, &sync.RWMutex{})
	service := services.NewDeviceLifetimeService(device, events, cache, deviceQuerier, mockClock)

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

		// Advance time 100ms between test cases so the 10ms brightness/color_temp
		// debounce expires, but the 5s battery/linkquality debounce stays active
		now = now.Add(100 * time.Millisecond)
		mockClock.SetMockTime(now)

		updatesCount = 0
		measurementsCount = 0

		service.Update(payload)

		totalFired := updatesCount + measurementsCount
		if totalFired != tc.expectedUpdates {
			t.Errorf("For payload %v, expected %d updates, got %d", payload, tc.expectedUpdates, totalFired)
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

	now := time.Now()
	mockClock := mocks.NewMockClock(func() time.Time {
		return now
	})

	app := settings.NewAppConfig()
	d1 := settings.NewDeviceConfig("x01234")
	d1.MetricsEnabled = true
	d1.DebounceOverrides["battery"] = utils.IntervalFromMilliseconds(5000)
	d1.DebounceOverrides["linkquality"] = utils.IntervalFromMilliseconds(5000)
	d1.DebounceOverrides["brightness"] = utils.IntervalFromMilliseconds(10)
	d1.DebounceOverrides["color_temp"] = utils.IntervalFromMilliseconds(10)

	app.AddDeviceConfig(d1)
	repo := mocks.NopSettingsrepo{}

	// enable automation for device so we can collect measurement data changes
	deviceQuerier := mocks.NewMockAutomationDeviceQuerierWithValues(map[string]bool{"x01234": true})
	cache := settings.NewDeviceConfigCache(&repo, app, &sync.RWMutex{})
	service := services.NewDeviceLifetimeService(device, events, cache, deviceQuerier, mockClock)

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

		// Advance time 100ms between test cases so the 10ms brightness/color_temp
		// debounce expires, but the 5s battery/linkquality debounce stays active
		now = now.Add(100 * time.Millisecond)
		mockClock.SetMockTime(now)

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
	testDeviceDisableTransition(t, false)
}

func TestOnConfigUpdated_ShouldDisableDevice_OnStartUp(t *testing.T) {
	testDeviceDisableTransition(t, true)
}

func testDeviceDisableTransition(t *testing.T, initiallyDisabled bool) {
	t.Helper()
	device := utils_test.CreateDevice("x01234", "testDevice", "brightness", 124.0, 0.0, 255.0)
	device.SetAvailability(devices.OnlineAvailability)
	added, updated := 0, 0
	events := &devices.DeviceRequestEvents{
		AvailabilityTimeout: time.Hour,
		OnNewDevice:         func(*devices.Device) { added++ },
		OnDeviceUpdated:     func(*devices.Device, *devices.UpdatePackage) { updated++ },
		OnDeviceMeasurementsUpdated: func(*devices.Device, map[string]interface{}) {
			t.Error("metrics are disabled")
		},
	}
	app := settings.NewAppConfig()
	cfg := settings.NewDeviceConfig(device.Id)
	cfg.Disabled = initiallyDisabled
	cfg.DebounceOverrides["brightness"] = utils.IntervalFromSeconds(0)
	app.AddDeviceConfig(cfg)
	cache := settings.NewDeviceConfigCache(&mocks.NopSettingsrepo{}, app, &sync.RWMutex{})
	service := services.NewDeviceLifetimeService(device, events, cache,
		mocks.NewMockAutomationDeviceQuerier(), utils.NewRealClock())
	stopLifetimeAfterTest(t, service, device.Id)
	service.Seed(map[string]interface{}{"brightness": 10.2})
	service.Update(map[string]interface{}{"brightness": 12.5})
	wantAdded, wantUpdated, wantValue := 1, 1, 12.5
	if initiallyDisabled {
		wantAdded, wantUpdated, wantValue = 0, 0, 124.0
	}
	if added != wantAdded || updated != wantUpdated || device.Exposes["brightness"].Data.Value() != wantValue {
		t.Fatalf("before config: added/updated/value = %d/%d/%v, want %d/%d/%v",
			added, updated, device.Exposes["brightness"].Data.Value(), wantAdded, wantUpdated, wantValue)
	}
	changed := settings.NewDeviceConfig(device.Id)
	changed.Disabled = !initiallyDisabled
	service.OnConfigUpdated(changed)
	service.Update(map[string]interface{}{"brightness": 20.5})
	if initiallyDisabled {
		wantUpdated, wantValue = 1, 20.5
	}
	if added != wantAdded || updated != wantUpdated || device.Exposes["brightness"].Data.Value() != wantValue {
		t.Fatalf("after config: added/updated/value = %d/%d/%v, want %d/%d/%v",
			added, updated, device.Exposes["brightness"].Data.Value(), wantAdded, wantUpdated, wantValue)
	}
}

func createTimestamp(hour, minute, second int) time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, now.Location())
}

func TestDeviceLifetimeService_UnchangedValuesNeverConsumeDebounce(t *testing.T) {
	// This test asserts that if the same value is sent repeatedly, it never
	// triggers an update — even after the debounce window has expired.
	// This verifies the ordering: value-check BEFORE debounce.

	updateCount := 0

	device := utils_test.CreateLightDevice("x01234", "testDevice", "brightness", 50.0)
	device.Availability = devices.OnlineAvailability

	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			updateCount++
		},
		OnDeviceMeasurementsUpdated: func(d *devices.Device, p map[string]interface{}) {},
	}

	now := time.Now()
	mockClock := mocks.NewMockClock(func() time.Time {
		return now
	})

	app := settings.NewAppConfig()
	d1 := settings.NewDeviceConfig("x01234")
	d1.DebounceOverrides = map[string]*utils.TimeInterval{
		"brightness": utils.IntervalFromSeconds(5),
	}
	app.AddDeviceConfig(d1)
	repo := mocks.NopSettingsrepo{}
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()

	cache := settings.NewDeviceConfigCache(&repo, app, &sync.RWMutex{})
	service := services.NewDeviceLifetimeService(device, events, cache, deviceQuerier, mockClock)

	// Send the same value 50.0 many times, advancing past the debounce window each time
	for i := 0; i < 10; i++ {
		now = now.Add(6 * time.Second) // well past the 5s debounce
		mockClock.SetMockTime(now)
		service.Update(map[string]interface{}{"brightness": 50.0})
	}

	if updateCount != 0 {
		t.Errorf("Expected 0 updates for unchanged values, got %d", updateCount)
	}

	// Now send an actually changed value — it SHOULD trigger an update
	now = now.Add(6 * time.Second)
	mockClock.SetMockTime(now)
	service.Update(map[string]interface{}{"brightness": 75.0})

	if updateCount != 1 {
		t.Errorf("Expected exactly 1 update after real value change, got %d", updateCount)
	}
}

func TestDeviceLifetimeService_NoisyValuesDebouncedCorrectly(t *testing.T) {
	// This test simulates a noisy sensor (like linkquality bouncing between 80 and 81)
	// and asserts that the debouncer correctly throttles even when values change.

	updateCount := 0

	brightness := utils_test.CreateEntity("brightness", "number", 80.0)
	brightness.Category = bridge.MeasurementCategory

	device := utils_test.CreateDeviceWithExposes("x01234", "testDevice", []*devices.Entity{brightness})
	device.Availability = devices.OnlineAvailability

	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {
			updateCount++
		},
		OnDeviceMeasurementsUpdated: func(d *devices.Device, p map[string]interface{}) {},
	}

	now := time.Now()
	mockClock := mocks.NewMockClock(func() time.Time {
		return now
	})

	app := settings.NewAppConfig()
	d1 := settings.NewDeviceConfig("x01234")
	d1.DebounceOverrides = map[string]*utils.TimeInterval{
		"brightness": utils.IntervalFromSeconds(60),
	}
	app.AddDeviceConfig(d1)
	repo := mocks.NopSettingsrepo{}
	deviceQuerier := mocks.NewMockAutomationDeviceQuerier()

	cache := settings.NewDeviceConfigCache(&repo, app, &sync.RWMutex{})
	service := services.NewDeviceLifetimeService(device, events, cache, deviceQuerier, mockClock)

	// First update: value changes from 80 -> 81, should go through (first event)
	service.Update(map[string]interface{}{"brightness": 81.0})
	if updateCount != 1 {
		t.Fatalf("Expected 1 update after first change, got %d", updateCount)
	}

	// Noisy updates every 10s with alternating values — all should be debounced
	values := []float64{80.0, 81.0, 80.0, 81.0, 80.0}
	for _, v := range values {
		now = now.Add(10 * time.Second)
		mockClock.SetMockTime(now)
		service.Update(map[string]interface{}{"brightness": v})
	}

	// Only the initial update should have gone through; the rest are debounced
	if updateCount != 1 {
		t.Errorf("Expected 1 total update (noisy values should be debounced), got %d", updateCount)
	}

	// After 60s, the next changed value should go through
	now = now.Add(61 * time.Second)
	mockClock.SetMockTime(now)
	service.Update(map[string]interface{}{"brightness": 82.0})

	if updateCount != 2 {
		t.Errorf("Expected 2 total updates after debounce window expired, got %d", updateCount)
	}
}

func TestDeviceLifetimeService_AutomationTriggersOnChanges(t *testing.T) {
	device := utils_test.CreateDoorSensorDevice("x01234", "testDoor", false)
	device.Availability = devices.OnlineAvailability

	var triggers int
	events := &devices.DeviceRequestEvents{
		OnDeviceUpdated: func(d *devices.Device, p *devices.UpdatePackage) {},
		OnDeviceAutomationTriggered: func(d *devices.Device, p map[string]interface{}) {
			triggers++
		},
	}

	app := settings.NewAppConfig()
	d1 := settings.NewDeviceConfig("x01234")
	app.AddDeviceConfig(d1)
	repo := mocks.NopSettingsrepo{}

	// Mock querier: automation is enabled
	deviceQuerier := mocks.NewMockAutomationDeviceQuerierWithValues(map[string]bool{"x01234": true})
	cache := settings.NewDeviceConfigCache(&repo, app, &sync.RWMutex{})
	mockClock := mocks.NewMockClock(func() time.Time { return time.Now() })

	service := services.NewDeviceLifetimeService(device, events, cache, deviceQuerier, mockClock)

	// Update 1: Should trigger automation
	service.Update(map[string]interface{}{"contact": true})

	// Update 2: Immediately after, should trigger again (no cooldown)
	service.Update(map[string]interface{}{"contact": false})

	// Update 3: Still immediately after, should trigger again
	service.Update(map[string]interface{}{"contact": true})

	if triggers != 3 {
		t.Fatalf("Expected exactly 3 triggers (no cooldown), got %d", triggers)
	}
}

// Characterize the distinct seed/update contracts before extracting their output
// handling: seed supplies the original payload, updates supply changed exposes,
// and automation precedes sampled metrics/UI even inside a debounce window.
func TestDeviceLifetimeService_SeedAndUpdateOutputSequence(t *testing.T) {
	brightness := utils_test.CreateEntity("brightness", "numeric", 0.0)
	brightness.Category = bridge.MeasurementCategory
	battery := utils_test.CreateEntity("battery", "numeric", 0.0)
	battery.Category = bridge.DiagnosticCategory
	device := utils_test.CreateDeviceWithExposes("baseline", "baseline", []*devices.Entity{brightness, battery})
	app := settings.NewAppConfig()
	cfg := settings.NewDeviceConfig(device.Id)
	cfg.MetricsEnabled = true
	for _, name := range []string{"brightness", "battery"} {
		cfg.DebounceOverrides[name] = utils.IntervalFromSeconds(10)
	}
	app.AddDeviceConfig(cfg)
	cache := settings.NewDeviceConfigCache(&mocks.NopSettingsrepo{}, app, &sync.RWMutex{})
	clock := mocks.NewMockClock(func() time.Time {
		return time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	})
	type output struct {
		stage string
		data  map[string]interface{}
	}
	var got []output
	record := func(stage string, data map[string]interface{}) {
		got = append(got, output{stage, maps.Clone(data)})
	}
	expectedBrightness := 10.0
	events := &devices.DeviceRequestEvents{
		AvailabilityTimeout: time.Hour,
		OnDeviceAutomationTriggered: func(d *devices.Device, data map[string]interface{}) {
			if value := d.Exposes["brightness"].Data.Value(); value != expectedBrightness {
				t.Errorf("automation read %v, want fresh value %v", value, expectedBrightness)
			}
			record("automation", data)
		},
		OnDeviceMeasurementsUpdated: func(_ *devices.Device, data map[string]interface{}) { record("metrics", data) },
		OnNewDevice:                 func(*devices.Device) { record("added", nil) },
		OnDeviceUpdated: func(_ *devices.Device, p *devices.UpdatePackage) {
			record("updated", p.Data)
		},
	}
	service := services.NewDeviceLifetimeService(device, events, cache,
		mocks.NewMockAutomationDeviceQuerierWithValues(map[string]bool{device.Id: true}), clock)
	stopLifetimeAfterTest(t, service, device.Id)
	seed := map[string]interface{}{"brightness": 10.0, "battery": 80.0, "unrecognized": "ignored"}
	service.Seed(seed)

	clock.Advance(time.Second)
	expectedBrightness = 11.0
	service.Update(map[string]interface{}{"brightness": 11.0, "battery": 80.0, "unrecognized": "ignored"})
	// Identical state must produce no automation, metrics or UI output.
	service.Update(map[string]interface{}{"brightness": 11.0, "battery": 80.0})
	clock.Advance(10 * time.Second)
	expectedBrightness = 12.0
	service.Update(map[string]interface{}{"brightness": 12.0, "battery": 80.0})

	want := []output{
		{"automation", seed},
		{"metrics", map[string]interface{}{"brightness": 10.0}},
		{"added", nil},
		{"automation", map[string]interface{}{"brightness": 11.0}},
		{"automation", map[string]interface{}{"brightness": 12.0}},
		{"metrics", map[string]interface{}{"brightness": 12.0}},
		{"updated", map[string]interface{}{"brightness": 12.0}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("output sequence:\n got: %#v\nwant: %#v", got, want)
	}
}
