package waste

// Validates provider field mappings used to interpret bin collection responses.

import (
	"bytes"
	"encoding/json"

	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
)

// Mapping converts provider JSON field locations into one typed waste schedule.
type Mapping struct {
	Metadata    MetadataMapping   `json:"metadata" yaml:"metadata"`
	Collections CollectionMapping `json:"collections" yaml:"collections"`
}

// MetadataMapping holds schedule-level JSON Pointers relative to the selected result root.
// These fields preserve provenance and quality; they do not manufacture missing evidence.
type MetadataMapping struct {
	PropertyID          string `json:"property_id" yaml:"property_id"`
	SourceRetrievedAt   string `json:"source_retrieved_at" yaml:"source_retrieved_at"`
	RefreshDueAt        string `json:"refresh_due_at" yaml:"refresh_due_at"`
	DateConfidence      string `json:"date_confidence" yaml:"date_confidence"`
	DateCompleteness    string `json:"date_completeness" yaml:"date_completeness"`
	AddressMatch        string `json:"address_match" yaml:"address_match"`
	EvidenceGranularity string `json:"evidence_granularity" yaml:"evidence_granularity"`
	Provisional         string `json:"provisional" yaml:"provisional"`
	Notes               string `json:"notes,omitempty" yaml:"notes,omitempty"`
	SourceURL           string `json:"source_url,omitempty" yaml:"source_url,omitempty"`
}

// CollectionMapping locates the collection list once; Fields applies to every returned service.
type CollectionMapping struct {
	Items  string           `json:"items" yaml:"items"`
	Fields CollectionFields `json:"fields" yaml:"fields"`
}

// CollectionFields names JSON locations relative to each collection, independently of bin colour.
type CollectionFields struct {
	Name                 string `json:"name" yaml:"name"`
	Dates                string `json:"dates" yaml:"dates"`
	Colour               string `json:"colour,omitempty" yaml:"colour,omitempty"`
	WasteType            string `json:"waste_type,omitempty" yaml:"waste_type,omitempty"`
	Container            string `json:"container,omitempty" yaml:"container,omitempty"`
	DatesComplete        string `json:"dates_complete,omitempty" yaml:"dates_complete,omitempty"`
	SubscriptionRequired string `json:"subscription_required,omitempty" yaml:"subscription_required,omitempty"`
	SourceServiceKey     string `json:"source_service_key,omitempty" yaml:"source_service_key,omitempty"`
}

// decodeMapping keeps unknown/misspelled domain fields out of active provider definitions.
func decodeMapping(raw map[string]any) (Mapping, error) {
	var mapping Mapping
	data, err := json.Marshal(raw)
	if err != nil {
		return mapping, model.InvalidDefinition("mapping", "Use JSON-compatible mapping values.")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&mapping); err != nil {
		return mapping, model.InvalidDefinition("mapping", "Unknown or invalid waste mapping field.")
	}
	for path, pointer := range map[string]string{
		"collections.items": mapping.Collections.Items, "collections.fields.name": mapping.Collections.Fields.Name,
		"collections.fields.dates": mapping.Collections.Fields.Dates, "metadata.property_id": mapping.Metadata.PropertyID,
		"metadata.address_match": mapping.Metadata.AddressMatch, "metadata.evidence_granularity": mapping.Metadata.EvidenceGranularity,
	} {
		if pointer == "" || !operations.ValidPointer(pointer) {
			return mapping, model.InvalidDefinition("mapping."+path, "Use a non-empty JSON Pointer.")
		}
	}
	var tree any
	_ = json.Unmarshal(data, &tree)
	if err = validateSelectors(tree, "mapping"); err != nil {
		return mapping, err
	}
	return mapping, nil
}

// validateSelectors rejects non-pointer values in the nested mapping tree.
func validateSelectors(value any, path string) error {
	switch item := value.(type) {
	case string:
		if !operations.ValidPointer(item) {
			return model.InvalidDefinition(path, "Use JSON Pointers for provider fields.")
		}
	case map[string]any:
		for key, child := range item {
			if err := validateSelectors(child, path+"."+key); err != nil {
				return err
			}
		}
	case nil:
		return model.InvalidDefinition(path, "Mapping fields cannot be null.")
	default:
		return model.InvalidDefinition(path, "Use field locations, not computed values.")
	}
	return nil
}
