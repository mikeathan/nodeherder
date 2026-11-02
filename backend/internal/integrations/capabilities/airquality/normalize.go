package airquality

// Builds a validated air-quality resource from provider data.

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"node-herder/internal/integrations/capabilities"
	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
)

const maxUnixTimestamp = 7258118400

// reader shares immutable selectors and timezone across one normalization.
type reader struct {
	root      any
	selectors mapping
	zone      *time.Location
	family    string
}

// parsedSeries keeps forecast points aligned with their optional index and pollen evidence.
type parsedSeries struct {
	points  []model.AirQualityPoint
	indices []*float64
	pollen  []model.AirQualityPollen
}

// newReader resolves the provider or default IANA timezone before reading air-quality data.
func newReader(root any, selectors mapping) (reader, error) {
	zoneName := text(pointer(root, selectors.Timezone))
	if zoneName == "" {
		zoneName = selectors.DefaultTimezone
	}
	if zoneName == "" {
		return reader{}, invalid("The provider omitted a timezone.")
	}
	zone, err := time.LoadLocation(zoneName)
	if err != nil {
		return reader{}, invalid("The provider returned an invalid timezone.")
	}
	return reader{root: root, selectors: selectors, zone: zone, family: selectors.IndexFamily}, nil
}

// forecast assembles a complete resource while preserving refresh identity and revision.
func (reader reader) forecast(context capabilities.Context) (model.AirQualityForecast, error) {
	forecast := model.AirQualityForecast{
		ContractVersion: model.AirQualityCapability,
		InstanceID:      context.InstanceID,
		ResourceID:      context.ResourceID,
		Revision:        1,
		Location:        text(pointer(reader.root, reader.selectors.Location)),
		Timezone:        reader.zone.String(),
		IndexFamily:     reader.family,
		Attribution:     reader.selectors.Attribution,
		CapturedAt:      context.Now.UTC(),
		LastCheckedAt:   context.Now.UTC(),
	}
	if context.Origin != nil {
		latitude, longitude := context.Origin.Latitude, context.Origin.Longitude
		forecast.Latitude = &latitude
		forecast.Longitude = &longitude
	}
	if forecast.ResourceID == "" {
		id, err := context.NewIdentity("resource")
		if err != nil {
			return forecast, err
		}
		forecast.ResourceID = id
	}
	if forecast.Location == "" {
		forecast.Location = context.SelectedLabel
	}
	if forecast.Location == "" {
		forecast.Location = context.SelectedValue
	}
	if context.Previous != nil {
		forecast.Revision = context.Previous.Revision + 1
	}
	hourly, err := reader.read(reader.selectors.Hourly)
	if err != nil {
		return forecast, err
	}
	forecast.Hourly = hourly.points
	cutoff := context.Now.Add(-24 * time.Hour)
	for _, point := range forecast.Hourly {
		if point.Time.Before(cutoff) {
			return forecast, invalid("The provider returned air quality older than 24 hours.")
		}
	}
	rootIndex := (*float64)(nil)
	rootPollen := model.AirQualityPollen{}
	if reader.selectors.Current != nil {
		current, err := reader.read(*reader.selectors.Current)
		if err != nil {
			return forecast, err
		}
		if len(current.points) != 1 {
			return forecast, invalid("The provider returned invalid current air quality.")
		}
		owned := current.points[0].AirQualityValues
		forecast.Current = &owned
		rootIndex = current.indices[0]
		rootPollen = current.pollen[0]
	} else if len(hourly.points) > 0 {
		index := latestPointIndex(hourly.points, context.Now)
		owned := hourly.points[index].AirQualityValues
		forecast.Current = &owned
		rootIndex = hourly.indices[index]
		rootPollen = hourly.pollen[index]
	}
	if !hasPollen(rootPollen) && len(hourly.pollen) > 0 {
		rootPollen = hourly.pollen[0]
	}
	forecast.Pollen = rootPollen
	forecast.Index = rootIndex
	forecast.Category = model.AirQualityCategoryFor(reader.family, rootIndex)
	if reader.selectors.Dominant != "" {
		forecast.Dominant = text(pointer(reader.root, reader.selectors.Dominant))
	}
	return forecast, forecast.Validate()
}

// read parses a series and keeps its index/pollen evidence aligned with each point.
func (reader reader) read(selectors series) (parsedSeries, error) {
	sourceRows, err := rows(reader.root, selectors)
	if err != nil {
		return parsedSeries{}, invalid("The provider returned invalid air-quality rows.")
	}
	result := parsedSeries{
		points:  make([]model.AirQualityPoint, 0, len(sourceRows)),
		indices: make([]*float64, 0, len(sourceRows)),
		pollen:  make([]model.AirQualityPollen, 0, len(sourceRows)),
	}
	for _, row := range sourceRows {
		timestamp, err := instant(row["time"], selectors.Format, reader.zone)
		if err != nil {
			return parsedSeries{}, invalid("The provider returned an invalid air-quality timestamp.")
		}
		result.points = append(result.points, model.AirQualityPoint{
			Time:             timestamp.UTC(),
			AirQualityValues: values(row),
		})
		result.indices = append(result.indices, quantity(row, fieldIndex))
		result.pollen = append(result.pollen, pollen(row))
	}
	return result, nil
}

// hasPollen reports whether a row supplied any species; all-null is a valid out-of-season state.
func hasPollen(value model.AirQualityPollen) bool {
	return value.Alder != nil || value.Birch != nil || value.Grass != nil ||
		value.Mugwort != nil || value.Olive != nil || value.Ragweed != nil
}

// latestPointIndex returns the latest elapsed point, or the first future point.
func latestPointIndex(points []model.AirQualityPoint, now time.Time) int {
	index := 0
	for position, point := range points {
		if !point.Time.After(now) {
			index = position
		}
	}
	return index
}

// rows reads column-oriented or object-oriented provider series.
func rows(root any, selectors series) ([]map[string]any, error) {
	source := pointer(root, selectors.Items)
	if selectors.Columns {
		return columnRows(source, selectors)
	}
	return objectRows(source, selectors)
}

// columnRows pairs equal-length field columns with their timestamps.
func columnRows(source any, selectors series) ([]map[string]any, error) {
	timestamps, ok := pointer(source, selectors.Time).([]any)
	if !ok || len(timestamps) > model.AirQualityMaxHourly {
		return nil, fmt.Errorf("invalid time column")
	}
	columns := make(map[field][]any, len(selectors.Fields))
	for name, selector := range selectors.Fields {
		column, ok := pointer(source, selector).([]any)
		if !ok || len(column) != len(timestamps) {
			return nil, fmt.Errorf("invalid air-quality column")
		}
		columns[field(name)] = column
	}
	result := make([]map[string]any, 0, len(timestamps))
	for index, timestamp := range timestamps {
		row := map[string]any{"time": timestamp}
		for key, column := range columns {
			row[string(key)] = column[index]
		}
		if err := validateRow(row); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, nil
}

// objectRows reads selectors from individual objects or one object.
func objectRows(source any, selectors series) ([]map[string]any, error) {
	if source == nil {
		return nil, fmt.Errorf("missing air-quality series")
	}
	items, ok := source.([]any)
	if !ok {
		items = []any{source}
	}
	if len(items) > model.AirQualityMaxHourly {
		return nil, fmt.Errorf("too many air-quality rows")
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		row := map[string]any{"time": pointer(item, selectors.Time)}
		for name, selector := range selectors.Fields {
			row[name] = pointer(item, selector)
		}
		if err := validateRow(row); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, nil
}

// validateRow rejects nonnumeric concentrations before normalization.
func validateRow(row map[string]any) error {
	for key, value := range row {
		if key == "time" {
			continue
		}
		if !supportedField(field(key)) {
			return fmt.Errorf("unknown air-quality field")
		}
		if value != nil && number(value) == nil {
			return fmt.Errorf("invalid air-quality quantity")
		}
	}
	return nil
}

// values reads pollutant fields, preserving absent values as nil.
func values(row map[string]any) model.AirQualityValues {
	return model.AirQualityValues{
		PM25:            quantity(row, fieldPM25),
		PM10:            quantity(row, fieldPM10),
		Ozone:           quantity(row, fieldOzone),
		NitrogenDioxide: quantity(row, fieldNitrogenDioxide),
		SulphurDioxide:  quantity(row, fieldSulphurDioxide),
		CarbonMonoxide:  quantity(row, fieldCarbonMonoxide),
	}
}

// pollen reads the six optional pollen species.
func pollen(row map[string]any) model.AirQualityPollen {
	return model.AirQualityPollen{
		Alder:   quantity(row, fieldAlder),
		Birch:   quantity(row, fieldBirch),
		Grass:   quantity(row, fieldGrass),
		Mugwort: quantity(row, fieldMugwort),
		Olive:   quantity(row, fieldOlive),
		Ragweed: quantity(row, fieldRagweed),
	}
}

// quantity reads one numeric provider value.
func quantity(row map[string]any, key field) *float64 {
	return number(row[string(key)])
}

// number accepts supported provider number representations and returns nil for other values.
func number(value any) *float64 {
	var result float64
	var err error
	switch typed := value.(type) {
	case json.Number:
		result, err = typed.Float64()
	case float64:
		result = typed
	case string:
		result, err = strconv.ParseFloat(typed, 64)
	default:
		return nil
	}
	if err != nil {
		return nil
	}
	return &result
}

// text converts supported provider scalar values into labels or pollutant names.
func text(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return ""
	}
}

// instant parses supported provider timestamp formats.
func instant(value any, format timeFormat, zone *time.Location) (time.Time, error) {
	if format == formatUnix {
		n := number(value)
		if n == nil || math.IsNaN(*n) || math.IsInf(*n, 0) || math.Trunc(*n) != *n || *n < 0 || *n > maxUnixTimestamp {
			return time.Time{}, fmt.Errorf("invalid time")
		}
		return time.Unix(int64(*n), 0).UTC(), nil
	}
	timestamp, err := time.Parse(time.RFC3339, text(value))
	if err == nil {
		return timestamp, nil
	}
	return time.ParseInLocation("2006-01-02T15:04", text(value), zone)
}

// pointer resolves one JSON pointer without exposing provider envelopes.
func pointer(root any, selector string) any {
	value, _ := operations.Pointer(root, selector)
	return value
}

// invalid returns a sanitized provider response failure.
func invalid(message string) error { return model.Fail("invalid_response", message) }
