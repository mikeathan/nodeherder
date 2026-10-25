package integrations

import "time"

// Quality preserves evidence for an exact property match and the provider's date confidence.
// Unknown confidence remains unknown; approximate data cannot become an exact-household snapshot.
type Quality struct {
	AddressMatch          string  `json:"address_match"`
	EvidenceGranularity   string  `json:"evidence_granularity"`
	DateConfidence        string  `json:"date_confidence"`
	Provisional           bool    `json:"provisional"`
	SourceConfidenceLabel *string `json:"source_confidence_label"`
}

// Provenance records the definition and property-binding revision that produced a snapshot.
// ProviderVersion is the exact source hash, not an inferred remote API version.
type Provenance struct {
	ProviderID      string  `json:"provider_id"`
	ProviderVersion string  `json:"provider_version"`
	ProfileRevision uint64  `json:"profile_revision"`
	BindingRevision uint64  `json:"binding_revision"`
	CouncilGSS      *string `json:"council_gss"`
	SourceURL       *string `json:"source_url"`
}

// Coverage describes the provider's known calendar horizon. Nullable bounds distinguish
// unknown coverage from a confirmed range; they are local YYYY-MM-DD dates, not UTC instants.
type Coverage struct {
	State   string  `json:"state"`
	From    *string `json:"from"`
	Through *string `json:"through"`
}

// WasteService is one collection stream inside a schedule, with a stable local service ID.
// Dates are Europe/London calendar dates; nil metadata means unknown, not a default measurement.
type WasteService struct {
	ServiceID        string   `json:"service_id"`
	SourceServiceKey *string  `json:"source_service_key"`
	Name             string   `json:"name"`
	WasteType        string   `json:"waste_type"`
	SourceWasteType  *string  `json:"source_waste_type"`
	BinColour        *string  `json:"bin_colour"`
	Container        *string  `json:"container"`
	Subscription     string   `json:"subscription"`
	DateState        string   `json:"date_state"`
	Dates            []string `json:"dates"`
	Coverage         Coverage `json:"coverage"`
}

// WasteSchedule is the versioned, validated resource exposed to authenticated consumers.
// Content revision/source age are independent of LastCheckedAt and read-time EvaluatedAt.
// Nullable timestamps preserve unknown freshness instead of declaring an unchanged response fresh.
type WasteSchedule struct {
	ContractVersion       string         `json:"contract_version"`
	InstanceID            string         `json:"instance_id"`
	ResourceID            string         `json:"resource_id"`
	Revision              uint64         `json:"revision"`
	Kind                  string         `json:"kind"`
	Timezone              string         `json:"timezone"`
	CapturedAt            time.Time      `json:"captured_at"`
	LastCheckedAt         time.Time      `json:"last_checked_at"`
	SourceRetrievedAt     *time.Time     `json:"source_retrieved_at"`
	FreshUntil            *time.Time     `json:"fresh_until"`
	EvaluatedAt           time.Time      `json:"evaluated_at"`
	Freshness             string         `json:"freshness"`
	Quality               Quality        `json:"quality"`
	Provenance            Provenance     `json:"provenance"`
	RepresentativeConsent string         `json:"representative_consent"`
	Notices               []string       `json:"notices"`
	Services              []WasteService `json:"services"`
}

// Project derives date-dependent metadata without changing content identity or source age.
func (s WasteSchedule) Project(now time.Time) WasteSchedule {
	s.EvaluatedAt = now.UTC()
	s.Freshness = "unknown"
	if s.SourceRetrievedAt != nil && s.FreshUntil != nil {
		s.Freshness = "fresh"
		if !now.Before(*s.FreshUntil) {
			s.Freshness = "stale"
		}
	}
	return s
}
