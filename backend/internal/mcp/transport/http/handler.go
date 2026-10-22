package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"node-herder/internal/auth"
	"sync"
	"time"

	"node-herder/utils"

	"github.com/google/uuid"
)

type MCPServer interface {
	HandleRequest(ctx context.Context, rawMessage json.RawMessage) json.RawMessage
	Running() bool
	RegisterNotificationListener(cb func(string)) string
	UnregisterNotificationListener(id string)
	OnClientConnectionChange(count int)
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

	if !h.server.Running() {
		writeJSONError(w, http.StatusServiceUnavailable, "MCP server is stopped")
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

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID != "" && (h.sse == nil || !h.sse.hasSession(sessionID, r)) {
		writeJSONError(w, http.StatusNotFound, "MCP session not found")
		return
	}
	response := h.server.HandleRequest(r.Context(), rawMessage)

	w.Header().Set("Content-Type", "application/json")
	if response != nil {
		// Legacy SSE replies are scoped to the advertised session endpoint.
		if h.sse != nil && sessionID != "" {
			h.sse.send(sessionID, response)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

const maxSSEClients = 64

type sseClient struct {
	messages  chan string
	principal string
	cancel    context.CancelFunc
}

type SSEHandler struct {
	server     MCPServer
	clients    map[string]*sseClient
	mu         sync.RWMutex
	statusMu   sync.Mutex
	listenerID string
	closed     bool
}

func NewSSEHandler(server MCPServer) *SSEHandler {
	h := &SSEHandler{server: server, clients: make(map[string]*sseClient)}
	h.listenerID = server.RegisterNotificationListener(h.broadcast)
	return h
}

// Close releases the app-owned notification listener and cancels active streams.
func (h *SSEHandler) Close() {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	h.closed = true
	for _, client := range h.clients {
		client.cancel()
	}
	h.mu.Unlock()
	h.server.UnregisterNotificationListener(h.listenerID)
}

func requestPrincipal(r *http.Request) string {
	if user, ok := auth.GetUserFromContext(r); ok {
		return user.UserID
	}
	return ""
}

func (h *SSEHandler) hasSession(id string, r *http.Request) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	client := h.clients[id]
	return client != nil && !h.closed && client.principal == requestPrincipal(r)
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

	for _, client := range h.clients {
		select {
		case client.messages <- message:
		default:
			// Client buffer full, skip
		}
	}
}

func (h *SSEHandler) send(id string, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if client := h.clients[id]; client != nil {
		select {
		case client.messages <- fmt.Sprintf("data: %s\n\n", data):
		default:
			client.cancel() // Reconnect rather than silently lose a response.
		}
	}
}

func (h *SSEHandler) notifyClientCount() {
	h.statusMu.Lock()
	defer h.statusMu.Unlock()
	h.server.OnClientConnectionChange(h.ClientCount())
}

// writeEvent refreshes the write deadline without limiting a healthy stream's lifetime.
func writeEvent(w http.ResponseWriter, event string) error {
	controller := http.NewResponseController(w)
	_ = controller.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err := io.WriteString(w, event); err != nil {
		return err
	}
	return controller.Flush()
}

func (h *SSEHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, ok := w.(http.Flusher); !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}
	if !h.server.Running() {
		http.Error(w, "MCP server is stopped", http.StatusServiceUnavailable)
		return
	}
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	stop := auth.WatchSession(r, cancel)
	defer stop()
	id := uuid.NewString()
	client := &sseClient{messages: make(chan string, 10), principal: requestPrincipal(r), cancel: cancel}
	h.mu.Lock()
	if h.closed || len(h.clients) >= maxSSEClients {
		h.mu.Unlock()
		http.Error(w, "MCP stream capacity reached", http.StatusServiceUnavailable)
		return
	}
	h.clients[id] = client
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.clients, id)
		h.mu.Unlock()
		h.notifyClientCount()
	}()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	defer func() { _ = http.NewResponseController(w).SetWriteDeadline(time.Time{}) }()
	if ctx.Err() != nil {
		return
	}
	initial := fmt.Sprintf("data: {\"type\":\"connected\"}\n\nevent: endpoint\ndata: /api/mcp?sessionId=%s\n\n", id)
	if err := writeEvent(w, initial); err != nil {
		return
	}
	h.notifyClientCount()
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case message := <-client.messages:
			if ctx.Err() != nil || !auth.SessionValid(r) {
				return
			}
			if err := writeEvent(w, message); err != nil {
				return
			}
		case <-ticker.C:
			if !h.server.Running() || !auth.SessionValid(r) {
				return
			}
			if err := writeEvent(w, ": heartbeat\n\n"); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (h *SSEHandler) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
