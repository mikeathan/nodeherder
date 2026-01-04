package query

import (
	"node-herder/internal/metrics/domain"
	"time"
)

type MetricsQueryRequest struct {
	DeviceIds []string
	Expose    string

	Time domain.TimeQuery

	Aggregation domain.AggregationType

	Filters []domain.MetricFilter

	Limit    int
	SortDesc bool
}

func ResolveTime(q domain.TimeQuery, now time.Time) (time.Time, time.Time) {
	// If user provided nothing at all, we choose a safe global window
	if q.Lookback == "" && q.From.IsZero() && q.To.IsZero() {
		// Safe default: last 30 days
		return now.Add(-30 * 24 * time.Hour), now
	}

	// If lookback provided, it overrides from/to
	if d, ok := parseLookback(q.Lookback); ok {
		return now.Add(-d), now
	}

	// Otherwise trust explicit from/to
	return q.From, q.To
}

func parseLookback(value string) (time.Duration, bool) {
	if value == "" {
		return 0, false
	}

	last := value[len(value)-1]
	if last == 'd' {
		days, err := time.ParseDuration(value[:len(value)-1] + "h")
		if err != nil {
			return 0, false
		}
		if days <= 0 {
			return 0, false
		}
		return days * 24, true
	}

	if last != 'm' && last != 'h' {
		return 0, false
	}

	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		return 0, false
	}

	return d, true
}
