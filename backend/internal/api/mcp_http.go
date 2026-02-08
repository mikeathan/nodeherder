package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"node-herder/utils"
)

// MCPServer defines the interface needed for HTTP transport.
type MCPServer interface {
	HandleRequest(ctx context.Context, rawMessage json.RawMessage) json.RawMessage
	RegisterNotificationListener(cb func(string))
	UnregisterNotificationListener(cb func(string))
}

// MCPHandler handles JSON-RPC requests over HTTP.
type MCPHandler struct {
	server MCPServer
	events *MCPEventsHandler
}

// NewMCPHandler creates a new MCPHandler.
func NewMCPHandler(server MCPServer, events *MCPEventsHandler) *MCPHandler {
	return &MCPHandler{server: server, events: events}
}

func (h *MCPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		utils.LogErrorf("MCPHandler: Invalid content type: %s", r.Header.Get("Content-Type"))
		writeJSONError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	var rawMessage json.RawMessage
	if err := json.Unmarshal(body, &rawMessage); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	response := h.server.HandleRequest(r.Context(), rawMessage)

	w.Header().Set("Content-Type", "application/json")
	if response != nil {
		// Dual-send: Broadcast via SSE if available, to support clients that expect it
		if h.events != nil {
			h.events.Send(response)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

// MCPEventsHandler handles SSE connections for MCP notifications.
type MCPEventsHandler struct {
	server  MCPServer
	clients map[chan string]struct{}
	mu      sync.RWMutex
}

// NewMCPEventsHandler creates a new MCPEventsHandler.
func NewMCPEventsHandler(server MCPServer) *MCPEventsHandler {
	h := &MCPEventsHandler{
		server:  server,
		clients: make(map[chan string]struct{}),
	}

	// Register for notifications from the MCP server
	server.RegisterNotificationListener(h.broadcast)

	return h
}

// broadcast sends a notification to all connected SSE clients.
func (h *MCPEventsHandler) broadcast(uri string) {
	notification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/resources/updated",
		"params": map[string]interface{}{
			"uri": uri,
		},
	}

	data, err := json.Marshal(notification)
	if err != nil {
		utils.LogErrorf("MCPEventsHandler: failed to marshal notification: %v", err)
		return
	}

	message := fmt.Sprintf("data: %s\n\n", string(data))

	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.clients {
		select {
		case ch <- message:
		default:
			// Client buffer full, skip
		}
	}
}

// Send broadcasts a raw message to all clients.
func (h *MCPEventsHandler) Send(data []byte) {
	message := fmt.Sprintf("data: %s\n\n", string(data))

	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.clients {
		select {
		case ch <- message:
		default:
			// Client buffer full, skip
		}
	}
}

func (h *MCPEventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if the client supports SSE
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for this client
	clientChan := make(chan string, 10)

	// Register the client
	h.mu.Lock()
	h.clients[clientChan] = struct{}{}
	h.mu.Unlock()

	// Cleanup on disconnect
	defer func() {
		h.mu.Lock()
		delete(h.clients, clientChan)
		h.mu.Unlock()
		close(clientChan)
	}()

	// Send initial connection event
	utils.LogInfo("MCPEventsHandler: Sending 'connected' event")
	fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")

	// Send endpoint event (Required by MCP spec for SSE)
	fmt.Fprintf(w, "event: endpoint\ndata: /api/mcp\n\n")

	flusher.Flush()
	utils.LogInfo("MCPEventsHandler: Flushed initial events")

	// Listen for notifications or client disconnect
	for {
		select {
		case msg := <-clientChan:
			fmt.Fprint(w, msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
