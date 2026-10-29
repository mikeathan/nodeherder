package capabilities_test

// Checks capability registration, dispatch and source discovery.

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

// This source capability deliberately has no waste/weather resource fields.
type opaqueSourceCapability struct{ waste.Adapter }

func (opaqueSourceCapability) Descriptor() capabilities.Descriptor {
	return capabilities.Descriptor{Capability: "opaque/v1", IntegrationType: "opaque", Name: "Opaque"}
}
func (opaqueSourceCapability) Sources(resource model.Resource) ([]model.SourceReference, error) {
	return []model.SourceReference{{Kind: "opaque_value", ResourceID: resource.ResourceID, ServiceID: "channel"}}, nil
}

func TestSourceDiscoveryUsesCapabilityInterface(t *testing.T) {
	registry, err := capabilities.NewRegistry(opaqueSourceCapability{})
	if err != nil {
		t.Fatal(err)
	}
	resource, err := model.NewResource(map[string]any{"contract_version": "opaque/v1", "resource_id": "reading", "revision": 1})
	if err != nil {
		t.Fatal(err)
	}
	sources, err := registry.Sources(resource)
	if err != nil || len(sources) != 1 || sources[0].Kind != "opaque_value" || sources[0].ServiceID != "channel" {
		t.Fatal("source discovery interpreted the payload", sources, err)
	}
	resource.ContractVersion = "removed/v1"
	if _, err = registry.Sources(resource); err == nil {
		t.Fatal("removed capability advertised a source")
	}
}
