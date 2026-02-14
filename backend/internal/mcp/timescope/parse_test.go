package timescope_test

import (
	"node-herder/internal/mcp/timescope"
	"testing"
	"time"
)

func TestTimeScope_Parse(t *testing.T) {
	// Fixed time for testing: 2024-01-15 14:30:00 UTC
	fixedTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	clock := func() time.Time { return fixedTime }

	ts := timescope.New(
		timescope.WithTimezone(time.UTC),
		timescope.WithClock(clock),
	)

	tests := []struct {
		name     string
		scope    string
		wantFrom time.Time
		wantTo   time.Time
	}{
		{
			name:     "today",
			scope:    "today",
			wantFrom: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantTo:   fixedTime,
		},
		{
			name:     "yesterday",
			scope:    "yesterday",
			wantFrom: time.Date(2024, 1, 14, 0, 0, 0, 0, time.UTC),
			wantTo:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "last_24_hours",
			scope:    "last_24_hours",
			wantFrom: fixedTime.Add(-24 * time.Hour),
			wantTo:   fixedTime,
		},
		{
			name:     "last_7_days",
			scope:    "last_7_days",
			wantFrom: fixedTime.Add(-7 * 24 * time.Hour),
			wantTo:   fixedTime,
		},
		{
			name:     "last_hour",
			scope:    "last_hour",
			wantFrom: fixedTime.Add(-1 * time.Hour),
			wantTo:   fixedTime,
		},
		{
			name:     "unknown scope defaults to last_24_hours",
			scope:    "unknown",
			wantFrom: fixedTime.Add(-24 * time.Hour),
			wantTo:   fixedTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ts.Parse(tt.scope)
			if !result.From.Equal(tt.wantFrom) {
				t.Errorf("Parse(%q).From = %v, want %v", tt.scope, result.From, tt.wantFrom)
			}
			if !result.To.Equal(tt.wantTo) {
				t.Errorf("Parse(%q).To = %v, want %v", tt.scope, result.To, tt.wantTo)
			}
		})
	}
}

func TestTimeScope_ParseCaseInsensitive(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	ts := timescope.New(
		timescope.WithClock(func() time.Time { return fixedTime }),
	)

	scopes := []string{"TODAY", "Today", "today", "ToDay"}
	for _, scope := range scopes {
		result := ts.Parse(scope)
		if result == nil {
			t.Errorf("Parse(%q) returned nil", scope)
		}
	}
}
