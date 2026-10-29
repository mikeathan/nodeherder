package definitions

// Parses provider documents and validates their requests, setup journeys and capabilities.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"

	"node-herder/internal/integrations/capabilities"
	"node-herder/internal/integrations/destination"
	"node-herder/internal/integrations/flows/searchselect"
	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
)

// MaxDefinitionBytes bounds provider source parsing and host definition files to 64 KiB.
const MaxDefinitionBytes = 64 << 10

var identifier = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

// Compiled is a validated, immutable provider revision.
type Compiled struct {
	Definition         model.Definition
	YAML, Hash, Origin string
	Revision           uint64
}

// ContentHash fingerprints exact UTF-8 host bytes, so even formatting changes cause edit conflicts.
func ContentHash(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// DefinitionKey creates the stable type/provider host-file identifier from a validated definition.
func DefinitionKey(d model.Definition) string { return d.IntegrationType + "/" + d.ID }

// StrongETag quotes the exact source hash for HTTP conditional host-file mutations.
func StrongETag(hash string) string { return `"sha256:` + hash + `"` }
func configError(path, message string) error {
	return model.InvalidDefinition(path, message)
}

// CompileYAML accepts one bounded document, without aliases, merge keys or unknown fields.
func (c *Compiler) CompileYAML(source string) (*Compiled, error) {
	if len(source) == 0 || len(source) > MaxDefinitionBytes || !utf8.ValidString(source) {
		return nil, configError("", "Use a UTF-8 document of at most 64 KiB.")
	}
	dec := yaml.NewDecoder(strings.NewReader(source))
	var root yaml.Node
	if err := dec.Decode(&root); err != nil {
		return nil, configError("", "Invalid YAML: "+err.Error())
	}
	var extra yaml.Node
	if err := dec.Decode(&extra); err != io.EOF {
		return nil, configError("", "Use exactly one YAML document.")
	}
	count := 0
	if err := checkYAML(&root, 0, &count); err != nil {
		return nil, err
	}
	var d model.Definition
	typed := yaml.NewDecoder(strings.NewReader(source))
	typed.KnownFields(true)
	if err := typed.Decode(&d); err != nil {
		return nil, configError("", err.Error())
	}
	// yaml.v3 deliberately treats null as a zero value before custom scalar decoding. Decode
	// the node's JSON representation with the same strict types used by the Form endpoint too.
	var raw any
	if err := root.Decode(&raw); err != nil {
		return nil, configError("", "Invalid YAML values.")
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil, configError("", "Use finite JSON-compatible scalar values.")
	}
	strict := json.NewDecoder(bytes.NewReader(encoded))
	strict.DisallowUnknownFields()
	if err := strict.Decode(&d); err != nil {
		return nil, configError("", err.Error())
	}
	if err := c.validateDefinition(&d); err != nil {
		return nil, err
	}
	u, _ := destination.Parse(d.BaseURL) // validateDefinition has already checked the URL.
	return &Compiled{Definition: d, YAML: source, Hash: ContentHash([]byte(source)), Origin: destination.Origin(u), Revision: 1}, nil
}

// checkYAML bounds document structure and rejects aliases, duplicate keys and unsupported scalars.
func checkYAML(n *yaml.Node, depth int, count *int) error {
	*count++
	if depth > 32 || *count > 8192 {
		return configError("", "The document is too deeply nested or large.")
	}
	if n.Kind == yaml.AliasNode || n.Anchor != "" {
		return configError("", "YAML anchors and aliases are unsupported.")
	}
	if n.Kind == yaml.ScalarNode {
		switch n.Tag {
		case "!!str", "!!null", "!!bool", "!!int", "!!float":
		default:
			return configError("", "Unsupported YAML value type.")
		}
	}
	if n.Kind == yaml.MappingNode {
		seen := map[string]bool{}
		for i := 0; i < len(n.Content); i += 2 {
			k := n.Content[i]
			if k.Tag != "!!str" || k.Value == "<<" || seen[k.Value] {
				return configError(k.Value, "Use unique text keys; merge keys are unsupported.")
			}
			seen[k.Value] = true
		}
	}
	for _, c := range n.Content {
		if err := checkYAML(c, depth+1, count); err != nil {
			return err
		}
	}
	return nil
}

// CompileDocument gives the form editor exactly the same validation as the YAML editor.
func (c *Compiler) CompileDocument(raw json.RawMessage) (*Compiled, error) {
	if len(raw) > MaxDefinitionBytes {
		return nil, configError("", "Document too large.")
	}
	var d model.Definition
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&d); err != nil {
		return nil, configError("", err.Error())
	}
	var extra any
	if dec.Decode(&extra) != io.EOF {
		return nil, configError("", "Use one document.")
	}
	b, err := yaml.Marshal(d)
	if err != nil {
		return nil, configError("", err.Error())
	}
	return c.CompileYAML(string(b))
}

func validPointer(p string) bool { return operations.ValidPointer(p) }
func requirePointer(path, p string) error {
	if p == "" || !validPointer(p) {
		return configError(path, "Use a JSON Pointer, for example /collections.")
	}
	return nil
}

// validateDefinition runs independent checks in dependency order. Earlier checks establish the
// invariants later checks rely on; this is also the single Form/YAML validation boundary.
func (c *Compiler) validateDefinition(d *model.Definition) error {
	if err := validateProviderMetadata(d); err != nil {
		return err
	}
	if err := c.capabilities.ValidateDefinition(*d); err != nil {
		return err
	}
	if err := validateSetupInputs(d); err != nil {
		return err
	}
	if err := searchselect.ValidateLinks(d); err != nil {
		return err
	}
	if err := validateRequestDeclarations(d); err != nil {
		return err
	}
	if err := searchselect.ValidateAvailability(d); err != nil {
		return err
	}
	if err := validateResultAndErrorMapping(d); err != nil {
		return err
	}
	return nil
}

// validateProviderMetadata validates shareable identity, destination, authentication and bounded policy defaults.
func validateProviderMetadata(d *model.Definition) error {
	if d.SchemaVersion != model.FormatVersion {
		return configError("schema_version", "Unsupported definition version.")
	}
	if !identifier.MatchString(d.IntegrationType) {
		return configError("integration_type", "Use a lowercase integration type.")
	}
	if !identifier.MatchString(d.ID) || strings.TrimSpace(d.Name) == "" || len(d.Name) > 120 {
		return configError("id", "Use a lowercase identifier and a provider name.")
	}
	if _, err := destination.Parse(d.BaseURL); err != nil {
		return configError("base_url", "Use an HTTPS address on port 443 without credentials, query or fragment.")
	}
	if d.Auth.Mode == "" {
		d.Auth.Mode = "none"
	}
	switch d.Auth.Mode {
	case "none":
	case "bearer", "optional_bearer", "apikey_header", "apikey_query":
		if !identifier.MatchString(d.Auth.SecretSlot) {
			return configError("auth.secret_slot", "Choose a named credential slot.")
		}
	default:
		return configError("auth.mode", "Supported modes are none, bearer, optional_bearer, apikey_header and apikey_query.")
	}
	if d.Policy.PollMinSeconds == 0 {
		d.Policy.PollMinSeconds = 5
	}
	if d.Policy.RenewalMinSeconds == 0 {
		d.Policy.RenewalMinSeconds = 86400
	}
	if d.Policy.PollMinSeconds < 5 || d.Policy.PollMinSeconds > 86400 || d.Policy.RenewalMinSeconds < 86400 || d.Policy.RenewalMinSeconds > 2592000 || d.Policy.IdempotencyWindowSeconds < 0 || d.Policy.IdempotencyWindowSeconds > 2592000 {
		return configError("policy", "Polling must be at least 5 seconds; renewal at least 24 hours; windows at most 30 days.")
	}
	return nil
}

// validateSetupInputs checks the finite user-input vocabulary before any recipe references it.
func validateSetupInputs(d *model.Definition) error {
	if len(d.Inputs) < 2 || len(d.Inputs) > 16 || len(d.Requests) == 0 || len(d.Requests) > 8 {
		return configError("requests", "Use 2–16 inputs and 1–8 requests.")
	}
	for id, v := range d.Inputs {
		if !identifier.MatchString(id) || strings.TrimSpace(v.Label) == "" || len(v.Label) > 120 {
			return configError("inputs."+id, "Use a named input and label.")
		}
		switch v.Type {
		case "postcode", "text", "select":
		default:
			return configError("inputs."+id, "Unsupported input type.")
		}
	}
	return nil
}

// validateRecipeLinks checks recipe links, mandatory JSON locations and disjoint provider statuses.
// validateRequestDeclarations checks each named HTTP operation and cross-request capture type consistency.
func validateRequestDeclarations(d *model.Definition) error {
	captureTypes := map[string]string{}
	count := 0
	for id, q := range d.Requests {
		path := "requests." + id
		if err := operations.ValidateCombination(*d, id); err != nil {
			return err
		}
		if !identifier.MatchString(id) || (q.Method != "GET" && q.Method != "POST") || len(q.Path) == 0 || len(q.Path) > 32 || len(q.Query) > 32 || len(q.Body) > 32 || (q.Method == "GET" && len(q.Body) > 0) || (bool(q.Idempotency) && q.Method != "POST") || (bool(q.Conditional) && q.Method != "GET") {
			return configError(path, "Use a named GET or POST request with bounded path, query and body.")
		}
		if !validPointer(q.ResponseRoot) {
			return configError(path+".response_root", "Invalid JSON Pointer.")
		}
		for name, c := range q.Captures {
			count++
			if count > 64 || !identifier.MatchString(name) || name == "etag" {
				return configError(path+".captures", "Invalid or reserved saved value name.")
			}
			if err := requirePointer(path+".captures."+name, c.Pointer); err != nil {
				return err
			}
			if c.When != "" && c.When != "complete" {
				return configError(path+".captures."+name, "Use when: complete or omit when.")
			}
			switch c.Type {
			case "string", "boolean", "number", "local-date", "utc-timestamp":
			default:
				return configError(path+".captures."+name, "Unsupported saved value type.")
			}
			if c.StripPrefix != "" && c.Type != "string" {
				return configError(path+".captures."+name, "strip_prefix requires a string.")
			}
			if old, ok := captureTypes[name]; ok && old != c.Type {
				return configError(path+".captures."+name, "Saved value types must agree.")
			}
			captureTypes[name] = c.Type
		}
	}
	return nil
}

// validateRequestAvailability proves every request value exists on all execution paths, including immediate and polled completion.
// validateResultAndErrorMapping checks output selectors and finite recovery rules without enabling arbitrary retry behavior.
func validateResultAndErrorMapping(d *model.Definition) error {
	if len(d.Errors.Rules) > 32 || !validPointer(d.Errors.Code) || len(d.Errors.Rules) > 0 && d.Errors.Code == "" {
		return configError("errors", "Invalid problem selector or too many error rules.")
	}
	seenRules := map[string]bool{}
	for _, r := range d.Errors.Rules {
		ruleKey := strconv.Itoa(r.Status) + ":" + r.Code
		if seenRules[ruleKey] {
			return configError("errors.rules", "Use one recovery action for each HTTP status/problem code.")
		}
		seenRules[ruleKey] = true
		if r.Status < 400 || r.Status > 599 || r.Code == "" || len(r.Code) > 128 {
			return configError("errors.rules", "Match an HTTP error and an exact problem code.")
		}
		switch r.Action {
		case "renewal":
			if r.Status >= 500 || r.Status == 401 || r.Status == 403 || r.Status == 429 {
				return configError("errors.rules", "Renewal cannot override authentication, rate-limit or server failures.")
			}
			if d.Refresh.RenewalRequest == "" {
				return configError("errors.rules", "Renewal needs a renewal request.")
			}
		case "restart_setup", "reconnect", "unsupported":
		default:
			return configError("errors.rules", "Unsupported recovery action.")
		}
	}
	return nil
}

// Compiler validates bounded syntax and delegates domain mappings to registered capabilities.
// Its registry is immutable after construction and shared safely with worker goroutines.
type Compiler struct{ capabilities *capabilities.Registry }

// NewCompiler binds capability validation without importing any concrete capability.
func NewCompiler(registry *capabilities.Registry) *Compiler { return &Compiler{capabilities: registry} }
