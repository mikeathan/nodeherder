package resources_test

import (
	"node-herder/internal/mcp/resources"
	"node-herder/models/devices"
	"node-herder/models/hub"
	"node-herder/store"
	"testing"
)

// mockDeviceStore implements the store interface for testing
type mockDeviceStore struct {
	devices []*devices.Device
	err     error
}

func (m *mockDeviceStore) LoadHubState() (*hub.HubState, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &hub.HubState{Devices: m.devices}, nil
}

func (m *mockDeviceStore) RegisterIsDirtyCallback(cb store.AppStoreDirtyFlagCallback) {
	// No-op for testing
}

func TestPromptResource_GetContent(t *testing.T) {
	// Create a mock store with some devices to verify they appear in the summary
	store := &mockDeviceStore{
		devices: []*devices.Device{
			{
				Id:           "dev-001",
				FriendlyName: "Attic Sensor",
				Exposes: map[string]*devices.Entity{
					"temperature": {Name: "temperature", Type: "numeric"},
				},
			},
		},
	}

	pr := resources.NewPromptResource(store)

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
		"Attic Sensor", // Verify device friendly name appears
		"dev-001",      // Verify device ID appears
		"temperature",  // Verify metric appears
	}

	for _, section := range expectedSections {
		if !contains(contentStr, section) {
			t.Errorf("GetContent() missing expected section: %q", section)
		}
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

func TestPromptResource_BinarySensorAnnotation(t *testing.T) {
	// Create a device with binary sensor that has an "on" value
	store := &mockDeviceStore{
		devices: []*devices.Device{
			{
				Id:           "dev-alarm",
				FriendlyName: "Attic alarm",
				Exposes: map[string]*devices.Entity{
					"alarm": {
						Name:   "alarm",
						Type:   "binary",
						Values: map[string]any{"on": true},
					},
					"temperature": {
						Name: "temperature",
						Type: "numeric",
					},
				},
			},
		},
	}

	pr := resources.NewPromptResource(store)
	content, err := pr.GetContent()
	if err != nil {
		t.Fatalf("GetContent() error = %v", err)
	}

	contentStr := string(content)

	// Binary sensor with "on" value should show annotation
	if !contains(contentStr, "alarm(on=true)") {
		t.Error("GetContent() should include 'alarm(on=true)' for binary sensor with on value")
	}

	// Numeric sensor should NOT have annotation
	if contains(contentStr, "temperature(on=") {
		t.Error("GetContent() should NOT include '(on=' for numeric sensors")
	}
}

func TestPromptResource_ContainsBooleanInterpretationSection(t *testing.T) {
	store := &mockDeviceStore{
		devices: []*devices.Device{
			{Id: "dev1", FriendlyName: "Test", Exposes: map[string]*devices.Entity{}},
		},
	}

	pr := resources.NewPromptResource(store)
	content, err := pr.GetContent()
	if err != nil {
		t.Fatalf("GetContent() error = %v", err)
	}

	contentStr := string(content)

	// Should contain the boolean interpretation section we added
	expectedSections := []string{
		"Interpreting Boolean",
		"true",
		"false",
		"active/triggered",
		"NOT active",
	}

	for _, section := range expectedSections {
		if !contains(contentStr, section) {
			t.Errorf("GetContent() missing expected boolean section: %q", section)
		}
	}
}
