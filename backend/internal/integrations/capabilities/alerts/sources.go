package alerts

// Exposes a validated alert snapshot as a Home placement source.

import model "node-herder/models/integrations"

// Sources exposes the snapshot's stable placement reference after domain validation.
func (Adapter) Sources(resource model.Resource) ([]model.SourceReference, error) {
	if _, err := decodeAlerts(resource); err != nil {
		return nil, err
	}
	return []model.SourceReference{{Kind: "weather_alerts", ResourceID: resource.ResourceID}}, nil
}
