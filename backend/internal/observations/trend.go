package observations

// This file derives proximity trend from the recent signal-strength window. It is a pure
// calculation so the rules can be tested directly: a trend claim is only useful if "steady"
// and "unknown" are as trustworthy as "approaching".

import (
	"slices"
	"time"

	model "node-herder/models/observations"
)

// trendResult is one identity's derived direction of change.
type trendResult struct {
	// Trend is one of the model.Trend* values.
	Trend string
	// DeltaDB is the smoothed change across the window, positive when the signal strengthened.
	// It is zero when Trend is unknown.
	DeltaDB float64
	// Signal is the window's signal strengths, oldest first, for display.
	Signal []int
	// Smoothed is the median signal strength over the window, or nil when there is no sample.
	// Median rather than mean: a single outlier packet should not move the estimate, and RSSI
	// outliers are common when a body passes between the device and the antenna.
	Smoothed *int
}

// appendReading adds one sample to the bounded window, dropping the oldest when full. The
// window is per identity, so memory stays proportional to identities, not to packet count.
func appendReading(window []rssiReading, rssi int, at time.Time) []rssiReading {
	window = append(window, rssiReading{RSSI: rssi, At: at})
	if len(window) > rssiWindowSize {
		// Copying the tail avoids retaining the dropped backing array for the life of the entry.
		window = append([]rssiReading(nil), window[len(window)-rssiWindowSize:]...)
	}
	return window
}

// computeTrend describes direction of change over the window.
//
// The comparison is between the mean of the older half and the mean of the newer half. Averaging
// is what makes the result stable: single-packet RSSI jitters by several dB, so any per-sample
// comparison would report movement constantly.
//
// Staleness matters as much as direction. A device that stopped advertising keeps its last
// window, and reporting "approaching" from data minutes old would be a lie, so an expired
// window yields unknown.
func computeTrend(window []rssiReading, now time.Time) trendResult {
	signal := make([]int, len(window))
	for i, reading := range window {
		signal[i] = reading.RSSI
	}
	smoothed := medianRSSI(window)
	if len(window) < rssiTrendMinReadings {
		return trendResult{Trend: model.TrendUnknown, Signal: signal, Smoothed: smoothed}
	}
	newest := window[len(window)-1]
	if now.Sub(newest.At) > rssiTrendHorizon {
		return trendResult{Trend: model.TrendUnknown, Signal: signal, Smoothed: smoothed}
	}

	half := len(window) / 2
	older := meanRSSI(window[:half])
	newer := meanRSSI(window[half:])
	delta := newer - older

	switch {
	case delta > rssiTrendDeadbandDB:
		return trendResult{Trend: model.TrendApproaching, DeltaDB: delta, Signal: signal, Smoothed: smoothed}
	case delta < -rssiTrendDeadbandDB:
		return trendResult{Trend: model.TrendReceding, DeltaDB: delta, Signal: signal, Smoothed: smoothed}
	default:
		return trendResult{Trend: model.TrendSteady, DeltaDB: delta, Signal: signal, Smoothed: smoothed}
	}
}

// medianRSSI returns the middle signal strength of the window, or nil when it is empty.
//
// Median is chosen over mean so one outlier packet cannot move the displayed distance: a body
// passing between device and antenna produces exactly that kind of outlier, and averaging would
// spread the error across the estimate instead of rejecting it. The slice is copied before
// sorting so the caller's window order is preserved.
func medianRSSI(readings []rssiReading) *int {
	if len(readings) == 0 {
		return nil
	}
	values := make([]int, len(readings))
	for i, reading := range readings {
		values[i] = reading.RSSI
	}
	slices.Sort(values)
	median := values[len(values)/2]
	return &median
}

// meanRSSI averages a slice of readings. An empty slice yields zero, which callers avoid by
// requiring rssiTrendMinReadings first.
func meanRSSI(readings []rssiReading) float64 {
	if len(readings) == 0 {
		return 0
	}
	total := 0
	for _, reading := range readings {
		total += reading.RSSI
	}
	return float64(total) / float64(len(readings))
}
