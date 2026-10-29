package weather

// Validates provider selectors, forecast shapes and unit conversions for weather data.

import (
	"bytes"
	"encoding/json"

	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
)

type series struct {
	// Offset is an optional root selector for fixed provider daily boundaries.
	Offset  string             `json:"offset,omitempty"`
	Items   string             `json:"items"`
	Parent  string             `json:"parent,omitempty"`
	Time    string             `json:"time"`
	Format  string             `json:"format"`
	Columns bool               `json:"columns"`
	Fields  map[string]string  `json:"fields"`
	Scales  map[string]float64 `json:"scales,omitempty"`
}
type mapping struct {
	Location        string            `json:"location"`
	Timezone        string            `json:"timezone"`
	DefaultTimezone string            `json:"default_timezone"`
	Attribution     string            `json:"attribution"`
	Conditions      map[string]string `json:"conditions"`
	Current         *series           `json:"current,omitempty"`
	Hourly          series            `json:"hourly"`
	Daily           series            `json:"daily"`
}

// decode strictly parses a weather mapping and validates its series before use.
func decode(raw map[string]any) (mapping, error) {
	var m mapping
	b, e := json.Marshal(raw)
	if e != nil {
		return m, e
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e = d.Decode(&m); e != nil {
		return m, model.InvalidDefinition("mapping", "Invalid weather mapping.")
	}
	if m.Attribution == "" || len(m.Attribution) > 512 {
		return m, model.InvalidDefinition("mapping.attribution", "Declare provider attribution.")
	}
	for _, selectors := range []*series{m.Current, &m.Hourly, &m.Daily} {
		if selectors != nil {
			if err := validateSeries(*selectors); err != nil {
				return m, err
			}
		}
	}
	return m, nil
}

// validateSeries checks field selectors, timestamp format and bounded unit scales.
func validateSeries(selectors series) error {
	if selectors.Time == "" {
		return model.InvalidDefinition("mapping.time", "Declare a timestamp selector.")
	}
	for _, selector := range []string{selectors.Items, selectors.Time, selectors.Parent, selectors.Offset} {
		if !operations.ValidPointer(selector) {
			return model.InvalidDefinition("mapping", "Use JSON pointers.")
		}
	}
	switch selectors.Format {
	case "unix", "rfc3339", "date", "local":
	default:
		return model.InvalidDefinition("mapping.format", "Declare a supported time format.")
	}
	for field, selector := range selectors.Fields {
		if !supportedWeatherField(field) {
			return model.InvalidDefinition("mapping.fields", "Unknown weather field.")
		}
		if !operations.ValidPointer(selector) {
			return model.InvalidDefinition("mapping.fields", "Use JSON pointers for weather fields.")
		}
	}
	for field, scale := range selectors.Scales {
		if !supportedWeatherField(field) {
			return model.InvalidDefinition("mapping.scales", "Unknown weather field.")
		}
		if scale <= 0 || scale > 1000 {
			return model.InvalidDefinition("mapping.scales", "Use bounded positive unit conversions.")
		}
	}
	return nil
}

// supportedWeatherField limits mappings to fields understood by the weather contract.
func supportedWeatherField(field string) bool {
	switch field {
	case "temperature", "feels_like", "wind", "gust", "direction", "humidity", "precipitation_probability", "precipitation", "condition", "high", "low":
		return true
	default:
		return false
	}
}
