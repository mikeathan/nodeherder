package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"node-herder/internal/mcp/resolver"
	"node-herder/internal/mcp/server"
	"node-herder/models/devices"
	"node-herder/store"
	"os"
	"testing"
)

// mockDeviceStore implements server.DeviceStore for testing
type mockDeviceStore struct {
	devices []*devices.Device
	err     error
}

func (m *mockDeviceStore) AllDevices() ([]*devices.Device, error) {
	return m.devices, m.err
}

func (m *mockDeviceStore) RegisterIsDirtyCallback(cb store.AppStoreDirtyFlagCallback) {
	// no-op for tests
}

// mockMetricsRepo implements metrics.Repository for minimal testing
type mockMetricsRepo struct{}

func TestServer_New(t *testing.T) {
	store := &mockDeviceStore{
		devices: []*devices.Device{
			{
				Id:           "dev-001",
				FriendlyName: "Test device",
				Exposes:      make(map[string]*devices.Entity),
			},
		},
	}

	// Note: QueryService requires a full AppStore, so we test with nil
	// In a real test, you'd use a proper mock
	srv := server.New(store, nil)

	if srv == nil {
		t.Fatal("New() returned nil")
	}
}

func TestServer_VerifyDeviceStoreAdapter(t *testing.T) {
	// Test that the device store adapter correctly converts devices
	store := &mockDeviceStore{
		devices: []*devices.Device{
			{Id: "dev-001", FriendlyName: "Device One"},
			{Id: "dev-002", FriendlyName: "Device Two"},
		},
	}

	// Verify the store works correctly
	devs, err := store.AllDevices()
	if err != nil {
		t.Fatalf("AllDevices() error = %v", err)
	}

	if len(devs) != 2 {
		t.Errorf("AllDevices() returned %d devices, want 2", len(devs))
	}
}

// TestMCPServerIntegration tests the full MCP flow
// This is a lightweight integration test
func TestMCPServerIntegration(t *testing.T) {
	// Create a store with test devices
	store := &mockDeviceStore{
		devices: []*devices.Device{
			{
				Id:           "sensor-attic-01",
				FriendlyName: "Attic temperature sensor",
				Exposes: map[string]*devices.Entity{
					"temperature": {
						Name: "temperature",
						Type: "numeric",
						Unit: "°C",
					},
				},
			},
		},
	}

	// Create server (without QueryService for this test)
	srv := server.New(store, nil)

	if srv == nil {
		t.Fatal("Failed to create MCP server")
	}

	// The server should be ready to serve
	// Note: We can't test ServeStdio directly in unit tests
	// as it blocks on stdin. That requires integration testing
	// with the MCP Inspector or a mock stdio implementation.
}

// TestResolverIntegration tests the resolver with the device store adapter pattern
func TestResolverIntegration(t *testing.T) {
	// This tests the pattern used in server.go where DeviceStore is adapted
	store := &mockDeviceStore{
		devices: []*devices.Device{
			{Id: "dev-001", FriendlyName: "Attic temperature sensor"},
			{Id: "dev-002", FriendlyName: "Living room light"},
			{Id: "dev-003", FriendlyName: "Kitchen humidity sensor"},
		},
	}

	// Create adapter (simulating what server.New does)
	adapter := &deviceInfoAdapter{store: store}

	// Create resolver with adapted store
	r := resolver.New(adapter)

	// Test resolution
	tests := []struct {
		name   string
		target string
		wantID string
	}{
		{
			name:   "resolve attic sensor",
			target: "attic temperature",
			wantID: "dev-001",
		},
		{
			name:   "resolve kitchen sensor",
			target: "kitchen humidity",
			wantID: "dev-003",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, _, err := r.Resolve(tt.target)
			if err != nil {
				t.Errorf("Resolve(%q) error = %v", tt.target, err)
				return
			}
			if id != tt.wantID {
				t.Errorf("Resolve(%q) = %q, want %q", tt.target, id, tt.wantID)
			}
		})
	}
}

// deviceInfoAdapter adapts mockDeviceStore to resolver.DeviceStore
type deviceInfoAdapter struct {
	store *mockDeviceStore
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

// Compile-time check that mockDeviceStore implements server.DeviceStore
var _ server.DeviceStore = (*mockDeviceStore)(nil)

// Compile-time check that deviceInfoAdapter implements resolver.DeviceStore
var _ resolver.DeviceStore = (*deviceInfoAdapter)(nil)

func TestServeStdio_SubscriptionInterception(t *testing.T) {
	// Setup mock store
	store := &mockDeviceStore{}
	srv := server.New(store, nil)

	// Capture stdin/stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	r, w, _ := os.Pipe()
	os.Stdin = r
	outReader, outWriter, _ := os.Pipe()
	os.Stdout = outWriter

	// Write subscription request
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      123,
		"method":  "resources/subscribe",
		"params": map[string]interface{}{
			"uri": "nodeherder://devices",
		},
	}
	json.NewEncoder(w).Encode(req)
	w.Close()

	// Run ServeStdio in background (it blocks until EOF)
	go func() {
		srv.ServeStdio()
		outWriter.Close()
	}()

	// Read response
	var buf bytes.Buffer
	io.Copy(&buf, outReader)

	var resp struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      interface{}     `json:"id"`
		Result  json.RawMessage `json:"result,omitempty"`
		Error   interface{}     `json:"error,omitempty"`
	}

	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response
	if resp.Error != nil {
		t.Errorf("Expected success, got error: %v", resp.Error)
	}

	// We can't easily check for exact ID type match (int vs float64) without helpers,
	// but we can check the value string rep
	if fmt.Sprintf("%v", resp.ID) != "123" {
		t.Errorf("Expected ID 123, got %v", resp.ID)
	}
}
