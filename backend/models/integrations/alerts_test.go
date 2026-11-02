package integrations

import (
	"encoding/json"
	"testing"
	"time"
)

func TestWeatherAlertsValidationAndProjection(t *testing.T) {
	now := time.Date(2026, 10, 29, 15, 0, 0, 0, time.UTC)
	fixture := func() WeatherAlerts {
		return WeatherAlerts{
			ContractVersion: WeatherAlertsCapability,
			InstanceID:      "instance",
			ResourceID:      "resource",
			Revision:        1,
			Status:          AlertStatusActive,
			Warnings: []WeatherWarning{
				testWarning("severe-warning", AlertSeveritySevere, now.Add(-time.Hour), now.Add(3*time.Hour)),
				testWarning("minor-warning", AlertSeverityMinor, now.Add(-2*time.Hour), now.Add(2*time.Hour)),
			},
			Attribution:   "Provider",
			CapturedAt:    now,
			LastCheckedAt: now,
		}
	}
	if err := fixture().Validate(); err != nil {
		t.Fatal(err)
	}
	projected := fixture().Project(now)
	if projected.Status != AlertStatusActive || len(projected.Warnings) != 2 || projected.Freshness != AlertFreshnessFresh {
		t.Fatalf("unexpected projection: %+v", projected)
	}
	stale := fixture().Project(now.Add(time.Hour))
	if stale.Freshness != AlertFreshnessStale {
		t.Fatalf("freshness = %q", stale.Freshness)
	}
	expired := fixture()
	expired.Warnings[0].ValidTo = now
	projected = expired.Project(now)
	if projected.Status != AlertStatusActive || len(projected.Warnings) != 1 {
		t.Fatalf("expired warning was not dropped: %+v", projected)
	}
	empty := WeatherAlerts{
		ContractVersion: WeatherAlertsCapability,
		InstanceID:      "instance",
		ResourceID:      "resource",
		Revision:        1,
		Status:          AlertStatusNone,
		Attribution:     "Provider",
		CapturedAt:      now,
		LastCheckedAt:   now,
	}
	if err := empty.Validate(); err != nil {
		t.Fatal(err)
	}
	if projected = empty.Project(now); projected.Status != AlertStatusNone {
		t.Fatalf("empty status = %q", projected.Status)
	}
}

func TestWeatherAlertsRejectInvalidWarnings(t *testing.T) {
	now := time.Date(2026, 10, 29, 15, 0, 0, 0, time.UTC)
	fixture := func() WeatherAlerts {
		return WeatherAlerts{
			ContractVersion: WeatherAlertsCapability,
			InstanceID:      "instance",
			ResourceID:      "resource",
			Revision:        1,
			Status:          AlertStatusActive,
			Warnings:        []WeatherWarning{testWarning("warning", AlertSeveritySevere, now, now.Add(time.Hour))},
			Attribution:     "Provider",
			CapturedAt:      now,
			LastCheckedAt:   now,
		}
	}
	tests := []struct {
		name   string
		change func(*WeatherAlerts)
	}{
		{"duplicate id", func(a *WeatherAlerts) { a.Warnings = append(a.Warnings, a.Warnings[0]) }},
		{"missing headline", func(a *WeatherAlerts) { a.Warnings[0].Headline = "" }},
		{"inverted window", func(a *WeatherAlerts) { a.Warnings[0].ValidTo = a.Warnings[0].ValidFrom }},
		{"unknown severity", func(a *WeatherAlerts) { a.Warnings[0].Severity = "catastrophic" }},
		{"unknown level", func(a *WeatherAlerts) { a.Warnings[0].Level = "purple" }},
		{"unordered", func(a *WeatherAlerts) {
			a.Warnings = append([]WeatherWarning{testWarning("minor", AlertSeverityMinor, now, now.Add(time.Hour))}, a.Warnings...)
		}},
		{"status mismatch", func(a *WeatherAlerts) { a.Status = AlertStatusNone }},
		{"bad geometry", func(a *WeatherAlerts) {
			a.Warnings[0].Polygon = &AlertPolygon{Type: "Polygon", Coordinates: json.RawMessage(`[[[200,0],[1,0],[1,1],[0,0]]]`)}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alerts := fixture()
			tt.change(&alerts)
			if alerts.Validate() == nil {
				t.Fatal("invalid alert snapshot accepted")
			}
		})
	}
}

func TestAlertPolygonBounds(t *testing.T) {
	valid := &AlertPolygon{
		Type:        "MultiPolygon",
		Coordinates: json.RawMessage(`[[[[-1.2,53.4],[-1.1,53.5],[-1.0,53.4],[-1.2,53.4]]]]`),
	}
	if !AlertPolygonValid(valid) {
		t.Fatal("valid polygon rejected")
	}
	if AlertPolygonValid(&AlertPolygon{Type: "Polygon", Coordinates: json.RawMessage(`[[[0,90.1],[1,0],[1,1],[0,0]]]`)}) {
		t.Fatal("out-of-range latitude accepted")
	}
	if AlertPolygonValid(&AlertPolygon{Type: "Point", Coordinates: json.RawMessage(`[0,0]`)}) {
		t.Fatal("unsupported geometry accepted")
	}
}

func TestOrderWarnings(t *testing.T) {
	now := time.Date(2026, 10, 29, 15, 0, 0, 0, time.UTC)
	warnings := []WeatherWarning{
		testWarning("minor", AlertSeverityMinor, now, now.Add(2*time.Hour)),
		testWarning("extreme", AlertSeverityExtreme, now.Add(time.Hour), now.Add(3*time.Hour)),
		testWarning("severe", AlertSeveritySevere, now, now.Add(time.Hour)),
	}
	ordered := OrderWarnings(warnings)
	if ordered[0].WarningID != "extreme" || ordered[1].WarningID != "severe" || ordered[2].WarningID != "minor" {
		t.Fatalf("unexpected order: %+v", ordered)
	}
	if warnings[0].WarningID != "minor" {
		t.Fatal("OrderWarnings mutated the input")
	}
}

func testWarning(id string, severity AlertSeverity, from, to time.Time) WeatherWarning {
	return WeatherWarning{
		WarningID: id,
		Severity:  severity,
		Level:     AlertLevelAmber,
		Type:      "Wind",
		Headline:  "Strong winds",
		ValidFrom: from,
		ValidTo:   to,
	}
}
