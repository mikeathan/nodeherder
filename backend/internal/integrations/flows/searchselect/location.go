package searchselect

// Validates location setup recipes and translates location search results into choices.

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"node-herder/internal/integrations/destination"
	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
)

// LocationCoordinates parses a selected latitude/longitude pair in decimal degrees.
// Both coordinates must be finite and within geographic bounds; invalid input returns
// a safe selection error without retaining provider text.
func LocationCoordinates(value string) (float64, float64, error) {
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return 0, 0, invalidLocationChoice()
	}
	latitude, err := parseCoordinate(parts[0], 90)
	if err != nil {
		return 0, 0, err
	}
	longitude, err := parseCoordinate(parts[1], 180)
	if err != nil {
		return 0, 0, err
	}
	return latitude, longitude, nil
}

// parseCoordinate rejects non-finite and out-of-range latitude or longitude values.
func parseCoordinate(value string, limit float64) (float64, error) {
	coordinate, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, invalidLocationChoice()
	}
	if math.IsNaN(coordinate) || math.IsInf(coordinate, 0) || math.Abs(coordinate) > limit {
		return 0, invalidLocationChoice()
	}
	return coordinate, nil
}

func invalidLocationChoice() error { return model.Fail("invalid_choice", "Choose a valid location.") }

// validateLocation checks the complete location selection recipe before activation.
func validateLocation(definition *model.Definition) error {
	setup := definition.Setup
	if err := validateLocationInputs(definition); err != nil {
		return err
	}
	if err := validateLocationRequests(definition); err != nil {
		return err
	}
	if err := validateLocationSelectors(setup.Search.Options); err != nil {
		return err
	}
	if setup.Search.BaseURL != "" {
		if _, err := destination.Parse(setup.Search.BaseURL); err != nil {
			return configError("setup.search.base_url", "Use a public HTTPS search service.")
		}
	}
	if alternative := setup.Search.Postcode; alternative != nil {
		if alternative.Request == setup.Search.Request || alternative.Request == setup.Select.Request || definition.Requests[alternative.Request].Method != "GET" {
			return configError("setup.search.postcode.request", "Use a separate named GET request.")
		}
		if _, err := destination.Parse(alternative.BaseURL); err != nil {
			return configError("setup.search.postcode.base_url", "Use a public HTTPS search service.")
		}
		if err := validateLocationSelectors(alternative.Options); err != nil {
			return err
		}
	}
	return validateLocationReferences(definition)
}

// validateLocationInputs checks the input types needed by location setup.
func validateLocationInputs(definition *model.Definition) error {
	setup := definition.Setup
	if setup.Search.Input == setup.Select.Input {
		return configError("setup", "Search and location selection need different inputs.")
	}
	if definition.Inputs[setup.Search.Input].Type != "text" {
		return configError("setup.search.input", "Location search requires a text input.")
	}
	if definition.Inputs[setup.Select.Input].Type != "select" {
		return configError("setup.select.input", "Location selection requires a select input.")
	}
	return nil
}

// validateLocationRequests checks the request references used by location setup.
func validateLocationRequests(definition *model.Definition) error {
	setup := definition.Setup
	if setup.Search.Request == setup.Select.Request || setup.Select.Request != definition.Refresh.Request {
		return configError("setup", "Use a separate location search and a shared selection/refresh request.")
	}
	if definition.Refresh.RenewalRequest != "" || setup.Status.Request != "" {
		return configError("setup", "Location selection completes synchronously without polling or renewal.")
	}
	for _, name := range []string{setup.Search.Request, setup.Select.Request, definition.Refresh.Request} {
		if definition.Requests[name].Method != "GET" {
			return configError("setup", "Location setup uses named GET requests.")
		}
	}
	return nil
}

// validateLocationSelectors checks the JSON selectors for location choices.
func validateLocationSelectors(options model.ChoiceMapping) error {
	if !operations.ValidPointer(options.Items) {
		return configError("setup.search.options.items", "Use a JSON pointer for the location list.")
	}
	for _, selector := range []string{options.Label, options.Latitude, options.Longitude} {
		if selector == "" || !operations.ValidPointer(selector) {
			return configError("setup.search.options", "Use nonempty location field selectors.")
		}
	}
	for _, selector := range options.Context {
		if selector == "" || !operations.ValidPointer(selector) {
			return configError("setup.search.options.context", "Use JSON pointers for location context.")
		}
	}
	return nil
}

// validateLocationReferences verifies which inputs and state are available to each request.
func validateLocationReferences(definition *model.Definition) error {
	setup := definition.Setup
	for name, request := range definition.Requests {
		inputs := map[string]bool{setup.Search.Input: true, setup.Select.Input: true}
		state := map[string]bool{"latitude": true, "longitude": true}
		if name == setup.Search.Request || (setup.Search.Postcode != nil && name == setup.Search.Postcode.Request) {
			inputs = map[string]bool{setup.Search.Input: true}
			state = map[string]bool{}
		}
		if err := validateRequestRefs(name, request, inputs, state); err != nil {
			return err
		}
	}
	return nil
}

// coordinateChoice converts a provider location into an opaque choice with a readable label.
func coordinateChoice(item any, step model.SearchStep) (model.Choice, error) {
	latitude, _ := operations.Pointer(item, step.Options.Latitude)
	longitude, _ := operations.Pointer(item, step.Options.Longitude)
	label, _ := operations.Pointer(item, step.Options.Label)
	coordinates := coordinateText(latitude) + "," + coordinateText(longitude)
	if _, _, err := LocationCoordinates(coordinates); err != nil {
		return model.Choice{}, model.Fail("invalid_response", "The provider returned invalid coordinates.")
	}
	name, ok := label.(string)
	if !ok || name == "" || len(name) > 256 {
		return model.Choice{}, model.Fail("invalid_response", "The provider returned an invalid location name.")
	}
	parts := []string{name}
	for _, selector := range step.Options.Context {
		value, _ := operations.Pointer(item, selector)
		context, ok := value.(string)
		if ok && context != "" && len(context) <= 256 && !strings.EqualFold(context, parts[len(parts)-1]) {
			parts = append(parts, context)
		}
	}
	return model.Choice{ID: coordinates, Label: strings.Join(parts, ", ")}, nil
}

// coordinateText formats supported provider coordinate values for validation.
func coordinateText(value any) string {
	switch coordinate := value.(type) {
	case json.Number:
		return coordinate.String()
	case float64:
		return strconv.FormatFloat(coordinate, 'f', -1, 64)
	default:
		return ""
	}
}

const selectionLabelKey = "nodeherder.selection_label"

// SelectionState returns an owned flow checkpoint for an accepted choice. The reserved
// label key cannot collide with provider capture identifiers and never enters shared YAML.
func SelectionState(recipe model.SetupRecipe, choice model.Choice, previous map[string]any) (map[string]any, error) {
	state := make(map[string]any, len(previous)+1)
	for key, value := range previous {
		state[key] = value
	}
	if recipe.Type == "location_select" {
		latitude, longitude, err := LocationCoordinates(choice.ID)
		if err != nil {
			return nil, err
		}
		state = map[string]any{"latitude": latitude, "longitude": longitude}
	}
	state[selectionLabelKey] = choice.Label
	return state, nil
}

// SelectionLabel restores a validated choice label from a flow checkpoint; older
// snapshots without that additive metadata return an empty label.
func SelectionLabel(state map[string]any) string {
	label, _ := state[selectionLabelKey].(string)
	return label
}
