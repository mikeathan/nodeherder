package timescope

import (
	"strings"
	"time"
)

// TimeRange represents a resolved time window for queries.
type TimeRange struct {
	From time.Time
	To   time.Time
}

// TimeScope handles timezone-aware time parsing.
type TimeScope struct {
	tz    *time.Location
	clock func() time.Time // injectable for testing
}

// Option configures a TimeScope.
type Option func(*TimeScope)

// WithTimezone sets the timezone for time calculations.
func WithTimezone(tz *time.Location) Option {
	return func(ts *TimeScope) {
		ts.tz = tz
	}
}

// WithClock sets a custom clock function for testing.
func WithClock(clock func() time.Time) Option {
	return func(ts *TimeScope) {
		ts.clock = clock
	}
}

// New creates a new TimeScope with the given options.
func New(opts ...Option) *TimeScope {
	ts := &TimeScope{
		tz:    time.UTC,
		clock: time.Now,
	}
	for _, opt := range opts {
		opt(ts)
	}
	return ts
}

// Parse converts a scope string to a TimeRange.
// Supported scopes: today, yesterday, last_24_hours, last_7_days, last_hour, last_30_days
func (ts *TimeScope) Parse(scope string) *TimeRange {
	now := ts.clock().In(ts.tz)

	switch strings.ToLower(scope) {
	case "today":
		return ts.today(now)
	case "yesterday":
		return ts.yesterday(now)
	case "last_24_hours", "last24hours":
		return ts.last24Hours(now)
	case "last_7_days", "last7days":
		return ts.last7Days(now)
	case "last_hour", "lasthour":
		return ts.lastHour(now)
	case "last_30_days", "last30days":
		return ts.last30Days(now)
	default:
		// Default to last 24 hours for unknown scopes
		return ts.last24Hours(now)
	}
}

func (ts *TimeScope) today(now time.Time) *TimeRange {
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, ts.tz)
	return &TimeRange{
		From: start,
		To:   now,
	}
}

func (ts *TimeScope) yesterday(now time.Time) *TimeRange {
	start := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, ts.tz)
	end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, ts.tz)
	return &TimeRange{
		From: start,
		To:   end,
	}
}

func (ts *TimeScope) last24Hours(now time.Time) *TimeRange {
	return &TimeRange{
		From: now.Add(-24 * time.Hour),
		To:   now,
	}
}

func (ts *TimeScope) last7Days(now time.Time) *TimeRange {
	return &TimeRange{
		From: now.Add(-7 * 24 * time.Hour),
		To:   now,
	}
}

func (ts *TimeScope) lastHour(now time.Time) *TimeRange {
	return &TimeRange{
		From: now.Add(-1 * time.Hour),
		To:   now,
	}
}

func (ts *TimeScope) last30Days(now time.Time) *TimeRange {
	return &TimeRange{
		From: now.Add(-30 * 24 * time.Hour),
		To:   now,
	}
}
