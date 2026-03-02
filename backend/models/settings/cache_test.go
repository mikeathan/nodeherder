package settings_test

import (
	"node-herder/mocks"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/settings"
	"node-herder/repository"
	"node-herder/utils"
	"os"
	"sync"
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
	cache := settings.NewDeviceConfigCache(&repo, appConfig, &sync.RWMutex{})

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

	cache := settings.NewDeviceConfigCache(&repo, settings.NewAppConfig(), &sync.RWMutex{})
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
	cache := settings.NewDeviceConfigCache(&repo, appConfig, &sync.RWMutex{})

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

	cache := settings.NewDeviceConfigCache(&repo, app, &sync.RWMutex{})
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

	// delete the debounce
	ok, err := cache.DeleteDebounce("device1", "expose2")
	if err != nil {
		t.Fatalf("DeleteDebounce returned error: %v", err)
	}
	if !ok {
		t.Errorf("Expected expose2 debounce to be deleted")
	}

	// assert that the expose2 debounce is deleted and falls back to default 60s
	debounce, ok = cache.GetDebounce("device1", "expose2", bridge.MeasurementCategory)
	if !ok || debounce != 60*time.Second {
		t.Errorf("Expected expose2 debounce to fall back to 60s default")
	}

	// assert that the expose1 debounce is still set
	_, ok = cache.GetDebounce("device1", "expose1", bridge.MeasurementCategory)
	if !ok {
		t.Errorf("Expected expose1 debounce to be set")
	}
}

func TestDeviceConfigCache_SetDebounce_Persists(t *testing.T) {
	repo := &mocks.TrackingSettingsRepo{}
	app := settings.NewAppConfig()
	device1 := settings.NewDeviceConfig("device1")
	app.AddDeviceConfig(device1)

	cache := settings.NewDeviceConfigCache(repo, app, &sync.RWMutex{})

	err := cache.SetDebounce("device1", "expose1", utils.IntervalFromMilliseconds(500))
	if err != nil {
		t.Fatalf("SetDebounce returned error: %v", err)
	}

	if repo.SaveDeviceConfigCallCount != 1 {
		t.Errorf("Expected SaveDeviceConfig to be called 1 time, got %d", repo.SaveDeviceConfigCallCount)
	}

	// Verify the value was set in memory
	debounce, ok := cache.GetDebounce("device1", "expose1", bridge.MeasurementCategory)
	if !ok {
		t.Errorf("Expected expose1 debounce to be found")
	}
	if debounce != 500*time.Millisecond {
		t.Errorf("Expected expose1 debounce to be 500ms, got %v", debounce)
	}
}

func TestDeviceConfigCache_DeleteDebounce_Persists(t *testing.T) {
	repo := &mocks.TrackingSettingsRepo{}
	app := settings.NewAppConfig()
	device1 := settings.NewDeviceConfig("device1")
	device1.DebounceOverrides = map[string]*utils.TimeInterval{
		"expose1": utils.IntervalFromMilliseconds(1000),
	}
	app.AddDeviceConfig(device1)

	cache := settings.NewDeviceConfigCache(repo, app, &sync.RWMutex{})

	ok, err := cache.DeleteDebounce("device1", "expose1")
	if err != nil {
		t.Fatalf("DeleteDebounce returned error: %v", err)
	}
	if !ok {
		t.Errorf("Expected DeleteDebounce to return true")
	}

	if repo.SaveDeviceConfigCallCount != 1 {
		t.Errorf("Expected SaveDeviceConfig to be called 1 time, got %d", repo.SaveDeviceConfigCallCount)
	}

	// Verify it was deleted from memory and fell back to 60s default
	debounce, ok := cache.GetDebounce("device1", "expose1", bridge.MeasurementCategory)
	if !ok || debounce != 60*time.Second {
		t.Errorf("Expected expose1 debounce to fall back to 60s default")
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

	cache := settings.NewDeviceConfigCache(&repo, appConfig, &sync.RWMutex{})

	debouncer := settings.NewDeviceDebouncer("device1", cache, mockClock)

	// Test First Event
	if debouncer.DebounceExpose(&devices.Entity{Name: "expose1", Category: bridge.MeasurementCategory, Type: bridge.NumericDataType}) == true { // First event should not be debounced
		t.Error("First event should not be debounced")
	}

	// Test Debounced Event (within duration)
	now = now.Add(500 * time.Millisecond)
	mockClock.SetMockTime(now)
	if debouncer.DebounceExpose(&devices.Entity{Name: "expose1", Category: bridge.MeasurementCategory, Type: bridge.NumericDataType}) == false { // Event within duration should be debounced
		t.Error("Event within duration should be debounced")
	}

	// Test After Duration
	now = now.Add(500 * time.Millisecond)
	mockClock.SetMockTime(now)
	if debouncer.DebounceExpose(&devices.Entity{Name: "expose1", Category: bridge.MeasurementCategory, Type: bridge.NumericDataType}) == true { // Event after duration should not be debounced
		t.Error("Event after duration should not be debounced")
	}

	// Test No Debounce Config
	if debouncer.DebounceExpose(&devices.Entity{Name: "expose3", Category: bridge.MeasurementCategory, Type: bridge.NumericDataType}) {
		t.Error("Event with no debounce config should not be debounced")
	}
	//Test multiple exposes.
	if debouncer.DebounceExpose(&devices.Entity{Name: "expose2", Category: bridge.MeasurementCategory, Type: bridge.NumericDataType}) {
		t.Error("First expose2 event should not be debounced")
	}

	now = now.Add(1 * time.Second)
	mockClock.SetMockTime(now)

	if !debouncer.DebounceExpose(&devices.Entity{Name: "expose2", Category: bridge.MeasurementCategory, Type: bridge.NumericDataType}) {
		t.Error("expose2 event within duration should be debounced")
	}

	// second device
	now2 := time.Now()
	mockClock.SetMockTime(now2)
	debouncer2 := settings.NewDeviceDebouncer("device2", cache, mockClock)

	if debouncer2.DebounceExpose(&devices.Entity{Name: "expose4", Category: bridge.MeasurementCategory, Type: bridge.NumericDataType}) {
		t.Error("First expose4 event should not be debounced")
	}

	now2 = now2.Add(3 * time.Second)
	mockClock.SetMockTime(now2)

	if !debouncer2.DebounceExpose(&devices.Entity{Name: "expose4", Category: bridge.MeasurementCategory, Type: bridge.NumericDataType}) {
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
	appConfig.Hub.Devices.Defaults.DefaultDebounceByCategory[bridge.DiagnosticCategory] = utils.IntervalFromSeconds(5)

	appConfig.AddDeviceConfig(d1)
	cache := settings.NewDeviceConfigCache(&repo, appConfig, &sync.RWMutex{})
	debouncer := settings.NewDeviceDebouncer("device1", cache, mockClock)

	// Test expose 1 event with debounce overrides
	// ############################################################################
	if debouncer.DebounceExpose(&devices.Entity{Name: "expose1", Category: bridge.MeasurementCategory, Type: bridge.NumericDataType}) == true { // First event should not be debounced
		t.Error("First expose event should not be debounced")
	}

	now = now.Add(100 * time.Millisecond)
	mockClock.SetMockTime(now)
	if debouncer.DebounceExpose(&devices.Entity{Name: "expose1", Category: bridge.MeasurementCategory, Type: bridge.NumericDataType}) == false {
		t.Error("Second expose event should be debounced")
	}

	now = now.Add(1 * time.Second)
	mockClock.SetMockTime(now)

	if debouncer.DebounceExpose(&devices.Entity{Name: "expose1", Category: bridge.MeasurementCategory, Type: bridge.NumericDataType}) == true {
		t.Error("Third expose event should be not debounced")
	}

	// now test diagnostics expose with no override but using default debounce
	// ############################################################################
	if debouncer.DebounceExpose(&devices.Entity{Name: "expose2", Category: bridge.DiagnosticCategory, Type: bridge.NumericDataType}) == true {
		t.Error("First expose2 event should not be debounced")
	}
	now = now.Add(2 * time.Second)
	mockClock.SetMockTime(now)

	if debouncer.DebounceExpose(&devices.Entity{Name: "expose2", Category: bridge.DiagnosticCategory, Type: bridge.NumericDataType}) == false {
		t.Error("Second expose2 event should be debounced")
	}

	now = now.Add(4 * time.Second)
	mockClock.SetMockTime(now)

	if debouncer.DebounceExpose(&devices.Entity{Name: "expose2", Category: bridge.DiagnosticCategory, Type: bridge.NumericDataType}) == true {
		t.Error("Third expose2 event should not be debounced")
	}

	// now test non-diagnostics expose with no override, it should be debounced with default 60s
	// ############################################################################
	if debouncer.DebounceExpose(&devices.Entity{Name: "expose3", Category: bridge.MeasurementCategory, Type: bridge.NumericDataType}) == true {
		t.Error("First expose3 event should not be debounced")
	}

	now = now.Add(1 * time.Second)
	mockClock.SetMockTime(now)

	if debouncer.DebounceExpose(&devices.Entity{Name: "expose3", Category: bridge.MeasurementCategory, Type: bridge.NumericDataType}) == false {
		t.Error("Second expose3 event SHOULD be debounced (60s default)")
	}

	now = now.Add(60 * time.Second)
	mockClock.SetMockTime(now)

	if debouncer.DebounceExpose(&devices.Entity{Name: "expose3", Category: bridge.MeasurementCategory, Type: bridge.NumericDataType}) == true {
		t.Error("Third expose3 event should NOT be debounced (past 60s default)")
	}
}

func TestDeviceConfigCache_UpdateDefaultsRefreshesCachedDevices(t *testing.T) {
	tempFile, err := os.CreateTemp("", "settings-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	settingsRepo, err := repository.NewFileSettingsRepoFromFile(tempFile.Name())
	if err != nil {
		t.Fatalf("failed to create settings repo: %v", err)
	}
	defer settingsRepo.Close()

	configCache, err := settings.NewAppConfigCache(settingsRepo, []settings.Task{})
	if err != nil {
		t.Fatalf("failed to create app config cache: %v", err)
	}

	deviceCache := configCache.GetDeviceConfigCache()

	// prime the cache with a default device config
	defaultId := "device-default"
	cfg, err := deviceCache.Get(defaultId)
	if err != nil {
		t.Fatalf("failed to load default device config: %v", err)
	}
	if cfg.MetricsEnabled {
		t.Fatalf("expected metrics to be disabled by default")
	}

	// add an explicit override that should not be changed by default updates
	override := settings.NewDeviceConfig("device-override")
	override.MetricsEnabled = false
	if err := configCache.SetDeviceConfigOverrides(override); err != nil {
		t.Fatalf("failed to set device override: %v", err)
	}

	// enable metrics via defaults
	newDefaults := settings.DefaultDeviceConfig()
	newDefaults.MetricsEnabled = true

	if err := configCache.SetDeviceConfigDefaults(newDefaults); err != nil {
		t.Fatalf("failed to update device defaults: %v", err)
	}

	updatedCfg, err := deviceCache.Get(defaultId)
	if err != nil {
		t.Fatalf("failed to load updated default device config: %v", err)
	}
	if !updatedCfg.MetricsEnabled {
		t.Fatalf("expected metrics to be enabled after updating defaults")
	}

	overrideCfg, err := deviceCache.Get("device-override")
	if err != nil {
		t.Fatalf("failed to load override device config: %v", err)
	}
	if overrideCfg.MetricsEnabled {
		t.Fatalf("expected override to remain unchanged after default update")
	}
}
