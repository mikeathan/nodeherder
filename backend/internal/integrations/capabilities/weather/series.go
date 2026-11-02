package weather

// Reads row, column and nested forecast shapes into a common sequence of weather values.

import (
	"fmt"

	model "node-herder/models/integrations"
)

// rows reads a forecast series and rejects nonnumeric quantities before normalization.
func rows(root any, s series) ([]map[string]any, error) {
	result, err := rawRows(root, s)
	if err != nil {
		return nil, err
	}
	for _, row := range result {
		for name, value := range row {
			field := weatherField(name)
			if field != fieldTime && field != fieldCondition && field != fieldNightCondition && !s.isClock(field) && value != nil && number(value) == nil {
				return nil, fmt.Errorf("invalid weather quantity")
			}
		}
	}
	return result, nil
}

// rawRows chooses the parser matching the configured provider response shape.
func rawRows(root any, selectors series) ([]map[string]any, error) {
	if selectors.Parent != "" {
		return nestedRows(root, selectors)
	}
	source := pointer(root, selectors.Items)
	if selectors.Columns {
		return columnRows(source, selectors)
	}
	return objectRows(source, selectors)
}

// nestedRows flattens bounded parent and child lists into a single forecast series.
func nestedRows(root any, selectors series) ([]map[string]any, error) {
	parents, ok := pointer(root, selectors.Parent).([]any)
	if !ok || len(parents) > maxNestedParents {
		return nil, fmt.Errorf("invalid parent rows")
	}
	childSelectors := selectors
	childSelectors.Parent = ""
	result := []map[string]any{}
	for _, parent := range parents {
		childRows, err := rows(parent, childSelectors)
		if err != nil {
			return nil, err
		}
		result = append(result, childRows...)
		if len(result) > maxSeriesRows {
			return nil, fmt.Errorf("too many hours")
		}
	}
	return result, nil
}

// columnRows pairs equal-length field columns with their corresponding timestamps.
func columnRows(source any, selectors series) ([]map[string]any, error) {
	timestamps, ok := pointer(source, selectors.Time).([]any)
	if !ok || len(timestamps) > maxSeriesRows {
		return nil, fmt.Errorf("invalid time column")
	}
	columns := make(map[string][]any, len(selectors.Fields))
	for field, selector := range selectors.Fields {
		column, ok := pointer(source, selector).([]any)
		if !ok || len(column) != len(timestamps) {
			return nil, fmt.Errorf("invalid forecast column")
		}
		columns[field] = column
	}
	result := make([]map[string]any, 0, len(timestamps))
	for index, timestamp := range timestamps {
		row := map[string]any{string(fieldTime): timestamp}
		for field, column := range columns {
			row[field] = column[index]
		}
		result = append(result, row)
	}
	return result, nil
}

// objectRows reads selectors from individual forecast objects or a single object.
func objectRows(source any, selectors series) ([]map[string]any, error) {
	if source == nil {
		return nil, fmt.Errorf("missing forecast")
	}
	items, ok := source.([]any)
	if !ok {
		items = []any{source}
	}
	if len(items) > maxSeriesRows {
		return nil, fmt.Errorf("too many forecast rows")
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		row := map[string]any{string(fieldTime): pointer(item, selectors.Time)}
		for field, selector := range selectors.Fields {
			row[field] = pointer(item, selector)
		}
		result = append(result, row)
	}
	return result, nil
}

// quantity applies the configured unit scale while preserving absent values as nil.
func quantity(row map[string]any, s series, key weatherField) *float64 {
	n := number(row[string(key)])
	if n != nil {
		if scale, ok := s.Scales[string(key)]; ok {
			*n *= scale
		}
	}
	return n
}
func values(row map[string]any, s series) model.WeatherValues {
	return model.WeatherValues{
		Temperature:              quantity(row, s, fieldTemperature),
		FeelsLike:                quantity(row, s, fieldFeelsLike),
		Wind:                     quantity(row, s, fieldWind),
		Gust:                     quantity(row, s, fieldGust),
		Direction:                quantity(row, s, fieldDirection),
		Humidity:                 quantity(row, s, fieldHumidity),
		PrecipitationProbability: quantity(row, s, fieldPrecipitationProbability),
		Precipitation:            quantity(row, s, fieldPrecipitation),
		UV:                       quantity(row, s, fieldUV),
	}
}

// condition maps a provider code to the shared weather condition or unknown.
func condition(row map[string]any, m mapping) string {
	if c, ok := m.Conditions[text(row[string(fieldCondition)])]; ok {
		return c
	}
	return model.WeatherConditionUnknown
}
