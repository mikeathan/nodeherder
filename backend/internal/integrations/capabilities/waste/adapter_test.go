package waste

// Checks the bin collection adapter and its resource boundaries.

import (
	"encoding/json"
	"testing"
	"time"

	"node-herder/internal/integrations/capabilities"
	model "node-herder/models/integrations"
)

func TestAdapterProjectionAnd304PreserveSourceEvidence(t *testing.T) {
	now := time.Date(2026, 10, 23, 13, 0, 0, 0, time.UTC)
	encoded, _ := json.Marshal(wasteMapping())
	var mapping map[string]any
	_ = json.Unmarshal(encoded, &mapping)
	adapter := Adapter{}
	result, err := adapter.Normalize(wasteFixture(), mapping, capabilities.Context{InstanceID: "instance", SelectedValue: "p:test", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	checked, err := adapter.Checked(result.Resource, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	projected, err := adapter.Project(checked, now.Add(48*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	schedule, err := decodeSchedule(projected)
	if err != nil {
		t.Fatal(err)
	}
	if projected.ResourceID != result.Resource.ResourceID || projected.Revision != result.Resource.Revision || !schedule.LastCheckedAt.Equal(now.Add(time.Hour)) || schedule.SourceRetrievedAt == nil || !schedule.SourceRetrievedAt.Equal(now.Add(-time.Hour)) || schedule.Freshness != "stale" {
		t.Fatal(schedule)
	}
	for _, raw := range []string{`{"contract_version":"waste_schedule/v1","resource_id":"r","revision":1,"timezone":"Europe/London","services":[]}`, `{"contract_version":"other/v1","resource_id":"r","revision":1}`} {
		var resource model.Resource
		_ = json.Unmarshal([]byte(raw), &resource)
		if _, err = adapter.Project(resource, now); err == nil {
			t.Fatal("corrupt resource projected")
		}
	}
}

func TestAdapterStrictMapping(t *testing.T) {
	encoded, _ := json.Marshal(wasteMapping())
	var mapping map[string]any
	_ = json.Unmarshal(encoded, &mapping)
	mapping["unexpected"] = map[string]any{"field": "/value"}
	if err := (Adapter{}).ValidateMapping(mapping); err == nil {
		t.Fatal("unknown mapping accepted")
	}
}
