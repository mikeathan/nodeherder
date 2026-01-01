package query_test

import (
	"testing"

	"node-herder/internal/metrics/domain"
	"node-herder/internal/metrics/query"
	"node-herder/utils"
)

func TestMatchesFilters_Numeric(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		filters  []domain.MetricFilter
		expected bool
	}{
		{
			name:  "greater than - match",
			value: float64(10),
			filters: []domain.MetricFilter{
				{Op: domain.OpGreaterThan, Value: float64(5)},
			},
			expected: true,
		},
		{
			name:  "greater than - no match",
			value: float64(3),
			filters: []domain.MetricFilter{
				{Op: domain.OpGreaterThan, Value: float64(5)},
			},
			expected: false,
		},
		{
			name:  "equals",
			value: float64(7),
			filters: []domain.MetricFilter{
				{Op: domain.OpEquals, Value: float64(7)},
			},
			expected: true,
		},
		{
			name:  "not equals",
			value: float64(7),
			filters: []domain.MetricFilter{
				{Op: domain.OpNotEquals, Value: float64(8)},
			},
			expected: true,
		},
		{
			name:  "invalid filter value type",
			value: float64(7),
			filters: []domain.MetricFilter{
				{Op: domain.OpEquals, Value: "7"},
			},
			expected: false,
		},
		{
			name:     "invalid numeric value type",
			value:    "7",
			filters:  []domain.MetricFilter{{Op: domain.OpEquals, Value: float64(7)}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueBytes, _ := utils.AnyToByteArray(tt.value)
			result := query.MatchesFilters(valueBytes, tt.filters, "numeric")
			if result != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMatchesFilters_Binary(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		filters  []domain.MetricFilter
		expected bool
	}{
		{
			name:  "equals true",
			value: true,
			filters: []domain.MetricFilter{
				{Op: domain.OpEquals, Value: true},
			},
			expected: true,
		},
		{
			name:  "equals false - no match",
			value: true,
			filters: []domain.MetricFilter{
				{Op: domain.OpEquals, Value: false},
			},
			expected: false,
		},
		{
			name:  "not equals",
			value: false,
			filters: []domain.MetricFilter{
				{Op: domain.OpNotEquals, Value: true},
			},
			expected: true,
		},
		{
			name:  "invalid filter value type",
			value: true,
			filters: []domain.MetricFilter{
				{Op: domain.OpEquals, Value: "true"},
			},
			expected: false,
		},
		{
			name:     "invalid binary value type",
			value:    "true",
			filters:  []domain.MetricFilter{{Op: domain.OpEquals, Value: true}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueBytes, _ := utils.AnyToByteArray(tt.value)
			result := query.MatchesFilters(valueBytes, tt.filters, "binary")
			if result != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMatchesFilters_Enum(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		filters  []domain.MetricFilter
		expected bool
	}{
		{
			name:  "equals",
			value: "ON",
			filters: []domain.MetricFilter{
				{Op: domain.OpEquals, Value: "ON"},
			},
			expected: true,
		},
		{
			name:  "not equals",
			value: "OFF",
			filters: []domain.MetricFilter{
				{Op: domain.OpNotEquals, Value: "ON"},
			},
			expected: true,
		},
		{
			name:  "invalid filter value type",
			value: "ON",
			filters: []domain.MetricFilter{
				{Op: domain.OpEquals, Value: true},
			},
			expected: false,
		},
		{
			name:     "invalid enum value type",
			value:    true,
			filters:  []domain.MetricFilter{{Op: domain.OpEquals, Value: "ON"}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueBytes, _ := utils.AnyToByteArray(tt.value)
			result := query.MatchesFilters(valueBytes, tt.filters, "enum")
			if result != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMatchesFilters_NoFilters(t *testing.T) {
	valueBytes, _ := utils.AnyToByteArray(float64(42))
	result := query.MatchesFilters(valueBytes, nil, "numeric")
	if result != true {
		t.Fatalf("expected true when no filters are provided")
	}
}

func TestMatchesFilters_UnknownExposeType(t *testing.T) {
	valueBytes, _ := utils.AnyToByteArray(float64(10))
	result := query.MatchesFilters(valueBytes, []domain.MetricFilter{
		{Op: domain.OpEquals, Value: float64(10)},
	}, "unknown")

	if result != false {
		t.Fatalf("expected false for unknown expose type")
	}
}
