// Package weather validates declaratively mapped forecasts without transport or storage ownership.
package weather

// Connects the weather contract to generic integration normalization and projection.

import (
	"time"

	"node-herder/internal/integrations/capabilities"
	model "node-herder/models/integrations"
)

// Adapter implements weather normalization and projection without owning I/O or persistence.
type Adapter struct{}

// Descriptor publishes the weather contract through the shared integration catalogue.
func (Adapter) Descriptor() capabilities.Descriptor {
	return capabilities.Descriptor{Capability: model.WeatherCapability, IntegrationType: "weather", Name: "Weather", Description: "Local conditions and forecasts. Refresh manually."}
}

// ValidateMapping rejects unsupported weather selectors before activation.
func (Adapter) ValidateMapping(raw map[string]any) error { _, err := decode(raw); return err }

// Normalize creates a bounded, owned resource from configured provider selectors.
// Existing identity is retained across refresh; failed normalization publishes nothing.
func (Adapter) Normalize(root any, raw map[string]any, context capabilities.Context) (capabilities.Result, error) {
	selectors, err := decode(raw)
	if err != nil {
		return capabilities.Result{}, err
	}
	reader, err := newForecastReader(root, selectors)
	if err != nil {
		return capabilities.Result{}, err
	}
	forecast, err := reader.forecast(context)
	if err != nil {
		return capabilities.Result{}, err
	}
	resource, err := model.NewResource(forecast.Project(context.Now))
	if err != nil {
		return capabilities.Result{}, err
	}
	return capabilities.Result{Resource: resource, Identities: context.Identities, Reservations: []capabilities.Identity{{ID: forecast.ResourceID, Kind: "resource"}}}, nil
}

// Project validates persisted weather and derives freshness without changing saved content.
func (Adapter) Project(resource model.Resource, now time.Time) (model.Resource, error) {
	forecast, err := decodeForecast(resource)
	if err != nil {
		return model.Resource{}, err
	}
	return model.NewResource(forecast.Project(now))
}

// Checked records successful validation without resetting the age of forecast content.
func (Adapter) Checked(resource model.Resource, now time.Time) (model.Resource, error) {
	forecast, err := decodeForecast(resource)
	if err != nil {
		return model.Resource{}, err
	}
	forecast.LastCheckedAt = now.UTC()
	return model.NewResource(forecast.Project(now))
}

// decodeForecast validates a stored forecast before it is projected or exposed.
func decodeForecast(resource model.Resource) (model.WeatherForecast, error) {
	var forecast model.WeatherForecast
	if err := resource.Decode(&forecast); err != nil {
		return forecast, err
	}
	return forecast, forecast.Validate()
}
