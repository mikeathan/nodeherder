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
	FirstSeen time.Time       `json:"firstSeen"`
	LastSeen  time.Time       `json:"lastSeen"`
	Packets   uint64          `json:"packets"`
	RSSI      *int            `json:"rssi,omitempty"`
	Sources   []SourceReading `json:"sources"`
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
}
