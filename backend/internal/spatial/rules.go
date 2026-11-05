// Package spatial owns the hub-side spatial model: geometry, radar tracking rules and the
// presence signals derived from them. It holds no transport or rendering code. Rules mirror
// the browser's spatial rules that matter for occupancy, because the hub owns the authoritative
// values when automations run without a browser.
package spatial

import (
	"math"
	"time"

	"node-herder/models/settings"
)

// Filter and tracking constants mirror frontend/src/domain/spatial so the dot and the
// automation agree on what counts as movement. They are documentation, not shared code.
const (
	deadbandMM        = 35.0
	maxSpeedMMPerS    = 3000.0
	gap               = 600 * time.Millisecond
	retain            = 8 * time.Second
	assocGateMM       = 700.0
	assocSpeedMMPerS  = 1500.0
	mergeMM           = 350.0
	mergeSamples      = 3
	zoneEnterSamples  = 2
	zoneExitGrace     = 800 * time.Millisecond
	filterMinCutoff   = 1.0
	filterBeta        = 0.01
	filterDerivCutoff = 1.0
	// maxDwellStep bounds how much dwell one update can add. A stalled feed must not turn a
	// missing frame gap into a large occupancy claim.
	maxDwellStep = 5 * time.Second
)

// Sample is one mapped slot position in millimetres with the time it was accepted.
type Sample struct {
	Slot int
	X    float64
	Y    float64
}

// Track is a deduplicated target with a stable local identifier.
type Track struct {
	ID   int
	Slot int
	X    float64
	Y    float64
	At   time.Time
}

// ZoneSignal is the derived state of one saved overlay.
type ZoneSignal struct {
	Occupied bool
	Count    int
	Dwell    time.Duration
}

// Signals is one computed presence snapshot for a source device.
type Signals struct {
	Tracks      []Track
	Occupied    bool
	TargetCount int
	Zones       map[string]ZoneSignal
	LastZone    string
	EmptyFor    time.Duration
}

// millimetres converts a declared length value to millimetres. The unit table matches the
// frame contract; unknown units are unusable rather than guessed.
func millimetres(value float64, unit string) (float64, bool) {
	switch unit {
	case "mm":
		return value, true
	case "cm":
		return value * 10, true
	case "m":
		return value * 1000, true
	}
	return 0, false
}

type axis struct {
	raw        float64
	smoothed   float64
	derivative float64
}

type slotFilter struct {
	x     axis
	y     axis
	last  time.Time
	ready bool
}

func alpha(cutoff, dt float64) float64 {
	timeConstant := 1 / (2 * math.Pi * cutoff)
	return 1 / (1 + timeConstant/dt)
}

func smoothAxis(a *axis, value, dt float64) {
	rawDerivative := (value - a.raw) / dt
	a.raw = value
	a.derivative += alpha(filterDerivCutoff, dt) * (rawDerivative - a.derivative)
	cutoff := filterMinCutoff + filterBeta*math.Abs(a.derivative)
	a.smoothed += alpha(cutoff, dt) * (value - a.smoothed)
}

// apply returns the smoothed position and velocity. Implausible jumps are rejected in favour
// of the previous position so one bad frame cannot move a track across the room.
func (f *slotFilter) apply(x, y float64, at time.Time) (float64, float64, float64, float64) {
	if !f.ready || at.Sub(f.last) > gap {
		f.x = axis{raw: x, smoothed: x}
		f.y = axis{raw: y, smoothed: y}
		f.last = at
		f.ready = true
		return x, y, 0, 0
	}
	dt := at.Sub(f.last).Seconds()
	if dt <= 0 {
		return f.x.smoothed, f.y.smoothed, f.x.derivative, f.y.derivative
	}
	distance := math.Hypot(x-f.x.raw, y-f.y.raw)
	if distance > 0 && distance/dt > maxSpeedMMPerS {
		return f.x.smoothed, f.y.smoothed, 0, 0
	}
	if dt*1000 >= float64(gap/time.Millisecond) {
		f.last = at
		return f.x.smoothed, f.y.smoothed, f.x.derivative, f.y.derivative
	}
	smoothAxis(&f.x, x, dt)
	smoothAxis(&f.y, y, dt)
	if math.Hypot(f.x.derivative, f.y.derivative) < 50 && math.Hypot(f.x.smoothed-f.x.raw, f.y.smoothed-f.y.raw) < deadbandMM {
		// Inside the dead band the reading is indistinguishable from jitter.
		f.x.smoothed = f.x.raw
		f.y.smoothed = f.y.raw
	}
	f.last = at
	return f.x.smoothed, f.y.smoothed, f.x.derivative, f.y.derivative
}

// slotPoint is one filtered sample offered to track association.
type slotPoint struct {
	slot, id     int
	x, y, vx, vy float64
	at           time.Time
}

type internalTrack struct {
	id, slot int
	x, y     float64
	at       time.Time
	vx, vy   float64
}

type zoneState struct {
	occupied     bool
	insideHits   int
	outsideSince time.Time
	dwell        time.Duration
	lastCount    int
}

// Tracker owns the per-source presence state: smoothing, identity, zone hysteresis and
// counters. It is not safe for concurrent use; the engine calls it from one worker.
type Tracker struct {
	filters         map[int]*slotFilter
	tracks          []internalTrack
	nextID          int
	mergeCount      map[[2]int]int
	zones           map[string]*zoneState
	emptiedAt       time.Time
	lastZone        string
	lastTargetCount int
	lastUpdate      time.Time
	started         bool
}

// NewTracker creates empty presence state for one source device.
func NewTracker() *Tracker {
	return &Tracker{filters: map[int]*slotFilter{}, mergeCount: map[[2]int]int{}, zones: map[string]*zoneState{}}
}

// Update computes the presence snapshot for one set of slot samples and the saved zones.
// Samples with non-finite coordinates are ignored; a slot that disappears keeps its identity
// for the retention window but is not drawn or counted.
func (t *Tracker) Update(samples []Sample, zones []settings.SpatialZone, at time.Time) Signals {
	if !t.started {
		t.started = true
		t.emptiedAt = at
	}
	elapsed := at.Sub(t.lastUpdate)
	if t.lastUpdate.IsZero() {
		elapsed = 0
	}
	t.lastUpdate = at

	points := make([]slotPoint, 0, len(samples))
	for _, sample := range samples {
		if math.IsNaN(sample.X) || math.IsInf(sample.X, 0) || math.IsNaN(sample.Y) || math.IsInf(sample.Y, 0) {
			continue
		}
		filter := t.filters[sample.Slot]
		if filter == nil {
			filter = &slotFilter{}
			t.filters[sample.Slot] = filter
		}
		x, y, vx, vy := filter.apply(sample.X, sample.Y, at)
		points = append(points, slotPoint{slot: sample.Slot, x: x, y: y, vx: vx, vy: vy, at: at})
	}

	t.assign(points, at)
	visible := t.visibleTracks(at)

	zonesState := make(map[string]ZoneSignal, len(zones))
	occupied := false
	for _, zone := range zones {
		state := t.zones[zone.ID]
		if state == nil {
			state = &zoneState{}
			t.zones[zone.ID] = state
		}
		count := 0
		for _, track := range visible {
			if insideZone(zone, zones, track.X, track.Y) {
				count++
			}
		}
		inside := count > 0
		if inside {
			state.outsideSince = time.Time{}
			if !state.occupied {
				state.insideHits++
				if state.insideHits >= zoneEnterSamples {
					state.occupied = true
					state.dwell = 0
					t.lastZone = zone.Name
				}
			} else {
				if elapsed > 0 {
					state.dwell += minDuration(elapsed, maxDwellStep)
				}
				t.lastZone = zone.Name
			}
		} else {
			state.insideHits = 0
			if state.occupied {
				if state.outsideSince.IsZero() {
					state.outsideSince = at
				} else if at.Sub(state.outsideSince) >= zoneExitGrace {
					state.occupied = false
					state.outsideSince = time.Time{}
					state.lastCount = 0
				}
			}
		}
		// Count follows debounced occupancy: an unoccupied zone never reports targets, and the
		// exit grace keeps the last real count instead of flickering to zero.
		publishedCount := 0
		if state.occupied {
			if count > 0 {
				state.lastCount = count
			}
			publishedCount = state.lastCount
		}
		zonesState[zone.ID] = ZoneSignal{Occupied: state.occupied, Count: publishedCount, Dwell: state.dwell}
		if zone.Kind == "detection" && state.occupied {
			occupied = true
		}
	}

	// target_count counts distinct tracks inside occupied detection zones, so overlapping zones
	// do not multiply one person and the count agrees with the debounced occupancy.
	targetTracks := map[int]struct{}{}
	if occupied {
		for _, zone := range zones {
			if zone.Kind != "detection" || !zonesState[zone.ID].Occupied {
				continue
			}
			for _, track := range visible {
				if insideZone(zone, zones, track.X, track.Y) {
					targetTracks[track.ID] = struct{}{}
				}
			}
		}
	}
	if len(targetTracks) > 0 {
		t.lastTargetCount = len(targetTracks)
	}
	targetCount := 0
	if occupied {
		targetCount = t.lastTargetCount
	} else {
		t.lastTargetCount = 0
	}

	if len(visible) > 0 {
		t.emptiedAt = at
	}
	emptyFor := time.Duration(0)
	if len(visible) == 0 && !t.emptiedAt.IsZero() {
		emptyFor = at.Sub(t.emptiedAt)
	}

	return Signals{
		Tracks:      visible,
		Occupied:    occupied,
		TargetCount: targetCount,
		Zones:       zonesState,
		LastZone:    t.lastZone,
		EmptyFor:    emptyFor,
	}
}

// assign matches points to retained tracks: the same slot first, then the nearest track
// within a speed-scaled gate. Unmatched points create identities; identities live for the
// retention window so a brief loss does not rename a person's dot.
func (t *Tracker) assign(points []slotPoint, at time.Time) {
	// keep identities that are still within retention
	kept := t.tracks[:0]
	for _, track := range t.tracks {
		if at.Sub(track.at) <= retain {
			kept = append(kept, track)
		}
	}
	t.tracks = kept

	used := make([]bool, len(t.tracks))
	for _, candidate := range points {
		match := -1
		for index, track := range t.tracks {
			if used[index] || !withinGate(track, candidate.x, candidate.y, candidate.vx, candidate.vy, at) {
				continue
			}
			if track.slot == candidate.slot {
				match = index
				break
			}
			if match < 0 || math.Hypot(track.x-candidate.x, track.y-candidate.y) < math.Hypot(t.tracks[match].x-candidate.x, t.tracks[match].y-candidate.y) {
				match = index
			}
		}
		if match < 0 {
			t.nextID++
			t.tracks = append(t.tracks, internalTrack{id: t.nextID, slot: candidate.slot, x: candidate.x, y: candidate.y, vx: candidate.vx, vy: candidate.vy, at: at})
			used = append(used, true)
			continue
		}
		used[match] = true
		t.tracks[match].slot = candidate.slot
		t.tracks[match].x = candidate.x
		t.tracks[match].y = candidate.y
		t.tracks[match].vx = candidate.vx
		t.tracks[match].vy = candidate.vy
		t.tracks[match].at = at
	}
	t.suppressDuplicates(at)
}

func withinGate(track internalTrack, x, y, vx, vy float64, at time.Time) bool {
	elapsed := math.Max(0, at.Sub(track.at).Seconds())
	gate := assocGateMM + math.Hypot(track.vx, track.vy)*elapsed
	return math.Hypot(track.x-x, track.y-y) <= gate
}

// suppressDuplicates hides the newer of two identities that stay within the merge distance,
// so one person reported by two slots is counted once. Separation reverses it.
func (t *Tracker) suppressDuplicates(at time.Time) {
	fresh := make([]internalTrack, 0, len(t.tracks))
	for _, track := range t.tracks {
		if at.Sub(track.at) <= time.Second {
			fresh = append(fresh, track)
		}
	}
	seen := map[[2]int]bool{}
	for i := 0; i < len(fresh); i++ {
		for j := i + 1; j < len(fresh); j++ {
			if math.Hypot(fresh[i].x-fresh[j].x, fresh[i].y-fresh[j].y) > mergeMM {
				continue
			}
			key := [2]int{fresh[i].id, fresh[j].id}
			if key[0] > key[1] {
				key[0], key[1] = key[1], key[0]
			}
			seen[key] = true
			t.mergeCount[key]++
		}
	}
	for key := range t.mergeCount {
		if !seen[key] {
			delete(t.mergeCount, key)
		}
	}
}

// visibleTracks returns recently updated identities, dropping the newer half of any duplicate
// pair that has persisted for the merge window. Tracks with no fresh sample are excluded.
func (t *Tracker) visibleTracks(at time.Time) []Track {
	suppressed := map[int]bool{}
	for key, count := range t.mergeCount {
		if count >= mergeSamples {
			suppressed[key[1]] = true
		}
	}
	visible := make([]Track, 0, len(t.tracks))
	for _, track := range t.tracks {
		if suppressed[track.id] || !track.at.Equal(at) {
			continue
		}
		visible = append(visible, Track{ID: track.id, Slot: track.slot, X: track.x, Y: track.y, At: track.at})
	}
	return visible
}

func minDuration(left, right time.Duration) time.Duration {
	if left < right {
		return left
	}
	return right
}

// containsPoint reports whether a point lies inside a polygon, including its edges.
func containsPoint(points []settings.SpatialPoint, x, y float64) bool {
	inside := false
	for i, j := 0, len(points)-1; i < len(points); j, i = i, i+1 {
		a, b := points[j], points[i]
		onEdge := math.Abs((b.X-a.X)*(y-a.Y)-(b.Y-a.Y)*(x-a.X)) < 0.001 &&
			x >= math.Min(a.X, b.X) && x <= math.Max(a.X, b.X) &&
			y >= math.Min(a.Y, b.Y) && y <= math.Max(a.Y, b.Y)
		if onEdge {
			return true
		}
		if (a.Y > y) != (b.Y > y) && x < (b.X-a.X)*(y-a.Y)/(b.Y-a.Y)+a.X {
			inside = !inside
		}
	}
	return inside
}

// insideZone applies exclusion masks to detection zones; other kinds test their own polygon.
func insideZone(zone settings.SpatialZone, zones []settings.SpatialZone, x, y float64) bool {
	if !containsPoint(zone.Points, x, y) {
		return false
	}
	if zone.Kind != "detection" {
		return true
	}
	for _, other := range zones {
		if other.Kind == "exclusion" && containsPoint(other.Points, x, y) {
			return false
		}
	}
	return true
}
