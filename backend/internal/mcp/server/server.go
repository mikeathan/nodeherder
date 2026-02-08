package server

import (
	"context"
	"encoding/json"
	"fmt"
	"node-herder/internal/mcp/protocol"
	"node-herder/internal/mcp/resolver"
	"node-herder/internal/mcp/resources"
	"node-herder/internal/mcp/tools"
	metrics "node-herder/internal/metrics/services"
	"node-herder/models/hub"
	"node-herder/store"
	"sync"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	ServerName    = "nodeherder"
	ServerVersion = "1.0.0"

	SystemPromptResourceURI = "nodeherder://system-prompt"

	QueryDeviceToolName = "query_device"

	MethodResourcesSubscribe   = "resources/subscribe"
	MethodResourcesUnsubscribe = "resources/unsubscribe"
)

type DeviceStore interface {
	LoadHubState() (*hub.HubState, error)
	RegisterIsDirtyCallback(cb store.AppStoreDirtyFlagCallback)
}

type Server struct {
	mcpServer      *server.MCPServer
	intentHandler  *tools.IntentHandler
	promptResource *resources.PromptResource

	// Notification listeners for HTTP/SSE transport
	listeners   map[string]func(string)
	listenersMu sync.RWMutex
}

func New(store DeviceStore, querier *metrics.QueryService) *Server {
	s := &Server{
		intentHandler:  tools.NewIntentHandler(resolver.New(&deviceInfoAdapter{store}), querier, &deviceLookupAdapter{store}),
		promptResource: resources.NewPromptResource(store),
		listeners:      make(map[string]func(string)),
	}

	// Register callback for updates - notify all listeners
	store.RegisterIsDirtyCallback(func() {
		s.notifyListeners(SystemPromptResourceURI)
	})

	s.mcpServer = server.NewMCPServer(
		ServerName,
		ServerVersion,
		server.WithResourceCapabilities(true, true),
		server.WithToolCapabilities(true),
	)

	s.registerTools()
	s.registerResources()

	return s
}

func (s *Server) notifyListeners(uri string) {
	s.listenersMu.RLock()
	defer s.listenersMu.RUnlock()

	for _, listener := range s.listeners {
		// invoke listener async to avoid blocking
		go listener(uri)
	}
}

// Returns a unique ID for unregistration.
func (s *Server) RegisterNotificationListener(cb func(string)) string {
	s.listenersMu.Lock()
	defer s.listenersMu.Unlock()

	id := uuid.New().String()
	s.listeners[id] = cb
	return id
}

func (s *Server) UnregisterNotificationListener(id string) {
	s.listenersMu.Lock()
	defer s.listenersMu.Unlock()

	delete(s.listeners, id)
}

// This is used by the HTTP transport.
func (s *Server) HandleRequest(ctx context.Context, rawMessage json.RawMessage) json.RawMessage {
	// Peek at the method to intercept subscribe/unsubscribe
	var baseMessage struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Method  string          `json:"method"`
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
	queryDeviceTool := mcp.NewTool(QueryDeviceToolName,
		mcp.WithDescription(fmt.Sprintf("Query device metrics. Read %s first to get exact metric names.", SystemPromptResourceURI)),
		mcp.WithString("target_name",
			mcp.Required(),
			mcp.Description("Natural language name for the device")),
		mcp.WithArray("metrics",
			mcp.Required(),
			mcp.Description(fmt.Sprintf("Metric names from %s resource", SystemPromptResourceURI))),
		mcp.WithString("time_scope",
			mcp.Description("Time range: today, yesterday, last_24_hours, last_7_days, last_hour")),
		mcp.WithString("aggregation",
			mcp.Description("Use 'last' for current value. Options: last, min, max, avg, count")),
	)

	s.mcpServer.AddTool(queryDeviceTool, s.handleDeclareIntent)
}

func (s *Server) registerResources() {
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

}

func (s *Server) handleDeclareIntent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, err := json.Marshal(req.Params.Arguments)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal arguments: %v", err)), nil
	}

	resp := s.intentHandler.Handle(ctx, args)
	return formatToolResponse(resp), nil
}

func formatToolResponse(resp *protocol.ToolResponse) *mcp.CallToolResult {
	data, _ := json.MarshalIndent(resp, "", "  ")
	if resp.Status == "error" {
		return mcp.NewToolResultError(string(data))
	}
	return mcp.NewToolResultText(string(data))
}
