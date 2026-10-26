package integrations

import (
	"encoding/json"
	"fmt"
)

// Resource preserves a capability's versioned JSON without teaching persistence its domain.
// Header fields are read-only indexes into the immutable payload. MarshalJSON retains the old
// flat waste resource shape, so existing encrypted records and browser contracts need no rewrite.
// A capability codec must validate the domain fields before publication or consumption.
type Resource struct {
	ContractVersion string `json:"contract_version"`
	ResourceID      string `json:"resource_id"`
	Revision        uint64 `json:"revision"`
	payload         json.RawMessage
}

// NewResource takes an owned JSON copy of a capability value and validates its common header.
// It does not replace capability validation; malformed or unversioned values return an error.
func NewResource(value any) (Resource, error) {
	var resource Resource
	data, err := json.Marshal(value)
	if err != nil {
		return resource, fmt.Errorf("encode integration resource: %w", err)
	}
	err = resource.UnmarshalJSON(data)
	return resource, err
}

// UnmarshalJSON accepts both existing flat snapshots and new capabilities with the same header.
// It copies the bytes because decoder buffers may be reused after this call returns.
func (r *Resource) UnmarshalJSON(data []byte) error {
	type header Resource
	var decoded header
	if len(data) > 2<<20 {
		return fmt.Errorf("integration resource exceeds 2 MiB")
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return fmt.Errorf("decode integration resource header: %w", err)
	}
	if decoded.ContractVersion == "" || decoded.ResourceID == "" || decoded.Revision == 0 {
		return fmt.Errorf("integration resource has an incomplete versioned header")
	}
	*r = Resource{ContractVersion: decoded.ContractVersion, ResourceID: decoded.ResourceID, Revision: decoded.Revision, payload: append(json.RawMessage(nil), data...)}
	return nil
}

// MarshalJSON returns an independent copy of the original flat payload, without a new wrapper.
// The zero value is invalid; callers cannot accidentally publish it as an empty resource.
func (r Resource) MarshalJSON() ([]byte, error) {
	if len(r.payload) == 0 {
		return nil, fmt.Errorf("integration resource has no payload")
	}
	return append([]byte(nil), r.payload...), nil
}

// Decode restores the concrete type owned by the capability; target must be a non-nil pointer.
// Codec implementations are responsible for semantic validation after decoding.
func (r Resource) Decode(target any) error {
	if len(r.payload) == 0 {
		return fmt.Errorf("integration resource has no payload")
	}
	return json.Unmarshal(r.payload, target)
}
