package spatial

import (
	"math"
	"testing"
	"time"

	"node-herder/models/settings"
)

var base = time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)

func deskZone() settings.SpatialZone {
	return settings.SpatialZone{
		ID: "desk", Name: "Desk", Kind: "detection", Shape: "rectangle",
		Points: []settings.SpatialPoint{{X: -500, Y: 1000}, {X: 500, Y: 1000}, {X: 500, Y: 2000}, {X: -500, Y: 2000}},
	}
}

func at(ms int) time.Time { return base.Add(time.Duration(ms) * time.Millisecond) }

func feed(t *Tracker, tracker func(int) Sample, from, to, step int, floor int) Signals {
	var last Signals
	for ms := from; ms <= to; ms += step {
		last = t.Update([]Sample{tracker(ms)}, []settings.SpatialZone{deskZone()}, at(ms))
	}
	return last
}

func TestOccupancyEntryAndExitGrace(t *testing.T) {
	tracker := NewTracker()
	inside := func(int) Sample { return Sample{Slot: 1, X: 0, Y: 1500} }
	first := tracker.Update([]Sample{inside(0)}, []settings.SpatialZone{deskZone()}, at(0))
	if first.Zones["desk"].Occupied {
		t.Fatal("one sample must not occupy a zone")
	}
	second := tracker.Update([]Sample{inside(250)}, []settings.SpatialZone{deskZone()}, at(250))
	if !second.Zones["desk"].Occupied {
		t.Fatal("second consecutive sample must occupy")
	}
	// A cleared slot stops counting immediately; the grace then runs from that update.
	grace := tracker.Update(nil, []settings.SpatialZone{deskZone()}, at(500))
	if !grace.Zones["desk"].Occupied {
		t.Fatal("occupancy must survive the exit grace")
	}
	missing := tracker.Update(nil, []settings.SpatialZone{deskZone()}, at(1000))
	if !missing.Zones["desk"].Occupied {
		t.Fatal("occupancy must survive inside the grace")
	}
	cleared := tracker.Update(nil, []settings.SpatialZone{deskZone()}, at(1400))
	if cleared.Zones["desk"].Occupied {
		t.Fatal("occupancy must clear after the grace")
	}
}

func TestBoundaryChatterTogglesAtMostOnce(t *testing.T) {
	tracker := NewTracker()
	transitions := 0
	previous := false
	for step := 0; step < 20; step++ {
		sample := Sample{Slot: 1, X: 0, Y: 1500}
		if step%2 == 1 {
			sample = Sample{Slot: 1, X: 0, Y: 500} // 500 mm outside the zone, 500 ms later
		}
		signals := tracker.Update([]Sample{sample}, []settings.SpatialZone{deskZone()}, at(step*500))
		if signals.Zones["desk"].Occupied != previous {
			transitions++
			previous = signals.Zones["desk"].Occupied
		}
	}
	if transitions > 1 {
		t.Fatalf("boundary chatter toggled %d times", transitions)
	}
}

func TestDuplicateSlotsCountOnce(t *testing.T) {
	tracker := NewTracker()
	samples := []Sample{{Slot: 1, X: 0, Y: 1500}, {Slot: 2, X: 120, Y: 1500}}
	var last Signals
	for step := 0; step < 4; step++ {
		last = tracker.Update(samples, []settings.SpatialZone{deskZone()}, at(step*250))
	}
	if len(last.Tracks) != 1 {
		t.Fatalf("visible tracks = %d, want 1", len(last.Tracks))
	}
	if last.Zones["desk"].Count != 1 || last.TargetCount != 1 {
		t.Fatalf("counts = %+v target=%d, want one", last.Zones["desk"], last.TargetCount)
	}
}

func TestDwellAccumulatesAndResetsPerEpisode(t *testing.T) {
	tracker := NewTracker()
	inside := func(ms int) Sample { return Sample{Slot: 1, X: 0, Y: 1500} }
	var signals Signals
	for ms := 0; ms <= 3000; ms += 250 {
		signals = tracker.Update([]Sample{inside(ms)}, []settings.SpatialZone{deskZone()}, at(ms))
	}
	if dwell := signals.Zones["desk"].Dwell; dwell < 2*time.Second || dwell > 3300*time.Millisecond {
		t.Fatalf("dwell = %v, want about three seconds", dwell)
	}

	tracker.Update(nil, []settings.SpatialZone{deskZone()}, at(3250))
	cleared := tracker.Update(nil, []settings.SpatialZone{deskZone()}, at(4150))
	if cleared.Zones["desk"].Occupied {
		t.Fatal("zone must clear before dwell resets")
	}
	reentered := tracker.Update([]Sample{inside(4400)}, []settings.SpatialZone{deskZone()}, at(4400))
	reentered = tracker.Update([]Sample{inside(4650)}, []settings.SpatialZone{deskZone()}, at(4650))
	if reentered.Zones["desk"].Dwell > time.Second {
		t.Fatalf("dwell carried across episodes: %v", reentered.Zones["desk"].Dwell)
	}
}

func TestEmptySecondsFollowTracks(t *testing.T) {
	tracker := NewTracker()
	empty := tracker.Update(nil, []settings.SpatialZone{deskZone()}, at(0))
	if empty.EmptyFor != 0 {
		t.Fatalf("empty at start = %v, want 0", empty.EmptyFor)
	}
	empty = tracker.Update(nil, []settings.SpatialZone{deskZone()}, at(2000))
	if empty.EmptyFor != 2*time.Second {
		t.Fatalf("empty = %v, want 2s", empty.EmptyFor)
	}
	tracker.Update([]Sample{{Slot: 1, X: 0, Y: 1500}}, []settings.SpatialZone{deskZone()}, at(2250))
	empty = tracker.Update([]Sample{{Slot: 1, X: 0, Y: 1500}}, []settings.SpatialZone{deskZone()}, at(2500))
	if empty.EmptyFor != 0 {
		t.Fatalf("empty while tracked = %v, want 0", empty.EmptyFor)
	}
}

func TestLastZoneRetainsMostRecentOccupation(t *testing.T) {
	tracker := NewTracker()
	for step := 0; step < 3; step++ {
		tracker.Update([]Sample{{Slot: 1, X: 0, Y: 1500}}, []settings.SpatialZone{deskZone()}, at(step*250))
	}
	for step := 3; step < 8; step++ {
		tracker.Update([]Sample{{Slot: 1, X: 0, Y: 1500}}, []settings.SpatialZone{deskZone()}, at(step*250))
	}
	outside := Sample{Slot: 1, X: 0, Y: 5000}
	signals := tracker.Update([]Sample{outside}, []settings.SpatialZone{deskZone()}, at(2500))
	if signals.LastZone != "Desk" {
		t.Fatalf("last zone = %q, want Desk", signals.LastZone)
	}
}

func TestExclusionMaskBlocksDetection(t *testing.T) {
	mask := settings.SpatialZone{
		ID: "fan", Name: "Fan", Kind: "exclusion", Shape: "rectangle",
		Points: []settings.SpatialPoint{{X: -500, Y: 1000}, {X: 0, Y: 1000}, {X: 0, Y: 1500}, {X: -500, Y: 1500}},
	}
	tracker := NewTracker()
	sample := []Sample{{Slot: 1, X: -400, Y: 1200}}
	var last Signals
	for step := 0; step < 3; step++ {
		last = tracker.Update(sample, []settings.SpatialZone{deskZone(), mask}, at(step*250))
	}
	if last.Zones["desk"].Occupied {
		t.Fatal("exclusion mask must block detection occupancy")
	}
}

func TestImplausibleJumpDoesNotMoveTheTrack(t *testing.T) {
	tracker := NewTracker()
	tracker.Update([]Sample{{Slot: 1, X: 0, Y: 1500}}, []settings.SpatialZone{deskZone()}, at(0))
	tracker.Update([]Sample{{Slot: 1, X: 0, Y: 1500}}, []settings.SpatialZone{deskZone()}, at(250))
	signals := tracker.Update([]Sample{{Slot: 1, X: 5000, Y: 1500}}, []settings.SpatialZone{deskZone()}, at(500))
	if math.Abs(signals.Tracks[0].X) > 500 {
		t.Fatalf("track teleported to %v", signals.Tracks[0].X)
	}
}

func TestMillimetresUsesDeclaredUnitsOnly(t *testing.T) {
	if value, ok := millimetres(120, "cm"); !ok || value != 1200 {
		t.Fatalf("cm = %v %t", value, ok)
	}
	if _, ok := millimetres(1, "ft"); ok {
		t.Fatal("unknown unit accepted")
	}
}

func TestOverlappingDetectionZonesCountOneTarget(t *testing.T) {
	near := deskZone()
	near.ID, near.Name = "sofa", "Sofa"
	near.Points = []settings.SpatialPoint{{X: 300, Y: 1200}, {X: 900, Y: 1200}, {X: 900, Y: 1800}, {X: 300, Y: 1800}}
	tracker := NewTracker()
	var last Signals
	for step := 0; step < 2; step++ {
		last = tracker.Update([]Sample{{Slot: 1, X: 500, Y: 1500}}, []settings.SpatialZone{deskZone(), near}, at(step*250))
	}
	if last.TargetCount != 1 {
		t.Fatalf("target count = %d, want one distinct target", last.TargetCount)
	}
	if last.Zones["desk"].Count != 1 || last.Zones["sofa"].Count != 1 {
		t.Fatalf("zone counts = %d/%d", last.Zones["desk"].Count, last.Zones["sofa"].Count)
	}
}

func TestClearedDuplicateDisappearsImmediately(t *testing.T) {
	tracker := NewTracker()
	for step := 0; step < 4; step++ {
		tracker.Update([]Sample{{Slot: 1, X: 0, Y: 1500}, {Slot: 2, X: 120, Y: 1500}}, []settings.SpatialZone{deskZone()}, at(step*250))
	}
	// Slot 2 stops reporting; only slot 1 remains and the count must not jump back to two.
	signals := tracker.Update([]Sample{{Slot: 1, X: 0, Y: 1500}}, []settings.SpatialZone{deskZone()}, at(1000))
	if signals.TargetCount != 1 || len(signals.Tracks) != 1 {
		t.Fatalf("cleared duplicate still visible: tracks=%d count=%d", len(signals.Tracks), signals.TargetCount)
	}
}
