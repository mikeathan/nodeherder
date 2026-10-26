package integrations

import (
	"testing"
	"time"
)

func TestWasteCalendarDate(t *testing.T) {
	// Source calendar semantics must not inherit the installation/process timezone.
	t.Setenv("TZ", "Pacific/Honolulu")
	for _, tc := range []struct {
		instant string
		date    string
	}{
		{"2026-10-24T22:30:00Z", "2026-10-24"},
		{"2026-10-24T23:30:00Z", "2026-10-25"},
		{"2026-10-26T00:30:00Z", "2026-10-26"},
	} {
		t.Run(tc.instant, func(t *testing.T) {
			now, err := time.Parse(time.RFC3339, tc.instant)
			if err != nil {
				t.Fatal(err)
			}
			date, err := WasteCalendarDate(now)
			if err != nil || date != tc.date {
				t.Fatalf("got %q, %v; want %q", date, err, tc.date)
			}
		})
	}
	if WasteTimezone != "Europe/London" {
		t.Fatal("v1 wire contract changed")
	}
}
