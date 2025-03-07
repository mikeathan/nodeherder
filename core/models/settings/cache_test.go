package settings

import (
	"node-herder/utils"
	"testing"
	"time"
)

func TestNewDeviceConfigCache(t *testing.T) {

	appConfig := NewAppConfig()

	d1 := NewDeviceConfig("device1")
	d1.Debounce = map[string]*utils.TimeInterval{
		"expose1": utils.IntervalFromMilliseconds(1000),
		"expose2": utils.IntervalFromMilliseconds(2000),
	}

	appConfig.AddDeviceConfig(d1)
	d2 := NewDeviceConfig("device2")
	d2.Debounce = map[string]*utils.TimeInterval{
		"expose3": utils.IntervalFromMilliseconds(3000),
		"expose4": utils.IntervalFromMilliseconds(4000),
	}
	appConfig.AddDeviceConfig(d2)

	cache := NewDeviceConfigCache(appConfig)

	if len(cache.deviceConfigs) != 2 {
		t.Errorf("Expected 2 device configs, got %d", len(cache.deviceConfigs))
	}

	expose1Debounce, ok := cache.GetDebounce("device1", "expose1")
	if !ok {
		t.Errorf("Expected expose1 debounce to be found, got %v", ok)
	}
	if expose1Debounce != 1000*time.Millisecond {
		t.Errorf("Expected expose1 debounce to be 1000ms, got %v", expose1Debounce)
	}
	expose2Debounce, ok := cache.GetDebounce("device1", "expose2")
	if !ok {
		t.Errorf("Expected expose2 debounce to be found, got %v", ok)
	}
	if expose2Debounce != 2000*time.Millisecond {
		t.Errorf("Expected expose2 debounce to be 2000ms, got %v", expose2Debounce)
	}

	expose3Debounce, ok := cache.GetDebounce("device2", "expose3")
	if !ok {
		t.Errorf("Expected expose3 debounce to be found, got %v", ok)
	}
	if expose3Debounce != 3000*time.Millisecond {
		t.Errorf("Expected expose3 debounce to be 3000ms, got %v", expose3Debounce)
	}
	expose4Debounce, ok := cache.GetDebounce("device2", "expose4")
	if !ok {
		t.Errorf("Expected expose4 debounce to be found, got %v", ok)
	}
	if expose4Debounce != 4000*time.Millisecond {
		t.Errorf("Expected expose4 debounce to be 4000ms, got %v", expose4Debounce)
	}
}

func TestDeviceConfigCache_Get(t *testing.T) {
	cache := NewDeviceConfigCache(NewAppConfig())
	device1 := NewDeviceConfig("device1")
	cache.Set(device1)

	device, ok := cache.Get("device1")
	if !ok || device == nil {
		t.Errorf("Expected to get device1, got nil")
	}

	_, ok = cache.Get("device2")
	if ok {
		t.Errorf("Expected to not get device2")
	}
}

func TestDeviceConfigCache_Set(t *testing.T) {

	app := NewAppConfig()
	device1 := NewDeviceConfig("device1")
	device1.Debounce = map[string]*utils.TimeInterval{
		"expose1": utils.IntervalFromMilliseconds(1000),
	}
	app.AddDeviceConfig(device1)
	cache := NewDeviceConfigCache(app)

	// add a second debounce to device1 after initialization
	device1.Debounce = map[string]*utils.TimeInterval{
		"expose2": utils.IntervalFromMilliseconds(2000),
	}
	cache.Set(device1)
	debounce, ok := cache.GetDebounce("device1", "expose1")
	if !ok {
		t.Errorf("Expected expose1 debounce to be set")
	}
	if debounce != 1*time.Second {
		t.Errorf("Expected expose1 debounce to be set")
	}
	debounce, ok = cache.GetDebounce("device1", "expose2")
	if !ok {
		t.Errorf("Expected expose2 debounce to be set")
	}
	if debounce != 2*time.Second {
		t.Errorf("Expected expose2 debounce to be set")
	}
}
