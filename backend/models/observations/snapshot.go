// Package observations defines the transport-agnostic record of a radio transmission that
// an enrolled device observed, and the bounded snapshot the browser renders.
package observations

import "time"

// MaxSnapshotIdentities bounds one broadcast snapshot. A busy household can hear hundreds
// of devices, most of them transient; the newest identities win, and Truncated reports the
// loss rather than growing the payload without limit (NH-06).
const MaxSnapshotIdentities = 256

// SourceDescriptor describes one observing device in a snapshot. Facets lists the filter
// dimensions this source can populate, so the viewer renders controls from data instead of
// hardcoding per-source knowledge.
type SourceDescriptor struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Media       []Medium          `json:"media"`
	Facets      []string          `json:"facets"`
	Diagnostics SourceDiagnostics `json:"diagnostics"`
}

// SourceDiagnostics reports what a source has actually delivered in this process. It exists to
// distinguish "connected but silent" from "not configured", so an empty viewer is diagnosable
// rather than mysterious. Counters reset on restart and are never metrics history.
type SourceDiagnostics struct {
	// Advertisements counts advertisement entries received from the source.
	Advertisements uint64 `json:"advertisements"`
	// Decoded counts entries that produced a usable identity.
	Decoded uint64 `json:"decoded"`
	// Dropped counts observations discarded because the aggregation queue was full.
	Dropped uint64 `json:"dropped"`
	// LastAdvertisement is when the most recent advertisement arrived; nil means never.
	LastAdvertisement *time.Time `json:"lastAdvertisement,omitempty"`
}

// SourceReading is one source's view of an identity. RSSI is absent when the source
// reported no value, which is distinct from a measured zero.
type SourceReading struct {
	ID   string `json:"id"`
	RSSI *int   `json:"rssi,omitempty"`
}

// Trend values reported in IdentityView.Trend. They describe the direction of change in
// received signal strength, which is a proximity proxy, never a velocity.
const (
	// TrendUnknown means there is not enough recent history to describe a direction.
	TrendUnknown = "unknown"
	// TrendSteady means any change stayed inside the jitter deadband.
	TrendSteady = "steady"
	// TrendApproaching means the signal strengthened: the device is likely closer.
	TrendApproaching = "approaching"
	// TrendReceding means the signal weakened: the device is likely further away.
	TrendReceding = "receding"
)

// IdentityView is the aggregated, renderable state of one identity.
type IdentityView struct {
	Key    string `json:"key"`
	Kind   string `json:"kind"`
	Vendor uint16 `json:"vendor,omitempty"`
	Stable bool   `json:"stable"`
	Label  string `json:"label,omitempty"`
	// Decision is the operator's triage state: unknown, known or ignored.
	Decision string `json:"decision"`
	// FirstSeen is when this process first heard the identity, so a viewer can offer
	// "recently appeared" instead of inferring novelty from the triage decision, which is
	// unknown for every identity on a first run and would therefore match everything.
	FirstSeen time.Time `json:"firstSeen"`
	LastSeen  time.Time `json:"lastSeen"`
	Packets   uint64    `json:"packets"`
	// RSSI is the newest single reading, kept for diagnostics.
	RSSI *int `json:"rssi,omitempty"`
	// RSSISmoothed is the median of the recent window and is what distance must be derived
	// from. A single reading swings several dB, which through the log-distance curve becomes a
	// multi-metre jump; the median removes that without lagging a real move.
	RSSISmoothed *int            `json:"rssiSmoothed,omitempty"`
	Sources      []SourceReading `json:"sources"`
	// Trend describes the direction of change over the recent window: approaching, receding,
	// steady or unknown. Unknown is never presented as steady.
	Trend string `json:"trend"`
	// TrendDeltaDB is the smoothed change in dB across the window. Positive means the signal
	// strengthened. It is zero when Trend is unknown.
	TrendDeltaDB float64 `json:"trendDeltaDb"`
	// Signal is the recent smoothed history as signal strengths in dBm, oldest first, for a
	// sparkline. It is bounded by the aggregation window and never persisted.
	Signal []int `json:"signal,omitempty"`

	// Classification is the advisory behavioural class, computed from stored evidence.
	//
	// It is omitted entirely when no summaries exist for the identity, so a consumer can tell
	// "not classified" from "classified as learning". The class is always advisory: the triage
	// decision above outranks it for display, and both are carried so the UI can show the
	// disagreement rather than silently picking one.
	Classification *ClassificationView `json:"classification,omitempty"`
}

// ClassificationView is the class and the evidence behind it, as rendered.
//
// The evidence travels with the class because a bare label cannot be argued with. An operator
// who disagrees with "infrastructure" needs to see "present 23 of 24 hours, never moved" to
// decide whether the system or their expectation is wrong.
type ClassificationView struct {
	// Class is the advisory class, or "learning" when evidence is insufficient.
	Class string `json:"class"`
	// Provisional is true when the class came from the first-packet pre-filter, so the UI can
	// present it as unsettled rather than as a verdict.
	Provisional bool `json:"provisional,omitempty"`
	// ShowByDefault reports whether the identity belongs in the default view. The pre-filter
	// decides this before any history exists, so it must travel with the class rather than being
	// re-derived by every consumer.
	ShowByDefault bool `json:"showByDefault"`
	// Candidate and ChangePending describe a class change held back by hysteresis.
	Candidate     string `json:"candidate,omitempty"`
	ChangePending bool   `json:"changePending,omitempty"`
	// Confidence is the margin over the runner-up class, 0-1.
	Confidence float64 `json:"confidence"`
	// HoursObserved is how much evidence exists, so the UI can say how far learning has to go.
	HoursObserved int `json:"hoursObserved"`
	// DaysConsidered is how many stored days contributed.
	DaysConsidered int `json:"daysConsidered,omitempty"`
	// Reason is the human-readable justification behind the class.
	Reason string `json:"reason,omitempty"`
	// Advisory is always true for a computed class, and states the precedence rule in the payload
	// rather than only in the UI: the operator's triage decision outranks this, and a consumer that
	// presents a class as a fact instead of a suggestion is misusing it.
	Advisory bool `json:"advisory"`
}

// Snapshot is one complete, self-contained broadcast. It is never a delta: a consumer that
// misses one snapshot reconverges on the next.
type Snapshot struct {
	At         time.Time          `json:"at"`
	Truncated  bool               `json:"truncated"`
	Sources    []SourceDescriptor `json:"sources"`
	Identities []IdentityView     `json:"identities"`
	// Arrivals counts stable identities first seen in this process, which is what an
	// operator means by "new".
	Arrivals int `json:"arrivals"`

	// Observing reports whether any integration currently wants this data.
	//
	// Without it, a consumer cannot tell "nobody is reading the radio" from "the radio hears
	// nothing", which are opposite problems with opposite fixes. It travels in the payload rather
	// than being inferred from an empty list, because an empty list is the normal state of a
	// working installation in a quiet room.
	Observing bool `json:"observing"`

	// Classification says how far the evidence needed to classify has got.
	//
	// It exists because a first day of observation looks identical to a broken classifier: both
	// show every row as "learning". Without this the operator cannot tell "still gathering" from
	// "not working", and the honest state is indistinguishable from a fault.
	Classification ClassificationProgress `json:"classification"`
}

// ClassificationProgress reports whether classification has enough evidence to run.
type ClassificationProgress struct {
	// Identities is how many identities are currently being tracked this process.
	Identities int `json:"identities"`
	// DaysStored is how many identities have at least one stored day of evidence.
	DaysStored int `json:"daysStored"`
	// HoursObserved is the best evidence any single identity has, in observed hours.
	//
	// It is a maximum rather than a total because classification is per identity: one device with
	// 24 hours is enough for that device, while forty devices with one hour each classify nothing.
	HoursObserved int `json:"hoursObserved"`
	// Ready is true once at least one identity has enough evidence to be classified.
	Ready bool `json:"ready"`
	// Note is a sentence the viewer can show verbatim, so the explanation of the waiting state
	// lives with the data that determines it.
	Note string `json:"note,omitempty"`
}
