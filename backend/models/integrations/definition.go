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

// Automatic-refresh bounds are shared by provider validation and runtime admission. They are
// seconds, matching the policy grammar and the persisted schedule.
const (
	MinimumRefreshSeconds = 60
	MaximumRefreshSeconds = 30 * 24 * 60 * 60
)

// WasteCapability identifies the provider-independent waste schedule output contract.
const WasteCapability = "waste_schedule/v1"

// JoinValue combines a bounded list of saved scalar references, for APIs that require one
// coordinate pair rather than separate parameters. It is not a template language.
type JoinValue struct {
	Separator string   `json:"separator" yaml:"separator"`
	Refs      []string `json:"refs" yaml:"refs"`
}

// Value is either a JSON scalar, an explicit reference, or a bounded scalar join.
// It is deliberately not a template or expression language.
type Value struct {
	Ref     string
	Literal any
	Join    *JoinValue
}

// UnmarshalYAML rejects implicit or executable YAML forms before accepting the typed value.
func (v *Value) UnmarshalYAML(n *yaml.Node) error {
	*v = Value{}
	if n.Kind == yaml.MappingNode {
		if len(n.Content) == 2 && n.Content[0].Value == "ref" && n.Content[1].Tag == "!!str" {
			v.Ref = n.Content[1].Value
			return nil
		}
		if len(n.Content) == 2 && n.Content[0].Value == "join" {
			var join JoinValue
			if err := n.Content[1].Decode(&join); err != nil {
				return fmt.Errorf("invalid join value")
			}
			v.Join = &join
			return nil
		}
		return fmt.Errorf("expected a scalar, {ref: input.name/state.name} or {join: ...}")
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
	if v.Join != nil {
		return map[string]JoinValue{"join": *v.Join}, nil
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
	if v.Join != nil {
		return json.Marshal(map[string]JoinValue{"join": *v.Join})
	}
	return json.Marshal(v.Literal)
}

// UnmarshalJSON rejects unsupported JSON shapes and preserves numeric/reference types.
func (v *Value) UnmarshalJSON(data []byte) error {
	*v = Value{}
	data = bytes.TrimSpace(data)
	if len(data) > 0 && data[0] == '{' {
		var union struct {
			Ref  *string    `json:"ref"`
			Join *JoinValue `json:"join"`
		}
		d := json.NewDecoder(bytes.NewReader(data))
		d.DisallowUnknownFields()
		if err := d.Decode(&union); err != nil {
			return err
		}
		if union.Ref != nil && union.Join != nil {
			return fmt.Errorf("choose one value form")
		}
		if union.Ref != nil {
			if *union.Ref == "" {
				return fmt.Errorf("reference is empty")
			}
			v.Ref = *union.Ref
			return nil
		}
		if union.Join != nil {
			v.Join = union.Join
			return nil
		}
		return fmt.Errorf("reference is empty")
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
	// Mapping is heterogeneous only at the configuration boundary; the capability strictly decodes it.
	Mapping map[string]any `json:"mapping" yaml:"mapping"`
}

// AuthConfig declares how to use separately stored credential slots; it never holds a token.
// OAuth client credentials use ClientIDSlot/ClientSecretSlot and a reviewed TokenURL.
type AuthConfig struct {
	Mode              string `json:"mode" yaml:"mode"`
	SecretSlot        string `json:"secret_slot,omitempty" yaml:"secret_slot,omitempty"`
	TokenURL          string `json:"token_url,omitempty" yaml:"token_url,omitempty"`
	Scope             string `json:"scope,omitempty" yaml:"scope,omitempty"`
	ClientIDSlot      string `json:"client_id_slot,omitempty" yaml:"client_id_slot,omitempty"`
	ClientSecretSlot  string `json:"client_secret_slot,omitempty" yaml:"client_secret_slot,omitempty"`
	TokenStyle        string `json:"token_style,omitempty" yaml:"token_style,omitempty"`
	TokenResponseRoot string `json:"token_response_root,omitempty" yaml:"token_response_root,omitempty"`
}

// RequiresCredential reports whether provider requests must carry declared private slots.
// Public cross-origin setup searches are handled separately by request destination policy.
func (auth AuthConfig) RequiresCredential() bool {
	switch auth.Mode {
	case "bearer", "apikey_header", "apikey_query", "oauth2_client_credentials":
		return true
	default:
		return false
	}
}

// Slots lists the credential slots declared by this auth mode.
func (auth AuthConfig) Slots() []string {
	switch auth.Mode {
	case "bearer", "optional_bearer", "apikey_header", "apikey_query":
		if auth.SecretSlot == "" {
			return nil
		}
		return []string{auth.SecretSlot}
	case "oauth2_client_credentials":
		if auth.ClientIDSlot == "" || auth.ClientSecretSlot == "" {
			return nil
		}
		return []string{auth.ClientIDSlot, auth.ClientSecretSlot}
	default:
		return nil
	}
}

// AllowsSlot reports whether a named slot belongs to this auth mode.
func (auth AuthConfig) AllowsSlot(slot string) bool {
	for _, declared := range auth.Slots() {
		if declared == slot {
			return true
		}
	}
	return false
}

// Policy bounds explicit status checks, fresh lookup creation and safe submission-key replay.
// All durations are seconds; the runtime also enforces stricter local ceilings.
// RefreshDefaultSeconds seeds a new instance's automatic schedule (0 keeps it manual-only).
// RefreshMinSeconds is the editor/API floor; it is configuration, not a trusted provider limit.
type Policy struct {
	PollMinSeconds           int `json:"poll_min_seconds" yaml:"poll_min_seconds"`
	RenewalMinSeconds        int `json:"renewal_min_seconds" yaml:"renewal_min_seconds"`
	IdempotencyWindowSeconds int `json:"idempotency_window_seconds" yaml:"idempotency_window_seconds"`
	RefreshDefaultSeconds    int `json:"refresh_default_seconds" yaml:"refresh_default_seconds"`
	RefreshMinSeconds        int `json:"refresh_min_seconds" yaml:"refresh_min_seconds"`
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
	// Combine names at most one additional GET; results are published together under primary/additional.
	Combine      string             `json:"combine,omitempty" yaml:"combine,omitempty"`
	Method       string             `json:"method" yaml:"method"`
	Path         []Value            `json:"path" yaml:"path"`
	Query        map[string]Value   `json:"query,omitempty" yaml:"query,omitempty"`
	Body         map[string]Value   `json:"body,omitempty" yaml:"body,omitempty"`
	Headers      map[string]string  `json:"headers,omitempty" yaml:"headers,omitempty"`
	Discover     *DiscoverConfig    `json:"discover,omitempty" yaml:"discover,omitempty"`
	Pagination   *PaginationConfig  `json:"pagination,omitempty" yaml:"pagination,omitempty"`
	Idempotency  StrictBool         `json:"idempotency,omitempty" yaml:"idempotency,omitempty"`
	Conditional  StrictBool         `json:"conditional,omitempty" yaml:"conditional,omitempty"`
	ResponseRoot string             `json:"response_root,omitempty" yaml:"response_root,omitempty"`
	Captures     map[string]Capture `json:"captures,omitempty" yaml:"captures,omitempty"`
}

// DiscoverConfig declares one bounded Atom link lookup immediately before a request.
type DiscoverConfig struct {
	Format     string     `json:"format" yaml:"format"`
	Rel        string     `json:"rel" yaml:"rel"`
	SameOrigin StrictBool `json:"same_origin" yaml:"same_origin"`
}

// PaginationConfig declares bounded sequential page collection for array responses.
type PaginationConfig struct {
	PageQuery string `json:"page_query" yaml:"page_query"`
	Start     int    `json:"start" yaml:"start"`
	MaxPages  int    `json:"max_pages" yaml:"max_pages"`
	// PageSize is an expected provider page-size hint; collection stops on an empty page or MaxPages.
	PageSize          int `json:"page_size,omitempty" yaml:"page_size,omitempty"`
	DelayMilliseconds int `json:"delay_ms,omitempty" yaml:"delay_ms,omitempty"`
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

// ChoiceMapping locates a selection array and its ID/label fields; child paths are relative to each item.
type ChoiceMapping struct {
	Context   []string `json:"context,omitempty" yaml:"context,omitempty"`
	Latitude  string   `json:"latitude,omitempty" yaml:"latitude,omitempty"`
	Longitude string   `json:"longitude,omitempty" yaml:"longitude,omitempty"`
	Items     string   `json:"items" yaml:"items"`
	Value     string   `json:"value" yaml:"value"`
	Label     string   `json:"label" yaml:"label"`
}

// Discriminator requires an exact provider field value before offering the supported setup journey.
type Discriminator struct {
	Pointer string `json:"pointer" yaml:"pointer"`
	Value   string `json:"value" yaml:"value"`
}

// SearchStep links a user input to a request and defines how to display its returned choices.
type SearchStep struct {
	// BaseURL optionally declares a public search service; cross-origin searches never receive credentials.
	Postcode *PostcodeSearch `json:"postcode,omitempty" yaml:"postcode,omitempty"`
	BaseURL  string          `json:"base_url,omitempty" yaml:"base_url,omitempty"`
	Input    string          `json:"input" yaml:"input"`
	Request  string          `json:"request" yaml:"request"`
	Expect   *Discriminator  `json:"expect,omitempty" yaml:"expect,omitempty"`
	Options  ChoiceMapping   `json:"options" yaml:"options"`
}

// PostcodeSearch declares an alternate public UK postcode lookup.
type PostcodeSearch struct {
	BaseURL string        `json:"base_url" yaml:"base_url"`
	Request string        `json:"request" yaml:"request"`
	Options ChoiceMapping `json:"options" yaml:"options"`
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

// RefreshRecipe reads a saved resource and optionally names the lookup used for bounded renewal.
type RefreshRecipe struct {
	Request        string `json:"request" yaml:"request"`
	RenewalRequest string `json:"renewal_request,omitempty" yaml:"renewal_request,omitempty"`
}

// InvalidDefinition returns a safe field-level error shared by syntax, flow and capability validators.
func InvalidDefinition(path, message string) *Failure {
	return &Failure{Code: "invalid_definition", Message: "Check the provider configuration.", Fields: []FieldError{{Path: path, Message: message}}}
}
