package settings_test

import (
	"node-herder/mocks"
	"node-herder/models/bridge"
	"node-herder/models/settings"
	"node-herder/utils"
	"testing"
	"time"
)

func TestNewDeviceConfigCache(t *testing.T) {

	appConfig := settings.NewAppConfig()

	d1 := settings.NewDeviceConfig("device1")
	d1.DebounceOverrides = map[string]*utils.TimeInterval{
		"expose1": utils.IntervalFromMilliseconds(1000),
		"expose2": utils.IntervalFromMilliseconds(2000),
	}

	appConfig.AddDeviceConfig(d1)
	d2 := settings.NewDeviceConfig("device2")
	d2.DebounceOverrides = map[string]*utils.TimeInterval{
		"expose3": utils.IntervalFromMilliseconds(3000),
		"expose4": utils.IntervalFromMilliseconds(4000),
	}
	appConfig.AddDeviceConfig(d2)

	repo := mocks.NopSettingsrepo{}
	cache := settings.NewDeviceConfigCache(&repo, appConfig)

	if cache.Size() != 2 {
		t.Errorf("Expected cache to have 2 devices, got %d", cache.Size())
	}

	expose1Debounce, ok := cache.GetDebounce("device1", "expose1", bridge.MeasurementCategory)
	if !ok {
		t.Errorf("Expected expose1 debounce to be found, got %v", ok)
	}
	if expose1Debounce != 1000*time.Millisecond {
		t.Errorf("Expected expose1 debounce to be 1000ms, got %v", expose1Debounce)
	}
	expose2Debounce, ok := cache.GetDebounce("device1", "expose2", bridge.MeasurementCategory)
	if !ok {
		t.Errorf("Expected expose2 debounce to be found, got %v", ok)
	}
	if expose2Debounce != 2000*time.Millisecond {
		t.Errorf("Expected expose2 debounce to be 2000ms, got %v", expose2Debounce)
	}

	expose3Debounce, ok := cache.GetDebounce("device2", "expose3", bridge.MeasurementCategory)
	if !ok {
		t.Errorf("Expected expose3 debounce to be found, got %v", ok)
	}
	if expose3Debounce != 3000*time.Millisecond {
		t.Errorf("Expected expose3 debounce to be 3000ms, got %v", expose3Debounce)
	}
	expose4Debounce, ok := cache.GetDebounce("device2", "expose4", bridge.MeasurementCategory)
	if !ok {
		t.Errorf("Expected expose4 debounce to be found, got %v", ok)
	}
	if expose4Debounce != 4000*time.Millisecond {
		t.Errorf("Expected expose4 debounce to be 4000ms, got %v", expose4Debounce)
	}
}

func TestDeviceConfigCache_Get(t *testing.T) {
	repo := mocks.NopSettingsrepo{}

	cache := settings.NewDeviceConfigCache(&repo, settings.NewAppConfig())
	device1 := settings.NewDeviceConfig("device1")
	cache.Set(device1)

	device, err := cache.Get("device1")
	if err != nil || device == nil {
		t.Errorf("Expected to get device1, got nil")
	}

	_, err = cache.Get("device2")
	if err != nil {
		t.Errorf("Expected to not get device2")
	}
}

func TestDeviceConfigCache_Set(t *testing.T) {

	repo := mocks.NopSettingsrepo{}
	appConfig := settings.NewAppConfig()
	device1 := settings.NewDeviceConfig("device1")
	device1.DebounceOverrides = map[string]*utils.TimeInterval{
		"expose1": utils.IntervalFromMilliseconds(1000),
	}
	appConfig.AddDeviceConfig(device1)
	cache := settings.NewDeviceConfigCache(&repo, appConfig)

	// add a second debounce to device1 after initialization
	device1.DebounceOverrides["expose2"] = utils.IntervalFromMilliseconds(2000)

	cache.Set(device1)
	debounce, ok := cache.GetDebounce("device1", "expose1", bridge.MeasurementCategory)
	if !ok {
		t.Errorf("Expected expose1 debounce to be set")
	}
	if debounce != 1*time.Second {
		t.Errorf("Expected expose1 debounce to be set")
	}
	debounce, ok = cache.GetDebounce("device1", "expose2", bridge.MeasurementCategory)
	if !ok {
		t.Errorf("Expected expose2 debounce to be set")
	}
	if debounce != 2*time.Second {
		t.Errorf("Expected expose2 debounce to be set")
	}
}

func TestDeviceConfigCache_DeleteDebounce(t *testing.T) {
	repo := mocks.NopSettingsrepo{}
	app := settings.NewAppConfig()
	device1 := settings.NewDeviceConfig("device1")
	device1.DebounceOverrides = map[string]*utils.TimeInterval{
		"expose1": utils.IntervalFromMilliseconds(1000),
		"expose2": utils.IntervalFromMilliseconds(2000),
	}
	app.AddDeviceConfig(device1)

	cache := settings.NewDeviceConfigCache(&repo, app)
	cache.Set(device1)

	// assert that the debounce is set
	debounce, ok := cache.GetDebounce("device1", "expose1", bridge.MeasurementCategory)
	if !ok {
		t.Errorf("Expected expose1 debounce to be set")
	}
	if debounce != 1*time.Second {
		t.Errorf("Expected expose1 debounce to be set")
	}
	debounce, ok = cache.GetDebounce("device1", "expose2", bridge.MeasurementCategory)
	if !ok {
		t.Errorf("Expected expose2 debounce to be set")
	}
	if debounce != 2*time.Second {
		t.Errorf("Expected expose2 debounce to be set")
	}

	// delete the deebounce
	ok = cache.DeleteDebounce("device1", "expose2")
	if !ok {
		t.Errorf("Expected expose1 debounce to be deleted")
	}

	// assert that the expose1 debounce is deleted
	_, ok = cache.GetDebounce("device1", "expose2", bridge.MeasurementCategory)
	if ok {
		t.Errorf("Expected expose2 debounce to be deleted")
	}

	// assert that the expose2 debounce is still set
	_, ok = cache.GetDebounce("device1", "expose1", bridge.MeasurementCategory)
	if !ok {
		t.Errorf("Expected expose1 debounce to be set")
	}
}

func TestDeviceDebouncer_DebounceExpose(t *testing.T) {
	repo := mocks.NopSettingsrepo{}
	now := time.Now()
	mockClock := mocks.NewMockClock(func() time.Time {
		return now
	})

	appConfig := settings.NewAppConfig()

	d1 := settings.NewDeviceConfig("device1")
	d1.DebounceOverrides = map[string]*utils.TimeInterval{
		"expose1": utils.IntervalFromSeconds(1),
		"expose2": utils.IntervalFromSeconds(2),
	}

	appConfig.AddDeviceConfig(d1)
	d2 := settings.NewDeviceConfig("device2")
	d2.DebounceOverrides = map[string]*utils.TimeInterval{
		"expose3": utils.IntervalFromSeconds(3),
		"expose4": utils.IntervalFromSeconds(4),
	}

	appConfig.AddDeviceConfig(d2)

	cache := settings.NewDeviceConfigCache(&repo, appConfig)

	debouncer := settings.NewDeviceDebouncer("device1", cache, mockClock)

	// Test First Event
	if debouncer.DebounceExpose("expose1", bridge.MeasurementCategory) == true { // First event should not be debounced
		t.Error("First event should not be debounced")
	}

	// Test Debounced Event (within duration)
	now = now.Add(500 * time.Millisecond)
	mockClock.SetMockTime(now)
	if debouncer.DebounceExpose("expose1", bridge.MeasurementCategory) == false { // Event within duration should be debounced
		t.Error("Event within duration should be debounced")
	}

	// Test After Duration
	now = now.Add(500 * time.Millisecond)
	mockClock.SetMockTime(now)
	if debouncer.DebounceExpose("expose1", bridge.MeasurementCategory) == true { // Event after duration should not be debounced
		t.Error("Event after duration should not be debounced")
	}

	// Test No Debounce Config
	if debouncer.DebounceExpose("expose3", bridge.MeasurementCategory) {
		t.Error("Event with no debounce config should not be debounced")
	}
	//Test multiple exposes.
	if debouncer.DebounceExpose("expose2", bridge.MeasurementCategory) {
		t.Error("First expose2 event should not be debounced")
	}

	now = now.Add(1 * time.Second)
	mockClock.SetMockTime(now)

	if !debouncer.DebounceExpose("expose2", bridge.MeasurementCategory) {
		t.Error("expose2 event within duration should be debounced")
	}

	// second device
	now2 := time.Now()
	mockClock.SetMockTime(now2)
	debouncer2 := settings.NewDeviceDebouncer("device2", cache, mockClock)

	if debouncer2.DebounceExpose("expose4", bridge.MeasurementCategory) {
		t.Error("First expose4 event should not be debounced")
	}

	now2 = now2.Add(3 * time.Second)
	mockClock.SetMockTime(now2)

	if !debouncer2.DebounceExpose("expose4", bridge.MeasurementCategory) {
		t.Error("Second expose4 event should be debounced")
	}
}

func TestDeviceDebouncer_DiagnosticsDebouncerWhenOverrideIsNotAvailable(t *testing.T) {

	repo := mocks.NopSettingsrepo{}
	now := time.Now()
	mockClock := mocks.NewMockClock(func() time.Time {
		return now
	})

	appConfig := settings.NewAppConfig()

	d1 := settings.NewDeviceConfig("device1")
	d1.DebounceOverrides = map[string]*utils.TimeInterval{
		"expose1": utils.IntervalFromSeconds(1),
	}
	appConfig.Hub.Devices.Default.DefaultDebounceByCategory[bridge.DiagnosticCategory] = utils.IntervalFromSeconds(5)

	appConfig.AddDeviceConfig(d1)
	cache := settings.NewDeviceConfigCache(&repo, appConfig)
	debouncer := settings.NewDeviceDebouncer("device1", cache, mockClock)

	// Test expose 1 event with debounce overrides
	// ############################################################################
	if debouncer.DebounceExpose("expose1", bridge.MeasurementCategory) == true { // First event should not be debounced
		t.Error("First expose event should not be debounced")
	}

	now = now.Add(100 * time.Millisecond)
	mockClock.SetMockTime(now)
	if debouncer.DebounceExpose("expose1", bridge.MeasurementCategory) == false {
		t.Error("Second expose event should be debounced")
	}

	now = now.Add(1 * time.Second)
	mockClock.SetMockTime(now)

	if debouncer.DebounceExpose("expose1", bridge.MeasurementCategory) == true {
		t.Error("Third expose event should be not debounced")
	}

	// now test diagnostics expose with no override but using default debounce
	// ############################################################################
	if debouncer.DebounceExpose("expose2", bridge.DiagnosticCategory) == true {
		t.Error("First expose2 event should not be debounced")
	}
	now = now.Add(2 * time.Second)
	mockClock.SetMockTime(now)

	if debouncer.DebounceExpose("expose2", bridge.DiagnosticCategory) == false {
		t.Error("Second expose2 event should be debounced")
	}

	now = now.Add(4 * time.Second)
	mockClock.SetMockTime(now)

	if debouncer.DebounceExpose("expose2", bridge.DiagnosticCategory) == true {
		t.Error("Third expose2 event should not be debounced")
	}

	// now test non-diagnostics expose with no override, it should not be debounced
	// ############################################################################
	if debouncer.DebounceExpose("expose3", bridge.MeasurementCategory) == true {
		t.Error("First expose3 event should not be debounced")
	}

	now = now.Add(1 * time.Second)
	mockClock.SetMockTime(now)

	if debouncer.DebounceExpose("expose3", bridge.MeasurementCategory) == true {
		t.Error("Second expose3 event should not be debounced")
	}

	now = now.Add(2 * time.Second)
	mockClock.SetMockTime(now)

	if debouncer.DebounceExpose("expose3", bridge.MeasurementCategory) == true {
		t.Error("Second expose3 event should not be debounced")
	}
}
