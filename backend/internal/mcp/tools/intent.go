package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"node-herder/internal/mcp/intent"
	"node-herder/internal/mcp/resolver"
	"node-herder/internal/metrics/query"
	metrics "node-herder/internal/metrics/services"
	"node-herder/models/devices"
)

// ToolResponse is the unified response type for all MCP tools.
type ToolResponse struct {
	Status     string      `json:"status"` // success, error, ambiguous
	Data       any         `json:"data,omitempty"`
	Hint       string      `json:"hint,omitempty"` // Hint for LLM when results are empty
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

// DeviceLookup provides access to device details for hints.
type DeviceLookup interface {
	FindDeviceByIds(ids []string) ([]*devices.Device, error)
}

// IntentHandler handles query_device tool calls.
type IntentHandler struct {
	resolver     *resolver.Resolver
	querier      *metrics.QueryService
	deviceLookup DeviceLookup
}

// NewIntentHandler creates a new IntentHandler.
func NewIntentHandler(resolver *resolver.Resolver, querier *metrics.QueryService, deviceLookup DeviceLookup) *IntentHandler {
	return &IntentHandler{
		resolver:     resolver,
		querier:      querier,
		deviceLookup: deviceLookup,
	}
}

// Handle processes a query_device tool call.
func (h *IntentHandler) Handle(ctx context.Context, args json.RawMessage) *ToolResponse {
	// Parse intent
	i, err := intent.Parse(args)
	if err != nil {
		return NewErrorResponse("parse_error", err.Error())
	}

	// Validate intent
	if err := intent.Validate(i, nil); err != nil {
		return NewErrorResponse("validation_error", err.Error())
	}

	// Resolve device
	deviceID, candidates, err := h.resolver.Resolve(i.TargetName)
	if err != nil {
		if _, ok := err.(*resolver.AmbiguousDeviceError); ok {
			return NewAmbiguousResponse(convertCandidates(candidates))
		}
		return NewErrorResponse("resolution_error", err.Error())
	}

	// Translate intent to query
	req := intent.Translate(i, deviceID, intent.DefaultTranslateOptions)

	// Execute query
	result, err := h.querier.Query(ctx, *req)
	if err != nil {
		return NewErrorResponse("query_error", err.Error())
	}

	resp := NewSuccessResponse(result)

	// Add hint if results are empty
	if h.isResultEmpty(result) && h.deviceLookup != nil {
		resp.Hint = h.buildMetricsHint(deviceID)
	}

	return resp
}

// isResultEmpty checks if query results have no data.
func (h *IntentHandler) isResultEmpty(result *[]query.MetricsQueryResponse) bool {
	if result == nil || len(*result) == 0 {
		return true
	}
	for _, r := range *result {
		if len(r.Values) > 0 {
			return false
		}
	}
	return true
}

// buildMetricsHint creates a hint showing available device metrics.
func (h *IntentHandler) buildMetricsHint(deviceID string) string {
	devices, err := h.deviceLookup.FindDeviceByIds([]string{deviceID})
	if err != nil || len(devices) == 0 {
		return ""
	}

	var metrics []string
	for name := range devices[0].Exposes {
		metrics = append(metrics, name)
	}

	if len(metrics) == 0 {
		return ""
	}

	return fmt.Sprintf("Available metrics for this device: %v", metrics)
}

func convertCandidates(candidates []resolver.Candidate) []Candidate {
	result := make([]Candidate, len(candidates))
	for i, c := range candidates {
		result[i] = Candidate{
			DeviceID: c.DeviceID,
			Name:     c.Name,
			Score:    c.Score,
		}
	}
	return result
}
