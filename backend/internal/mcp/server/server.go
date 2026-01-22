package server

import (
	"context"
	"encoding/json"
	"fmt"
	"node-herder/internal/mcp/resolver"
	"node-herder/internal/mcp/resources"
	"node-herder/internal/mcp/tools"
	metrics "node-herder/internal/metrics/services"
	"node-herder/models/devices"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// DeviceStore provides access to devices for the MCP server.
type DeviceStore interface {
	AllDevices() ([]*devices.Device, error)
}

// deviceInfoAdapter adapts DeviceStore to resolver.DeviceStore
type deviceInfoAdapter struct {
	store DeviceStore
}

func (a *deviceInfoAdapter) AllDeviceInfo() ([]resolver.DeviceInfo, error) {
	devs, err := a.store.AllDevices()
	if err != nil {
		return nil, err
	}
	result := make([]resolver.DeviceInfo, len(devs))
	for i, d := range devs {
		result[i] = resolver.DeviceInfo{
			ID:   d.Id,
			Name: d.FriendlyName,
		}
	}
	return result, nil
}

// Server is the MCP server for nodeherder.
type Server struct {
	mcpServer       *server.MCPServer
	intentHandler   *tools.IntentHandler
	promptResource  *resources.PromptResource
	devicesResource *resources.DevicesResource
}

// New creates a new MCP server.
func New(store DeviceStore, querier *metrics.QueryService) *Server {
	s := &Server{
		intentHandler:   tools.NewIntentHandler(resolver.New(&deviceInfoAdapter{store}), querier),
		promptResource:  resources.NewPromptResource(),
		devicesResource: resources.NewDevicesResource(store),
	}

	// Create MCP server
	s.mcpServer = server.NewMCPServer(
		"nodeherder",
		"1.0.0",
		server.WithResourceCapabilities(true, false),
		server.WithToolCapabilities(true),
	)

	// Register tools
	s.registerTools()

	// Register resources
	s.registerResources()

	return s
}

func (s *Server) registerTools() {
	// query_device tool
	queryDeviceTool := mcp.NewTool("query_device",
		mcp.WithDescription("Query device metrics using natural language descriptions"),
		mcp.WithString("target_name",
			mcp.Required(),
			mcp.Description("Natural language name for the device")),
		mcp.WithArray("metrics",
			mcp.Required(),
			mcp.Description("Array of metric names to query")),
		mcp.WithString("time_scope",
			mcp.Description("Time range: today, yesterday, last_24_hours, last_7_days, last_hour")),
		mcp.WithString("aggregation",
			mcp.Description("Aggregation: latest_value, min_value, max_value, avg_value, count_events")),
	)

	s.mcpServer.AddTool(queryDeviceTool, s.handleDeclareIntent)
}

func (s *Server) registerResources() {
	// System prompt resource
	promptResource := mcp.NewResource(
		"nodeherder://system-prompt",
		"System Prompt",
		mcp.WithResourceDescription("Domain rules and guidance for LLM interactions"),
		mcp.WithMIMEType("text/plain"),
	)

	s.mcpServer.AddResource(promptResource, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		content, err := s.promptResource.GetContent()
		if err != nil {
			return nil, err
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "nodeherder://system-prompt",
				MIMEType: "text/plain",
				Text:     string(content),
			},
		}, nil
	})

	// Devices resource
	devicesResource := mcp.NewResource(
		"nodeherder://devices",
		"Device Context",
		mcp.WithResourceDescription("Available devices and their metrics"),
		mcp.WithMIMEType("application/json"),
	)

	s.mcpServer.AddResource(devicesResource, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		content, err := s.devicesResource.GetContent()
		if err != nil {
			return nil, err
		}
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "nodeherder://devices",
				MIMEType: "application/json",
				Text:     string(content),
			},
		}, nil
	})
}

func (s *Server) handleDeclareIntent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := json.Marshal(req.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal arguments: %v", err)), nil
	}

	resp := s.intentHandler.Handle(ctx, args)
	return formatToolResponse(resp), nil
}

func formatToolResponse(resp *tools.ToolResponse) *mcp.CallToolResult {
	data, _ := json.MarshalIndent(resp, "", "  ")
	if resp.Status == "error" {
		return mcp.NewToolResultError(string(data))
	}
	return mcp.NewToolResultText(string(data))
}

// ServeStdio starts the MCP server in stdio mode.
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.mcpServer)
}
