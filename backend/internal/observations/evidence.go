package observations

// This file accumulates the evidence classification needs for one identity. It stores no
// packets: every field is a counter or a statistic, so memory grows with the number of
// identities rather than with traffic, and a day of history costs one small row.
//
// The rule that shapes everything here: classify from evidence, never from a conclusion.
// Nothing in this file decides what a device *is*. It records what the device *did*, so the
// class can be recomputed later from the same evidence when the rules change, without a
// migration and without a reclassification job.

import (
	"math"
	"math/bits"
	"time"

	model "node-herder/models/observations"
)

// Accumulation thresholds. They are named constants rather than literals because each one is a
// judgement that a reader should be able to find and argue with.
const (
	// writeFloorHours is how many distinct hours a day must contain before it is worth storing.
	// One sighting is noise: an identity that appears once is a passer-by, and keeping a row per
	// passer-by per day is exactly the database-filling the maintainer rejected.
	writeFloorHours = 2

	// writeFloorPackets is the packet count below which a day is not stored, for the same reason.
	writeFloorPackets = 20

	// mobilityCrossingDB is how far the signal must move before the change counts as movement.
	//
	// Chosen against two measured numbers: adjacent distance bands are 7.5-10 dB apart at the
	// default path-loss assumptions, and ordinary RSSI jitter is around 3 dB. A 6 dB threshold
	// therefore sits above jitter and below band spacing, so it counts real changes of range and
	// ignores a body passing the antenna.
	//
	// Counted in dB rather than in distance bands deliberately. Distance is derived in the
	// frontend from txPowerAt1m and pathLossExponent, which are presentation assumptions; storing
	// history keyed to them would duplicate a path-loss model in Go and silently invalidate every
	// stored summary the day those constants changed.
	mobilityCrossingDB = 6

	// maxHourly is hour 23, the highest bit the hour mask holds.
	maxHourly = 23
)

// evidence accumulates the aggregates classification needs for one identity.
//
// It holds no packets and no identity strings: the caller supplies the identity key when the
// day is promoted to a summary, so an accumulator for an identity that ages out can be dropped
// without leaving identifiers behind.
type evidence struct {
	firstSeen time.Time
	lastSeen  time.Time

	// packets and lifetimePackets are the same value across a day rollover, but they serve
	// different questions. packets answers "how much on this day"; lifetimePackets answers "how
	// long-standing is this identity", which survives a reset.
	packets         uint64
	lifetimePackets uint64

	// hours holds the hours of day this identity has been heard in during the current day. A
	// bitmask rather than a slice: it is one machine word, copies cheaply, and encodes exactly
	// what classification asks (was it heard at 04:00, how many distinct hours).
	hours uint32

	rssi      welfordState
	strongest int
	weakest   int
	hasRSSI   bool

	mobility mobilityState
}

// welfordState is a running mean and variance using Welford's algorithm.
//
// Chosen over "sum of squares" because RSSI is a small number repeated many times, which is
// exactly the shape that loses precision: with ten thousand readings at -70 dB a naive
// implementation drifts, and a variance that should be zero comes out slightly negative. Welford
// updates incrementally and stays accurate, and it needs only two floats, so an accumulator for
// every identity stays cheap.
type welfordState struct {
	Count uint64
	Mean  float64
	M2    float64
}

// observe folds one value into the running statistics.
func (w *welfordState) observe(value float64) {
	w.Count++
	delta := value - w.Mean
	w.Mean += delta / float64(w.Count)
	w.M2 += delta * (value - w.Mean)
}

// mean returns the running mean, or zero when nothing has been observed.
func (w *welfordState) mean() float64 {
	if w.Count == 0 {
		return 0
	}
	return w.Mean
}

// variance returns the population variance, or zero when fewer than two values have been
// observed. It is clamped at zero because floating-point cancellation can produce a tiny
// negative number for identical readings, and a negative variance is nonsense to a consumer.
func (w *welfordState) variance() float64 {
	if w.Count < 2 {
		return 0
	}
	variance := w.M2 / float64(w.Count)
	if variance < 0 {
		return 0
	}
	return variance
}

// mobilityState counts how often the signal moved by more than mobilityCrossingDB.
//
// It keeps an anchor level and a direction. Noise alone cannot produce a crossing, because the
// signal must travel 6 dB away from the anchor before anything is counted; once counted, the
// anchor moves to the new level, so a monotonic walk across the room counts once per 6 dB
// rather than once per packet.
type mobilityState struct {
	anchored  bool
	anchor    float64
	crossings uint32
}

// observe folds one smoothed reading into the crossing count.
func (m *mobilityState) observe(smoothed float64) {
	if !m.anchored {
		m.anchored = true
		m.anchor = smoothed
		return
	}
	if math.Abs(smoothed-m.anchor) < mobilityCrossingDB {
		return
	}
	m.crossings++
	m.anchor = smoothed
}

// observe folds one observation into the accumulator.
//
// A reading of RSSI 0 is skipped: it means the source could not measure the signal, not that
// the device is at zero dB. Counting it would pull the mean upward and present a distant device
// as extremely close, which is worse than having no reading at all.
func (e *evidence) observe(observation model.Observation) {
	at := observation.ObservedAt
	if e.firstSeen.IsZero() {
		e.firstSeen = at
	}
	e.lastSeen = at

	hour := at.UTC().Hour()
	if hour >= 0 && hour <= maxHourly {
		e.hours |= 1 << uint(hour)
	}

	e.packets++
	e.lifetimePackets++

	if observation.RSSI == 0 {
		return
	}
	if !e.hasRSSI {
		e.strongest = observation.RSSI
		e.weakest = observation.RSSI
		e.hasRSSI = true
	}
	if observation.RSSI > e.strongest {
		e.strongest = observation.RSSI
	}
	if observation.RSSI < e.weakest {
		e.weakest = observation.RSSI
	}

	e.rssi.observe(float64(observation.RSSI))
	e.mobility.observe(float64(observation.RSSI))
}

// hourMask returns the hours of day with at least one observation.
func (e *evidence) hourMask() uint32 { return e.hours }

// activeHours returns how many distinct hours of the day were observed.
func (e *evidence) activeHours() uint8 { return uint8(bits.OnesCount32(e.hours)) }

// meetsWriteFloor reports whether this day is worth storing.
//
// The floor is enforced here, at the write boundary, rather than by reviewing what gets stored.
// A rule that lives in code cannot be forgotten; a rule that lives in a convention can.
func (e *evidence) meetsWriteFloor() bool {
	return e.activeHours() >= writeFloorHours && e.packets >= writeFloorPackets
}

// resetDaily clears the per-day counters at a day boundary, keeping the lifetime fields.
//
// Packet counts and the hour mask describe one day, so carrying them across midnight would
// double-count the previous day into the next. firstSeen and lifetimePackets describe the
// identity, so they survive; losing them would make a long-standing device look like an arrival
// every midnight and destroy the evidence classification depends on.
func (e *evidence) resetDaily() {
	e.packets = 0
	e.hours = 0
	e.rssi = welfordState{}
	e.mobility = mobilityState{}
	e.strongest = 0
	e.weakest = 0
	e.hasRSSI = false
}

// summary renders the current day as a storable row.
//
// The identity key and family are passed in rather than held on the accumulator so that an
// accumulator belonging to an identity that aged out carries no identifier of its own.
func (e *evidence) summary(key, day, family string) model.DailySummary {
	return model.DailySummary{
		Key:         key,
		Day:         day,
		Packets:     e.packets,
		ActiveHours: e.activeHours(),
		HourMask:    e.hourMask(),
		RSSIMean:    e.rssi.mean(),
		RSSIM2:      e.rssi.M2,
		Strongest:   e.strongest,
		Weakest:     e.weakest,
		Crossings:   e.mobility.crossings,
		Family:      family,
	}
}
