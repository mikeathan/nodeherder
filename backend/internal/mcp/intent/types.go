package intent

import "time"

type Intent struct {
	TargetName       string   `json:"target_name"`
	Metrics          []string `json:"metrics"`
	TimeScope        string   `json:"time_scope"`
	Aggregation      string   `json:"aggregation"`
	AggregationValue any      `json:"aggregation_value,omitempty"`
	Filters          []Filter `json:"filters,omitempty"`
}

type Filter struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value any    `json:"value"`
}

type DeviceContext struct {
	ID      string
	Name    string
	Metrics map[string]MetricInfo
}

type MetricInfo struct {
	Name         string
	Type         string // numeric, binary, enum
	Aggregations []string
}

type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}
