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
	"node-herder/models/settings"
	"node-herder/store"
	"node-herder/utils"
	"sync"
	"sync/atomic"

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
	AppConfig() *settings.AppConfigCache
}

// MCPStatusProvider abstracts MCP server status queries.
type MCPStatusProvider interface {
	Status() MCPStatusInfo
	Stop() error
	Start() error
	Restart() error
	Running() bool
	SetOnStatusChange(cb func())
}

// MCPStatusInfo represents the MCP server status for the settings UI.
type MCPStatusInfo struct {
	Running          bool   `json:"running"`
	Enabled          bool   `json:"enabled"`
	Name             string `json:"name"`
	Version          string `json:"version"`
	ConnectedClients int    `json:"connectedClients"`
}

type Server struct {
	mcpServer      *server.MCPServer
	intentHandler  *tools.IntentHandler
	promptResource *resources.PromptResource
	store          DeviceStore

	// Notification listeners for HTTP/SSE transport
	listeners   map[string]func(string)
	listenersMu sync.RWMutex

	// Atomic running state: 1 = running, 0 = stopped
	running          int32
	connectedClients int
	clientCountMu    sync.RWMutex
	onStatusChange   func()
	configCache      *settings.AppConfigCache
}

func New(store DeviceStore, configCache *settings.AppConfigCache, querier *metrics.QueryService) *Server {
	s := &Server{
		intentHandler:  tools.NewIntentHandler(resolver.New(&deviceInfoAdapter{store}), querier, &deviceLookupAdapter{store}),
		promptResource: resources.NewPromptResource(store),
		listeners:      make(map[string]func(string)),
		store:          store,
		configCache:    configCache,
	}

	// Server starts in running state
	atomic.StoreInt32(&s.running, 1)

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

func (s *Server) OnClientConnectionChange(count int) {
	s.clientCountMu.Lock()
	s.connectedClients = count
	cb := s.onStatusChange
	s.clientCountMu.Unlock()

	if cb != nil {
		cb()
	}
}

func (s *Server) SetOnStatusChange(cb func()) {
	s.clientCountMu.Lock()
	defer s.clientCountMu.Unlock()
	s.onStatusChange = cb
}

func (s *Server) Stop() error {
	atomic.StoreInt32(&s.running, 0)
	return nil
}

func (s *Server) Start() error {
	atomic.StoreInt32(&s.running, 1)
	return nil
}

func (s *Server) Restart() error {
	if err := s.Stop(); err != nil {
		return err
	}
	return s.Start()
}

func (s *Server) Initialize() error {
	config, err := s.store.AppConfig().LoadAppConfig()
	if err != nil {
		return err
	}

	if config.Hub.MCP.Enabled {
		utils.LogInfo("MCP Server starting")
		return s.Start()
	}
	utils.LogInfo("MCP Server stopping")
	return s.Stop()
}

func (s *Server) Running() bool {
	return atomic.LoadInt32(&s.running) == 1
}

func (s *Server) Status() MCPStatusInfo {
	s.clientCountMu.RLock()
	clients := s.connectedClients
	s.clientCountMu.RUnlock()

	enabled := false
	if config, err := s.configCache.LoadAppConfig(); err == nil && config != nil {
		enabled = config.Hub.MCP.Enabled
	}

	return MCPStatusInfo{
		Running:          s.Running(),
		Enabled:          enabled,
		Name:             ServerName,
		Version:          ServerVersion,
		ConnectedClients: clients,
	}
}

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
