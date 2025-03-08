package services_test

import (
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/settings"
	"node-herder/utils"
	"testing"
	"time"
)


todo
func TestDeviceDebouncer_DebounceExpose(t *testing.T) {

	now := time.Now()
	mockClock := mocks.NewMockClock(func() time.Time {
		return now
	})

	appConfig := settings.NewAppConfig()

	d1 := settings.NewDeviceConfig("device1")
	d1.Debounce = map[string]*utils.TimeInterval{
		"expose1": utils.IntervalFromSeconds(1),
		"expose2": utils.IntervalFromSeconds(2),
	}

	appConfig.AddDeviceConfig(d1)
	d2 := settings.NewDeviceConfig("device2")
	d2.Debounce = map[string]*utils.TimeInterval{
		"expose3": utils.IntervalFromMilliseconds(3000),
		"expose4": utils.IntervalFromMilliseconds(4000),
	}
	appConfig.AddDeviceConfig(d2)

	cache := settings.NewDeviceConfigCache(appConfig)

	debouncer := services.NewDeviceDebouncer("device1", cache, mockClock)

	// Test First Event
	if debouncer.DebounceExpose("expose1") {
		t.Error("First event should not be debounced")
	}

	// Test Debounced Event (within duration)
	mockClock.SetMockTime(now.Add(500 * time.Millisecond))
	if !debouncer.DebounceExpose("expose1") {
		t.Error("Event within duration should be debounced")
	}

	// Test After Duration
	mockClock.SetMockTime(now.Add(600 * time.Millisecond))
	if debouncer.DebounceExpose("expose1") {
		t.Error("Event after duration should not be debounced")
	}

	// Test No Debounce Config
	if debouncer.DebounceExpose("expose3") {
		t.Error("Event with no debounce config should not be debounced")
	}
	//Test multiple exposes.
	if debouncer.DebounceExpose("expose2") {
		t.Error("First expose2 event should not be debounced")
	}
	mockClock.SetMockTime(now.Add(1 * time.Second))

	if !debouncer.DebounceExpose("expose2") {
		t.Error("expose2 event within duration should be debounced")
	}
}
