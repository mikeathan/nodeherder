package integrations

// Defines the provider-independent severe-weather alert resource and its projection rules.

import (
	"encoding/json"
	"math"
	"slices"
	"sort"
	"time"
)

// WeatherAlertsCapability is the normalized severe-weather alert contract.
const WeatherAlertsCapability = "weather_alerts/v1"

// Alert status values.
type AlertStatus string

const (
	AlertStatusActive AlertStatus = "active"
	AlertStatusNone   AlertStatus = "none"
)

// Normalized severity values; the provider scale is preserved separately in Level.
type AlertSeverity string

const (
	AlertSeverityMinor    AlertSeverity = "minor"
	AlertSeverityModerate AlertSeverity = "moderate"
	AlertSeveritySevere   AlertSeverity = "severe"
	AlertSeverityExtreme  AlertSeverity = "extreme"
	AlertSeverityUnknown  AlertSeverity = "unknown"
)

// Provider warning-level values.
type AlertLevel string

const (
	AlertLevelYellow  AlertLevel = "yellow"
	AlertLevelAmber   AlertLevel = "amber"
	AlertLevelRed     AlertLevel = "red"
	AlertLevelUnknown AlertLevel = "unknown"
)

// Read-time freshness values.
type AlertFreshness string

const (
	AlertFreshnessFresh AlertFreshness = "fresh"
	AlertFreshnessStale AlertFreshness = "stale"
)

// Resource bounds shared by validation and provider normalization.
const (
	AlertMaxWarnings           = 50
	AlertMaxAreas              = 50
	AlertMaxAreaLength         = 120
	AlertMaxHeadlineLength     = 300
	AlertMaxDescriptionLength  = 4000
	AlertMaxInstructionLength  = 2000
	AlertMaxTypeLength         = 80
	AlertMaxIDLength           = 128
	AlertMaxAttributionLength  = 512
	AlertMaxVerticesPerWarning = 2000
	AlertMaxTotalVertices      = 100000
	alertFreshFor              = time.Hour
)

// AlertPolygon preserves a bounded GeoJSON Polygon or MultiPolygon.
// Coordinates remain raw JSON so the provider shape is retained without a geometry library.
type AlertPolygon struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

// WeatherWarning is one active warning with plain-text presentation fields.
type WeatherWarning struct {
	WarningID   string        `json:"warning_id"`
	Severity    AlertSeverity `json:"severity"`
	Level       AlertLevel    `json:"level"`
	Type        string        `json:"type"`
	Headline    string        `json:"headline"`
	Description string        `json:"description"`
	Instruction string        `json:"instruction"`
	Areas       []string      `json:"areas"`
	ValidFrom   time.Time     `json:"valid_from"`
	ValidTo     time.Time     `json:"valid_to"`
	IssuedAt    *time.Time    `json:"issued_at"`
	ModifiedAt  *time.Time    `json:"modified_at"`
	Polygon     *AlertPolygon `json:"polygon"`
}

// WeatherAlerts is the normalized, bounded alert snapshot.
type WeatherAlerts struct {
	ContractVersion string           `json:"contract_version"`
	InstanceID      string           `json:"instance_id"`
	ResourceID      string           `json:"resource_id"`
	Revision        uint64           `json:"revision"`
	Status          AlertStatus      `json:"status"`
	Warnings        []WeatherWarning `json:"warnings"`
	Attribution     string           `json:"attribution"`
	CapturedAt      time.Time        `json:"captured_at"`
	LastCheckedAt   time.Time        `json:"last_checked_at"`
	EvaluatedAt     time.Time        `json:"evaluated_at"`
	Freshness       AlertFreshness   `json:"freshness"`
}

// Validate rejects invalid saved or normalized alert snapshots independently of provider decoding.
func (a WeatherAlerts) Validate() error {
	invalid := Fail("invalid_resource", "The weather-alert data is invalid. Refresh or check the provider configuration.")
	if a.ContractVersion != WeatherAlertsCapability || a.Revision == 0 {
		return invalid
	}
	for _, value := range []string{a.InstanceID, a.ResourceID, a.Attribution} {
		if value == "" {
			return invalid
		}
	}
	if len(a.Attribution) > AlertMaxAttributionLength || a.CapturedAt.IsZero() || a.LastCheckedAt.IsZero() {
		return invalid
	}
	if a.Status != AlertStatusActive && a.Status != AlertStatusNone {
		return invalid
	}
	if len(a.Warnings) > AlertMaxWarnings {
		return invalid
	}
	if (a.Status == AlertStatusNone) != (len(a.Warnings) == 0) {
		return invalid
	}
	seen := make(map[string]bool, len(a.Warnings))
	totalVertices := 0
	for index, warning := range a.Warnings {
		if !validWarning(warning, seen) {
			return invalid
		}
		if warning.Polygon != nil {
			vertices := alertPolygonVertexCount(warning.Polygon)
			totalVertices += vertices
			if totalVertices > AlertMaxTotalVertices {
				return invalid
			}
		}
		if index > 0 && !warningBefore(a.Warnings[index-1], warning) {
			return invalid
		}
	}
	return nil
}

// validWarning checks one warning and records its id for duplicate detection.
func validWarning(warning WeatherWarning, seen map[string]bool) bool {
	if warning.WarningID == "" || len(warning.WarningID) > AlertMaxIDLength || seen[warning.WarningID] {
		return false
	}
	seen[warning.WarningID] = true
	if !validSeverity(warning.Severity) || !validLevel(warning.Level) {
		return false
	}
	if len(warning.Type) > AlertMaxTypeLength || warning.Headline == "" || len(warning.Headline) > AlertMaxHeadlineLength {
		return false
	}
	if len(warning.Description) > AlertMaxDescriptionLength || len(warning.Instruction) > AlertMaxInstructionLength {
		return false
	}
	if len(warning.Areas) > AlertMaxAreas {
		return false
	}
	for _, area := range warning.Areas {
		if len(area) > AlertMaxAreaLength {
			return false
		}
	}
	if warning.ValidFrom.IsZero() || warning.ValidTo.IsZero() || !warning.ValidTo.After(warning.ValidFrom) {
		return false
	}
	if warning.IssuedAt != nil && warning.IssuedAt.IsZero() {
		return false
	}
	if warning.ModifiedAt != nil && warning.ModifiedAt.IsZero() {
		return false
	}
	return warning.Polygon == nil || AlertPolygonValid(warning.Polygon)
}

// AlertPolygonValid reports whether a GeoJSON polygon is bounded and finite.
// Callers drop invalid geometry while retaining the warning itself.
func AlertPolygonValid(polygon *AlertPolygon) bool {
	if polygon == nil {
		return true
	}
	_, ok := alertPolygonVertices(polygon)
	return ok
}

// alertPolygonVertexCount returns the vertex count for an already-valid polygon.
func alertPolygonVertexCount(polygon *AlertPolygon) int {
	count, _ := alertPolygonVertices(polygon)
	return count
}

// alertPolygonVertices validates coordinates and counts positions.
func alertPolygonVertices(polygon *AlertPolygon) (int, bool) {
	if polygon == nil {
		return 0, true
	}
	switch polygon.Type {
	case "Polygon":
		var rings [][][]float64
		if json.Unmarshal(polygon.Coordinates, &rings) != nil {
			return 0, false
		}
		count := 0
		for _, ring := range rings {
			if len(ring) < 4 || !validPositions(ring) {
				return 0, false
			}
			count += len(ring)
			if count > AlertMaxVerticesPerWarning {
				return 0, false
			}
		}
		return count, count > 0
	case "MultiPolygon":
		var polygons [][][][]float64
		if json.Unmarshal(polygon.Coordinates, &polygons) != nil {
			return 0, false
		}
		count := 0
		for _, rings := range polygons {
			for _, ring := range rings {
				if len(ring) < 4 || !validPositions(ring) {
					return 0, false
				}
				count += len(ring)
				if count > AlertMaxVerticesPerWarning {
					return 0, false
				}
			}
		}
		return count, count > 0
	default:
		return 0, false
	}
}

// validPositions checks a ring of finite WGS84 longitude/latitude pairs.
func validPositions(positions [][]float64) bool {
	for _, position := range positions {
		if len(position) != 2 ||
			math.IsNaN(position[0]) || math.IsInf(position[0], 0) ||
			math.IsNaN(position[1]) || math.IsInf(position[1], 0) ||
			position[0] < -180 || position[0] > 180 ||
			position[1] < -90 || position[1] > 90 {
			return false
		}
	}
	return true
}

// validSeverity checks the normalized severity enum.
func validSeverity(value AlertSeverity) bool {
	switch value {
	case AlertSeverityMinor, AlertSeverityModerate, AlertSeveritySevere, AlertSeverityExtreme, AlertSeverityUnknown:
		return true
	default:
		return false
	}
}

// validLevel checks the provider level enum.
func validLevel(value AlertLevel) bool {
	switch value {
	case AlertLevelYellow, AlertLevelAmber, AlertLevelRed, AlertLevelUnknown:
		return true
	default:
		return false
	}
}

// warningBefore reports ordered severity descending, valid-from ascending, then id for stability.
func warningBefore(left, right WeatherWarning) bool {
	leftRank, rightRank := alertSeverityRank(left.Severity), alertSeverityRank(right.Severity)
	if leftRank != rightRank {
		return leftRank > rightRank
	}
	if !left.ValidFrom.Equal(right.ValidFrom) {
		return left.ValidFrom.Before(right.ValidFrom)
	}
	return left.WarningID < right.WarningID
}

// alertSeverityRank maps severity to its display order.
func alertSeverityRank(value AlertSeverity) int {
	switch value {
	case AlertSeverityExtreme:
		return 4
	case AlertSeveritySevere:
		return 3
	case AlertSeverityModerate:
		return 2
	case AlertSeverityMinor:
		return 1
	default:
		return 0
	}
}

// OrderWarnings returns a sorted copy; callers never reorder the saved slice in place.
func OrderWarnings(warnings []WeatherWarning) []WeatherWarning {
	ordered := slices.Clone(warnings)
	sort.SliceStable(ordered, func(left, right int) bool {
		return warningBefore(ordered[left], ordered[right])
	})
	return ordered
}

// Project drops expired warnings, refreshes ordering/status and derives freshness.
// It does not change content revision, capture time or last-check time.
func (a WeatherAlerts) Project(now time.Time) WeatherAlerts {
	a.EvaluatedAt = now.UTC()
	a.Freshness = AlertFreshnessFresh
	if !now.Before(a.CapturedAt.Add(alertFreshFor)) {
		a.Freshness = AlertFreshnessStale
	}
	active := make([]WeatherWarning, 0, len(a.Warnings))
	for _, warning := range a.Warnings {
		if warning.ValidTo.After(now) {
			active = append(active, warning)
		}
	}
	a.Warnings = OrderWarnings(active)
	if len(a.Warnings) == 0 {
		a.Status = AlertStatusNone
	} else {
		a.Status = AlertStatusActive
	}
	return a
}
