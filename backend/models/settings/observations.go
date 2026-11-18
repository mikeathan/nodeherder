package settings

// This file holds operator decisions about identities that wireless observation sources
// reported. Decisions are user intent, so they persist; raw observations never do.

// ObservationDecision is the operator's triage state for one observed identity. It is a
// closed set: an unknown value is rejected rather than silently treated as Unknown.
type ObservationDecision string

const (
	// DecisionUnknown is the default: reported but not yet triaged.
	DecisionUnknown ObservationDecision = "unknown"
	// DecisionKnown marks an identity the operator recognises; it stays visible.
	DecisionKnown ObservationDecision = "known"
	// DecisionIgnored hides an identity from broadcasts and from the viewer.
	DecisionIgnored ObservationDecision = "ignored"
)

// ValidObservationDecision reports whether value is a supported decision.
func ValidObservationDecision(value ObservationDecision) bool {
	switch value {
	case DecisionUnknown, DecisionKnown, DecisionIgnored:
		return true
	}
	return false
}

// ObservationIdentity is one persisted triage decision. Key is the opaque identity key
// produced by the observation source and must be treated as immutable, because changing a
// key derivation would orphan every decision.
type ObservationIdentity struct {
	// Decision is the operator's choice. Empty is normalised to DecisionUnknown on read.
	Decision ObservationDecision `json:"decision"`
	// Label is an operator-supplied display name; empty keeps the reported name.
	Label string `json:"label,omitempty"`
}
