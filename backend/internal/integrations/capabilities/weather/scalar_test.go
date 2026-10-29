package weather

// Checks weather scalar and timestamp parsing.

import (
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
