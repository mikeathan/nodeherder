// Package integrations defines provider-independent connection configuration and resources.
package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// FormatVersion identifies the supported shareable provider-definition grammar.
const FormatVersion = "nodeherder.integration/v1"

// WasteCapability identifies the provider-independent waste schedule output contract.
const WasteCapability = "waste_schedule/v1"

// Value is either a JSON scalar or an explicit reference to a setup input/captured state.
// It is deliberately not a template or expression language.
type Value struct {
	Ref     string
	Literal any
}

// UnmarshalYAML rejects implicit or executable YAML forms before accepting the typed value.
func (v *Value) UnmarshalYAML(n *yaml.Node) error {
	*v = Value{}
	if n.Kind == yaml.MappingNode {
		if len(n.Content) != 2 || n.Content[0].Value != "ref" || n.Content[1].Tag != "!!str" {
			return fmt.Errorf("expected a scalar or {ref: input.name/state.name}")
		}
		v.Ref = n.Content[1].Value
		return nil
	}
	if n.Kind != yaml.ScalarNode {
		return fmt.Errorf("request values must be scalars or references")
	}
	switch n.Tag {
	case "!!str", "!!bool", "!!int", "!!float", "!!null":
	default:
		return fmt.Errorf("unsupported scalar type %s", n.Tag)
	}
	return n.Decode(&v.Literal)
}

// MarshalYAML preserves scalar types and explicit references when saving the host definition.
func (v Value) MarshalYAML() (any, error) {
	if v.Ref != "" {
		return map[string]string{"ref": v.Ref}, nil
	}
	if n, ok := v.Literal.(json.Number); ok {
		if i, err := n.Int64(); err == nil {
			return i, nil
		}
		return n.Float64()
	}
	return v.Literal, nil
}

// MarshalJSON encodes the scalar/reference union without exposing implementation fields.
func (v Value) MarshalJSON() ([]byte, error) {
	if v.Ref != "" {
		return json.Marshal(map[string]string{"ref": v.Ref})
	}
	return json.Marshal(v.Literal)
}

// UnmarshalJSON rejects unsupported JSON shapes and preserves numeric/reference types.
func (v *Value) UnmarshalJSON(data []byte) error {
	*v = Value{}
	data = bytes.TrimSpace(data)
	if len(data) > 0 && data[0] == '{' {
		var ref struct {
			Ref string `json:"ref"`
		}
		d := json.NewDecoder(bytes.NewReader(data))
		d.DisallowUnknownFields()
		if err := d.Decode(&ref); err != nil {
			return err
		}
		if ref.Ref == "" {
			return fmt.Errorf("reference is empty")
		}
		v.Ref = ref.Ref
		return nil
	}
	var x any
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	if err := d.Decode(&x); err != nil {
		return err
	}
	switch x.(type) {
	case nil, string, bool, json.Number:
		v.Literal = x
		return nil
	}
	return fmt.Errorf("expected scalar or reference")
}

// StrictBool keeps YAML and JSON form validation aligned; legacy yes/no coercion is unsupported.
type StrictBool bool

// UnmarshalYAML accepts only explicit true/false, keeping Form and YAML validation aligned.
func (b *StrictBool) UnmarshalYAML(n *yaml.Node) error {
	if n.Tag != "!!bool" || (n.Value != "true" && n.Value != "false") {
		return fmt.Errorf("use true or false")
	}
	*b = n.Value == "true"
	return nil
}

// UnmarshalJSON accepts only explicit true/false, keeping Form and YAML validation aligned.
func (b *StrictBool) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if !bytes.Equal(data, []byte("true")) && !bytes.Equal(data, []byte("false")) {
		return fmt.Errorf("use true or false")
	}
	*b = StrictBool(bytes.Equal(data, []byte("true")))
	return nil
}

// Definition is the complete finite v1 provider grammar; it contains no private instance settings.
type Definition struct {
	SchemaVersion   string             `json:"schema_version" yaml:"schema_version"`
	IntegrationType string             `json:"integration_type" yaml:"integration_type"`
	ID              string             `json:"id" yaml:"id"`
	Name            string             `json:"name" yaml:"name"`
	Capability      string             `json:"capability" yaml:"capability"`
	BaseURL         string             `json:"base_url" yaml:"base_url"`
	Auth            AuthConfig         `json:"auth" yaml:"auth"`
	Policy          Policy             `json:"policy" yaml:"policy"`
	Errors          ErrorRules         `json:"errors" yaml:"errors"`
	Inputs          map[string]Input   `json:"inputs" yaml:"inputs"`
	Requests        map[string]Request `json:"requests" yaml:"requests"`
	Setup           SetupRecipe        `json:"setup" yaml:"setup"`
	Refresh         RefreshRecipe      `json:"refresh" yaml:"refresh"`
	Mapping         Mapping            `json:"mapping" yaml:"mapping"`
}

// AuthConfig declares how to use a separately stored credential slot; it never holds a token.
type AuthConfig struct {
	Mode       string `json:"mode" yaml:"mode"`
	SecretSlot string `json:"secret_slot,omitempty" yaml:"secret_slot,omitempty"`
}

// Policy bounds explicit status checks, fresh lookup creation and safe submission-key replay.
// All durations are seconds; the runtime also enforces stricter local ceilings.
type Policy struct {
	PollMinSeconds           int `json:"poll_min_seconds" yaml:"poll_min_seconds"`
	RenewalMinSeconds        int `json:"renewal_min_seconds" yaml:"renewal_min_seconds"`
	IdempotencyWindowSeconds int `json:"idempotency_window_seconds" yaml:"idempotency_window_seconds"`
}

// ErrorRules locates the provider problem code and lists exact status/code recovery matches.
type ErrorRules struct {
	Code  string      `json:"code" yaml:"code"`
	Rules []ErrorRule `json:"rules" yaml:"rules"`
}

// ErrorRule maps one exact provider failure to a supported recovery action, never an arbitrary retry.
type ErrorRule struct {
	Status int    `json:"status" yaml:"status"`
	Code   string `json:"code" yaml:"code"`
	Action string `json:"action" yaml:"action"`
}

// Input describes a setup question. Private answers are stored in Setup/Instance, not this definition.
type Input struct {
	Type     string     `json:"type" yaml:"type"`
	Label    string     `json:"label" yaml:"label"`
	Required StrictBool `json:"required" yaml:"required"`
}

// Request describes one named GET/POST operation. Path segments and fields are typed Values;
// ResponseRoot selects the result object, while Captures read the original response envelope.
type Request struct {
	Method       string             `json:"method" yaml:"method"`
	Path         []Value            `json:"path" yaml:"path"`
	Query        map[string]Value   `json:"query,omitempty" yaml:"query,omitempty"`
	Body         map[string]Value   `json:"body,omitempty" yaml:"body,omitempty"`
	Idempotency  StrictBool         `json:"idempotency,omitempty" yaml:"idempotency,omitempty"`
	Conditional  StrictBool         `json:"conditional,omitempty" yaml:"conditional,omitempty"`
	ResponseRoot string             `json:"response_root,omitempty" yaml:"response_root,omitempty"`
	Captures     map[string]Capture `json:"captures,omitempty" yaml:"captures,omitempty"`
}

// Capture saves a typed response field for a later request. Pointer is an RFC 6901 JSON
// location (for example /result/property_id), not a Go address. When=complete defers extraction
// until successful lookup completion; StripPrefix validates/removes a provider namespace.
type Capture struct {
	Pointer     string     `json:"pointer" yaml:"pointer"`
	Type        string     `json:"type" yaml:"type"`
	When        string     `json:"when,omitempty" yaml:"when,omitempty"`
	StripPrefix string     `json:"strip_prefix,omitempty" yaml:"strip_prefix,omitempty"`
	Optional    StrictBool `json:"optional,omitempty" yaml:"optional,omitempty"`
}

// ChoiceMapping locates an address array and its ID/label fields; child paths are relative to each item.
type ChoiceMapping struct {
	Items string `json:"items" yaml:"items"`
	Value string `json:"value" yaml:"value"`
	Label string `json:"label" yaml:"label"`
}

// Discriminator requires an exact provider field value before offering the supported setup journey.
type Discriminator struct {
	Pointer string `json:"pointer" yaml:"pointer"`
	Value   string `json:"value" yaml:"value"`
}

// SearchStep links a user input to a request and defines how to display its returned choices.
type SearchStep struct {
	Input   string         `json:"input" yaml:"input"`
	Request string         `json:"request" yaml:"request"`
	Expect  *Discriminator `json:"expect,omitempty" yaml:"expect,omitempty"`
	Options ChoiceMapping  `json:"options" yaml:"options"`
}

// SelectionStep links the explicit selected candidate to the first lookup submission.
type SelectionStep struct {
	Input   string `json:"input" yaml:"input"`
	Request string `json:"request" yaml:"request"`
}

// StatusStep classifies lookup progress using configured exact strings and a named status request.
type StatusStep struct {
	Request string   `json:"request" yaml:"request"`
	Pointer string   `json:"pointer" yaml:"pointer"`
	Pending []string `json:"pending" yaml:"pending"`
	Done    []string `json:"done" yaml:"done"`
	Failed  []string `json:"failed" yaml:"failed"`
}

// SetupRecipe wires the supported search/select/status stages to arbitrary request identifiers.
type SetupRecipe struct {
	Type   string        `json:"type" yaml:"type"`
	Search SearchStep    `json:"search" yaml:"search"`
	Select SelectionStep `json:"select" yaml:"select"`
	Status StatusStep    `json:"status" yaml:"status"`
}

// RefreshRecipe reads a saved schedule and optionally names the lookup used for bounded renewal.
type RefreshRecipe struct {
	Request        string `json:"request" yaml:"request"`
	RenewalRequest string `json:"renewal_request,omitempty" yaml:"renewal_request,omitempty"`
}

// Mapping converts provider JSON field locations into one typed waste schedule.
type Mapping struct {
	Metadata    MetadataMapping   `json:"metadata" yaml:"metadata"`
	Collections CollectionMapping `json:"collections" yaml:"collections"`
}

// MetadataMapping holds schedule-level JSON Pointers relative to the selected result root.
// These fields preserve provenance and quality; they do not manufacture missing evidence.
type MetadataMapping struct {
	PropertyID          string `json:"property_id" yaml:"property_id"`
	SourceRetrievedAt   string `json:"source_retrieved_at" yaml:"source_retrieved_at"`
	RefreshDueAt        string `json:"refresh_due_at" yaml:"refresh_due_at"`
	DateConfidence      string `json:"date_confidence" yaml:"date_confidence"`
	DateCompleteness    string `json:"date_completeness" yaml:"date_completeness"`
	AddressMatch        string `json:"address_match" yaml:"address_match"`
	EvidenceGranularity string `json:"evidence_granularity" yaml:"evidence_granularity"`
	Provisional         string `json:"provisional" yaml:"provisional"`
	Notes               string `json:"notes,omitempty" yaml:"notes,omitempty"`
	SourceURL           string `json:"source_url,omitempty" yaml:"source_url,omitempty"`
}

// CollectionMapping locates the collection list once; Fields applies to every returned service.
type CollectionMapping struct {
	Items  string           `json:"items" yaml:"items"`
	Fields CollectionFields `json:"fields" yaml:"fields"`
}

// CollectionFields names JSON locations relative to each collection, independently of bin colour.
type CollectionFields struct {
	Name                 string `json:"name" yaml:"name"`
	Dates                string `json:"dates" yaml:"dates"`
	Colour               string `json:"colour,omitempty" yaml:"colour,omitempty"`
	WasteType            string `json:"waste_type,omitempty" yaml:"waste_type,omitempty"`
	Container            string `json:"container,omitempty" yaml:"container,omitempty"`
	DatesComplete        string `json:"dates_complete,omitempty" yaml:"dates_complete,omitempty"`
	SubscriptionRequired string `json:"subscription_required,omitempty" yaml:"subscription_required,omitempty"`
	SourceServiceKey     string `json:"source_service_key,omitempty" yaml:"source_service_key,omitempty"`
}
