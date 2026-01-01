package query

import (
	"fmt"
	"node-herder/internal/metrics/domain"
)

func AggregateExposeValue(aggregator domain.AggregationType, expose domain.ExposeResult) (any, int64, bool, error) {
	switch r := expose.(type) {
	case *domain.ExposeNumericMetricsResult:
		return aggregateNumericValues(aggregator, r)

	case *domain.ExposeBinaryEventsResult:
		return aggregateBinaryValues(aggregator, r)

	case *domain.ExposeTimeRangeMetricsResult:
		return aggregateTimeRangeValues(aggregator, r)

	default:
		return nil, 0, false, unsupportedExposeAgg(aggregator, "unknown")
	}
}

func aggregateNumericValues(aggregator domain.AggregationType, result *domain.ExposeNumericMetricsResult) (any, int64, bool, error) {

	if len(result.Data) == 0 {
		return nil, 0, false, nil
	}

	switch aggregator {
	case domain.AggNone:
		return result.Data, 0, true, nil

	case domain.AggLast:
		last := result.Data[len(result.Data)-1]
		return float64(last.Y), last.X, true, nil

	case domain.AggMin:
		return numericMin(result.Data)

	case domain.AggMax:
		return numericMax(result.Data)

	case domain.AggAvg:
		return numericAvg(result.Data)

	case domain.AggCount:
		return len(result.Data), 0, true, nil

	default:
		return nil, 0, false, unsupportedExposeAgg(aggregator, result.Name)
	}
}

func aggregateBinaryValues(aggregator domain.AggregationType, result *domain.ExposeBinaryEventsResult) (any, int64, bool, error) {

	count := len(result.Data)
	if count == 0 {
		return nil, 0, false, nil
	}

	switch aggregator {
	case domain.AggNone:
		return result.Data, 0, true, nil

	case domain.AggLast:

		last := result.Data[count-1]
		return last.Value, last.Timestamp, true, nil

	case domain.AggCount:
		return count, 0, true, nil
	default:
		return nil, 0, false, unsupportedExposeAgg(aggregator, result.Name)
	}
}

func aggregateTimeRangeValues(aggregator domain.AggregationType, result *domain.ExposeTimeRangeMetricsResult) (any, int64, bool, error) {

	count := len(result.Data)
	if count == 0 {
		return nil, 0, false, nil
	}

	switch aggregator {
	case domain.AggNone:
		return result.Data, 0, true, nil

	case domain.AggLast:
		last := result.Data[count-1]
		timestamp := last.Y[1]
		if timestamp == 0 {
			timestamp = last.Y[0]
		}
		return last.X, timestamp, true, nil

	case domain.AggCount:
		return count, 0, true, nil
	default:
		return nil, 0, false, unsupportedExposeAgg(aggregator, result.Name)
	}
}

func unsupportedExposeAgg(agg domain.AggregationType, expose string) error {
	return fmt.Errorf("aggregation %q unsupported for %s expose", agg, expose)
}

func numericMin(data []*domain.NumericValue) (any, int64, bool, error) {
	min := data[0]
	for _, p := range data[1:] {
		if p.Y < min.Y {
			min = p
		}
	}
	return float64(min.Y), min.X, true, nil
}

func numericMax(data []*domain.NumericValue) (any, int64, bool, error) {
	max := data[0]
	for _, p := range data[1:] {
		if p.Y > max.Y {
			max = p
		}
	}
	return float64(max.Y), max.X, true, nil
}

func numericAvg(data []*domain.NumericValue) (any, int64, bool, error) {
	var sum float64
	for _, p := range data {
		sum += float64(p.Y)
	}
	return sum / float64(len(data)), 0, true, nil
}
