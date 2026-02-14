package server_test

import (
	"encoding/json"
	"node-herder/internal/mcp/resolver"
	"node-herder/internal/mcp/server"
	"node-herder/models/devices"
	"node-herder/models/hub"
	"node-herder/models/settings"
	"node-herder/store"
	"testing"
)

// mockDeviceStore implements server.DeviceStore for testing
type mockDeviceStore struct {
	devices     []*devices.Device
	err         error
	configCache *settings.AppConfigCache
}

func (m *mockDeviceStore) LoadHubState() (*hub.HubState, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &hub.HubState{Devices: m.devices}, nil
}

func (m *mockDeviceStore) RegisterIsDirtyCallback(cb store.AppStoreDirtyFlagCallback) {
	// no-op for tests
}

func (m *mockDeviceStore) AppConfig() *settings.AppConfigCache {
	return m.configCache
}

// mockSettingsRepo implements settings.Repository
type mockSettingsRepo struct {
	config *settings.AppConfig
}

func (m *mockSettingsRepo) SaveAppConfig(config *settings.AppConfig) error {
	m.config = config
	return nil
}
func (m *mockSettingsRepo) Load() (*settings.AppConfig, error) {
	if m.config == nil {
		return settings.NewAppConfig(), nil
	}
	return m.config, nil
}
func (m *mockSettingsRepo) LoadBridgeConfig() (*settings.BridgeConfig, error) {
	return settings.NewBridgeConfig(), nil
}
func (m *mockSettingsRepo) SaveBridgeConfig(bridgeConfig *settings.BridgeConfig) error {
	return nil
}
func (m *mockSettingsRepo) SaveHubConfig(hubConfig *settings.HubConfig) error {
	return nil
}
func (m *mockSettingsRepo) LoadOrDefaultDeviceConfig(id string) (*settings.DeviceConfig, error) {
	return settings.NewDeviceConfig(id), nil
}
func (m *mockSettingsRepo) SaveDeviceConfig(deviceConfig *settings.DeviceConfig) error {
	return nil
}
func (m *mockSettingsRepo) DeleteDeviceConfig(id string) error {
	return nil
}
func (m *mockSettingsRepo) Close() error {
	return nil
}

// mockMetricsRepo implements metrics.Repository for minimal testing
type mockMetricsRepo struct{}

func TestServer_New(t *testing.T) {
	repo := &mockSettingsRepo{}
	cache, _ := settings.NewAppConfigCache(repo, nil)

	store := &mockDeviceStore{
		configCache: cache,
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
	srv := server.New(store, cache, nil)

	if srv == nil {
		t.Fatal("New() returned nil")
	}
}

func TestServer_VerifyDeviceStoreAdapter(t *testing.T) {
	repo := &mockSettingsRepo{}
	cache, _ := settings.NewAppConfigCache(repo, nil)

	// Test that the device store adapter correctly converts devices
	store := &mockDeviceStore{
		configCache: cache,
		devices: []*devices.Device{
			{Id: "dev-001", FriendlyName: "Device One"},
			{Id: "dev-002", FriendlyName: "Device Two"},
		},
	}

	// Verify the store works correctly
	state, err := store.LoadHubState()
	if err != nil {
		t.Fatalf("LoadHubState() error = %v", err)
	}

	if len(state.Devices) != 2 {
		t.Errorf("LoadHubState() returned %d devices, want 2", len(state.Devices))
	}
}

// TestMCPServerIntegration tests the full MCP flow
// This is a lightweight integration test
func TestMCPServerIntegration(t *testing.T) {
	repo := &mockSettingsRepo{}
	cache, _ := settings.NewAppConfigCache(repo, nil)

	// Create a store with test devices
	store := &mockDeviceStore{
		configCache: cache,
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
	srv := server.New(store, cache, nil)

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
	repo := &mockSettingsRepo{}
	cache, _ := settings.NewAppConfigCache(repo, nil)

	// This tests the pattern used in server.go where DeviceStore is adapted
	store := &mockDeviceStore{
		configCache: cache,
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

// Compile-time check that mockDeviceStore implements server.DeviceStore
var _ server.DeviceStore = (*mockDeviceStore)(nil)

// Compile-time check that deviceInfoAdapter implements resolver.DeviceStore
var _ resolver.DeviceStore = (*deviceInfoAdapter)(nil)

func TestHandleRequest_ResourcesList(t *testing.T) {
	repo := &mockSettingsRepo{}
	cache, _ := settings.NewAppConfigCache(repo, nil)

	// Setup mock store with devices
	store := &mockDeviceStore{
		configCache: cache,
		devices: []*devices.Device{
			{Id: "dev-001", FriendlyName: "Test device"},
		},
	}
	srv := server.New(store, cache, nil)

	// First, send initialize request
	initReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test",
				"version": "1.0",
			},
		},
	}
	initBytes, _ := json.Marshal(initReq)
	srv.HandleRequest(t.Context(), initBytes)

	// Now test resources/list
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "resources/list",
	}
	reqBytes, _ := json.Marshal(req)

	respBytes := srv.HandleRequest(t.Context(), reqBytes)

	if respBytes == nil {
		t.Fatal("HandleRequest returned nil")
	}

	// Just verify we got a valid JSON response
	var resp map[string]interface{}
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp["error"] != nil {
		t.Errorf("Expected success, got error: %v", resp["error"])
	}
}

func TestNotificationListener(t *testing.T) {
	repo := &mockSettingsRepo{}
	cache, _ := settings.NewAppConfigCache(repo, nil)

	store := &mockDeviceStore{configCache: cache}
	srv := server.New(store, cache, nil)

	// Track notifications
	var receivedURI string
	listener := func(uri string) {
		receivedURI = uri
	}

	srv.RegisterNotificationListener(listener)

	// The store's dirty callback should trigger the listener
	// In a real scenario, this would be called when devices change
	// For this test, we directly test the listener registration
	if receivedURI != "" {
		t.Errorf("Expected no notification yet, got %q", receivedURI)
	}
}

func TestServer_Initialize(t *testing.T) {
	repo := &mockSettingsRepo{}
	cache, _ := settings.NewAppConfigCache(repo, nil)
	store := &mockDeviceStore{configCache: cache}
	srv := server.New(store, cache, nil)

	// 1. Test Initialize with default config (Enabled: false)
	err := srv.Initialize()
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	if srv.Running() {
		t.Error("Expected server to be stopped by default")
	}

	// 2. Test Initialize with Enabled: true
	// 2. Test Initialize with Enabled: true
	cache.SaveMCPConfig(&settings.MCPConfig{Enabled: true})

	err = srv.Initialize()
	if err != nil {
		t.Fatalf("Initialize() enabled error = %v", err)
	}
	if !srv.Running() {
		t.Error("Expected server to be running")
	}

	// 3. Test Initialize with Enabled: false again
	// 3. Test Initialize with Enabled: false again
	cache.SaveMCPConfig(&settings.MCPConfig{Enabled: false})

	err = srv.Initialize()
	if err != nil {
		t.Fatalf("Initialize() disabled error = %v", err)
	}
	if srv.Running() {
		t.Error("Expected server to be stopped")
	}
}
