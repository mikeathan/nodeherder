package services_test

import (
	"node-herder/internal/services"
	"node-herder/mocks"
	"node-herder/models/settings"
	"node-herder/utils"
	"testing"
	"time"
)

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
		"expose3": utils.IntervalFromSeconds(3),
		"expose4": utils.IntervalFromSeconds(4),
	}

	appConfig.AddDeviceConfig(d2)

	cache := settings.NewDeviceConfigCache(appConfig)

	debouncer := services.NewDeviceDebouncer("device1", cache, mockClock)

	// Test First Event
	if debouncer.DebounceExpose("expose1") == true { // First event should not be debounced
		t.Error("First event should not be debounced")
	}

	// Test Debounced Event (within duration)
	now = now.Add(500 * time.Millisecond)
	mockClock.SetMockTime(now)
	if debouncer.DebounceExpose("expose1") == false { // Event within duration should be debounced
		t.Error("Event within duration should be debounced")
	}

	// Test After Duration
	now = now.Add(500 * time.Millisecond)
	mockClock.SetMockTime(now)
	if debouncer.DebounceExpose("expose1") == true { // Event after duration should not be debounced
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

	now = now.Add(1 * time.Second)
	mockClock.SetMockTime(now)

	if !debouncer.DebounceExpose("expose2") {
		t.Error("expose2 event within duration should be debounced")
	}

	// second device
	now2 := time.Now()
	mockClock.SetMockTime(now2)
	debouncer2 := services.NewDeviceDebouncer("device2", cache, mockClock)

	if debouncer2.DebounceExpose("expose4") {
		t.Error("First expose4 event should not be debounced")
	}

	now2 = now2.Add(3 * time.Second)
	mockClock.SetMockTime(now2)

	if !debouncer2.DebounceExpose("expose4") {
		t.Error("Second expose4 event should be debounced")
	}
}
