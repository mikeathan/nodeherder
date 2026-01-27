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
	"node-herder/models/hub"
	"sync"

	"node-herder/store"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	ServerName    = "nodeherder"
	ServerVersion = "1.0.0"

	DevicesResourceURI      = "nodeherder://devices"
	SystemPromptResourceURI = "nodeherder://system-prompt"

	QueryDeviceToolName = "query_device"

	MethodResourcesSubscribe   = "resources/subscribe"
	MethodResourcesUnsubscribe = "resources/unsubscribe"
)

// DeviceStore provides access to devices for the MCP server.
type DeviceStore interface {
	LoadHubState() (*hub.HubState, error)
	RegisterIsDirtyCallback(cb store.AppStoreDirtyFlagCallback)
}

// deviceInfoAdapter adapts DeviceStore to resolver.DeviceStore
type deviceInfoAdapter struct {
	store DeviceStore
}

func (a *deviceInfoAdapter) AllDeviceInfo() ([]resolver.DeviceInfo, error) {
	state, err := a.store.LoadHubState()
	if err != nil {
		return nil, err
	}
	result := make([]resolver.DeviceInfo, len(state.Devices))
	for i, d := range state.Devices {
		result[i] = resolver.DeviceInfo{
			ID:   d.Id,
			Name: d.FriendlyName,
		}
	}
	return result, nil
}

// deviceLookupAdapter adapts DeviceStore to tools.DeviceLookup
type deviceLookupAdapter struct {
	store DeviceStore
}

func (a *deviceLookupAdapter) FindDeviceByIds(ids []string) ([]*devices.Device, error) {
	state, err := a.store.LoadHubState()
	if err != nil {
		return nil, err
	}

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	var result []*devices.Device
	for _, d := range state.Devices {
		if idSet[d.Id] {
			result = append(result, d)
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

	// Notification listeners for HTTP/SSE transport
	listeners   []func(string)
	listenersMu sync.RWMutex
}

// New creates a new MCP server.
func New(store DeviceStore, querier *metrics.QueryService) *Server {
	s := &Server{
		intentHandler:   tools.NewIntentHandler(resolver.New(&deviceInfoAdapter{store}), querier, &deviceLookupAdapter{store}),
		promptResource:  resources.NewPromptResource(),
		devicesResource: resources.NewDevicesResource(store),
		listeners:       []func(string){},
	}

	// Register callback for updates - notify all listeners
	store.RegisterIsDirtyCallback(func() {
		s.notifyListeners(DevicesResourceURI)
	})

	// Create MCP server
	s.mcpServer = server.NewMCPServer(
		ServerName,
		ServerVersion,
		server.WithResourceCapabilities(true, true),
		server.WithToolCapabilities(true),
	)

	// Register tools
	s.registerTools()

	// Register resources
	s.registerResources()

	return s
}

// notifyListeners calls all registered notification listeners.
func (s *Server) notifyListeners(uri string) {
	s.listenersMu.RLock()
	defer s.listenersMu.RUnlock()

	for _, listener := range s.listeners {
		listener(uri)
	}
}

// RegisterNotificationListener adds a callback for resource notifications.
func (s *Server) RegisterNotificationListener(cb func(string)) {
	s.listenersMu.Lock()
	defer s.listenersMu.Unlock()
	s.listeners = append(s.listeners, cb)
}

// UnregisterNotificationListener removes a previously registered callback.
// Note: This uses function pointer comparison which may not work for closures.
// For production use, consider using a unique ID-based approach.
func (s *Server) UnregisterNotificationListener(cb func(string)) {
	s.listenersMu.Lock()
	defer s.listenersMu.Unlock()

	for i, listener := range s.listeners {
		if &listener == &cb {
			s.listeners = append(s.listeners[:i], s.listeners[i+1:]...)
			return
		}
	}
}

// HandleRequest processes a JSON-RPC request and returns the response.
// This is used by the HTTP transport.
func (s *Server) HandleRequest(ctx context.Context, rawMessage json.RawMessage) json.RawMessage {
	// Peek at the method to intercept subscribe/unsubscribe
	var baseMessage struct {
		JSONRPC string      `json:"jsonrpc"`
		ID      interface{} `json:"id"`
		Method  string      `json:"method"`
	}
	if err := json.Unmarshal(rawMessage, &baseMessage); err != nil {
		return s.errorResponse(nil, -32700, "Parse error")
	}

	// Intercept subscribe/unsubscribe - return success
	// The actual subscription is handled by the SSE connection
	if baseMessage.Method == MethodResourcesSubscribe || baseMessage.Method == MethodResourcesUnsubscribe {
		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      baseMessage.ID,
			"result":  map[string]interface{}{},
		}
		data, _ := json.Marshal(response)
		return data
	}

	// Delegate to library for everything else
	resp := s.mcpServer.HandleMessage(ctx, rawMessage)
	if resp == nil {
		return nil
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return s.errorResponse(baseMessage.ID, -32603, "Internal error")
	}
	return data
}

// errorResponse creates a JSON-RPC error response.
func (s *Server) errorResponse(id interface{}, code int, message string) json.RawMessage {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	data, _ := json.Marshal(response)
	return data
}

func (s *Server) registerTools() {
	// query_device tool
	queryDeviceTool := mcp.NewTool(QueryDeviceToolName,
		mcp.WithDescription(fmt.Sprintf("Query device metrics. Read %s first to get exact metric names.", DevicesResourceURI)),
		mcp.WithString("target_name",
			mcp.Required(),
			mcp.Description("Natural language name for the device")),
		mcp.WithArray("metrics",
			mcp.Required(),
			mcp.Description(fmt.Sprintf("Metric names from %s resource", DevicesResourceURI))),
		mcp.WithString("time_scope",
			mcp.Description("Time range: today, yesterday, last_24_hours, last_7_days, last_hour")),
		mcp.WithString("aggregation",
			mcp.Description("Aggregation: last, min, max, avg, count, last_event")),
	)

	s.mcpServer.AddTool(queryDeviceTool, s.handleDeclareIntent)
}

func (s *Server) registerResources() {
	// System prompt resource
	promptResource := mcp.NewResource(
		SystemPromptResourceURI,
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
				URI:      SystemPromptResourceURI,
				MIMEType: "text/plain",
				Text:     string(content),
			},
		}, nil
	})

	// Devices resource
	devicesResource := mcp.NewResource(
		DevicesResourceURI,
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
				URI:      DevicesResourceURI,
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
