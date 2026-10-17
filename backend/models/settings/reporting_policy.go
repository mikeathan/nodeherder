package settings

// Reporting policies control observation output, metric eligibility and how long a
// command waits for device evidence. They never change raw state or automation values.
// Property names and operational values come from saved hub/device configuration.

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/utils"
	"strings"
	"time"
)

// The compatibility default is declarative data. Once configured, the hub's saved
// defaults and device overrides are authoritative; no property-specific rules live here.
//
//go:embed reporting-defaults.json
var reportingDefaultsJSON []byte

var reportingDefaults = loadReportingDefaults()

type reportingBootstrap struct {
	LiveIntervalByCategory       map[bridge.ExposeCategory]*utils.TimeInterval `json:"liveIntervalByCategory"`
	CommandConfirmationTimeoutMs int                                           `json:"commandConfirmationTimeoutMs"`
}

func loadReportingDefaults() reportingBootstrap {
	var defaults reportingBootstrap
	if err := json.Unmarshal(reportingDefaultsJSON, &defaults); err != nil || (defaults.CommandConfirmationTimeoutMs <= 0 || defaults.CommandConfirmationTimeoutMs > math.MaxInt32) {
		panic("invalid embedded reporting defaults")
	}
	return defaults
}

// ReportingPolicy describes observations, never raw state or automation inputs.
type ReportingPolicy struct {
	LiveInterval  time.Duration
	MinimumChange float64
	RecordMetrics bool
}

// A present zero/false is an explicit override, not a missing value.
// resolveReportingValue preserves explicit zero/false settings at the narrowest scope.
func resolveReportingValue[T any](device, defaults map[string]T, name string, fallback T) T {
	if value, ok := device[name]; ok {
		return value
	}
	if value, ok := defaults[name]; ok {
		return value
	}
	return fallback
}

// GetReportingPolicy is the single resolution path for observation and history rules.
// Missing entries inherit category intervals and schema-defined metric eligibility.
func (d *DeviceConfigCache) GetReportingPolicy(id string, expose *devices.Entity) ReportingPolicy {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	override := d.appConfig.Hub.Devices.Overrides[id]
	if override == nil {
		override = &DeviceConfig{}
	}
	defaults := d.appConfig.Hub.Devices.Defaults
	return ReportingPolicy{
		LiveInterval:  resolveInterval(override.LiveIntervalOverrides, defaults.LiveIntervalOverrides, override.LiveIntervalByCategory, defaults.LiveIntervalByCategory, expose.Name, expose.Category).Duration(),
		MinimumChange: resolveReportingValue(override.ReportingThresholds, defaults.ReportingThresholds, expose.Name, 0),
		RecordMetrics: resolveReportingValue(override.MetricOverrides, defaults.MetricOverrides, expose.Name, expose.Category == bridge.MeasurementCategory),
	}
}

// GetCommandConfirmationTimeout is shared policy for the command handler and the
// frontend's pending-command deadline. The snapshot exposes the configured values;
// the bootstrap JSON covers older settings that predate this field.
func (d *DeviceConfigCache) GetCommandConfirmationTimeout(id string) time.Duration {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	milliseconds := reportingDefaults.CommandConfirmationTimeoutMs
	if configured := d.appConfig.Hub.Devices.Defaults.CommandConfirmationTimeoutMs; configured != nil {
		milliseconds = *configured
	}
	if override := d.appConfig.Hub.Devices.Overrides[id]; override != nil && override.CommandConfirmationTimeoutMs != nil {
		milliseconds = *override.CommandConfirmationTimeoutMs
	}
	return time.Duration(milliseconds) * time.Millisecond
}

// ValidateReportingSettings runs before publishing or persisting configuration.
func (d *DeviceConfig) ValidateReportingSettings() error {
	if d == nil {
		return fmt.Errorf("device config is required")
	}
	for _, intervals := range []map[string]*utils.TimeInterval{d.LiveIntervalByCategory, d.LiveIntervalOverrides, d.DefaultDebounceByCategory, d.DebounceOverrides} {
		for name, interval := range intervals {
			if strings.TrimSpace(name) == "" || !validReportingInterval(interval) {
				return fmt.Errorf("reporting interval for %q must use a supported unit and a nonnegative value within one year", name)
			}
		}
	}
	for property, delta := range d.ReportingThresholds {
		if property == "" || delta < 0 || math.IsNaN(delta) || math.IsInf(delta, 0) {
			return fmt.Errorf("reporting threshold for %q must be finite and nonnegative", property)
		}
	}
	for property := range d.MetricOverrides {
		if property == "" {
			return fmt.Errorf("metric override property is required")
		}
	}
	if value := d.CommandConfirmationTimeoutMs; value != nil && (*value <= 0 || *value > math.MaxInt32) {
		return fmt.Errorf("command confirmation timeout must fit a positive browser timer in milliseconds")
	}
	return nil
}

func newDefaultConfirmationTimeout() *int {
	value := reportingDefaults.CommandConfirmationTimeoutMs
	return &value
}

// resolveInterval is shared by live and history policies. A missing rule means
// immediate output; explicit zero overrides a slower inherited rule. Specific field
// rules outrank category rules; device scope wins within each level.
func resolveInterval(device, global map[string]*utils.TimeInterval, deviceCategories, categories map[bridge.ExposeCategory]*utils.TimeInterval, name string, category bridge.ExposeCategory) *utils.TimeInterval {
	fallback := resolveReportingValue(deviceCategories, categories, category, (*utils.TimeInterval)(nil))
	if fallback == nil {
		fallback = utils.IntervalFromSeconds(0)
	}
	return resolveReportingValue(device, global, name, fallback)
}

func validReportingInterval(interval *utils.TimeInterval) bool {
	if interval == nil || interval.Value < 0 {
		return false
	}
	units := map[string]int64{utils.UnitMilliseconds: 1, utils.UnitSeconds: 1000, utils.UnitMinutes: 60000, utils.UnitHours: 3600000, utils.UnitDays: 86400000}
	multiplier, ok := units[interval.Unit]
	return ok && int64(interval.Value) <= int64(365*24*60*60*1000)/multiplier
}

func defaultLiveIntervals() map[bridge.ExposeCategory]*utils.TimeInterval {
	return cloneIntervals(reportingDefaults.LiveIntervalByCategory)
}

// Interval structs are copied too: callers editing a snapshot cannot mutate policy.
func cloneIntervals[K comparable](values map[K]*utils.TimeInterval) map[K]*utils.TimeInterval {
	result := maps.Clone(values)
	for key, interval := range result {
		result[key] = clonePointer(interval)
	}
	return result
}
