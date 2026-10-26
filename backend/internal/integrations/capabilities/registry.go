package capabilities

import (
	"fmt"
	"sort"
	"time"

	model "node-herder/models/integrations"
)

// Descriptor identifies one installed output contract and its discovery category.
type Descriptor struct {
	Capability, IntegrationType, Name, Description string
}

// Context carries publication evidence and tentative identity allocation, never credentials.
// Previous is nil for previews/new bindings. Identity maps are read-only inputs; results must
// return their own maps. NewIdentity reserves nothing until runtime commits publication.
type Context struct {
	NewIdentity                                         func(kind string) (string, error)
	InstanceID, ResourceID, ProviderID, ProviderVersion string
	ProfileRevision, BindingRevision                    uint64
	SelectedValue                                       string
	Now                                                 time.Time
	Previous                                            *model.Resource
	Identities                                          map[string]string
}

// Identity is a capability-owned semantic identifier that runtime reserves atomically.
type Identity struct{ ID, Kind string }

// Result contains a validated consumer-safe resource and identity mapping for one atomic
// publication. Resource must exclude credentials, private setup inputs and raw provider envelopes.
// Renew requests a bounded renewal; it never schedules work or performs provider I/O itself.
type Result struct {
	Resource     model.Resource
	Identities   map[string]string
	Reservations []Identity
	Renew        bool
}

// Handler owns domain validation/projection. Implementations are immutable and safe for
// concurrent calls; they must not write storage, start goroutines or perform network I/O.
type Handler interface {
	// Descriptor returns stable version/category metadata for discovery and validation.
	Descriptor() Descriptor
	// ValidateMapping strictly decodes domain selectors before a definition can activate.
	ValidateMapping(map[string]any) error
	// Normalize creates owned consumer-safe output without mutating the context's prior state.
	Normalize(any, map[string]any, Context) (Result, error)
	// Project derives current read-time values without changing persisted content identity.
	Project(model.Resource, time.Time) (model.Resource, error)
	// Checked records successful HTTP 304 observation without inventing new provider evidence.
	Checked(model.Resource, time.Time) (model.Resource, error)
}

// RenewalPolicy is optional domain recovery after failed normalization. The runtime still
// enforces refresh-only admission, one renewal, last-good data and provider cooldowns.
type RenewalPolicy interface{ RenewAfter(error) bool }

// Registry is an immutable capability lookup. Registration happens once at composition time;
// workers only read it. Duplicate/invalid registrations fail rather than replace an adapter.
type Registry struct{ handlers map[string]Handler }

// NewRegistry copies registrations into a privately owned lookup for concurrent readers.
func NewRegistry(handlers ...Handler) (*Registry, error) {
	r := &Registry{handlers: make(map[string]Handler, len(handlers))}
	for _, handler := range handlers {
		if handler == nil {
			return nil, fmt.Errorf("nil integration capability")
		}
		d := handler.Descriptor()
		if d.Capability == "" || d.IntegrationType == "" || d.Name == "" {
			return nil, fmt.Errorf("incomplete integration capability descriptor")
		}
		if _, exists := r.handlers[d.Capability]; exists {
			return nil, fmt.Errorf("duplicate integration capability %q", d.Capability)
		}
		r.handlers[d.Capability] = handler
	}
	return r, nil
}

// Resolve fails closed for missing/removed capability versions, including persisted records.
func (r *Registry) Resolve(id string) (Handler, error) {
	if r != nil {
		if handler, ok := r.handlers[id]; ok {
			return handler, nil
		}
	}
	return nil, model.Fail("unsupported_capability", "This integration's data type is not installed.")
}

// Descriptors returns stable catalogue order without exposing the registry's internal map.
func (r *Registry) Descriptors() []Descriptor {
	out := []Descriptor{}
	if r == nil {
		return out
	}
	for _, handler := range r.handlers {
		out = append(out, handler.Descriptor())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Capability < out[j].Capability })
	return out
}

// ValidateDefinition checks capability/category agreement and delegates mapping validation.
func (r *Registry) ValidateDefinition(d model.Definition) error {
	h, err := r.Resolve(d.Capability)
	if err != nil {
		return err
	}
	if h.Descriptor().IntegrationType != d.IntegrationType {
		return model.Fail("invalid_definition", "The integration type does not match its capability.")
	}
	return h.ValidateMapping(d.Mapping)
}

// Normalize dispatches the same validated result boundary for live execution and previews.
// A codec cannot publish a different capability version than the admitted definition.
func (r *Registry) Normalize(d model.Definition, root any, context Context) (Result, error) {
	handler, err := r.Resolve(d.Capability)
	if err != nil {
		return Result{}, err
	}
	result, err := handler.Normalize(root, d.Mapping, context)
	if err != nil {
		return Result{}, err
	}
	if result.Resource.ContractVersion != d.Capability {
		return Result{}, model.Fail("invalid_resource", "The capability returned a different data contract.")
	}
	return result, nil
}
