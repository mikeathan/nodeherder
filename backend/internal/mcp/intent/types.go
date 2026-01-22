package intent

// Intent represents an LLM's declared query intent.
type Intent struct {
	TargetName  string   `json:"target_name"`
	Metrics     []string `json:"metrics"`
	TimeScope   string   `json:"time_scope"`
	Aggregation string   `json:"aggregation"`
	Filters     []Filter `json:"filters,omitempty"`
}

// Filter represents a metric filter constraint.
type Filter struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value any    `json:"value"`
}

// DeviceContext provides device metadata for validation.
type DeviceContext struct {
	ID      string
	Name    string
	Metrics map[string]MetricInfo
}

// MetricInfo contains metadata about a device metric.
type MetricInfo struct {
	Name         string
	Type         string // numeric, binary, enum
	Aggregations []string
}
