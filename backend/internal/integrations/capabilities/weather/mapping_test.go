package weather

// Checks weather mapping validation for daily clock selectors.

import "testing"

// mappingFixture returns a minimal valid mapping whose daily fields can be extended per test.
func mappingFixture() map[string]any {
	return map[string]any{
		"location":         "",
		"timezone":         "",
		"default_timezone": "UTC",
		"attribution":      "Test",
		"conditions":       map[string]any{},
		"hourly": map[string]any{
			"items":  "/hourly",
			"time":   "/time",
			"format": "unix",
			"fields": map[string]any{"temperature": "/temperature"},
		},
		"daily": map[string]any{
			"items":  "/daily",
			"time":   "/time",
			"format": "date",
			"fields": map[string]any{"high": "/high"},
		},
	}
}

func TestDailyClockMappingValidation(t *testing.T) {
	t.Run("accepts sunrise and sunset", func(t *testing.T) {
		raw := mappingFixture()
		daily := raw["daily"].(map[string]any)
		daily["fields"] = map[string]any{"high": "/high", "sunrise": "/sunrise", "sunset": "/sunset"}
		daily["clocks"] = []any{"sunrise", "sunset"}
		if _, err := decode(raw); err != nil {
			t.Fatal(err)
		}
	})

	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"rejects clocks on hourly", func(raw map[string]any) {
			raw["hourly"].(map[string]any)["fields"] = map[string]any{"temperature": "/temperature", "sunrise": "/sunrise"}
			raw["hourly"].(map[string]any)["clocks"] = []any{"sunrise"}
		}},
		{"rejects unknown clock fields", func(raw map[string]any) {
			daily := raw["daily"].(map[string]any)
			daily["fields"] = map[string]any{"high": "/high", "uv_max": "/uv"}
			daily["clocks"] = []any{"uv_max"}
		}},
		{"rejects clocks without a selector", func(raw map[string]any) {
			raw["daily"].(map[string]any)["clocks"] = []any{"sunrise"}
		}},
		{"rejects duplicate clocks", func(raw map[string]any) {
			daily := raw["daily"].(map[string]any)
			daily["fields"] = map[string]any{"high": "/high", "sunrise": "/sunrise"}
			daily["clocks"] = []any{"sunrise", "sunrise"}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := mappingFixture()
			tt.mutate(raw)
			if _, err := decode(raw); err == nil {
				t.Fatal("invalid clock mapping accepted")
			}
		})
	}
}
