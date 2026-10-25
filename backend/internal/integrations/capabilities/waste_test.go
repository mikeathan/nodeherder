package capabilities

import (
	"encoding/json"
	model "node-herder/models/integrations"
	"testing"
	"time"
)

func wasteFixture() map[string]any {
	return map[string]any{"property_id": "p:test", "address_match": "exact", "evidence_granularity": "property", "retrieved_at": "2026-10-23T12:00:00Z", "refresh_due_at": "2026-10-24T12:00:00Z", "date_completeness": "limited_horizon", "collections": []any{map[string]any{"name": "Paper", "waste_type": "paper_card", "container": "wheelie_bin", "dates": []any{"2026-10-27", "2026-10-27", "2026-10-26"}, "dates_complete": true}}}
}
func wasteMapping() model.Mapping {
	return model.Mapping{Metadata: model.MetadataMapping{PropertyID: "/property_id", AddressMatch: "/address_match", EvidenceGranularity: "/evidence_granularity", Provisional: "/provisional", SourceRetrievedAt: "/retrieved_at", RefreshDueAt: "/refresh_due_at", DateCompleteness: "/date_completeness"}, Collections: model.CollectionMapping{Items: "/collections", Fields: model.CollectionFields{Name: "/name", Dates: "/dates", WasteType: "/waste_type", Container: "/container", Colour: "/bin_colour", DatesComplete: "/dates_complete"}}}
}
func TestNormalizeWaste(t *testing.T) {
	now := time.Date(2026, 10, 23, 13, 0, 0, 0, time.UTC)
	c := WasteContext{InstanceID: "instance", ProviderID: "provider", ProfileRevision: 1, BindingRevision: 1, SelectedProperty: "p:test", Now: now}
	r, e := NormalizeWaste(wasteFixture(), wasteMapping(), c)
	if e != nil {
		t.Fatal(e)
	}
	s := r.Snapshot.Services[0]
	if len(s.Dates) != 2 || s.Dates[0] != "2026-10-26" || s.Coverage.From != nil || s.Subscription != "unknown" || r.Snapshot.Freshness != "fresh" {
		t.Fatal(s)
	}
	c.ServiceIDs = r.ServiceIDs
	c.Previous = &r.Snapshot
	c.ResourceID = r.Snapshot.ResourceID
	f := wasteFixture()
	f["collections"].([]any)[0].(map[string]any)["bin_colour"] = "blue"
	r2, e := NormalizeWaste(f, wasteMapping(), c)
	if e != nil || r2.Snapshot.Services[0].ServiceID != s.ServiceID {
		t.Fatal("colour changed identity", e)
	}
	f["collections"].([]any)[0].(map[string]any)["name"] = "Renamed"
	r3, e := NormalizeWaste(f, wasteMapping(), c)
	if e != nil || len(r3.Snapshot.Services) != 2 || r3.Snapshot.Kind != "partial" || r3.Snapshot.Services[1].DateState != "unavailable" {
		t.Fatal("lost old service", e)
	}
	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
	}{{"wrong property", func(f map[string]any) { f["property_id"] = "other" }}, {"provisional", func(f map[string]any) { f["provisional"] = true }}, {"nonexact", func(f map[string]any) { f["address_match"] = "postcode" }}, {"invalid date", func(f map[string]any) { f["collections"].([]any)[0].(map[string]any)["dates"] = []any{"2026-02-30"} }}, {"duplicate service", func(f map[string]any) { list := f["collections"].([]any); f["collections"] = append(list, list[0]) }}, {"empty services", func(f map[string]any) { f["collections"] = []any{} }}} {
		t.Run(tc.name, func(t *testing.T) {
			f := wasteFixture()
			tc.mutate(f)
			if _, e := NormalizeWaste(f, wasteMapping(), c); e == nil {
				b, _ := json.Marshal(f)
				t.Fatalf("accepted %s", b)
			}
		})
	}
}
