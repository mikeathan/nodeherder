package searchselect

// Interprets provider search choices and lookup status for selection-based setup.

import (
	"slices"

	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
)

// Progress classifies a configured lookup response, independently of its output capability.
// A pending result requires an explicit later continuation; this function never polls.
func Progress(body any, step model.StatusStep) (string, error) {
	value, ok := operations.Pointer(body, step.Pointer)
	status, isString := value.(string)
	if !ok || !isString {
		return "", model.Fail("invalid_response", "The provider omitted its lookup status.")
	}
	switch {
	case slices.Contains(step.Pending, status):
		return "pending", nil
	case slices.Contains(step.Done, status):
		return "done", nil
	case slices.Contains(step.Failed, status):
		return "", model.Fail("lookup_failed", "The provider could not retrieve the requested data.")
	default:
		return "", model.Fail("invalid_response", "The provider returned an unknown lookup status.")
	}
}

// Choices validates and copies a bounded search result using provider-defined ID/label paths.
// Values remain opaque: an address, station or account identifier has no meaning to this flow.
func Choices(body any, step model.SearchStep) ([]model.Choice, error) {
	if step.Expect != nil {
		value, ok := operations.Pointer(body, step.Expect.Pointer)
		if !ok || value != step.Expect.Value {
			return nil, model.Fail("unsupported_setup", "The provider requires a different setup for this search.")
		}
	}
	value, ok := operations.Pointer(body, step.Options.Items)
	if value == nil && step.Options.Latitude != "" {
		return []model.Choice{}, nil
	}
	list, isList := value.([]any)
	if !ok || !isList || len(list) > 2000 {
		return nil, model.Fail("invalid_response", "The provider returned an invalid selection list.")
	}
	choices := make([]model.Choice, 0, len(list))
	seen := make(map[string]bool, len(list))
	for _, item := range list {
		if step.Options.Latitude != "" {
			choice, err := coordinateChoice(item, step)
			if err != nil {
				return nil, err
			}
			if !seen[choice.ID] {
				choices = append(choices, choice)
				seen[choice.ID] = true
			}
			continue
		}
		id, _ := operations.Pointer(item, step.Options.Value)
		label, _ := operations.Pointer(item, step.Options.Label)
		idText, idOK := id.(string)
		labelText, labelOK := label.(string)
		if !idOK || !labelOK || idText == "" || labelText == "" || len(idText) > 2048 || len(labelText) > 512 || seen[idText] {
			return nil, model.Fail("invalid_response", "The provider returned invalid or duplicate selections.")
		}
		seen[idText] = true
		choices = append(choices, model.Choice{ID: idText, Label: labelText})
	}
	return choices, nil
}
