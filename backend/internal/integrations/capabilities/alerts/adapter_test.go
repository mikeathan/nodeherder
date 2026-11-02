package alerts

// Checks the alert mapping, normalization, geometry fallback and projection.

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	defaults "node-herder/configs/integrations"
	"node-herder/internal/integrations/capabilities"
	"node-herder/internal/integrations/definitions"
	model "node-herder/models/integrations"
)

const metOfficePayload = `{
  "features":[{
    "type":"Feature",
    "geometry":{"type":"MultiPolygon","coordinates":[[[[-1.2,53.4],[-1.1,53.5],[-1.0,53.4],[-1.2,53.4]]]]},
    "properties":{
      "warningId":"w1",
      "warningLevel":"AMBER",
      "weatherType":["Wind","Rain"],
      "warningHeadline":"Strong winds may cause disruption",
      "warningFurtherDetails":"<p>A deep area of low pressure brings gusts of <b>60-70 mph</b>.</p>",
      "whatToExpect":["Gusts may cause delays."],
      "whatShouldIDo":"Secure loose objects and allow extra travel time.",
      "affectedAreas":[{"regionName":"Yorkshire & Humber"},{"regionName":"North East England"}],
      "validFromDate":"2026-10-30T06:00:00Z",
      "validToDate":"2026-10-30T18:00:00Z",
      "issuedDate":"2026-10-29T10:12:00Z",
      "modifiedDate":"2026-10-29T10:12:00Z"
    }
  }]
}`

func metOfficeMapping() mapping {
	return mapping{
		Attribution: "Warnings: Met Office National Severe Weather Warning Service.",
		Warnings: warningSeries{
			Items:    "/features",
			Format:   alertFormatRFC3339,
			AreaName: "/regionName",
			Fields: map[string]string{
				"warning_id":  "/properties/warningId",
				"level":       "/properties/warningLevel",
				"type":        "/properties/weatherType",
				"headline":    "/properties/warningHeadline",
				"description": "/properties/warningFurtherDetails",
				"instruction": "/properties/whatShouldIDo",
				"areas":       "/properties/affectedAreas",
				"valid_from":  "/properties/validFromDate",
				"valid_to":    "/properties/validToDate",
				"issued_at":   "/properties/issuedDate",
				"modified_at": "/properties/modifiedDate",
				"polygon":     "/geometry",
			},
		},
	}
}

func normalizeMetOffice(t *testing.T, payload string) (model.WeatherAlerts, error) {
	t.Helper()
	var root any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		t.Fatal(err)
	}
	raw := rawMapping(t, metOfficeMapping())
	now := time.Date(2026, 10, 29, 15, 0, 0, 0, time.UTC)
	result, err := (Adapter{}).Normalize(root, raw, capabilities.Context{
		InstanceID: "instance",
		ProviderID: "met_office",
		Now:        now,
		NewIdentity: func(string) (string, error) {
			return "resource", nil
		},
	})
	if err != nil {
		return model.WeatherAlerts{}, err
	}
	var alerts model.WeatherAlerts
	if err = result.Resource.Decode(&alerts); err != nil {
		t.Fatal(err)
	}
	return alerts, nil
}

func rawMapping(t *testing.T, value mapping) map[string]any {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err = json.Unmarshal(encoded, &raw); err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestShippedAlertProvidersCompile(t *testing.T) {
	registry, err := capabilities.NewRegistry(Adapter{})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"alerts/met_office.yaml", "alerts/nws.yaml"} {
		source, err := defaults.Files.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := definitions.NewCompiler(registry).CompileYAML(string(source)); err != nil {
			if failure, ok := err.(*model.Failure); ok {
				t.Fatalf("%s: %+v", name, failure.Fields)
			}
			t.Fatalf("%s: %v", name, err)
		}
	}
}

func TestMetOfficeAlertNormalization(t *testing.T) {
	alerts, err := normalizeMetOffice(t, metOfficePayload)
	if err != nil {
		t.Fatal(err)
	}
	if alerts.Status != model.AlertStatusActive || len(alerts.Warnings) != 1 {
		t.Fatalf("unexpected status: %+v", alerts)
	}
	warning := alerts.Warnings[0]
	if warning.WarningID != "w1" || warning.Severity != model.AlertSeveritySevere || warning.Level != model.AlertLevelAmber {
		t.Fatalf("unexpected severity: %+v", warning)
	}
	if warning.Type != "Wind, Rain" || warning.Headline != "Strong winds may cause disruption" {
		t.Fatalf("unexpected warning text: %+v", warning)
	}
	if !strings.Contains(warning.Description, "deep area of low pressure") ||
		!strings.Contains(warning.Description, "60-70 mph") ||
		strings.Contains(warning.Description, "<") ||
		strings.Contains(warning.Description, "Gusts may cause delays") {
		t.Fatalf("description used the wrong provider field: %q", warning.Description)
	}
	if len(warning.Areas) != 2 || warning.Areas[0] != "Yorkshire & Humber" {
		t.Fatalf("unexpected areas: %+v", warning.Areas)
	}
	if warning.Polygon == nil || warning.Polygon.Type != "MultiPolygon" {
		t.Fatalf("missing polygon: %+v", warning.Polygon)
	}
	if alerts.Freshness != model.AlertFreshnessFresh {
		t.Fatalf("freshness = %q", alerts.Freshness)
	}
}

func TestAlertNormalizationEmptyAndExpired(t *testing.T) {
	empty, err := normalizeMetOffice(t, `{"features":[]}`)
	if err != nil {
		t.Fatal(err)
	}
	if empty.Status != model.AlertStatusNone || len(empty.Warnings) != 0 {
		t.Fatalf("empty snapshot = %+v", empty)
	}
	nullFeatures, err := normalizeMetOffice(t, `{"features":null}`)
	if err != nil || nullFeatures.Status != model.AlertStatusNone {
		t.Fatalf("null feature list = %+v, %v", nullFeatures, err)
	}
	expiredPayload := strings.Replace(metOfficePayload, `"validFromDate":"2026-10-30T06:00:00Z"`, `"validFromDate":"2026-10-29T06:00:00Z"`, 1)
	expiredPayload = strings.Replace(expiredPayload, `"validToDate":"2026-10-30T18:00:00Z"`, `"validToDate":"2026-10-29T14:00:00Z"`, 1)
	expired, err := normalizeMetOffice(t, expiredPayload)
	if err != nil {
		t.Fatal(err)
	}
	if expired.Status != model.AlertStatusNone || len(expired.Warnings) != 0 {
		t.Fatalf("expired snapshot = %+v", expired)
	}
}

func TestAlertNormalizationKeepsWarningWithBadGeometry(t *testing.T) {
	payload := strings.Replace(metOfficePayload, `[-1.2,53.4]`, `[200,53.4]`, 1)
	alerts, err := normalizeMetOffice(t, payload)
	if err != nil {
		t.Fatal(err)
	}
	if alerts.Status != model.AlertStatusActive || len(alerts.Warnings) != 1 || alerts.Warnings[0].Polygon != nil {
		t.Fatalf("bad geometry was not dropped: %+v", alerts)
	}
}

func TestAlertNormalizationRejectsMissingHeadline(t *testing.T) {
	payload := strings.Replace(metOfficePayload, `"warningHeadline":"Strong winds may cause disruption"`, `"warningHeadline":""`, 1)
	if _, err := normalizeMetOffice(t, payload); err == nil {
		t.Fatal("missing headline accepted")
	}
}

func TestAlertAdapterSourceAndProjection(t *testing.T) {
	alerts, err := normalizeMetOffice(t, metOfficePayload)
	if err != nil {
		t.Fatal(err)
	}
	resource, err := model.NewResource(alerts)
	if err != nil {
		t.Fatal(err)
	}
	sources, err := (Adapter{}).Sources(resource)
	if err != nil || len(sources) != 1 || sources[0].Kind != "weather_alerts" || sources[0].ResourceID != "resource" {
		t.Fatalf("unexpected sources: %+v, %v", sources, err)
	}
	projected, err := (Adapter{}).Project(resource, time.Date(2026, 10, 29, 16, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if err = projected.Decode(&alerts); err != nil || alerts.Freshness != model.AlertFreshnessStale {
		t.Fatalf("expected stale projection: %+v, %v", alerts, err)
	}
}
