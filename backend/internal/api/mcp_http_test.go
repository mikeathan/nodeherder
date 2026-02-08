package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"node-herder/internal/api"
)

// mockMCPServer implements api.MCPServer for testing
type mockMCPServer struct {
	handleRequestFn func(ctx context.Context, rawMessage json.RawMessage) json.RawMessage
	listeners       []func(string)
}

func (m *mockMCPServer) HandleRequest(ctx context.Context, rawMessage json.RawMessage) json.RawMessage {
	if m.handleRequestFn != nil {
		return m.handleRequestFn(ctx, rawMessage)
	}
	return []byte(`{"jsonrpc":"2.0","id":1,"result":{}}`)
}

func (m *mockMCPServer) RegisterNotificationListener(cb func(string)) {
	m.listeners = append(m.listeners, cb)
}

func (m *mockMCPServer) UnregisterNotificationListener(cb func(string)) {
	// no-op for tests
}

func (m *mockMCPServer) triggerNotification(uri string) {
	for _, cb := range m.listeners {
		cb(uri)
	}
}

func TestMCPHandler_ValidRequest(t *testing.T) {
	server := &mockMCPServer{
		handleRequestFn: func(ctx context.Context, rawMessage json.RawMessage) json.RawMessage {
			return []byte(`{"jsonrpc":"2.0","id":1,"result":{"resources":[]}}`)
		},
	}

	handler := api.NewMCPHandler(server, nil)

	reqBody := `{"jsonrpc":"2.0","id":1,"method":"resources/list"}`
	req := httptest.NewRequest(http.MethodPost, "/api/mcp", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp["result"] == nil {
		t.Error("Expected result in response")
	}
}

func TestMCPHandler_InvalidContentType(t *testing.T) {
	server := &mockMCPServer{}
	handler := api.NewMCPHandler(server, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/mcp", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("Expected status %d, got %d", http.StatusUnsupportedMediaType, rec.Code)
	}
}

func TestMCPHandler_InvalidJSON(t *testing.T) {
	server := &mockMCPServer{}
	handler := api.NewMCPHandler(server, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/mcp", bytes.NewBufferString(`not json`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestMCPHandler_NilResponse(t *testing.T) {
	server := &mockMCPServer{
		handleRequestFn: func(ctx context.Context, rawMessage json.RawMessage) json.RawMessage {
			return nil // Notification, no response expected
		},
	}

	handler := api.NewMCPHandler(server, nil)

	reqBody := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	req := httptest.NewRequest(http.MethodPost, "/api/mcp", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestMCPEventsHandler_Connection(t *testing.T) {
	server := &mockMCPServer{}
	handler := api.NewMCPEventsHandler(server)

	req := httptest.NewRequest(http.MethodGet, "/api/mcp/events", nil)

	// Create a context with timeout to avoid blocking forever
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	// Run handler in goroutine since it blocks
	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(rec, req)
		close(done)
	}()

	// Wait for handler to complete (due to context timeout)
	<-done

	// Check headers
	if rec.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected Content-Type text/event-stream, got %s", rec.Header().Get("Content-Type"))
	}

	// Check that initial connection event was sent
	body := rec.Body.String()
	if !bytes.Contains([]byte(body), []byte(`"type":"connected"`)) {
		t.Errorf("Expected connected event, got: %s", body)
	}
}

func TestMCPEventsHandler_ReceivesNotification(t *testing.T) {
	server := &mockMCPServer{}
	handler := api.NewMCPEventsHandler(server)

	req := httptest.NewRequest(http.MethodGet, "/api/mcp/events", nil)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	// Run handler in goroutine
	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(rec, req)
		close(done)
	}()

	// Give handler time to start and register listener
	time.Sleep(50 * time.Millisecond)

	// Trigger a notification
	server.triggerNotification("nodeherder://devices")

	// Wait for handler to complete
	<-done

	// Check that notification was received
	body := rec.Body.String()
	if !bytes.Contains([]byte(body), []byte(`nodeherder://devices`)) {
		t.Errorf("Expected devices notification, got: %s", body)
	}
}
