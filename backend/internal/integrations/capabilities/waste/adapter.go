package waste

import (
	"errors"
	"time"

	"node-herder/internal/integrations/capabilities"
	model "node-herder/models/integrations"
)

// Adapter implements the waste contract without owning execution, storage or credentials.
// Its zero value is immutable and ready for concurrent use by integration workers.
type Adapter struct{}

// Descriptor supplies waste discovery metadata; no catalogue branch is needed in runtime.
func (Adapter) Descriptor() capabilities.Descriptor {
	return capabilities.Descriptor{Capability: model.WasteCapability, IntegrationType: "waste_collection", Name: "Bin collections", Description: "Collection dates for your address."}
}

// ValidateMapping rejects unknown fields and invalid/missing waste selectors before activation.
func (Adapter) ValidateMapping(raw map[string]any) error {
	_, err := decodeMapping(raw)
	return err
}

// Normalize converts a provider result using waste-specific evidence and identity rules.
// Failed normalization never mutates previous resources or the supplied identity map.
func (Adapter) Normalize(root any, raw map[string]any, c capabilities.Context) (capabilities.Result, error) {
	var out capabilities.Result
	mapping, err := decodeMapping(raw)
	if err != nil {
		return out, err
	}
	var previous *model.WasteSchedule
	if c.Previous != nil {
		decoded, err := decodeSchedule(*c.Previous)
		if err != nil {
			return out, err
		}
		previous = &decoded
	}
	result, err := NormalizeWaste(root, mapping, WasteContext{
		NewIdentity: c.NewIdentity, InstanceID: c.InstanceID, ResourceID: c.ResourceID,
		ProviderID: c.ProviderID, ProviderVersion: c.ProviderVersion,
		ProfileRevision: c.ProfileRevision, BindingRevision: c.BindingRevision,
		SelectedProperty: c.SelectedValue, Now: c.Now, Previous: previous, ServiceIDs: c.Identities,
	})
	if err != nil {
		return out, err
	}
	out.Resource, err = model.NewResource(result.Snapshot)
	if err != nil {
		return out, err
	}
	out.Identities = result.ServiceIDs
	out.Reservations = []capabilities.Identity{{ID: result.Snapshot.ResourceID, Kind: "resource"}}
	for _, service := range result.Snapshot.Services {
		out.Reservations = append(out.Reservations, capabilities.Identity{ID: service.ServiceID, Kind: "service"})
	}
	out.Renew = horizonExhausted(result.Snapshot, c.Now)
	return out, nil
}

// Project validates stored waste data and derives current ordering/freshness without writes.
func (Adapter) Project(resource model.Resource, now time.Time) (model.Resource, error) {
	schedule, err := decodeSchedule(resource)
	if err != nil {
		return model.Resource{}, err
	}
	return model.NewResource(schedule.Project(now))
}

// Checked advances only local check time after HTTP 304, preserving source age/content revision.
func (Adapter) Checked(resource model.Resource, now time.Time) (model.Resource, error) {
	schedule, err := decodeSchedule(resource)
	if err != nil {
		return model.Resource{}, err
	}
	schedule.LastCheckedAt = now.UTC()
	return model.NewResource(schedule)
}

// RenewAfter allows a bounded refresh renewal when the provider's cached waste horizon is empty.
func (Adapter) RenewAfter(err error) bool {
	var failure *model.Failure
	return errors.As(err, &failure) && failure.Code == "no_collection_dates"
}

// decodeSchedule guards persisted/domain boundaries independently of provider normalization.
// Invalid saved payloads remain in storage for recovery and cannot turn into empty schedules.
func decodeSchedule(resource model.Resource) (model.WasteSchedule, error) {
	var schedule model.WasteSchedule
	err := resource.Decode(&schedule)
	if err != nil || schedule.ContractVersion != model.WasteCapability || schedule.Timezone != model.WasteTimezone || len(schedule.Services) == 0 || len(schedule.Services) > 64 {
		return schedule, model.Fail("invalid_resource", "The saved collection data needs review. Refresh or check its provider configuration.")
	}
	seen := map[string]bool{schedule.ResourceID: true}
	for _, service := range schedule.Services {
		if service.ServiceID == "" || seen[service.ServiceID] || len(service.Dates) > 730 {
			return schedule, model.Fail("invalid_resource", "The saved collection services need review.")
		}
		seen[service.ServiceID] = true
		for _, date := range service.Dates {
			if !localDate(date) {
				return schedule, model.Fail("invalid_resource", "The saved collection dates need review.")
			}
		}
	}
	return schedule, nil
}
