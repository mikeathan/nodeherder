package observations

import (
	"testing"
	"time"

	model "node-herder/models/observations"
)

// day is a fixed UTC day so tests never depend on when they run.
var day = time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC)

// at returns a time on the fixture day.
func at(hour, minute int) time.Time {
	return day.Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)
}

// observe folds one reading into an accumulator at the given time.
func observe(e *evidence, rssi int, when time.Time) {
	e.observe(model.Observation{Identity: model.Identity{Key: "addr:aa:bb:cc:dd:ee:01"}, RSSI: rssi, ObservedAt: when})
}

func TestEvidenceCountsPacketsAndBounds(t *testing.T) {
	var e evidence
	observe(&e, -60, at(9, 0))
	observe(&e, -62, at(9, 1))
	observe(&e, -58, at(10, 0))

	if e.packets != 3 {
		t.Fatalf("packets = %d, want 3", e.packets)
	}
	if !e.firstSeen.Equal(at(9, 0)) {
		t.Errorf("firstSeen = %v, want %v", e.firstSeen, at(9, 0))
	}
	if !e.lastSeen.Equal(at(10, 0)) {
		t.Errorf("lastSeen = %v, want %v", e.lastSeen, at(10, 0))
	}
}

func TestEvidenceHourMaskRecordsDistinctHours(t *testing.T) {
	var e evidence
	observe(&e, -60, at(4, 0))
	observe(&e, -61, at(4, 30))
	observe(&e, -62, at(17, 15))

	if got := e.hourMask(); got != (1<<4)|(1<<17) {
		t.Errorf("hourMask = %#b, want bits 4 and 17", got)
	}
	if got := e.activeHours(); got != 2 {
		t.Errorf("activeHours = %d, want 2", got)
	}
}

func TestEvidenceWelfordMeanAndVariance(t *testing.T) {
	var e evidence
	for _, rssi := range []int{-70, -60, -80} {
		observe(&e, rssi, at(9, 0))
	}

	// Welford must match the population variance: mean -70, variance 200/3.
	if got := e.rssi.mean(); got < -70.0001 || got > -69.9999 {
		t.Errorf("mean = %f, want -70", got)
	}
	want := 200.0 / 3.0
	if got := e.rssi.variance(); got < want-0.001 || got > want+0.001 {
		t.Errorf("variance = %f, want %f", got, want)
	}
}

func TestEvidenceWelfordIsStableOverManyReadings(t *testing.T) {
	// A naive sum-of-squares implementation loses precision here; Welford must not.
	var e evidence
	for i := 0; i < 10000; i++ {
		observe(&e, -70, at(9, 0))
	}

	if got := e.rssi.mean(); got < -70.0001 || got > -69.9999 {
		t.Errorf("mean = %f after 10k identical readings, want -70", got)
	}
	// Zero variance must be exactly zero, not a small negative from cancellation.
	if got := e.rssi.variance(); got < -0.001 || got > 0.001 {
		t.Errorf("variance = %f, want 0", got)
	}
}

func TestEvidenceTracksStrongestAndWeakest(t *testing.T) {
	var e evidence
	for _, rssi := range []int{-80, -55, -90, -70} {
		observe(&e, rssi, at(9, 0))
	}

	if e.strongest != -55 {
		t.Errorf("strongest = %d, want -55", e.strongest)
	}
	if e.weakest != -90 {
		t.Errorf("weakest = %d, want -90", e.weakest)
	}
}

func TestEvidenceMobilityCountsOnlyRealChanges(t *testing.T) {
	tests := []struct {
		name     string
		readings []int
		want     uint32
	}{
		{
			name:     "steady signal produces no crossings",
			readings: []int{-70, -71, -70, -72, -69, -71},
			want:     0,
		},
		{
			// Jitter of a few dB must never be counted as movement, or every device looks mobile.
			name:     "jitter inside the threshold produces no crossings",
			readings: []int{-70, -67, -73, -68, -72, -70},
			want:     0,
		},
		{
			name:     "one clear move counts once",
			readings: []int{-70, -70, -70, -82, -82, -82},
			want:     1,
		},
		{
			name:     "moving and returning counts twice",
			readings: []int{-70, -70, -70, -82, -82, -82, -70, -70, -70},
			want:     2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var e evidence
			for _, rssi := range test.readings {
				observe(&e, rssi, at(9, 0))
			}
			if got := e.mobility.crossings; got != test.want {
				t.Errorf("crossings = %d, want %d", got, test.want)
			}
		})
	}
}

func TestEvidencePassesTheWriteFloor(t *testing.T) {
	tests := []struct {
		name     string
		hours    []int
		packets  int
		wantKeep bool
	}{
		{name: "one hour is noise", hours: []int{9}, packets: 100, wantKeep: false},
		{name: "two hours but too few packets", hours: []int{9, 10}, packets: 19, wantKeep: false},
		{name: "two hours and enough packets", hours: []int{9, 10}, packets: 20, wantKeep: true},
		{name: "many hours", hours: []int{9, 10, 11, 12}, packets: 400, wantKeep: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var e evidence
			remaining := test.packets
			for _, hour := range test.hours {
				// Spread the readings across the declared hours.
				for i := 0; i < remaining/len(test.hours) || i == 0; i++ {
					observe(&e, -70, at(hour, 0))
					remaining--
					if remaining <= 0 {
						break
					}
				}
			}
			// Top up any remainder inside the last hour so the packet count is exact.
			for remaining > 0 {
				observe(&e, -70, at(test.hours[len(test.hours)-1], 0))
				remaining--
			}

			if got := e.meetsWriteFloor(); got != test.wantKeep {
				t.Errorf("meetsWriteFloor = %v, want %v (packets %d, %d hours)",
					got, test.wantKeep, e.packets, e.activeHours())
			}
		})
	}
}

func TestEvidenceDaySummaryCarriesOnlyAggregates(t *testing.T) {
	var e evidence
	observe(&e, -70, at(9, 0))
	observe(&e, -80, at(9, 30))
	observe(&e, -60, at(15, 0))

	summary := e.summary("addr:aa:bb:cc:dd:ee:01", "2026-11-13", "0x004c:0215")

	if summary.Key != "addr:aa:bb:cc:dd:ee:01" {
		t.Errorf("Key = %q", summary.Key)
	}
	if summary.Day != "2026-11-13" {
		t.Errorf("Day = %q", summary.Day)
	}
	if summary.Packets != 3 {
		t.Errorf("Packets = %d, want 3", summary.Packets)
	}
	if summary.ActiveHours != 2 {
		t.Errorf("ActiveHours = %d, want 2", summary.ActiveHours)
	}
	if summary.HourMask != (1<<9)|(1<<15) {
		t.Errorf("HourMask = %#b", summary.HourMask)
	}
	if summary.Strongest != -60 || summary.Weakest != -80 {
		t.Errorf("Strongest/Weakest = %d/%d, want -60/-80", summary.Strongest, summary.Weakest)
	}
	if summary.Family != "0x004c:0215" {
		t.Errorf("Family = %q", summary.Family)
	}
}

func TestEvidenceResetsForANewDay(t *testing.T) {
	var e evidence
	observe(&e, -70, at(23, 0))
	before := e.summary("k", "2026-11-13", "")
	firstSeen := e.firstSeen

	e.resetDaily()

	// The day's counters must clear, or the next summary would double-count yesterday.
	if e.packets != 0 || e.activeHours() != 0 {
		t.Errorf("after reset packets = %d, hours = %d, want 0/0", e.packets, e.activeHours())
	}
	// The lifetime fields must survive, because they describe the identity rather than the day.
	if !e.firstSeen.Equal(firstSeen) {
		t.Errorf("firstSeen changed across a day reset")
	}
	if e.lifetimePackets != before.Packets {
		t.Errorf("lifetimePackets = %d, want %d", e.lifetimePackets, before.Packets)
	}
}

func TestEvidenceIgnoresUnusableReadings(t *testing.T) {
	// An RSSI of 0 means the source could not measure signal, not that the device is at zero dB.
	// The device was still heard, so it counts as activity — but it must contribute no signal
	// evidence, or it would drag the mean upward and present a distant device as very close.
	var e evidence
	observe(&e, 0, at(9, 0))
	observe(&e, -70, at(9, 1))

	if e.packets != 2 {
		t.Errorf("packets = %d, want 2 (the device was heard both times)", e.packets)
	}
	if e.rssi.Count != 1 {
		t.Errorf("rssi count = %d, want 1 (unusable reading excluded from signal)", e.rssi.Count)
	}
	if got := e.rssi.mean(); got != -70 {
		t.Errorf("mean = %f, want -70 (the unusable reading must not move it)", got)
	}
	if e.strongest != -70 {
		t.Errorf("strongest = %d, want -70 (0 is not the strongest reading)", e.strongest)
	}
}
