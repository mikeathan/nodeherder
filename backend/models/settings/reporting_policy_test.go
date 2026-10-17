package settings_test

import (
	"math"
	"node-herder/mocks"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/settings"
	"node-herder/utils"
	"sync"
	"testing"
	"time"
)

func TestValidateReportingSettings(t *testing.T) {
	for _, value := range []float64{-1, math.NaN(), math.Inf(1), math.Inf(-1)} {
		cfg := settings.NewDeviceConfig("sensor")
		cfg.ReportingThresholds = map[string]float64{"target_distance": value}
		if cfg.ValidateReportingSettings() == nil {
			t.Errorf("accepted %v", value)
		}
	}
	cfg := settings.NewDeviceConfig("sensor")
	cfg.ReportingThresholds = map[string]float64{"": 0}
	if cfg.ValidateReportingSettings() == nil {
		t.Fatal("accepted blank property")
	}
	var missing *settings.DeviceConfig
	if missing.ValidateReportingSettings() == nil {
		t.Fatal("nil config accepted")
	}
}

func TestConfirmationTimeoutValidation(t *testing.T) {
	for _, value := range []int{0, -1, int(math.MaxInt32) + 1} {
		cfg := settings.NewDeviceConfig("sensor")
		cfg.CommandConfirmationTimeoutMs = &value
		if cfg.ValidateReportingSettings() == nil {
			t.Fatalf("accepted timeout %d", value)
		}
	}
	cfg := settings.NewDeviceConfig("sensor")
	cfg.MetricOverrides = map[string]bool{"": true}
	if cfg.ValidateReportingSettings() == nil {
		t.Fatal("blank history field accepted")
	}
}

func TestReportingLiveIntervalPrecedence(t *testing.T) {
	app := settings.NewAppConfig()
	expose := devices.NewEntity("new_health_reading")
	expose.Category = bridge.DiagnosticCategory
	cfg := settings.NewDeviceConfig("sensor")
	app.AddDeviceConfig(cfg)
	cache := settings.NewDeviceConfigCache(&mocks.NopSettingsrepo{}, app, &sync.RWMutex{})
	if got := cache.GetReportingPolicy("sensor", expose).LiveInterval; got != 30*time.Second {
		t.Fatalf("category default %v", got)
	}
	app.Hub.Devices.Defaults.LiveIntervalOverrides = map[string]*utils.TimeInterval{expose.Name: utils.IntervalFromSeconds(10)}
	if got := cache.GetReportingPolicy("sensor", expose).LiveInterval; got != 10*time.Second {
		t.Fatalf("global field %v", got)
	}
	cfg.LiveIntervalOverrides = map[string]*utils.TimeInterval{expose.Name: utils.IntervalFromSeconds(0)}
	if got := cache.GetReportingPolicy("sensor", expose).LiveInterval; got != 0 {
		t.Fatalf("explicit zero %v", got)
	}
	delete(cfg.LiveIntervalOverrides, expose.Name)
	if got := cache.GetReportingPolicy("sensor", expose).LiveInterval; got != 10*time.Second {
		t.Fatalf("remove inheritance %v", got)
	}
	expose.Name, expose.Category = "temperature", bridge.MeasurementCategory
	if got := cache.GetReportingPolicy("sensor", expose).LiveInterval; got != 0 {
		t.Fatalf("measurement slowed %v", got)
	}
}

func TestReportingIntervalValidation(t *testing.T) {
	for _, interval := range []*utils.TimeInterval{nil, {Value: -1, Unit: "seconds"}, {Value: 1, Unit: "nonsense"}, {Value: math.MaxInt, Unit: "days"}} {
		cfg := settings.NewDeviceConfig("sensor")
		cfg.LiveIntervalOverrides = map[string]*utils.TimeInterval{"distance": interval}
		if cfg.ValidateReportingSettings() == nil {
			t.Fatalf("accepted invalid interval %+v", interval)
		}
	}
	cfg := settings.NewDeviceConfig("sensor")
	cfg.LiveIntervalByCategory = map[string]*utils.TimeInterval{"": utils.IntervalFromSeconds(1)}
	if cfg.ValidateReportingSettings() == nil {
		t.Fatal("accepted blank category")
	}
	cfg.LiveIntervalByCategory = map[string]*utils.TimeInterval{"diagnostic": utils.IntervalFromSeconds(0)}
	if err := cfg.ValidateReportingSettings(); err != nil {
		t.Fatal(err)
	}
}

func TestMeasurementDefaultsDoNotSuppressChanges(t *testing.T) {
	app := settings.NewAppConfig()
	cache := settings.NewDeviceConfigCache(&mocks.NopSettingsrepo{}, app, &sync.RWMutex{})
	for _, name := range []string{"target_distance", "temperature", "unknown_measurement"} {
		expose := devices.NewEntity(name)
		expose.Category = bridge.MeasurementCategory
		policy := cache.GetReportingPolicy("sensor", expose)
		if policy.LiveInterval != 0 || policy.MinimumChange != 0 {
			t.Fatalf("measurement %s has default suppression: %+v", name, policy)
		}
	}
}

func TestDeviceCategoryIntervalPrecedence(t *testing.T) {
	app := settings.NewAppConfig()
	cfg := settings.NewDeviceConfig("sensor")
	cfg.LiveIntervalByCategory = map[string]*utils.TimeInterval{"diagnostic": utils.IntervalFromSeconds(10), "measurement": utils.IntervalFromSeconds(5)}
	cfg.DefaultDebounceByCategory = map[string]*utils.TimeInterval{"diagnostic": utils.IntervalFromSeconds(20)}
	app.AddDeviceConfig(cfg)
	cache := settings.NewDeviceConfigCache(&mocks.NopSettingsrepo{}, app, &sync.RWMutex{})
	e := devices.NewEntity("signal")
	e.Category = bridge.DiagnosticCategory
	check := func(id string, live, history time.Duration) {
		t.Helper()
		if got := cache.GetReportingPolicy(id, e).LiveInterval; got != live {
			t.Fatalf("%s live %v; want %v", id, got, live)
		}
		if got, _ := cache.GetDebounce(id, e.Name, e.Category); got != history {
			t.Fatalf("%s history %v; want %v", id, got, history)
		}
	}
	check("sensor", 10*time.Second, 20*time.Second)
	check("other", 30*time.Second, 300*time.Second)
	cfg.LiveIntervalByCategory[e.Category] = utils.IntervalFromSeconds(0)
	check("sensor", 0, 20*time.Second)
	app.Hub.Devices.Defaults.LiveIntervalOverrides = map[string]*utils.TimeInterval{e.Name: utils.IntervalFromSeconds(7)}
	check("sensor", 7*time.Second, 20*time.Second)
	cfg.LiveIntervalOverrides = map[string]*utils.TimeInterval{e.Name: utils.IntervalFromSeconds(2)}
	check("sensor", 2*time.Second, 20*time.Second)
	delete(cfg.LiveIntervalOverrides, e.Name)
	delete(app.Hub.Devices.Defaults.LiveIntervalOverrides, e.Name)
	delete(cfg.LiveIntervalByCategory, e.Category)
	delete(cfg.DefaultDebounceByCategory, e.Category)
	app.Hub.Devices.Defaults.LiveIntervalByCategory[e.Category] = utils.IntervalFromSeconds(12)
	check("sensor", 12*time.Second, 300*time.Second)
	e.Name, e.Category = "distance", bridge.MeasurementCategory
	if got := cache.GetReportingPolicy("sensor", e).LiveInterval; got != 5*time.Second {
		t.Fatalf("device measurement %v", got)
	}
	if got := cache.GetReportingPolicy("other", e).LiveInterval; got != 0 {
		t.Fatalf("other measurement slowed: %v", got)
	}
}
