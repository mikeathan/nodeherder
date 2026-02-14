package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"node-herder/internal/mcp/intent"
	"node-herder/internal/mcp/protocol"
	"node-herder/internal/mcp/resolver"
	"node-herder/internal/metrics/query"
	metrics "node-herder/internal/metrics/services"
	"node-herder/models/devices"
)

type DeviceLookup interface {
	FindDeviceByIds(ids []string) ([]*devices.Device, error)
}

type IntentHandler struct {
	resolver     *resolver.Resolver
	querier      *metrics.QueryService
	deviceLookup DeviceLookup
}

func NewIntentHandler(resolver *resolver.Resolver, querier *metrics.QueryService, deviceLookup DeviceLookup) *IntentHandler {
	return &IntentHandler{
		resolver:     resolver,
		querier:      querier,
		deviceLookup: deviceLookup,
	}
}

func (h *IntentHandler) Handle(ctx context.Context, args json.RawMessage) *protocol.ToolResponse {
	// Parse intent
	i, err := intent.Parse(args)
	if err != nil {
		return protocol.NewErrorResponse("parse_error", err.Error())
	}

	// Validate intent
	if err := intent.Validate(i, nil); err != nil {
		return protocol.NewErrorResponse("validation_error", err.Error())
	}

	// Resolve device
	deviceID, candidates, err := h.resolver.Resolve(i.TargetName)
	if err != nil {
		if _, ok := err.(*resolver.AmbiguousDeviceError); ok {
			return protocol.NewAmbiguousResponse(convertCandidates(candidates))
		}
		return protocol.NewErrorResponse("resolution_error", err.Error())
	}

	// Translate intent to query
	req := intent.Translate(i, deviceID, intent.DefaultTranslateOptions)

	// Execute query
	result, err := h.querier.Query(ctx, *req)
	if err != nil {
		return protocol.NewErrorResponse("query_error", err.Error())
	}

	resp := protocol.NewSuccessResponse(result)

	// Add hint if results are empty
	if h.isResultEmpty(result) && h.deviceLookup != nil {
		resp.Hint = h.buildMetricsHint(deviceID)
	}

	return resp
}

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

func convertCandidates(candidates []resolver.Candidate) []protocol.Candidate {
	result := make([]protocol.Candidate, len(candidates))
	for i, c := range candidates {
		result[i] = protocol.Candidate{
			DeviceID: c.DeviceID,
			Name:     c.Name,
			Score:    c.Score,
		}
	}
	return result
}
