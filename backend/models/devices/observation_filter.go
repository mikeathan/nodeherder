package devices

import (
	"math"
	"node-herder/models/bridge"
)

// ObservationFilter decides which numeric readings are useful to live observers.
// DeviceLifetimeService has already accepted raw state for automation before calling
// this filter. Only the last published number is retained for each known device field,
// so small changes accumulate instead of being lost between successive noisy samples.
// The live publisher owns and serializes this state; the filter starts no background work.
type ObservationFilter struct{ published map[string]float64 }

func NewObservationFilter() *ObservationFilter {
	return &ObservationFilter{published: make(map[string]float64)}
}

// Prime records a value already delivered in the initial device snapshot.
func (f *ObservationFilter) Prime(expose *Entity, value any) {
	if number, ok := reportNumber(value); ok {
		f.published[expose.Name] = number
	} else {
		delete(f.published, expose.Name)
	}
}

// Allows checks without advancing the baseline, so a delayed reading is compared
// with what observers actually saw rather than with intermediate incoming samples.
func (f *ObservationFilter) Allows(expose *Entity, value any, threshold float64) bool {
	if !IsNumericObservation(expose, value) {
		return true
	}
	number, _ := reportNumber(value)
	previous, exists := f.published[expose.Name]
	return !exists || (number != previous && (threshold <= 0 || !belowReportThreshold(math.Abs(number-previous), threshold)))
}

// Accept advances the baseline only when published. Unknown readings clear it so
// the next valid value can be delivered immediately, even inside a live interval.
func (f *ObservationFilter) Accept(expose *Entity, value any, threshold float64) bool {
	if !f.Allows(expose, value, threshold) {
		return false
	}
	if number, ok := reportNumber(value); ok {
		f.published[expose.Name] = number
	} else {
		delete(f.published, expose.Name)
	}
	return true
}

// IsNumericObservation excludes physical events, writable confirmation evidence
// and unknown/invalid readings from both delta and interval filtering.
func IsNumericObservation(expose *Entity, value any) bool {
	_, numeric := reportNumber(value)
	return numeric && expose.Type == bridge.NumericDataType && expose.AccessMode == bridge.ReadAccessMode && !expose.IsEventValue(value)
}

func belowReportThreshold(difference, threshold float64) bool {
	// Account only for floating-point rounding at the exact configured boundary.
	return difference < threshold && threshold-difference > threshold*1e-12
}

func reportNumber(value any) (float64, bool) {
	var number float64
	switch v := value.(type) {
	case float64:
		number = v
	case float32:
		number = float64(v)
	case int:
		number = float64(v)
	case int64:
		number = float64(v)
	default:
		return 0, false
	}
	return number, !math.IsNaN(number) && !math.IsInf(number, 0)
}

// HasBaseline is false after unknown data, allowing a valid recovery immediately.
func (f *ObservationFilter) HasBaseline(name string) bool {
	_, ok := f.published[name]
	return ok
}
