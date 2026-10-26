package integrations

import (
	"context"
	"encoding/json"

	defaults "node-herder/configs/integrations"
	"node-herder/internal/integrations/capabilities"
	"node-herder/internal/integrations/capabilities/waste"
	"node-herder/internal/integrations/definitions"
	"node-herder/internal/integrations/runtime"
)

// Service is the application-facing integration runtime. Composition here installs supported
// capabilities; runtime and its coordinator have no dependency on concrete capability packages.
type Service = runtime.Service

// Options injects repositories, HTTP and clock; nil Capabilities installs the shipped adapters.
type Options = runtime.Options

// NewService composes shipped capabilities unless callers explicitly supply their own registry.
// Runtime takes lifecycle ownership of its repository/client and performs no startup provider I/O.
func NewService(ctx context.Context, options Options) (*Service, error) {
	if options.Shipped == nil {
		options.Shipped = defaults.Files
	}
	if options.Capabilities == nil {
		registry, err := shippedCapabilities()
		if err != nil {
			return nil, err
		}
		options.Capabilities = registry
	}
	return runtime.NewService(ctx, options)
}

// DegradedService provides readable availability when private startup storage cannot open.
func DegradedService(code string) *Service { return runtime.DegradedService(code) }

// CompileYAML validates host YAML against the shipped capabilities without activating it.
func CompileYAML(source string) (*definitions.Compiled, error) {
	registry, err := shippedCapabilities()
	if err != nil {
		return nil, err
	}
	return definitions.NewCompiler(registry).CompileYAML(source)
}

// CompileDocument validates editor JSON with the same shipped capability and YAML rules.
func CompileDocument(raw json.RawMessage) (*definitions.Compiled, error) {
	registry, err := shippedCapabilities()
	if err != nil {
		return nil, err
	}
	return definitions.NewCompiler(registry).CompileDocument(raw)
}

// shippedCapabilities is the single registration point used by runtime and standalone editors.
func shippedCapabilities() (*capabilities.Registry, error) {
	return capabilities.NewRegistry(waste.Adapter{})
}
