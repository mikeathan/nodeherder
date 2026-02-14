package timescope_test

import (
	"node-herder/internal/mcp/timescope"
	"testing"
	"time"
)

func TestToday(t *testing.T) {
	tests := []struct {
		name     string
		timezone *time.Location
	}{
		{
			name:     "UTC timezone",
			timezone: time.UTC,
		},
		{
			name:     "fixed offset timezone",
			timezone: time.FixedZone("EST", -5*60*60),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := timescope.Today(tt.timezone)

			// Should be start of day (00:00:00)
			if result.Hour() != 0 || result.Minute() != 0 || result.Second() != 0 {
				t.Errorf("Today() should be start of day, got %v", result)
			}

			// Should be in the correct timezone
			if result.Location().String() != tt.timezone.String() {
				t.Errorf("Today() location = %v, want %v", result.Location(), tt.timezone)
			}
		})
	}
}

func TestYesterday(t *testing.T) {
	tests := []struct {
		name     string
		timezone *time.Location
	}{
		{
			name:     "UTC timezone",
			timezone: time.UTC,
		},
		{
			name:     "fixed offset timezone",
			timezone: time.FixedZone("PST", -8*60*60),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yesterday := timescope.Yesterday(tt.timezone)
			today := timescope.Today(tt.timezone)

			// Yesterday should be exactly one day before today
			diff := today.Sub(yesterday)
			if diff != 24*time.Hour {
				t.Errorf("Yesterday() should be 24 hours before Today(), got diff = %v", diff)
			}

			// Should be start of day (00:00:00)
			if yesterday.Hour() != 0 || yesterday.Minute() != 0 || yesterday.Second() != 0 {
				t.Errorf("Yesterday() should be start of day, got %v", yesterday)
			}
		})
	}
}

func TestStartOfDay(t *testing.T) {
	tests := []struct {
		name  string
		input time.Time
		want  time.Time
	}{
		{
			name:  "mid day time",
			input: time.Date(2024, 6, 15, 14, 30, 45, 123456789, time.UTC),
			want:  time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "already start of day",
			input: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			want:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "end of day",
			input: time.Date(2024, 12, 31, 23, 59, 59, 999999999, time.UTC),
			want:  time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "preserves timezone",
			input: time.Date(2024, 6, 15, 14, 30, 0, 0, time.FixedZone("CET", 1*60*60)),
			want:  time.Date(2024, 6, 15, 0, 0, 0, 0, time.FixedZone("CET", 1*60*60)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := timescope.StartOfDay(tt.input)
			if !got.Equal(tt.want) {
				t.Errorf("StartOfDay(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestEndOfDay(t *testing.T) {
	tests := []struct {
		name  string
		input time.Time
		want  time.Time
	}{
		{
			name:  "mid day time",
			input: time.Date(2024, 6, 15, 14, 30, 45, 0, time.UTC),
			want:  time.Date(2024, 6, 15, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:  "start of day",
			input: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			want:  time.Date(2024, 1, 1, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:  "preserves timezone",
			input: time.Date(2024, 6, 15, 10, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
			want:  time.Date(2024, 6, 15, 23, 59, 59, 999999999, time.FixedZone("JST", 9*60*60)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := timescope.EndOfDay(tt.input)
			if !got.Equal(tt.want) {
				t.Errorf("EndOfDay(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestStartOfDay_EndOfDay_Consistency(t *testing.T) {
	// Verify that EndOfDay is always after StartOfDay for the same input
	input := time.Date(2024, 7, 20, 12, 0, 0, 0, time.UTC)

	start := timescope.StartOfDay(input)
	end := timescope.EndOfDay(input)

	if !end.After(start) {
		t.Errorf("EndOfDay should be after StartOfDay: start=%v, end=%v", start, end)
	}

	// The difference should be almost 24 hours (minus 1 nanosecond)
	diff := end.Sub(start)
	expected := 24*time.Hour - time.Nanosecond
	if diff != expected {
		t.Errorf("EndOfDay - StartOfDay = %v, want %v", diff, expected)
	}
}
