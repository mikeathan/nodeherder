package utils_test

import (
	"node-herder/utils"
	"testing"
)

func TestCompareValuesCaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		val1     any
		val2     any
		expected bool
	}{
		{
			name:     "Both strings exactly matching",
			val1:     "TOGGLE",
			val2:     "TOGGLE",
			expected: true,
		},
		{
			name:     "Both strings mixed casing",
			val1:     "Toggle",
			val2:     "toggle",
			expected: true,
		},
		{
			name:     "Both strings entirely different casing",
			val1:     "TOGGLE",
			val2:     "toggle",
			expected: true,
		},
		{
			name:     "Different strings",
			val1:     "ON",
			val2:     "OFF",
			expected: false,
		},
		{
			name:     "Integer vs Float",
			val1:     1,
			val2:     1.0,
			expected: true, // fmt.Sprintf("%v", 1) is "1" and fmt.Sprintf("%v", 1.0) is also "1" without explicit float precision
		},
		{
			name:     "String vs integer exact match text",
			val1:     "123",
			val2:     123,
			expected: true, // "123" == "123"
		},
		{
			name:     "String vs boolean true match",
			val1:     "true",
			val2:     true,
			expected: true, // "true" == "true"
		},
		{
			name:     "String vs boolean false mismatch",
			val1:     "true",
			val2:     false,
			expected: false,
		},
		{
			name:     "Nil values",
			val1:     nil,
			val2:     nil,
			expected: true,
		},
		{
			name:     "One nil value",
			val1:     "nil",
			val2:     nil,
			expected: false, // "<nil>" != "nil" for fmt.Sprintf
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.CompareValuesCaseInsensitive(tt.val1, tt.val2)
			if result != tt.expected {
				t.Errorf("CompareValuesCaseInsensitive(%v, %v) = %v; want %v", tt.val1, tt.val2, result, tt.expected)
			}
		})
	}
}
