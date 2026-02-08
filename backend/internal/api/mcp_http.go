package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"node-herder/utils"
)

type MCPServer interface {
	HandleRequest(ctx context.Context, rawMessage json.RawMessage) json.RawMessage
	RegisterNotificationListener(cb func(string)) string
	UnregisterNotificationListener(id string)
}

type MCPHandler struct {
	server MCPServer
	sse    *SSEHandler
}

func NewMCPHandler(server MCPServer, sse *SSEHandler) *MCPHandler {
	return &MCPHandler{server: server, sse: sse}
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
		if h.sse != nil {
			h.sse.Send(response)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

type SSEHandler struct {
	server  MCPServer
	clients map[chan string]struct{}
	mu      sync.RWMutex
}

func NewSSEHandler(server MCPServer) *SSEHandler {
	h := &SSEHandler{
		server:  server,
		clients: make(map[chan string]struct{}),
	}

	// We discard the ID because the global handler persists for the app lifetime
	server.RegisterNotificationListener(h.broadcast)

	return h
}

type notification struct {
	JSONRPC string             `json:"jsonrpc"`
	Method  string             `json:"method"`
	Params  notificationParams `json:"params"`
}

type notificationParams struct {
	URI string `json:"uri"`
}

func (h *SSEHandler) broadcast(uri string) {
	msg := notification{
		JSONRPC: "2.0",
		Method:  "notifications/resources/updated",
		Params:  notificationParams{URI: uri},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		utils.LogErrorf("SSEHandler: failed to marshal notification: %v", err)
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

func (h *SSEHandler) Send(data []byte) {
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

func (h *SSEHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientChan := make(chan string, 10)

	h.mu.Lock()
	h.clients[clientChan] = struct{}{}
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, clientChan)
		h.mu.Unlock()
		close(clientChan)
	}()

	utils.LogInfo("SSEHandler: Sending 'connected' event")
	fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")

	// Send endpoint event (Required by MCP spec for SSE)
	fmt.Fprintf(w, "event: endpoint\ndata: /api/mcp\n\n")

	flusher.Flush()
	utils.LogInfo("SSEHandler: Flushed initial events")

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-clientChan:
			fmt.Fprint(w, msg)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprint(w, ": heartbeat\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (r *Router) RegisterMCP(server MCPServer) {
	if server == nil {
		return
	}
	sseHandler := NewSSEHandler(server)
	r.PublicPOST("/api/mcp", NewMCPHandler(server, sseHandler))
	r.PublicGET("/api/mcp/events", sseHandler)
}
