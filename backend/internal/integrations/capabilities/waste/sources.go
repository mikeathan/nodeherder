package waste

// Exposes validated collection schedules and services as Home placement sources.

import model "node-herder/models/integrations"

// Sources validates the waste resource and exposes stable summary/service references.
func (Adapter) Sources(resource model.Resource) ([]model.SourceReference, error) {
	schedule, err := decodeSchedule(resource)
	if err != nil {
		return nil, err
	}
	result := []model.SourceReference{{Kind: "waste_summary", ResourceID: resource.ResourceID}}
	for _, service := range schedule.Services {
		result = append(result, model.SourceReference{Kind: "waste_service", ResourceID: resource.ResourceID, ServiceID: service.ServiceID})
	}
	return result, nil
}
