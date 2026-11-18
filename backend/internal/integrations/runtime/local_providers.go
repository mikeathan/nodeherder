package runtime

// Resolves providers that are built in code rather than read from a document.
//
// A local provider has no base URL, no requests and no inputs, so it cannot be expressed in the
// provider grammar and must not be stored as one. It is instead registered with the runtime at
// construction and resolved by key wherever a stored record would otherwise be read.
//
// Everything downstream — an instance, its enable state, its snapshot, its board entry — works
// exactly as it does for an authored provider, because the only thing that differs is where the
// definition comes from.

import (
	"fmt"

	"node-herder/internal/integrations/definitions"
)

// installLocalDefinitions compiles the registered local providers once at startup.
//
// Compiled rather than re-derived per call: they cannot change at runtime, and a provider that
// silently vanished between two reads would look like a deleted integration.
func (s *Service) installLocalDefinitions() error {
	for _, definition := range s.localDefinitions {
		key := definitions.DefinitionKey(definition)
		compiled, err := s.compiler.CompileLocal(definition)
		if err != nil {
			return fmt.Errorf("compile local provider %s: %w", key, err)
		}
		s.localCompiled[key] = compiled
	}
	return nil
}

// localDefinition resolves a registered local provider by key.
func (s *Service) localDefinition(key string) (*definitions.Compiled, bool) {
	compiled, ok := s.localCompiled[key]
	return compiled, ok
}

// localProviders lists the registered local providers for the catalogue.
func (s *Service) localProviders() map[string]*definitions.Compiled {
	return s.localCompiled
}

// SetLocalSource supplies the value a local provider publishes.
//
// Wiring happens after construction because the source is the hub's live view, which is created
// after the integration runtime. Refusing a late install would mean the two could only be composed
// in one order, which is a constraint with no benefit.
func (s *Service) SetLocalSource(source func() (any, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.localSource = source
}
