package mcp

import "time"

// ToolResponse is the unified response type for all MCP tools.
type ToolResponse struct {
	Status     string      `json:"status"` // success, error, ambiguous
	Data       any         `json:"data,omitempty"`
	Error      *ToolError  `json:"error,omitempty"`
	Candidates []Candidate `json:"candidates,omitempty"`
}

// ToolError represents an error returned by an MCP tool.
type ToolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Candidate represents a device candidate when resolution is ambiguous.
type Candidate struct {
	DeviceID string  `json:"device_id"`
	Name     string  `json:"name"`
	Score    float64 `json:"score"`
}

// TimeRange represents a resolved time window for queries.
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

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

// AggregationMap maps intent aggregation strings to domain aggregation types.
var AggregationMap = map[string]string{
	"count_events": "count",
	"latest_value": "last",
	"last_event":   "last_event",
	"min_value":    "min",
	"max_value":    "max",
	"avg_value":    "avg",
}

// NewSuccessResponse creates a successful tool response.
func NewSuccessResponse(data any) *ToolResponse {
	return &ToolResponse{
		Status: "success",
		Data:   data,
	}
}

// NewErrorResponse creates an error tool response.
func NewErrorResponse(code, message string) *ToolResponse {
	return &ToolResponse{
		Status: "error",
		Error: &ToolError{
			Code:    code,
			Message: message,
		},
	}
}

// NewAmbiguousResponse creates an ambiguous tool response with candidates.
func NewAmbiguousResponse(candidates []Candidate) *ToolResponse {
	return &ToolResponse{
		Status:     "ambiguous",
		Candidates: candidates,
	}
}
