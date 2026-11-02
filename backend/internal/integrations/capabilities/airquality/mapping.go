package airquality

// Validates air-quality provider selectors and response shapes.

import (
	"bytes"
	"encoding/json"

	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
)

// field identifies one quantity the air-quality contract understands.
type field string

const (
	fieldIndex           field = "index"
	fieldPM25            field = "pm2_5"
	fieldPM10            field = "pm10"
	fieldOzone           field = "ozone"
	fieldNitrogenDioxide field = "nitrogen_dioxide"
	fieldSulphurDioxide  field = "sulphur_dioxide"
	fieldCarbonMonoxide  field = "carbon_monoxide"
	fieldAlder           field = "alder"
	fieldBirch           field = "birch"
	fieldGrass           field = "grass"
	fieldMugwort         field = "mugwort"
	fieldOlive           field = "olive"
	fieldRagweed         field = "ragweed"
)

// timeFormat identifies how the provider encodes timestamps.
type timeFormat string

const (
	formatUnix    timeFormat = "unix"
	formatRFC3339 timeFormat = "rfc3339"
)

// series describes one current or hourly provider series.
type series struct {
	Items   string            `json:"items"`
	Time    string            `json:"time"`
	Format  timeFormat        `json:"format"`
	Columns bool              `json:"columns"`
	Fields  map[string]string `json:"fields"`
}

// mapping is the capability-owned provider recipe, including the declared index family.
type mapping struct {
	Location        string  `json:"location"`
	Timezone        string  `json:"timezone"`
	DefaultTimezone string  `json:"default_timezone"`
	Attribution     string  `json:"attribution"`
	IndexFamily     string  `json:"index_family"`
	Dominant        string  `json:"dominant,omitempty"`
	Current         *series `json:"current,omitempty"`
	Hourly          series  `json:"hourly"`
}

// decode strictly parses an air-quality mapping and validates its selectors.
func decode(raw map[string]any) (mapping, error) {
	var m mapping
	encoded, err := json.Marshal(raw)
	if err != nil {
		return m, err
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&m); err != nil {
		return m, model.InvalidDefinition("mapping", "Invalid air-quality mapping.")
	}
	if m.Attribution == "" || len(m.Attribution) > model.AirQualityMaxAttributionLength {
		return m, model.InvalidDefinition("mapping.attribution", "Declare provider attribution.")
	}
	if m.IndexFamily != model.AirQualityFamilyEuropean && m.IndexFamily != model.AirQualityFamilyUS {
		return m, model.InvalidDefinition("mapping.index_family", "Declare european or us air-quality index bands.")
	}
	if m.Timezone == "" && m.DefaultTimezone == "" {
		return m, model.InvalidDefinition("mapping.timezone", "Declare a timezone selector or default.")
	}
	if m.Dominant != "" && !operations.ValidPointer(m.Dominant) {
		return m, model.InvalidDefinition("mapping.dominant", "Use a JSON pointer for the dominant pollutant.")
	}
	if m.Current != nil {
		if err := validateSeries(*m.Current); err != nil {
			return m, err
		}
	}
	if err := validateSeries(m.Hourly); err != nil {
		return m, err
	}
	return m, nil
}

// validateSeries checks item/time selectors, format and supported field pointers.
func validateSeries(selectors series) error {
	for _, selector := range []string{selectors.Items, selectors.Time} {
		if selector == "" || !operations.ValidPointer(selector) {
			return model.InvalidDefinition("mapping", "Use JSON pointers for air-quality items and time.")
		}
	}
	switch selectors.Format {
	case formatUnix, formatRFC3339:
	default:
		return model.InvalidDefinition("mapping.format", "Declare a supported time format.")
	}
	for name, selector := range selectors.Fields {
		if !supportedField(field(name)) {
			return model.InvalidDefinition("mapping.fields", "Unknown air-quality field.")
		}
		if selector == "" || !operations.ValidPointer(selector) {
			return model.InvalidDefinition("mapping.fields", "Use JSON pointers for air-quality fields.")
		}
	}
	return nil
}

// supportedField limits mappings to normalized air-quality quantities.
func supportedField(value field) bool {
	switch value {
	case fieldIndex, fieldPM25, fieldPM10, fieldOzone, fieldNitrogenDioxide,
		fieldSulphurDioxide, fieldCarbonMonoxide, fieldAlder, fieldBirch,
		fieldGrass, fieldMugwort, fieldOlive, fieldRagweed:
		return true
	default:
		return false
	}
}
