// Package observations defines the transport-agnostic record of a radio transmission that an
// enrolled device observed, and the bounded snapshot the browser renders.
package observations

// DailySummary is one identity's behaviour on one UTC day: the evidence classification reads,
// and nothing else.
//
// It deliberately contains no raw observations, packet contents or payloads. A summary says how
// often, how strong and how mobile a device was; it does not say what it transmitted. That is
// what keeps retaining it defensible, since it cannot reconstruct anything about the traffic
// itself.
//
// The class is absent by design. Classification is computed from this evidence each time it is
// needed, so changing the rules needs no migration and no reclassification job, and a device
// whose behaviour changes is reclassified from the days that actually happened rather than from
// a stale verdict.
type DailySummary struct {
	// Key is the identity the summary belongs to, matching Identity.Key in the live snapshot.
	Key string `json:"key"`
	// Day is the UTC day in YYYY-MM-DD form, and is part of the storage key with Key.
	Day string `json:"day"`

	// Packets is how many usable observations this identity produced on this day.
	Packets uint64 `json:"packets"`
	// ActiveHours is how many distinct hours of the day contained at least one observation.
	ActiveHours uint8 `json:"activeHours"`
	// HourMask has bit N set when hour N (UTC) contained at least one observation. The mask is
	// stored rather than ActiveHours alone because classification distinguishes when a device
	// was heard (04:00 every night) from merely how often.
	HourMask uint32 `json:"hourMask"`

	// RSSIMean is the mean signal strength in dBm over the day.
	RSSIMean float64 `json:"rssiMean"`
	// RSSIM2 is Welford's accumulated sum of squared deviations, not a variance. It is stored
	// raw because it can be merged exactly across sources and partitions; derive the variance
	// with Var() rather than dividing by Packets at each call site.
	RSSIM2 float64 `json:"rssiM2"`
	// Strongest and Weakest bound the day's readings, so "never heard strongly" survives
	// storage as a fact rather than an inference from the mean.
	Strongest int `json:"strongest"`
	Weakest   int `json:"weakest"`

	// Crossings counts how often the smoothed signal moved by more than the mobility threshold.
	// It is the mobility signal: variance conflates a device that moved with a body passing the
	// antenna, whereas a threshold crossing is a real change of range. Counted in dB, not in
	// distance bands, so stored history does not depend on a presentation constant.
	Crossings uint32 `json:"crossings"`

	// Family groups rotating addresses that share a vendor and payload signature, so a phone
	// advertising forty rotating addresses is classified once rather than forty times.
	Family string `json:"family,omitempty"`
}

// Var returns the population variance of the day's signal strength, or zero when fewer than two
// readings contributed. It is clamped at zero because floating-point cancellation can leave a
// tiny negative value for identical readings, which is not a meaningful variance.
func (s DailySummary) Var() float64 {
	if s.Packets < 2 {
		return 0
	}
	variance := s.RSSIM2 / float64(s.Packets)
	if variance < 0 {
		return 0
	}
	return variance
}

// Transition records that an identity's class changed.
//
// It is stored rather than derived because a change is an event: the evidence that produced it may
// later be pruned, and "this became personal on Tuesday" is exactly what an operator needs when a
// device behaves unexpectedly.
//
// It is a log rather than a derivation, and deliberately self-contained: it states the reason at
// the time, because the summaries that justified it will eventually be gone. Retention matches the
// summaries so that every transition still displayable is still explainable.
type Transition struct {
	// Key is the identity the change belongs to.
	Key string `json:"key"`
	// Day is the UTC day of the change, and is part of the storage key with Key.
	Day string `json:"day"`
	// At is the exact time of the change in RFC 3339 form.
	At string `json:"at"`
	// From and To are the classes before and after.
	From string `json:"from"`
	To   string `json:"to"`
	// Reason states why, in the operator's terms, so the row explains itself.
	Reason string `json:"reason"`
}
