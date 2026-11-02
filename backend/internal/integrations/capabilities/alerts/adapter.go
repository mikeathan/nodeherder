// Package alerts validates normalized severe-weather alerts without transport or storage ownership.
package alerts

// Connects the severe-weather-alert contract to generic integration normalization and projection.

import (
	"time"

	"node-herder/internal/integrations/capabilities"
	model "node-herder/models/integrations"
)

// Adapter implements alert normalization and projection without owning I/O or persistence.
type Adapter struct{}

// Descriptor publishes the alert contract through the shared integration catalogue.
func (Adapter) Descriptor() capabilities.Descriptor {
	return capabilities.Descriptor{
		Capability:      model.WeatherAlertsCapability,
		IntegrationType: "alerts",
		Name:            "Severe weather alerts",
		Description:     "Active regional weather warnings and instructions.",
	}
}

// ValidateMapping rejects unsupported alert selectors before activation.
func (Adapter) ValidateMapping(raw map[string]any) error { _, err := decode(raw); return err }

// Normalize creates a bounded, owned alert resource from configured provider selectors.
// Failed normalization publishes nothing and preserves the last-good snapshot.
func (Adapter) Normalize(root any, raw map[string]any, context capabilities.Context) (capabilities.Result, error) {
	selectors, err := decode(raw)
	if err != nil {
		return capabilities.Result{}, err
	}
	alerts, err := NormalizeAlerts(root, selectors, context)
	if err != nil {
		return capabilities.Result{}, err
	}
	resource, err := model.NewResource(alerts)
	if err != nil {
		return capabilities.Result{}, err
	}
	return capabilities.Result{
		Resource:     resource,
		Identities:   context.Identities,
		Reservations: []capabilities.Identity{{ID: alerts.ResourceID, Kind: "resource"}},
	}, nil
}

// Project validates persisted data and drops warnings expired at read time.
func (Adapter) Project(resource model.Resource, now time.Time) (model.Resource, error) {
	alerts, err := decodeAlerts(resource)
	if err != nil {
		return model.Resource{}, err
	}
	return model.NewResource(alerts.Project(now))
}

// Checked records successful validation without resetting the age of provider content.
func (Adapter) Checked(resource model.Resource, now time.Time) (model.Resource, error) {
	alerts, err := decodeAlerts(resource)
	if err != nil {
		return model.Resource{}, err
	}
	alerts.LastCheckedAt = now.UTC()
	return model.NewResource(alerts.Project(now))
}

// decodeAlerts validates a stored snapshot before it is projected or exposed.
func decodeAlerts(resource model.Resource) (model.WeatherAlerts, error) {
	var alerts model.WeatherAlerts
	if err := resource.Decode(&alerts); err != nil {
		return alerts, err
	}
	return alerts, alerts.Validate()
}
