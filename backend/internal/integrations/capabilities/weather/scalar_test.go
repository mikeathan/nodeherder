package weather

// Checks weather scalar and timestamp parsing.

import (
	"encoding/json"
	"testing"
	"time"
)

func TestInstantISOTimePrecision(t *testing.T) {
	for _, value := range []string{"2026-10-28T12:00Z", "2026-10-28T12:00:00Z", "2026-10-28T13:00+01:00"} {
		got, err := instant(value, "rfc3339", time.UTC)
		if err != nil || !got.Equal(time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)) {
			t.Fatalf("instant(%q) = %v, %v", value, got, err)
		}
	}
	for _, value := range []string{"2026-10-28T12:00", "2026-10-28T12:00Zgarbage", "bad"} {
		if _, err := instant(value, "rfc3339", time.UTC); err == nil {
			t.Fatalf("accepted invalid timestamp %q", value)
		}
	}
}

func TestLocalClock(t *testing.T) {
	tests := []struct{ name, value, want string }{
		{"clock", "07:12", "07:12"},
		{"local minute", "2026-10-28T07:12", "07:12"},
		{"local second", "2026-10-28T07:12:30", "07:12"},
		{"utc instant", "2026-10-28T07:12:00Z", "07:12"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := localClock(tt.value, time.UTC)
			if err != nil || got != tt.want {
				t.Fatalf("localClock(%q) = %q, %v", tt.value, got, err)
			}
		})
	}
	for _, value := range []string{"7:12", "25:00", "bad", ""} {
		if _, err := localClock(value, time.UTC); err == nil {
			t.Fatalf("accepted invalid clock %q", value)
		}
	}
}

func TestClockValueAcceptsUnixInstants(t *testing.T) {
	for _, value := range []any{json.Number("1793171520"), float64(1793171520)} {
		got, err := clockValue(value, time.UTC)
		if err != nil || got != "07:12" {
			t.Fatalf("clockValue(%v) = %q, %v", value, got, err)
		}
	}
	for _, value := range []any{json.Number("1.5"), float64(-1), true, json.Number("not-a-number")} {
		if _, err := clockValue(value, time.UTC); err == nil {
			t.Fatalf("accepted invalid clock value %v", value)
		}
	}
}
