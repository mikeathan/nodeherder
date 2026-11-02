// Package fuel validates normalized fuel prices and computes origin-relative distance ordering.
package fuel

// Connects the fuel-price contract to generic integration normalization and projection.

import (
	"time"

	"node-herder/internal/integrations/capabilities"
	model "node-herder/models/integrations"
)

// Adapter implements fuel-price normalization and projection without owning I/O or persistence.
type Adapter struct{}

// Descriptor publishes the fuel-price contract through the shared integration catalogue.
func (Adapter) Descriptor() capabilities.Descriptor {
	return capabilities.Descriptor{
		Capability:      model.FuelPricesCapability,
		IntegrationType: "fuel_prices",
		Name:            "Fuel prices",
		Description:     "Nearby forecourt prices for the selected origin.",
	}
}

// ValidateMapping rejects unsupported fuel selectors before activation.
func (Adapter) ValidateMapping(raw map[string]any) error { _, err := decode(raw); return err }

// Normalize creates a bounded, owned fuel snapshot from configured provider selectors.
func (Adapter) Normalize(root any, raw map[string]any, context capabilities.Context) (capabilities.Result, error) {
	selectors, err := decode(raw)
	if err != nil {
		return capabilities.Result{}, err
	}
	fuel, err := NormalizeFuel(root, selectors, context)
	if err != nil {
		return capabilities.Result{}, err
	}
	resource, err := model.NewResource(fuel)
	if err != nil {
		return capabilities.Result{}, err
	}
	return capabilities.Result{
		Resource:     resource,
		Identities:   context.Identities,
		Reservations: []capabilities.Identity{{ID: fuel.ResourceID, Kind: "resource"}},
	}, nil
}

// Project validates persisted data and derives distance ordering/freshness without writes.
func (Adapter) Project(resource model.Resource, now time.Time) (model.Resource, error) {
	fuel, err := decodeFuel(resource)
	if err != nil {
		return model.Resource{}, err
	}
	return model.NewResource(fuel.Project(now))
}

// Checked records successful validation without resetting the age of provider content.
func (Adapter) Checked(resource model.Resource, now time.Time) (model.Resource, error) {
	fuel, err := decodeFuel(resource)
	if err != nil {
		return model.Resource{}, err
	}
	fuel.LastCheckedAt = now.UTC()
	return model.NewResource(fuel.Project(now))
}

// decodeFuel validates a stored snapshot before it is projected or exposed.
func decodeFuel(resource model.Resource) (model.FuelPrices, error) {
	var fuel model.FuelPrices
	if err := resource.Decode(&fuel); err != nil {
		return fuel, err
	}
	return fuel, fuel.Validate()
}
