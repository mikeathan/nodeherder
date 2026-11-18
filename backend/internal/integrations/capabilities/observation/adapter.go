// Package observation publishes the hub's wireless observation census as an integration resource.
//
// It is the only local capability in this tree: every other one normalises an HTTP response, while
// this one is handed a value the hub already produced. It still satisfies the same contract, which
// is what lets the catalog, the setup flow, the board, the enable toggle and the remove action all
// work without knowing that difference.
package observation

import (
	"time"

	"node-herder/internal/integrations/capabilities"
	model "node-herder/models/integrations"
	observationsmodel "node-herder/models/observations"
)

// ContractVersion identifies the payload shape this capability owns.
const ContractVersion = "observation_census/v1"

// IntegrationType is the catalog category this capability belongs to.
const IntegrationType = "wireless_observation"

// Source is one observing device, as published to consumers.
//
// It is deliberately narrow: the viewer needs to say what is reporting and whether it is healthy.
// Anything a consumer would have to join against live state for is left out, because a stored
// resource that disagrees with the live view is worse than one that omits the field.
type Source struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Advertisements and Decoded are lifetime counters for this process, and reset on restart. They
	// exist to distinguish "connected but silent" from "not configured".
	Advertisements uint64 `json:"advertisements"`
	Decoded        uint64 `json:"decoded"`
	Dropped        uint64 `json:"dropped"`
}

// Census is the published value: what the hub is currently hearing, and through what.
type Census struct {
	ContractVersion string `json:"contract_version"`
	ResourceID      string `json:"resource_id"`
	Revision        uint64 `json:"revision"`
	// Identities is how many devices are being tracked, which is the number the operator acts on.
	Identities int `json:"identities"`
	// Hidden is how many the class filter is demoting, so the summary can explain the difference
	// between "few devices" and "many devices, most hidden".
	Hidden  int      `json:"hidden"`
	Sources []Source `json:"sources"`
	// ObservedAt is when this census was taken. A consumer must treat it as a snapshot, never as a
	// live value: the radar is the live view.
	ObservedAt time.Time `json:"observed_at"`
}

// Adapter implements the local observation capability.
type Adapter struct{}

// Descriptor publishes the capability through the shared integration catalogue.
func (Adapter) Descriptor() capabilities.Descriptor {
	return capabilities.Descriptor{
		Capability:      ContractVersion,
		IntegrationType: IntegrationType,
		Name:            "Wireless observation",
		Description:     "Read nearby Bluetooth devices through an enrolled sensor. Shows distance and movement, and classifies what it hears.",
	}
}

// ValidateMapping accepts only an empty mapping.
//
// A local provider has nothing to select, and accepting arbitrary selectors would create the
// impression there is something to configure. Refusing them makes a mistaken configuration fail
// at activation rather than sit there doing nothing.
func (Adapter) ValidateMapping(raw map[string]any) error {
	if len(raw) != 0 {
		return model.Fail("invalid_mapping", "Wireless observation has nothing to configure.")
	}
	return nil
}

// Normalize converts the hub's snapshot into the published census.
//
// It receives the snapshot through the generic `any` boundary because that is the shared contract.
// A wrong type is an error rather than an empty census: publishing zero devices because the input
// was the wrong shape would look exactly like a silent sensor.
func (Adapter) Normalize(root any, raw map[string]any, context capabilities.Context) (capabilities.Result, error) {
	if err := (Adapter{}).ValidateMapping(raw); err != nil {
		return capabilities.Result{}, err
	}
	snapshot, ok := root.(observationsmodel.Snapshot)
	if !ok {
		return capabilities.Result{}, model.Fail("invalid_input", "Wireless observation needs an observation snapshot.")
	}

	sources := make([]Source, 0, len(snapshot.Sources))
	for _, source := range snapshot.Sources {
		sources = append(sources, Source{
			ID:             source.ID,
			Name:           source.Name,
			Advertisements: source.Diagnostics.Advertisements,
			Decoded:        source.Diagnostics.Decoded,
			Dropped:        source.Diagnostics.Dropped,
		})
	}

	hidden := 0
	for _, identity := range snapshot.Identities {
		if identity.Classification != nil && !identity.Classification.ShowByDefault {
			hidden++
		}
	}

	census := Census{
		ContractVersion: ContractVersion,
		ResourceID:      context.ResourceID,
		Revision:        context.ProfileRevision,
		Identities:      len(snapshot.Identities),
		Hidden:          hidden,
		Sources:         sources,
		ObservedAt:      snapshot.At,
	}
	resource, err := model.NewResource(census)
	if err != nil {
		return capabilities.Result{}, err
	}
	return capabilities.Result{
		Resource:     resource,
		Identities:   context.Identities,
		Reservations: []capabilities.Identity{{ID: context.ResourceID, Kind: "resource"}},
	}, nil
}

// Project re-derives the census without changing saved content identity.
//
// A census ages: its counters and identity list describe a moment, so a projected one is the stored
// value plus the fact that it is no longer current. Presenting stale numbers as live would be the
// same failure as a stale distance.
func (Adapter) Project(resource model.Resource, _ time.Time) (model.Resource, error) {
	var census Census
	if err := resource.Decode(&census); err != nil {
		return model.Resource{}, err
	}
	if census.ContractVersion != ContractVersion {
		return model.Resource{}, model.Fail("invalid_resource", "Unexpected observation payload.")
	}
	return model.NewResource(census)
}

// Checked records a successful observation refresh without inventing new evidence.
func (Adapter) Checked(resource model.Resource, now time.Time) (model.Resource, error) {
	return (Adapter{}).Project(resource, now)
}

// Definition builds the local provider definition.
//
// Built in Go rather than parsed from YAML because it has no base URL, no requests and no inputs —
// the grammar those fields belong to is for providers that fetch. `Local` is what tells the runtime
// to publish from the hub rather than perform an operation, and what the compiler refuses to accept
// from an authored document.
func Definition() model.Definition {
	return model.Definition{
		Local:           true,
		SchemaVersion:   model.FormatVersion,
		IntegrationType: IntegrationType,
		ID:              "hub",
		Name:            "Enrolled sensors",
		Capability:      ContractVersion,
		Mapping:         map[string]any{},
	}
}

// DefinitionKey is the stable host-file identifier for this provider, matching the shape the
// runtime derives for authored definitions.
func DefinitionKey() string { return IntegrationType + "/hub" }
