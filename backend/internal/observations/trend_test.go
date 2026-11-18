package observations

// Trend rules are a claim about physical movement, so each rule is tested directly: steady and
// unknown must be as trustworthy as approaching, or the indicator is worse than nothing.

import (
	"testing"
	"time"

	model "node-herder/models/observations"
)

var trendNow = time.Date(2026, 11, 12, 12, 0, 0, 0, time.UTC)

// readingsAt builds a window from signal strengths, spaced one second apart and ending now.
func readingsAt(values ...int) []rssiReading {
	window := make([]rssiReading, len(values))
	for i, value := range values {
		offset := time.Duration(len(values)-1-i) * time.Second
		window[i] = rssiReading{RSSI: value, At: trendNow.Add(-offset)}
	}
	return window
}

func TestTrendUnknownWithoutEnoughHistory(t *testing.T) {
	for _, window := range [][]rssiReading{
		nil,
		readingsAt(-60),
		readingsAt(-60, -62),
	} {
		if got := computeTrend(window, trendNow); got.Trend != model.TrendUnknown {
			t.Fatalf("window %v gave %q, want unknown", window, got.Trend)
		}
	}
}

func TestTrendUnknownWhenWindowIsStale(t *testing.T) {
	window := readingsAt(-70, -68, -66, -64)
	// The device stopped advertising: reporting a direction from old data would be a lie.
	if got := computeTrend(window, trendNow.Add(rssiTrendHorizon+time.Second)); got.Trend != model.TrendUnknown {
		t.Fatalf("stale window gave %q, want unknown", got.Trend)
	}
	// Just inside the horizon it is still usable.
	if got := computeTrend(window, trendNow.Add(time.Second)); got.Trend == model.TrendUnknown {
		t.Fatal("fresh window should describe a direction")
	}
}

func TestTrendSteadyInsideDeadband(t *testing.T) {
	// A 2 dB drift is ordinary jitter and must not be reported as movement.
	if got := computeTrend(readingsAt(-60, -60, -61, -59, -60, -61), trendNow); got.Trend != model.TrendSteady {
		t.Fatalf("trend = %q (delta %.1f), want steady", got.Trend, got.DeltaDB)
	}
}

func TestTrendApproachingWhenSignalStrengthens(t *testing.T) {
	got := computeTrend(readingsAt(-80, -78, -72, -70, -66, -64), trendNow)
	if got.Trend != model.TrendApproaching {
		t.Fatalf("trend = %q (delta %.1f), want approaching", got.Trend, got.DeltaDB)
	}
	if got.DeltaDB <= 0 {
		t.Fatalf("delta = %.1f, want positive for a strengthening signal", got.DeltaDB)
	}
}

func TestTrendRecedingWhenSignalWeakens(t *testing.T) {
	got := computeTrend(readingsAt(-60, -62, -70, -72, -78, -80), trendNow)
	if got.Trend != model.TrendReceding {
		t.Fatalf("trend = %q (delta %.1f), want receding", got.Trend, got.DeltaDB)
	}
	if got.DeltaDB >= 0 {
		t.Fatalf("delta = %.1f, want negative for a weakening signal", got.DeltaDB)
	}
}

func TestTrendSignalMirrorsTheWindow(t *testing.T) {
	got := computeTrend(readingsAt(-70, -68, -66), trendNow)
	want := []int{-70, -68, -66}
	if len(got.Signal) != len(want) {
		t.Fatalf("signal = %v, want %v", got.Signal, want)
	}
	for i := range want {
		if got.Signal[i] != want[i] {
			t.Fatalf("signal = %v, want %v", got.Signal, want)
		}
	}
}

// The displayed distance is derived from this value, so it must reject single outliers: one bad
// packet moving the estimate several metres is the jitter users notice.
func TestMedianRejectsASingleOutlier(t *testing.T) {
	got := computeTrend(readingsAt(-60, -61, -60, -59, -95, -60, -61), trendNow)
	if got.Smoothed == nil {
		t.Fatal("no smoothed value")
	}
	// The mean would be about -65; the median stays with the cluster.
	if *got.Smoothed < -62 || *got.Smoothed > -59 {
		t.Fatalf("smoothed = %d, want a value from the -59..-61 cluster", *got.Smoothed)
	}
}

func TestMedianIsNilWithoutSamples(t *testing.T) {
	if got := computeTrend(nil, trendNow); got.Smoothed != nil {
		t.Fatalf("smoothed = %v, want nil so the UI shows a missing value", *got.Smoothed)
	}
}

func TestMedianSurvivesATinyWindow(t *testing.T) {
	// One or two readings cannot describe a trend, but they can still describe a position.
	got := computeTrend(readingsAt(-70), trendNow)
	if got.Trend != model.TrendUnknown {
		t.Fatalf("trend = %q, want unknown", got.Trend)
	}
	if got.Smoothed == nil || *got.Smoothed != -70 {
		t.Fatalf("smoothed = %v, want -70", got.Smoothed)
	}
}

func TestMedianDoesNotReorderTheWindow(t *testing.T) {
	window := readingsAt(-70, -60, -80)
	before := []int{window[0].RSSI, window[1].RSSI, window[2].RSSI}
	computeTrend(window, trendNow)
	after := []int{window[0].RSSI, window[1].RSSI, window[2].RSSI}
	for i := range before {
		if before[i] != after[i] {
			t.Fatalf("window reordered: %v became %v", before, after)
		}
	}
}

func TestTrendUnknownDeltaIsZero(t *testing.T) {
	// A caller must not read a leftover delta alongside unknown, which would imply movement.
	if got := computeTrend(nil, trendNow); got.DeltaDB != 0 {
		t.Fatalf("delta = %.1f with unknown trend, want 0", got.DeltaDB)
	}
}

func TestAppendReadingKeepsWindowBounded(t *testing.T) {
	var window []rssiReading
	for i := 0; i < rssiWindowSize+15; i++ {
		window = appendReading(window, -60-i, trendNow)
	}
	if len(window) != rssiWindowSize {
		t.Fatalf("window length = %d, want the bounded %d", len(window), rssiWindowSize)
	}
	// The newest sample must survive; the oldest are the ones dropped.
	if newest := window[len(window)-1].RSSI; newest != -60-(rssiWindowSize+14) {
		t.Fatalf("newest = %d, want the last appended", newest)
	}
}

func TestAppendReadingPreservesOrder(t *testing.T) {
	window := appendReading(appendReading(nil, -70, trendNow), -60, trendNow)
	if window[0].RSSI != -70 || window[1].RSSI != -60 {
		t.Fatalf("window = %v, want oldest first", window)
	}
}
