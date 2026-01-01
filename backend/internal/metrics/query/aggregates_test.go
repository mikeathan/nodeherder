package query_test

import (
	"testing"

	"node-herder/internal/metrics/domain"
	"node-herder/internal/metrics/query"
)

func TestAggregateExposeValue_Numeric(t *testing.T) {
	expose := &domain.ExposeNumericMetricsResult{
		Name: "temperature",
		Data: []*domain.NumericValue{
			{X: 1, Y: 10},
			{X: 2, Y: 20},
			{X: 3, Y: 5},
		},
	}

	tests := []struct {
		name      string
		agg       domain.AggregationType
		wantValue any
		wantTs    int64
		wantOk    bool
	}{
		{"last", domain.AggLast, float64(5), int64(3), true},
		{"min", domain.AggMin, float64(5), int64(3), true},
		{"max", domain.AggMax, float64(20), int64(2), true},
		{"avg", domain.AggAvg, float64(35.0 / 3.0), 0, true},
		{"count", domain.AggCount, 3, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ts, ok, err := query.AggregateExposeValue(tt.agg, expose)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOk)
			}
			if val != tt.wantValue {
				t.Fatalf("value = %v, want %v", val, tt.wantValue)
			}
			if ts != tt.wantTs {
				t.Fatalf("ts = %d, want %d", ts, tt.wantTs)
			}
		})
	}
}

func TestAggregateExposeValue_Binary(t *testing.T) {
	expose := &domain.ExposeBinaryEventsResult{
		Name: "door",
		Data: []*domain.BinaryEvent{
			{Timestamp: 10, Value: false},
			{Timestamp: 20, Value: true},
		},
	}

	val, ts, ok, err := query.AggregateExposeValue(domain.AggLast, expose)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok=true")
	}
	if val != true {
		t.Fatalf("value = %v, want true", val)
	}
	if ts != int64(20) {
		t.Fatalf("ts = %d, want 20", ts)
	}
}

func TestAggregateExposeValue_TimeRange(t *testing.T) {
	expose := &domain.ExposeTimeRangeMetricsResult{
		Name: "mode",
		Data: []*domain.TimeRangeValue{
			{X: "auto", Y: [2]int64{10, 0}},
			{X: "manual", Y: [2]int64{20, 30}},
		},
	}

	val, ts, ok, err := query.AggregateExposeValue(domain.AggLast, expose)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok=true")
	}
	if val != "manual" {
		t.Fatalf("value = %v, want manual", val)
	}
	if ts != int64(30) {
		t.Fatalf("ts = %d, want 30", ts)
	}
}

func TestAggregateExposeValue_EmptyData(t *testing.T) {
	expose := &domain.ExposeNumericMetricsResult{
		Name: "empty",
		Data: []*domain.NumericValue{},
	}

	_, _, ok, err := query.AggregateExposeValue(domain.AggLast, expose)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for empty data")
	}
}

func TestAggregateExposeValue_UnsupportedAggregation(t *testing.T) {
	expose := &domain.ExposeBinaryEventsResult{
		Name: "door",
		Data: []*domain.BinaryEvent{
			{Timestamp: 1, Value: true},
		},
	}

	_, _, _, err := query.AggregateExposeValue(domain.AggAvg, expose)
	if err == nil {
		t.Fatal("expected error for unsupported aggregation")
	}
}
