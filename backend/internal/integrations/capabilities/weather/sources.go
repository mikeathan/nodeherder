package weather

// Exposes a validated forecast as a Home placement source.

import model "node-herder/models/integrations"

// Sources exposes the forecast's stable placement reference after domain validation.
func (Adapter) Sources(resource model.Resource) ([]model.SourceReference, error) {
	if _, err := decodeForecast(resource); err != nil {
		return nil, err
	}
	return []model.SourceReference{{Kind: "weather", ResourceID: resource.ResourceID}}, nil
}
