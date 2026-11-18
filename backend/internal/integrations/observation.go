package integrations

// Connects the shipped observation capability to the hub's live observation service.
//
// The dependency runs one way on purpose: the integration runtime asks the hub for a snapshot, and
// the hub asks the runtime whether anything wants observation. Neither imports the other's
// implementation — the runtime knows only `func() (any, error)` and `func() bool`, so it stays
// transport-agnostic and testable without a radio.

import (
	"errors"

	"node-herder/internal/integrations/capabilities/observation"
)

// Enabled returns a gate reporting whether an observation integration is enabled.
//
// Nil means observation is always on, which is the behaviour for a hub whose integration runtime is
// unavailable: refusing to observe because the *integration* layer is degraded would silently break
// a working sensor.
func ObservationEnabled(service *Service) func() bool {
	if service == nil {
		return nil
	}
	return func() bool { return service.LocalEnabled(observation.DefinitionKey()) }
}

// WithObservationSource lets the runtime publish the hub's live observation view.
//
// It must be called before the integration is added, otherwise the first publish reports an
// unavailable source and the operator sees an error on a card that would have worked.
func WithObservationSource(service *Service, source func() (any, error)) error {
	if service == nil || source == nil {
		return errors.New("observation source needs an integration service and a source")
	}
	service.SetLocalSource(source)
	return nil
}
