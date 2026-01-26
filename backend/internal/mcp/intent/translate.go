package intent

import (
	"node-herder/internal/mcp/timescope"
	"node-herder/internal/metrics/domain"
	"node-herder/internal/metrics/query"
	"time"
)

// TranslateOptions configures intent translation.
type TranslateOptions struct {
	Timezone *time.Location
	Clock    func() time.Time
}

// DefaultTranslateOptions provides sensible defaults.
var DefaultTranslateOptions = TranslateOptions{
	Timezone: time.UTC,
	Clock:    time.Now,
}

// Translate converts an Intent to a MetricsQueryRequest.
func Translate(intent *Intent, deviceID string, opts TranslateOptions) *query.MetricsQueryRequest {
	ts := timescope.New(
		timescope.WithTimezone(opts.Timezone),
		timescope.WithClock(opts.Clock),
	)

	timeRange := ts.Parse(intent.TimeScope)

	req := &query.MetricsQueryRequest{
		DeviceIds: []string{deviceID},
		Exposes:   intent.Metrics, // All requested metrics
		Time: domain.TimeQuery{
			From: timeRange.From,
			To:   timeRange.To,
		},
		Aggregation: mapAggregation(intent.Aggregation),
		Filters:     translateFilters(intent.Filters),
	}

	return req
}

// AggregationMap maps intent aggregation strings to domain aggregation types.
var AggregationMap = map[string]domain.AggregationType{
	"count_events": domain.AggCount,
	"latest_value": domain.AggLast,
	"last_event":   domain.LastEvent,
	"min_value":    domain.AggMin,
	"max_value":    domain.AggMax,
	"avg_value":    domain.AggAvg,
}

func mapAggregation(agg string) domain.AggregationType {
	if mapped, ok := AggregationMap[agg]; ok {
		return mapped
	}
	return domain.AggNone
}

func translateFilters(filters []Filter) []domain.MetricFilter {
	result := make([]domain.MetricFilter, len(filters))
	for i, f := range filters {
		result[i] = domain.MetricFilter{
			Field: mapField(f.Field),
			Op:    mapOp(f.Op),
			Value: f.Value,
		}
	}
	return result
}

func mapField(field string) domain.MetricField {
	switch field {
	case "value":
		return domain.FieldValue
	case "state":
		return domain.FieldState
	case "timestamp":
		return domain.FieldTimestamp
	default:
		return domain.FieldValue
	}
}

func mapOp(op string) domain.OperationType {
	switch op {
	case "=", "==":
		return domain.OpEquals
	case "!=":
		return domain.OpNotEquals
	case ">":
		return domain.OpGreaterThan
	case "<":
		return domain.OpLessThan
	default:
		return domain.OpEquals
	}
}
