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

	if len(cache.devicesConfigs) != 2 {
		t.Errorf("Expected 2 device configs, got %d", len(cache.devicesConfigs))
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

func TestDeviceConfigCache_Delete(t *testing.T) {

	app := NewAppConfig()
	device1 := NewDeviceConfig("device1")
	device1.Debounce = map[string]*utils.TimeInterval{
		"expose1": utils.IntervalFromMilliseconds(1000),
		"expose2": utils.IntervalFromMilliseconds(2000),
	}
	device2 := NewDeviceConfig("device2")
	device2.Debounce = map[string]*utils.TimeInterval{
		"expose3": utils.IntervalFromMilliseconds(1000),
		"expose4": utils.IntervalFromMilliseconds(2000),
	}
	app.AddDeviceConfig(device1)
	app.AddDeviceConfig(device2)

	cache := NewDeviceConfigCache(app)
	cache.Set(device1)
	cache.Set(device2)

	// assert that the debounce is set
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

	// delete the device
	cache.Delete("device1")

	// assert that the device is deleted
	_, ok = cache.Get("device1")
	if ok {
		t.Errorf("Expected device1 to be deleted")
	}

	// assertt that the all expose debounces are deleted
	_, ok = cache.GetDebounce("device1", "expose1")
	if ok {
		t.Errorf("Expected expose1 debounce to be deleted")
	}
	_, ok = cache.GetDebounce("device1", "expose2")
	if ok {
		t.Errorf("Expected expose2 debounce to be deleted")
	}

	// assert that the other device is still there
	_, ok = cache.Get("device2")
	if !ok {
		t.Errorf("Expected device2 to be still there")
	}

	debounce, ok = cache.GetDebounce("device2", "expose3")
	if !ok {
		t.Errorf("Expected expose3 debounce to be set")
	}
	if debounce != 1*time.Second {
		t.Errorf("Expected expose3 debounce to be set")
	}
	debounce, ok = cache.GetDebounce("device2", "expose4")
	if !ok {
		t.Errorf("Expected expose4 debounce to be set")
	}
	if debounce != 2*time.Second {
		t.Errorf("Expected expose4 debounce to be set")
	}
}

func TestDeviceConfigCache_DeleteDebounce(t *testing.T) {
	app := NewAppConfig()
	device1 := NewDeviceConfig("device1")
	device1.Debounce = map[string]*utils.TimeInterval{
		"expose1": utils.IntervalFromMilliseconds(1000),
		"expose2": utils.IntervalFromMilliseconds(2000),
	}
	app.AddDeviceConfig(device1)

	cache := NewDeviceConfigCache(app)
	cache.Set(device1)

	// assert that the debounce is set
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

	// delete the deebounce
	ok = cache.DeleteDebounce("device1", "expose2")
	if !ok {
		t.Errorf("Expected expose1 debounce to be deleted")
	}

	// assert that the expose1 debounce is deleted
	_, ok = cache.GetDebounce("device1", "expose2")
	if ok {
		t.Errorf("Expected expose2 debounce to be deleted")
	}

	// assert that the expose2 debounce is still set
	_, ok = cache.GetDebounce("device1", "expose1")
	if !ok {
		t.Errorf("Expected expose1 debounce to be set")
	}

}
