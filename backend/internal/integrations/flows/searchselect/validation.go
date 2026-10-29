package searchselect

// Checks that selection-based recipes have valid steps, references and saved values.

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
)

// ValidateLinks checks the named search/select/status journey and disjoint progress states.
func ValidateLinks(d *model.Definition) error {
	s := d.Setup
	if s.Type == "location_select" {
		return validateLocation(d)
	}
	if s.Search.BaseURL != "" || s.Search.Postcode != nil {
		return configError("setup.search.base_url", "Separate search services require location_select.")
	}
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
	for path, p := range map[string]string{
		"setup.search.options.items": s.Search.Options.Items,
		"setup.search.options.value": s.Search.Options.Value,
		"setup.search.options.label": s.Search.Options.Label,
		"setup.status.pointer":       s.Status.Pointer,
	} {
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
	return nil
}

// ValidateAvailability proves saved values exist on immediate and polled completion paths.
func ValidateAvailability(d *model.Definition) error {
	s := d.Setup
	if s.Type == "location_select" {
		return validateLocation(d)
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
	return nil
}

// captureNames returns the required state fields available at a particular completion phase.
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

// validateRequestRefs ensures a request uses only inputs and state available at its step.
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

func configError(path, message string) error { return model.InvalidDefinition(path, message) }
func requirePointer(path, p string) error {
	if p == "" || !operations.ValidPointer(p) {
		return configError(path, "Use a JSON Pointer, for example /items.")
	}
	return nil
}
