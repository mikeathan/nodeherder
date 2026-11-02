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
	Offset string     `json:"offset,omitempty"`
	Items  string     `json:"items"`
	Parent string     `json:"parent,omitempty"`
	Time   string     `json:"time"`
	Format timeFormat `json:"format"`
	// Clocks lists fields whose values are local clock timestamps reduced to HH:MM in the
	// resource timezone. Only daily series may declare clocks; only sunrise/sunset are clocks.
	Clocks  []weatherField     `json:"clocks,omitempty"`
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
	if m.Attribution == "" || len(m.Attribution) > model.WeatherMaxAttributionLength {
		return m, model.InvalidDefinition("mapping.attribution", "Declare provider attribution.")
	}
	for _, selectors := range []*series{m.Current, &m.Hourly, &m.Daily} {
		if selectors != nil {
			if err := validateSeries(*selectors); err != nil {
				return m, err
			}
		}
	}
	if m.Current != nil && len(m.Current.Clocks) > 0 {
		return m, model.InvalidDefinition("mapping.current.clocks", "Clock selectors are only supported for daily weather.")
	}
	if len(m.Hourly.Clocks) > 0 {
		return m, model.InvalidDefinition("mapping.hourly.clocks", "Clock selectors are only supported for daily weather.")
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
	case formatUnix, formatRFC3339, formatDate, formatLocal:
	default:
		return model.InvalidDefinition("mapping.format", "Declare a supported time format.")
	}
	for name, selector := range selectors.Fields {
		field := weatherField(name)
		if !supportedWeatherField(field) {
			return model.InvalidDefinition("mapping.fields", "Unknown weather field.")
		}
		if !operations.ValidPointer(selector) {
			return model.InvalidDefinition("mapping.fields", "Use JSON pointers for weather fields.")
		}
	}
	for name, scale := range selectors.Scales {
		if !supportedWeatherField(weatherField(name)) {
			return model.InvalidDefinition("mapping.scales", "Unknown weather field.")
		}
		if scale <= 0 || scale > maxUnitScale {
			return model.InvalidDefinition("mapping.scales", "Use bounded positive unit conversions.")
		}
	}
	seen := map[weatherField]bool{}
	for _, field := range selectors.Clocks {
		if !clockField(field) {
			return model.InvalidDefinition("mapping.clocks", "Declare only sunrise or sunset clock selectors.")
		}
		if seen[field] {
			return model.InvalidDefinition("mapping.clocks", "Declare each clock selector once.")
		}
		if _, ok := selectors.Fields[string(field)]; !ok {
			return model.InvalidDefinition("mapping.clocks", "Declare a selector for each clock field.")
		}
		seen[field] = true
	}
	return nil
}

// supportedWeatherField limits mappings to fields understood by the weather contract.
func supportedWeatherField(field weatherField) bool {
	switch field {
	case fieldTemperature, fieldFeelsLike, fieldWind, fieldGust, fieldDirection, fieldHumidity,
		fieldPressure, fieldVisibility, fieldPrecipitationProbability, fieldPrecipitation,
		fieldUV, fieldCondition, fieldNightCondition, fieldHigh, fieldLow, fieldFeelsLikeHigh,
		fieldFeelsLikeLow, fieldWindGust, fieldWindDirection, fieldRainProbability,
		fieldSnowProbability, fieldUVMax, fieldSunrise, fieldSunset:
		return true
	default:
		return false
	}
}

// clockField reports whether a weather field is a local clock time rather than a quantity.
func clockField(field weatherField) bool { return field == fieldSunrise || field == fieldSunset }

// isClock reports whether the field is normalized from a local clock timestamp.
func (s series) isClock(field weatherField) bool {
	for _, clock := range s.Clocks {
		if clock == field {
			return true
		}
	}
	return false
}
