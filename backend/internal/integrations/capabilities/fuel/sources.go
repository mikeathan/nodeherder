package fuel

// Exposes a validated fuel snapshot as a Home placement source.

import model "node-herder/models/integrations"

// Sources exposes the snapshot's stable placement reference after domain validation.
func (Adapter) Sources(resource model.Resource) ([]model.SourceReference, error) {
	if _, err := decodeFuel(resource); err != nil {
		return nil, err
	}
	return []model.SourceReference{{Kind: "fuel_prices", ResourceID: resource.ResourceID}}, nil
}
