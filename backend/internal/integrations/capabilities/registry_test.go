package capabilities_test

import (
	"testing"

	"node-herder/internal/integrations/capabilities"
	"node-herder/internal/integrations/capabilities/waste"
	model "node-herder/models/integrations"
)

func TestRegistryRejectsInvalidRegistrations(t *testing.T) {
	if _, err := capabilities.NewRegistry(nil); err == nil {
		t.Fatal("nil adapter accepted")
	}
	if _, err := capabilities.NewRegistry(waste.Adapter{}, waste.Adapter{}); err == nil {
		t.Fatal("duplicate adapter replaced registration")
	}
	registry, err := capabilities.NewRegistry(waste.Adapter{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = registry.Resolve("missing/v1"); err == nil {
		t.Fatal("unknown capability accepted")
	}
	if err = registry.ValidateDefinition(model.Definition{Capability: model.WasteCapability, IntegrationType: "different"}); err == nil {
		t.Fatal("mismatched category accepted")
	}
	descriptors := registry.Descriptors()
	descriptors[0].Name = "changed"
	if registry.Descriptors()[0].Name == "changed" {
		t.Fatal("caller changed catalogue metadata")
	}
}
