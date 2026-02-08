package tools_test

import (
	"context"
	"encoding/json"
	"node-herder/internal/mcp/protocol"
	"node-herder/internal/mcp/resolver"
	"node-herder/internal/mcp/tools"
	"testing"
)

// mockDeviceStore implements resolver.DeviceStore for testing
type mockDeviceStore struct {
	devices []resolver.DeviceInfo
}

func (m *mockDeviceStore) AllDeviceInfo() ([]resolver.DeviceInfo, error) {
	return m.devices, nil
}

// mockQueryService implements metrics querying for testing
type mockQueryService struct {
	result interface{}
	err    error
}

func TestIntentHandler_Handle(t *testing.T) {
	store := &mockDeviceStore{
		devices: []resolver.DeviceInfo{
			{ID: "dev-001", Name: "Attic temperature sensor"},
			{ID: "dev-002", Name: "Living room presence sensor"},
			{ID: "dev-003", Name: "Kitchen power switch"},
		},
	}

	r := resolver.New(store)

	// Note: In a real test, you'd mock the QueryService
	// For now, we test the parsing and resolution logic
	handler := tools.NewIntentHandler(r, nil, nil)

	tests := []struct {
		name       string
		input      string
		wantStatus string
	}{
		{
			name: "missing target_name",
			input: `{
				"metrics": ["temperature"]
			}`,
			wantStatus: "error",
		},
		{
			name: "missing metrics",
			input: `{
				"target_name": "attic sensor"
			}`,
			wantStatus: "error",
		},
		{
			name:       "invalid JSON",
			input:      `not valid json`,
			wantStatus: "error",
		},
		{
			name: "empty metrics array",
			input: `{
				"target_name": "attic sensor",
				"metrics": []
			}`,
			wantStatus: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := handler.Handle(context.Background(), json.RawMessage(tt.input))
			if resp.Status != tt.wantStatus {
				t.Errorf("Handle() status = %q, want %q", resp.Status, tt.wantStatus)
			}
		})
	}
}

func TestIntentHandler_DeviceResolution(t *testing.T) {
	t.Run("device not found returns error", func(t *testing.T) {
		store := &mockDeviceStore{
			devices: []resolver.DeviceInfo{
				{ID: "dev-001", Name: "Attic temperature sensor"},
				{ID: "dev-002", Name: "Kitchen humidity sensor"},
			},
		}
		r := resolver.New(store)
		handler := tools.NewIntentHandler(r, nil, nil)

		input := `{
			"target_name": "bedroom light",
			"metrics": ["state"]
		}`
		resp := handler.Handle(context.Background(), json.RawMessage(input))

		if resp.Status != "error" {
			t.Errorf("Handle() status = %q, want %q", resp.Status, "error")
		}
		if resp.Error == nil {
			t.Error("Handle() should return an error")
		}
		if resp.Error != nil && resp.Error.Code != "resolution_error" {
			t.Errorf("Handle() error code = %q, want %q", resp.Error.Code, "resolution_error")
		}
	})

	// Note: Ambiguous resolution is tested in resolver/resolver_test.go
	// Full integration tests with query execution require the MCP Inspector
	// or a complete mock of the QueryService
}

func TestToolResponse_Factories(t *testing.T) {
	t.Run("NewSuccessResponse", func(t *testing.T) {
		data := map[string]interface{}{"temperature": 22.5}
		resp := protocol.NewSuccessResponse(data)

		if resp.Status != "success" {
			t.Errorf("Status = %q, want %q", resp.Status, "success")
		}
		if resp.Data == nil {
			t.Error("Data should not be nil")
		}
		if resp.Error != nil {
			t.Error("Error should be nil")
		}
	})

	t.Run("NewErrorResponse", func(t *testing.T) {
		resp := protocol.NewErrorResponse("validation_error", "target_name is required")

		if resp.Status != "error" {
			t.Errorf("Status = %q, want %q", resp.Status, "error")
		}
		if resp.Error == nil {
			t.Fatal("Error should not be nil")
		}
		if resp.Error.Code != "validation_error" {
			t.Errorf("Error.Code = %q, want %q", resp.Error.Code, "validation_error")
		}
	})

	t.Run("NewAmbiguousResponse", func(t *testing.T) {
		candidates := []protocol.Candidate{
			{DeviceID: "dev-001", Name: "Attic sensor 1", Score: 0.85},
			{DeviceID: "dev-002", Name: "Attic sensor 2", Score: 0.80},
		}
		resp := protocol.NewAmbiguousResponse(candidates)

		if resp.Status != "ambiguous" {
			t.Errorf("Status = %q, want %q", resp.Status, "ambiguous")
		}
		if len(resp.Candidates) != 2 {
			t.Errorf("Candidates length = %d, want 2", len(resp.Candidates))
		}
	})
}

func TestToolResponse_HintField(t *testing.T) {
	t.Run("success response can have hint", func(t *testing.T) {
		resp := protocol.NewSuccessResponse([]int{})
		resp.Hint = "Available metrics: [temperature, humidity]"

		if resp.Status != "success" {
			t.Errorf("Status = %q, want %q", resp.Status, "success")
		}
		if resp.Hint == "" {
			t.Error("Hint should not be empty")
		}
	})

	t.Run("hint is empty by default", func(t *testing.T) {
		resp := protocol.NewSuccessResponse(map[string]int{"count": 5})

		if resp.Hint != "" {
			t.Errorf("Hint = %q, want empty", resp.Hint)
		}
	})

	t.Run("error response does not need hint", func(t *testing.T) {
		resp := protocol.NewErrorResponse("query_error", "something went wrong")

		if resp.Hint != "" {
			t.Errorf("Error response should not have hint, got %q", resp.Hint)
		}
	})
}

func TestToolResponse_JSONSerialization(t *testing.T) {
	t.Run("success response serializes correctly", func(t *testing.T) {
		resp := protocol.NewSuccessResponse(map[string]float64{"temp": 22.5})

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var decoded protocol.ToolResponse
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if decoded.Status != "success" {
			t.Errorf("decoded.Status = %q, want %q", decoded.Status, "success")
		}
	})

	t.Run("error response serializes correctly", func(t *testing.T) {
		resp := protocol.NewErrorResponse("test_error", "test message")

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var decoded protocol.ToolResponse
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if decoded.Status != "error" {
			t.Errorf("decoded.Status = %q, want %q", decoded.Status, "error")
		}
		if decoded.Error == nil || decoded.Error.Code != "test_error" {
			t.Error("decoded error should have correct code")
		}
	})

	t.Run("ambiguous response serializes candidates", func(t *testing.T) {
		candidates := []protocol.Candidate{
			{DeviceID: "d1", Name: "Device 1", Score: 0.9},
			{DeviceID: "d2", Name: "Device 2", Score: 0.8},
		}
		resp := protocol.NewAmbiguousResponse(candidates)

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var decoded protocol.ToolResponse
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if len(decoded.Candidates) != 2 {
			t.Errorf("decoded.Candidates length = %d, want 2", len(decoded.Candidates))
		}
	})
}

func TestToolResponse_OmitsEmptyFields(t *testing.T) {
	t.Run("success response omits error and candidates", func(t *testing.T) {
		resp := protocol.NewSuccessResponse("data")

		data, _ := json.Marshal(resp)
		str := string(data)

		if contains(str, "error") {
			t.Error("success response should not include 'error' field")
		}
		if contains(str, "candidates") {
			t.Error("success response should not include 'candidates' field")
		}
	})

	t.Run("error response omits data and candidates", func(t *testing.T) {
		resp := protocol.NewErrorResponse("err", "msg")

		data, _ := json.Marshal(resp)
		str := string(data)

		if contains(str, "\"data\"") {
			t.Error("error response should not include 'data' field")
		}
		if contains(str, "candidates") {
			t.Error("error response should not include 'candidates' field")
		}
	})
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
