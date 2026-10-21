package http_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	mcphttp "node-herder/internal/mcp/transport/http"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// mockMCPServer implements mcphttp.MCPServer for testing
type mockMCPServer struct {
	handleRequestFn        func(ctx context.Context, rawMessage json.RawMessage) json.RawMessage
	listeners              map[string]func(string)
	running                bool
	unregisteredListenerID string
}

func (m *mockMCPServer) Running() bool {
	return m.running
}

func (m *mockMCPServer) HandleRequest(ctx context.Context, rawMessage json.RawMessage) json.RawMessage {
	if m.handleRequestFn != nil {
		return m.handleRequestFn(ctx, rawMessage)
	}
	return []byte(`{"jsonrpc":"2.0","id":1,"result":{}}`)
}

func (m *mockMCPServer) RegisterNotificationListener(cb func(string)) string {
	if m.listeners == nil {
		m.listeners = make(map[string]func(string))
	}
	id := uuid.New().String()
	m.listeners[id] = cb
	return id
}

func (m *mockMCPServer) UnregisterNotificationListener(id string) {
	m.unregisteredListenerID = id
}

func (m *mockMCPServer) OnClientConnectionChange(count int) {
	// no-op for mock
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
		running: true,
	}

	handler := mcphttp.NewMCPHandler(server, nil)

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
	server := &mockMCPServer{running: true}
	handler := mcphttp.NewMCPHandler(server, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/mcp", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("Expected status %d, got %d", http.StatusUnsupportedMediaType, rec.Code)
	}
}

func TestMCPHandler_InvalidJSON(t *testing.T) {
	server := &mockMCPServer{running: true}
	handler := mcphttp.NewMCPHandler(server, nil)

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
		running: true,
	}

	handler := mcphttp.NewMCPHandler(server, nil)

	reqBody := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	req := httptest.NewRequest(http.MethodPost, "/api/mcp", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestSSEHandler_Connection(t *testing.T) {
	server := &mockMCPServer{running: true}
	handler := mcphttp.NewSSEHandler(server)

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

func TestSSEHandler_ReceivesNotification(t *testing.T) {
	server := &mockMCPServer{running: true}
	handler := mcphttp.NewSSEHandler(server)

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

func readSSEData(t *testing.T, reader *bufio.Reader) string {
	t.Helper()
	var data string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(line, "data: ") {
			data = strings.TrimSpace(strings.TrimPrefix(line, "data: "))
		}
		if line == "\n" && data != "" {
			return data
		}
	}
}
func TestMCPRepliesAreIsolatedAndSessionsCleanUp(t *testing.T) {
	backend := &mockMCPServer{running: true}
	sse := mcphttp.NewSSEHandler(backend)
	defer sse.Close()
	mux := http.NewServeMux()
	mux.Handle("/api/mcp/events", sse)
	mux.Handle("/api/mcp", mcphttp.NewMCPHandler(backend, sse))
	server := httptest.NewServer(mux)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	open := func() (*http.Response, *bufio.Reader, string) {
		req, _ := http.NewRequestWithContext(ctx, "GET", server.URL+"/api/mcp/events", nil)
		response, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		reader := bufio.NewReader(response.Body)
		if data := readSSEData(t, reader); !strings.Contains(data, "connected") {
			t.Fatal(data)
		}
		endpoint := readSSEData(t, reader)
		if !strings.HasPrefix(endpoint, "/api/mcp?sessionId=") {
			t.Fatalf("unscoped endpoint: %s", endpoint)
		}
		return response, reader, endpoint
	}
	a, ar, endpoint := open()
	defer a.Body.Close()
	b, br, _ := open()
	defer b.Body.Close()
	post := func(endpoint string) *http.Response {
		response, err := http.Post(server.URL+endpoint, "application/json", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"resources/list"}`))
		if err != nil {
			t.Fatal(err)
		}
		return response
	}
	response := post(endpoint)
	io.Copy(io.Discard, response.Body)
	response.Body.Close()
	if response.StatusCode != 200 {
		t.Fatal(response.StatusCode)
	}
	if data := readSSEData(t, ar); !strings.Contains(data, `"result"`) {
		t.Fatal(data)
	}
	backend.triggerNotification("nodeherder://barrier")
	for _, reader := range []*bufio.Reader{ar, br} {
		if data := readSSEData(t, reader); !strings.Contains(data, "nodeherder://barrier") {
			t.Fatalf("another client's response leaked: %s", data)
		}
	}
	// HTTP-only responses must not be duplicated into any SSE stream.
	response = post("/api/mcp")
	response.Body.Close()
	backend.triggerNotification("nodeherder://http-only-barrier")
	for _, reader := range []*bufio.Reader{ar, br} {
		if data := readSSEData(t, reader); !strings.Contains(data, "http-only-barrier") {
			t.Fatal(data)
		}
	}
	response = post("/api/mcp?sessionId=unknown")
	response.Body.Close()
	if response.StatusCode != 404 {
		t.Fatalf("unknown session status=%d", response.StatusCode)
	}
	cancel()
	a.Body.Close()
	b.Body.Close()
	deadline := time.Now().Add(time.Second)
	for sse.ClientCount() != 0 {
		if time.Now().After(deadline) {
			t.Fatal("disconnected streams retained")
		}
		runtime.Gosched()
	}
	sse.Close()
	if backend.unregisteredListenerID == "" {
		t.Fatal("notification listener retained")
	}
}

func TestLegacySSESDKUsesAdvertisedSessionEndpoint(t *testing.T) {
	protocol := mcpserver.NewMCPServer("test-server", "1.0.0", mcpserver.WithResourceCapabilities(true, true))
	backend := &mockMCPServer{running: true, handleRequestFn: func(ctx context.Context, raw json.RawMessage) json.RawMessage {
		response := protocol.HandleMessage(ctx, raw)
		if response == nil {
			return nil
		}
		body, err := json.Marshal(response)
		if err != nil {
			panic(err)
		}
		return body
	}}
	sse := mcphttp.NewSSEHandler(backend)
	mux := http.NewServeMux()
	mux.Handle("/api/mcp/events", sse)
	mux.Handle("/api/mcp", mcphttp.NewMCPHandler(backend, sse))
	server := httptest.NewServer(mux)
	defer server.Close()
	defer sse.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := mcpclient.NewSSEMCPClient(server.URL + "/api/mcp/events")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if err := client.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if mcpclient.GetEndpoint(client).Query().Get("sessionId") == "" {
		t.Fatal("SDK ignored session endpoint")
	}
	request := mcp.InitializeRequest{}
	request.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	request.Params.ClientInfo = mcp.Implementation{Name: "test-client", Version: "1.0.0"}
	initialized, err := client.Initialize(ctx, request)
	if err != nil || initialized.ServerInfo.Name != "test-server" {
		t.Fatalf("initialize: %+v %v", initialized, err)
	}
	resources, err := client.ListResources(ctx, mcp.ListResourcesRequest{})
	if err != nil || len(resources.Resources) != 0 {
		t.Fatalf("list resources: %+v %v", resources, err)
	}
}
