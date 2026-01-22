package resources_test

import (
	"encoding/json"
	"node-herder/internal/mcp/resources"
	"node-herder/models/devices"
	"testing"
)

// mockDeviceStore implements the store interface for testing
type mockDeviceStore struct {
	devices []*devices.Device
	err     error
}

func (m *mockDeviceStore) AllDevices() ([]*devices.Device, error) {
	return m.devices, m.err
}

func TestPromptResource_GetContent(t *testing.T) {
	pr := resources.NewPromptResource()

	content, err := pr.GetContent()
	if err != nil {
		t.Fatalf("GetContent() error = %v", err)
	}

	if len(content) == 0 {
		t.Error("GetContent() returned empty content")
	}

	// Verify key sections are present
	contentStr := string(content)
	expectedSections := []string{
		"query_device",
		"target_name",
		"metrics",
		"time_scope",
		"aggregation",
		"nodeherder://devices",
	}

	for _, section := range expectedSections {
		if !contains(contentStr, section) {
			t.Errorf("GetContent() missing expected section: %q", section)
		}
	}
}

func TestDevicesResource_GetContent(t *testing.T) {
	store := &mockDeviceStore{
		devices: []*devices.Device{
			{
				Id:           "dev-001",
				FriendlyName: "Attic temperature sensor",
				Exposes: map[string]*devices.Entity{
					"temperature": {
						Name:        "temperature",
						Type:        "numeric",
						Unit:        "°C",
						Description: "Temperature reading",
					},
				},
			},
			{
				Id:           "dev-002",
				FriendlyName: "Kitchen light",
				Exposes: map[string]*devices.Entity{
					"state": {
						Name:        "state",
						Type:        "binary",
						Description: "Light state",
					},
				},
			},
		},
	}

	dr := resources.NewDevicesResource(store)
	content, err := dr.GetContent()
	if err != nil {
		t.Fatalf("GetContent() error = %v", err)
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(content, &result); err != nil {
		t.Fatalf("GetContent() returned invalid JSON: %v", err)
	}

	// Check structure
	if _, ok := result["version"]; !ok {
		t.Error("GetContent() missing 'version' field")
	}

	devicesArr, ok := result["devices"].([]interface{})
	if !ok {
		t.Fatal("GetContent() missing or invalid 'devices' array")
	}

	if len(devicesArr) != 2 {
		t.Errorf("GetContent() devices count = %d, want 2", len(devicesArr))
	}
}

func TestDevicesResource_EmptyStore(t *testing.T) {
	store := &mockDeviceStore{
		devices: []*devices.Device{},
	}

	dr := resources.NewDevicesResource(store)
	content, err := dr.GetContent()
	if err != nil {
		t.Fatalf("GetContent() error = %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(content, &result); err != nil {
		t.Fatalf("GetContent() returned invalid JSON: %v", err)
	}

	devicesArr, ok := result["devices"].([]interface{})
	if !ok {
		t.Fatal("GetContent() missing 'devices' array")
	}

	if len(devicesArr) != 0 {
		t.Errorf("GetContent() devices count = %d, want 0", len(devicesArr))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
