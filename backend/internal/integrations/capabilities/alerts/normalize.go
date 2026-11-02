package alerts

// Builds a validated severe-weather-alert resource from provider data.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"node-herder/internal/integrations/capabilities"
	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
)

const maxAlertUnixTimestamp = 7258118400

var alertHTMLTag = regexp.MustCompile(`<[^>]*>`)

// NormalizeAlerts converts one provider snapshot into a bounded, ordered alert resource.
func NormalizeAlerts(root any, selectors mapping, context capabilities.Context) (model.WeatherAlerts, error) {
	alerts := model.WeatherAlerts{
		ContractVersion: model.WeatherAlertsCapability,
		InstanceID:      context.InstanceID,
		ResourceID:      context.ResourceID,
		Revision:        1,
		Status:          model.AlertStatusNone,
		Attribution:     selectors.Attribution,
		CapturedAt:      context.Now.UTC(),
		LastCheckedAt:   context.Now.UTC(),
	}
	if alerts.ResourceID == "" {
		id, err := context.NewIdentity("resource")
		if err != nil {
			return alerts, err
		}
		alerts.ResourceID = id
	}
	if context.Previous != nil {
		alerts.Revision = context.Previous.Revision + 1
	}
	rawItems := pointer(root, selectors.Warnings.Items)
	items, ok := rawItems.([]any)
	if rawItems == nil {
		items, ok = []any{}, true
	}
	if !ok || len(items) > model.AlertMaxWarnings {
		return alerts, invalid("The provider returned an invalid warning list.")
	}
	alerts.Warnings = make([]model.WeatherWarning, 0, len(items))
	for _, item := range items {
		warning, err := normalizeWarning(item, selectors.Warnings, context)
		if err != nil {
			return alerts, err
		}
		alerts.Warnings = append(alerts.Warnings, warning)
	}
	alerts.Warnings = model.OrderWarnings(alerts.Warnings)
	if len(alerts.Warnings) > 0 {
		alerts.Status = model.AlertStatusActive
	}
	alerts = alerts.Project(context.Now)
	return alerts, alerts.Validate()
}

// normalizeWarning maps one provider feature and drops unusable geometry while keeping the warning.
func normalizeWarning(item any, selectors warningSeries, context capabilities.Context) (model.WeatherWarning, error) {
	var warning model.WeatherWarning
	fields := selectors.Fields
	validFrom, err := alertInstant(pointer(item, fields[string(alertFieldValidFrom)]), selectors.Format)
	if err != nil {
		return warning, invalid("The provider returned an invalid warning start time.")
	}
	validTo, err := alertInstant(pointer(item, fields[string(alertFieldValidTo)]), selectors.Format)
	if err != nil || !validTo.After(validFrom) {
		return warning, invalid("The provider returned an invalid warning end time.")
	}
	headline := limitText(clean(textList(pointer(item, fields[string(alertFieldHeadline)]), " ")), model.AlertMaxHeadlineLength)
	if headline == "" {
		return warning, invalid("The provider returned a warning without a headline.")
	}
	severity, level := warningSeverity(item, fields)
	warningType := limitText(clean(textList(pointer(item, fields[string(alertFieldType)]), ", ")), model.AlertMaxTypeLength)
	if warningType == "" {
		warningType = "Weather warning"
	}
	issuedAt, err := optionalInstant(pointer(item, fields[string(alertFieldIssuedAt)]), selectors.Format)
	if err != nil {
		return warning, invalid("The provider returned an invalid warning issue time.")
	}
	modifiedAt, err := optionalInstant(pointer(item, fields[string(alertFieldModifiedAt)]), selectors.Format)
	if err != nil {
		return warning, invalid("The provider returned an invalid warning modification time.")
	}
	warning = model.WeatherWarning{
		Severity:    severity,
		Level:       level,
		Type:        warningType,
		Headline:    headline,
		Description: limitText(clean(textList(pointer(item, fields[string(alertFieldDescription)]), " ")), model.AlertMaxDescriptionLength),
		Instruction: limitText(clean(text(pointer(item, fields[string(alertFieldInstruction)]))), model.AlertMaxInstructionLength),
		Areas:       warningAreas(pointer(item, fields[string(alertFieldAreas)]), selectors.AreaName),
		ValidFrom:   validFrom.UTC(),
		ValidTo:     validTo.UTC(),
		IssuedAt:    issuedAt,
		ModifiedAt:  modifiedAt,
		Polygon:     warningPolygon(pointer(item, fields[string(alertFieldPolygon)])),
	}
	warning.WarningID = limitText(clean(text(pointer(item, fields[string(alertFieldWarningID)]))), model.AlertMaxIDLength)
	if warning.WarningID == "" {
		warning.WarningID = derivedWarningID(context.ProviderID, warningType, headline, warning.ValidFrom)
	}
	return warning, nil
}

// warningSeverity prefers an explicit normalized severity, then maps the provider level.
func warningSeverity(item any, fields map[string]string) (model.AlertSeverity, model.AlertLevel) {
	level := model.AlertLevelUnknown
	if selector := fields[string(alertFieldLevel)]; selector != "" {
		value := model.AlertLevel(strings.ToLower(clean(text(pointer(item, selector)))))
		switch value {
		case model.AlertLevelYellow, model.AlertLevelAmber, model.AlertLevelRed:
			level = value
		}
	}
	severity := model.AlertSeverityUnknown
	if selector := fields[string(alertFieldSeverity)]; selector != "" {
		switch strings.ToLower(clean(text(pointer(item, selector)))) {
		case string(model.AlertSeverityMinor):
			severity = model.AlertSeverityMinor
		case string(model.AlertSeverityModerate):
			severity = model.AlertSeverityModerate
		case string(model.AlertSeveritySevere):
			severity = model.AlertSeveritySevere
		case string(model.AlertSeverityExtreme):
			severity = model.AlertSeverityExtreme
		}
	}
	if severity == model.AlertSeverityUnknown {
		switch level {
		case model.AlertLevelYellow:
			severity = model.AlertSeverityMinor
		case model.AlertLevelAmber:
			severity = model.AlertSeveritySevere
		case model.AlertLevelRed:
			severity = model.AlertSeverityExtreme
		}
	}
	return severity, level
}

// warningAreas accepts a comma-separated label or an array of strings/objects.
func warningAreas(value any, areaName string) []string {
	var raw []string
	switch typed := value.(type) {
	case string:
		raw = strings.Split(typed, ",")
	case []any:
		raw = make([]string, 0, len(typed))
		for _, item := range typed {
			if areaName != "" {
				raw = append(raw, text(pointer(item, areaName)))
			} else {
				raw = append(raw, text(item))
			}
		}
	default:
		return []string{}
	}
	areas := make([]string, 0, len(raw))
	for _, area := range raw {
		cleaned := limitText(clean(area), model.AlertMaxAreaLength)
		if cleaned == "" {
			continue
		}
		areas = append(areas, cleaned)
		if len(areas) == model.AlertMaxAreas {
			break
		}
	}
	return areas
}

// warningPolygon retains bounded GeoJSON geometry; invalid geometry drops to nil.
func warningPolygon(value any) *model.AlertPolygon {
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	coordinates, ok := object["coordinates"]
	if !ok {
		return nil
	}
	encoded, err := json.Marshal(coordinates)
	if err != nil {
		return nil
	}
	polygon := &model.AlertPolygon{Type: text(object["type"]), Coordinates: encoded}
	if !model.AlertPolygonValid(polygon) {
		return nil
	}
	return polygon
}

// textList joins provider string arrays while preserving scalar values.
func textList(value any, separator string) string {
	items, ok := value.([]any)
	if !ok {
		return text(value)
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		if part := text(item); part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, separator)
}

// optionalInstant parses a present timestamp and preserves absence as nil.
func optionalInstant(value any, format alertTimeFormat) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := alertInstant(value, format)
	if err != nil {
		return nil, err
	}
	parsed = parsed.UTC()
	return &parsed, nil
}

// alertInstant parses RFC3339 or integer Unix timestamps.
func alertInstant(value any, format alertTimeFormat) (time.Time, error) {
	if format == alertFormatUnix {
		n := number(value)
		if n == nil || math.IsNaN(*n) || math.IsInf(*n, 0) || math.Trunc(*n) != *n || *n < 0 || *n > maxAlertUnixTimestamp {
			return time.Time{}, fmt.Errorf("invalid warning time")
		}
		return time.Unix(int64(*n), 0).UTC(), nil
	}
	parsed, err := time.Parse(time.RFC3339, text(value))
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

// number accepts supported provider number representations.
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

// text converts supported provider scalar values into plain text.
func text(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case float64:
		return fmt.Sprintf("%g", typed)
	default:
		return ""
	}
}

// clean strips HTML, control characters and repeated whitespace from provider text.
func clean(value string) string {
	value = html.UnescapeString(value)
	value = alertHTMLTag.ReplaceAllString(value, " ")
	value = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || !unicode.IsControl(r) {
			return r
		}
		return -1
	}, value)
	return strings.Join(strings.Fields(value), " ")
}

// limitText bounds a string by rune count without splitting UTF-8.
func limitText(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	// Validation counts bytes; trim at a UTF-8 boundary so non-ASCII text cannot fail later.
	trimmed := []byte(value)[:max]
	for !utf8.Valid(trimmed) {
		trimmed = trimmed[:len(trimmed)-1]
	}
	return string(trimmed)
}

// derivedWarningID builds a stable local id from the canonical warning tuple.
func derivedWarningID(providerID, warningType, headline string, validFrom time.Time) string {
	canonical := providerID + "\x00" + warningType + "\x00" + headline + "\x00" + validFrom.UTC().Format(time.RFC3339)
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:16])
}

// pointer resolves one JSON pointer without exposing provider envelopes.
func pointer(root any, selector string) any {
	if selector == "" {
		return nil
	}
	value, _ := operations.Pointer(root, selector)
	return value
}

// invalid returns a sanitized provider response failure.
func invalid(message string) error { return model.Fail("invalid_response", message) }
