package alerts

// Validates severe-weather-alert provider selectors and response shapes.

import (
	"bytes"
	"encoding/json"

	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
)

// alertField identifies one warning property the alerts contract understands.
type alertField string

const (
	alertFieldWarningID   alertField = "warning_id"
	alertFieldSeverity    alertField = "severity"
	alertFieldLevel       alertField = "level"
	alertFieldType        alertField = "type"
	alertFieldHeadline    alertField = "headline"
	alertFieldDescription alertField = "description"
	alertFieldInstruction alertField = "instruction"
	alertFieldAreas       alertField = "areas"
	alertFieldValidFrom   alertField = "valid_from"
	alertFieldValidTo     alertField = "valid_to"
	alertFieldIssuedAt    alertField = "issued_at"
	alertFieldModifiedAt  alertField = "modified_at"
	alertFieldPolygon     alertField = "polygon"
)

// alertTimeFormat identifies supported provider timestamp encodings.
type alertTimeFormat string

const (
	alertFormatRFC3339 alertTimeFormat = "rfc3339"
	alertFormatUnix    alertTimeFormat = "unix"
)

// warningSeries describes how one provider's warning collection is shaped.
type warningSeries struct {
	Items    string            `json:"items"`
	Format   alertTimeFormat   `json:"format"`
	AreaName string            `json:"area_name,omitempty"`
	Fields   map[string]string `json:"fields"`
}

// mapping is the capability-owned provider recipe for severe-weather alerts.
type mapping struct {
	Attribution string        `json:"attribution"`
	Warnings    warningSeries `json:"warnings"`
}

// decode strictly parses an alerts mapping and validates its selectors.
func decode(raw map[string]any) (mapping, error) {
	var m mapping
	encoded, err := json.Marshal(raw)
	if err != nil {
		return m, err
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&m); err != nil {
		return m, model.InvalidDefinition("mapping", "Invalid alerts mapping.")
	}
	if m.Attribution == "" || len(m.Attribution) > model.AlertMaxAttributionLength {
		return m, model.InvalidDefinition("mapping.attribution", "Declare provider attribution.")
	}
	if err = validateSeries(m.Warnings); err != nil {
		return m, err
	}
	return m, nil
}

// validateSeries checks item/time selectors, required fields and supported pointers.
func validateSeries(selectors warningSeries) error {
	if selectors.Items == "" || !operations.ValidPointer(selectors.Items) {
		return model.InvalidDefinition("mapping.warnings.items", "Use a JSON pointer for warning items.")
	}
	switch selectors.Format {
	case alertFormatRFC3339, alertFormatUnix:
	default:
		return model.InvalidDefinition("mapping.warnings.format", "Declare a supported warning time format.")
	}
	if selectors.AreaName != "" && !operations.ValidPointer(selectors.AreaName) {
		return model.InvalidDefinition("mapping.warnings.area_name", "Use a JSON pointer for area names.")
	}
	for name, selector := range selectors.Fields {
		if !supportedField(alertField(name)) {
			return model.InvalidDefinition("mapping.warnings.fields", "Unknown alert field.")
		}
		if selector == "" || !operations.ValidPointer(selector) {
			return model.InvalidDefinition("mapping.warnings.fields", "Use JSON pointers for alert fields.")
		}
	}
	for _, required := range []alertField{alertFieldHeadline, alertFieldValidFrom, alertFieldValidTo} {
		if _, ok := selectors.Fields[string(required)]; !ok {
			return model.InvalidDefinition("mapping.warnings.fields", "Declare required alert fields.")
		}
	}
	if _, severity := selectors.Fields[string(alertFieldSeverity)]; !severity {
		if _, level := selectors.Fields[string(alertFieldLevel)]; !level {
			return model.InvalidDefinition("mapping.warnings.fields", "Declare a severity or provider level selector.")
		}
	}
	return nil
}

// supportedField limits mappings to normalized alert properties.
func supportedField(value alertField) bool {
	switch value {
	case alertFieldWarningID, alertFieldSeverity, alertFieldLevel, alertFieldType,
		alertFieldHeadline, alertFieldDescription, alertFieldInstruction, alertFieldAreas,
		alertFieldValidFrom, alertFieldValidTo, alertFieldIssuedAt, alertFieldModifiedAt,
		alertFieldPolygon:
		return true
	default:
		return false
	}
}
