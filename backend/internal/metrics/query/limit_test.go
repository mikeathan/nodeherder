package query_test

import (
	"node-herder/internal/metrics/domain"
	"node-herder/internal/metrics/query"
	"testing"
)

func TestApplyLimitSortNumeric(t *testing.T) {
	result := &domain.ExposeNumericMetricsResult{
		Data: []*domain.NumericValue{
			{X: 1, Y: 1},
			{X: 2, Y: 2},
			{X: 3, Y: 3},
		},
	}

	query.ApplyLimitSortBy(result, 2, true)

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 points, got %d", len(result.Data))
	}
	if result.Data[0].X != 3 || result.Data[1].X != 2 {
		t.Fatalf("unexpected order after sort/limit: %+v", result.Data)
	}
}

func TestApplyLimitSortBinary(t *testing.T) {
	result := &domain.ExposeBinaryEventsResult{
		Data: []*domain.BinaryEvent{
			{Timestamp: 1, Value: true},
			{Timestamp: 3, Value: false},
			{Timestamp: 2, Value: true},
		},
	}

	query.ApplyLimitSortBy(result, 2, true)

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 points, got %d", len(result.Data))
	}
	if result.Data[0].Timestamp != 3 || result.Data[1].Timestamp != 2 {
		t.Fatalf("unexpected order after sort/limit: %+v", result.Data)
	}
}

func TestApplyLimitSortTimeRange(t *testing.T) {
	result := &domain.ExposeTimeRangeMetricsResult{
		Data: []*domain.TimeRangeValue{
			{Y: [2]int64{100, 110}},
			{Y: [2]int64{200, 210}},
			{Y: [2]int64{150, 160}},
		},
	}

	query.ApplyLimitSortBy(result, 2, true)

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 points, got %d", len(result.Data))
	}
	if result.Data[0].Y[0] != 200 || result.Data[1].Y[0] != 150 {
		t.Fatalf("unexpected order after sort/limit: %+v", result.Data)
	}
}
