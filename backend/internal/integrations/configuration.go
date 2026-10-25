// Package integrations owns the manual connection lifecycle and bounded provider execution.
package integrations

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"math"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
)

const MaxDefinitionBytes = 64 << 10

var identifier = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

// CompiledDefinition is a validated, immutable provider revision.
type CompiledDefinition struct {
	Definition         model.Definition
	YAML, Hash, Origin string
	Revision           uint64
}

func ContentHash(b []byte) string             { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func DefinitionKey(d model.Definition) string { return d.IntegrationType + "/" + d.ID }
func StrongETag(hash string) string           { return `"sha256:` + hash + `"` }
func configError(path, message string) error {
	return &model.Failure{Code: "invalid_definition", Message: "Check the provider configuration.", Fields: []model.FieldError{{Path: path, Message: message}}}
}

// CompileYAML accepts one bounded document, without aliases, merge keys or unknown fields.
func CompileYAML(source string) (*CompiledDefinition, error) {
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
	if err := validateDefinition(&d); err != nil {
		return nil, err
	}
	u, _ := url.Parse(d.BaseURL)
	return &CompiledDefinition{Definition: d, YAML: source, Hash: ContentHash([]byte(source)), Origin: Origin(u), Revision: 1}, nil
}
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
func CompileDocument(raw json.RawMessage) (*CompiledDefinition, error) {
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
	return CompileYAML(string(b))
}

// Origin canonicalizes the credential destination without including a path.
func Origin(u *url.URL) string {
	host := strings.ToLower(u.Hostname())
	port := u.Port()
	if port != "" && port != "443" {
		host = net.JoinHostPort(host, port)
	} else if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	return strings.ToLower(u.Scheme) + "://" + host
}
func validPointer(p string) bool {
	if p == "" {
		return true
	}
	if len(p) > 256 || p[0] != '/' {
		return false
	}
	for i := 0; i < len(p); i++ {
		if p[i] == '~' {
			i++
			if i == len(p) || (p[i] != '0' && p[i] != '1') {
				return false
			}
		}
	}
	return true
}
func requirePointer(path, p string) error {
	if p == "" || !validPointer(p) {
		return configError(path, "Use a JSON Pointer, for example /collections.")
	}
	return nil
}
func validateDefinition(d *model.Definition) error {
	if d.SchemaVersion != model.FormatVersion {
		return configError("schema_version", "Unsupported definition version.")
	}
	if d.IntegrationType != "waste_collection" || d.Capability != model.WasteCapability {
		return configError("capability", "This release supports waste_collection and waste_schedule/v1.")
	}
	if !identifier.MatchString(d.ID) || strings.TrimSpace(d.Name) == "" || len(d.Name) > 120 {
		return configError("id", "Use a lowercase identifier and a provider name.")
	}
	u, err := url.Parse(d.BaseURL)
	if err != nil || u.Scheme != "https" || (u.Hostname() == "" || u.Port() != "" && u.Port() != "443") || u.User != nil || u.RawQuery != "" || u.Fragment != "" || len(d.BaseURL) > 2048 {
		return configError("base_url", "Use an HTTPS address on port 443 without credentials, query or fragment.")
	}
	if d.Auth.Mode == "" {
		d.Auth.Mode = "none"
	}
	switch d.Auth.Mode {
	case "none":
	case "bearer", "optional_bearer":
		if !identifier.MatchString(d.Auth.SecretSlot) {
			return configError("auth.secret_slot", "Choose a named credential slot.")
		}
	default:
		return configError("auth.mode", "Supported modes are none, bearer and optional_bearer.")
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
	s := d.Setup
	if s.Type != "search_select" || s.Search.Input == s.Select.Input {
		return configError("setup", "Use a search_select recipe with two different inputs.")
	}
	if _, ok := d.Inputs[s.Search.Input]; !ok {
		return configError("setup.search.input", "Unknown search input.")
	}
	if _, ok := d.Inputs[s.Select.Input]; !ok {
		return configError("setup.select.input", "Unknown selection input.")
	}
	for _, id := range []string{s.Search.Request, s.Select.Request, s.Status.Request, d.Refresh.Request} {
		if _, ok := d.Requests[id]; !ok {
			return configError("setup", "A request reference does not exist: "+id)
		}
	}
	if d.Refresh.RenewalRequest != "" {
		if _, ok := d.Requests[d.Refresh.RenewalRequest]; !ok {
			return configError("refresh.renewal_request", "Unknown request.")
		}
	}
	for path, p := range map[string]string{"setup.search.options.items": s.Search.Options.Items, "setup.search.options.value": s.Search.Options.Value, "setup.search.options.label": s.Search.Options.Label, "setup.status.pointer": s.Status.Pointer, "mapping.collections.items": d.Mapping.Collections.Items, "mapping.collections.fields.name": d.Mapping.Collections.Fields.Name, "mapping.collections.fields.dates": d.Mapping.Collections.Fields.Dates, "mapping.metadata.property_id": d.Mapping.Metadata.PropertyID, "mapping.metadata.address_match": d.Mapping.Metadata.AddressMatch, "mapping.metadata.evidence_granularity": d.Mapping.Metadata.EvidenceGranularity} {
		if err := requirePointer(path, p); err != nil {
			return err
		}
	}
	if s.Search.Expect != nil {
		if err := requirePointer("setup.search.expect.pointer", s.Search.Expect.Pointer); err != nil {
			return err
		}
		if s.Search.Expect.Value == "" {
			return configError("setup.search.expect.value", "Choose the supported input discriminator.")
		}
	}
	seenStatuses := map[string]bool{}
	for _, list := range [][]string{s.Status.Pending, s.Status.Done, s.Status.Failed} {
		if len(list) == 0 || len(list) > 16 {
			return configError("setup.status", "Declare bounded pending, done and failed states.")
		}
		for _, v := range list {
			if v == "" || len(v) > 64 || seenStatuses[v] {
				return configError("setup.status", "Status values must be nonempty and disjoint.")
			}
			seenStatuses[v] = true
		}
	}
	captureTypes := map[string]string{}
	count := 0
	for id, q := range d.Requests {
		path := "requests." + id
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
	if s.Search.Request == s.Select.Request || s.Search.Request == s.Status.Request || s.Select.Request == s.Status.Request || d.Refresh.Request == s.Search.Request || d.Refresh.Request == s.Select.Request || d.Refresh.Request == s.Status.Request {
		return configError("setup", "Search, selection, status and refresh need distinct requests.")
	}
	if d.Requests[s.Search.Request].Method != "GET" || d.Requests[s.Status.Request].Method != "GET" || d.Requests[d.Refresh.Request].Method != "GET" {
		return configError("setup", "Search, status and refresh must use GET in this recipe.")
	}
	if d.Refresh.RenewalRequest != "" && d.Refresh.RenewalRequest != s.Select.Request {
		return configError("refresh.renewal_request", "Renewal must reuse the selection request in search_select/v1.")
	}

	searchState := captureNames(d.Requests[s.Search.Request], true)
	selectState := copyNames(searchState)
	for k := range captureNames(d.Requests[s.Select.Request], false) {
		selectState[k] = true
	}
	// A schedule must work whether selection finished immediately or after status checks.
	selectDone := copyNames(searchState)
	for k := range captureNames(d.Requests[s.Select.Request], true) {
		selectDone[k] = true
	}
	statusDone := copyNames(selectState)
	for k := range captureNames(d.Requests[s.Status.Request], true) {
		statusDone[k] = true
	}
	allState := map[string]bool{}
	for k := range selectDone {
		if statusDone[k] {
			allState[k] = true
		}
	}

	for id, q := range d.Requests {
		inputs := map[string]bool{s.Search.Input: true, s.Select.Input: true}
		states := allState
		switch id {
		case s.Search.Request:
			inputs = map[string]bool{s.Search.Input: true}
			states = map[string]bool{}
		case s.Select.Request:
			states = searchState
		case s.Status.Request:
			states = selectState
		}
		if err := validateRequestRefs(id, q, inputs, states); err != nil {
			return err
		}
	}
	raw, _ := json.Marshal(d.Mapping)
	var mapping any
	_ = json.Unmarshal(raw, &mapping)
	if err := checkMapping(mapping); err != nil {
		return err
	}
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
func captureNames(q model.Request, complete bool) map[string]bool {
	m := map[string]bool{}
	for k, c := range q.Captures {
		if !bool(c.Optional) && (complete || c.When != "complete") {
			m[k] = true
		}
	}
	return m
}
func copyNames(m map[string]bool) map[string]bool {
	n := map[string]bool{}
	for k, v := range m {
		n[k] = v
	}
	return n
}
func validateRequestRefs(id string, q model.Request, inputs, states map[string]bool) error {
	check := func(v model.Value, path bool) error {
		if v.Ref != "" {
			parts := strings.Split(v.Ref, ".")
			if len(parts) != 2 {
				return configError("requests."+id, "Use input.name or state.name.")
			}
			allowed := inputs
			if parts[0] == "state" {
				allowed = states
			} else if parts[0] != "input" {
				return configError("requests."+id, "Only input and saved state references are supported.")
			}
			if !allowed[parts[1]] {
				return configError("requests."+id, "Value is unavailable at this step: "+v.Ref)
			}
			return nil
		}
		encoded, err := json.Marshal(v.Literal)
		if err != nil {
			return configError("requests."+id, "Use a finite scalar value.")
		}
		if text, ok := v.Literal.(string); ok && len(text) > 2048 {
			return configError("requests."+id, "Use fixed text values of at most 2048 bytes.")
		}
		if len(encoded) > 0 && encoded[0] != '"' && string(encoded) != "true" && string(encoded) != "false" && string(encoded) != "null" {
			number, err := strconv.ParseFloat(string(encoded), 64)
			if err != nil || math.IsInf(number, 0) || math.Abs(number) > 9007199254740991 {
				return configError("requests."+id, "Use text for large numeric identifiers so Form and YAML keep their exact value.")
			}
		}
		if path {
			s, ok := v.Literal.(string)
			if !ok || s == "" || s == "." || s == ".." || len(s) > 256 || strings.ContainsAny(s, "/\\?#") {
				return configError("requests."+id+".path", "Use separate, nonempty path segments.")
			}
		}
		return nil
	}
	for _, v := range q.Path {
		if err := check(v, true); err != nil {
			return err
		}
	}
	for _, m := range []map[string]model.Value{q.Query, q.Body} {
		for k, v := range m {
			if k == "" || len(k) > 128 {
				return configError("requests."+id, "Invalid parameter name.")
			}
			if err := check(v, false); err != nil {
				return err
			}
		}
	}
	return nil
}
func checkMapping(v any) error { return checkMappingAt(v, "mapping") }
func checkMappingAt(v any, path string) error {
	switch x := v.(type) {
	case string:
		if !validPointer(x) {
			return configError(path, "Use JSON Pointers for provider fields.")
		}
	case map[string]any:
		for key, c := range x {
			if err := checkMappingAt(c, path+"."+key); err != nil {
				return err
			}
		}
	}
	return nil
}

// Pointer reads a decoded provider JSON value.
func Pointer(root any, p string) (any, bool) { return operations.Pointer(root, p) }
