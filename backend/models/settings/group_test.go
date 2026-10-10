package settings_test

import (
	"encoding/json"
	"strings"
	"testing"

	"node-herder/models/settings"
)

// Spec 007 FR-09: dashboard groups carry an optional display order. Absent stays absent
// (older data and older frontends), zero is a real position.
func TestDashboardGroupOrderJSON(t *testing.T) {
	var unordered settings.DashboardGroup
	if err := json.Unmarshal([]byte(`{"name":"Kitchen","deviceGroup":{}}`), &unordered); err != nil {
		t.Fatal(err)
	}
	if unordered.Order != nil {
		t.Fatalf("missing order must decode as nil, got %v", *unordered.Order)
	}
	raw, err := json.Marshal(unordered)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "order") {
		t.Fatalf("absent order must not be emitted, got %s", raw)
	}

	var first settings.DashboardGroup
	if err := json.Unmarshal([]byte(`{"name":"Attic","deviceGroup":{},"order":0}`), &first); err != nil {
		t.Fatal(err)
	}
	if first.Order == nil || *first.Order != 0 {
		t.Fatalf("order 0 must be kept, got %v", first.Order)
	}
	raw, err = json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"order":0`) {
		t.Fatalf("order 0 must be emitted, got %s", raw)
	}
}

func TestDashboardGroupOrderIsSavedAndSnapshotOwned(t *testing.T) {
	repo := &bridgeSnapshotRepo{bridge: settings.DefaultBridgeConfig()}
	cache, err := settings.NewAppConfigCache(repo, nil)
	if err != nil {
		t.Fatal(err)
	}
	two := 2
	if err := cache.SaveDashboardGroup(&settings.DashboardGroup{Name: "living", DeviceGroup: map[string]*settings.DeviceGroup{}, Order: &two}); err != nil {
		t.Fatal(err)
	}
	snapshot, err := cache.LoadAppConfig()
	if err != nil {
		t.Fatal(err)
	}
	got := snapshot.Hub.DashboardGroups["living"].Order
	if got == nil || *got != 2 {
		t.Fatalf("saved order lost, got %v", got)
	}
	// mutating the snapshot must not reach the cache
	*got = 9
	again, err := cache.LoadAppConfig()
	if err != nil {
		t.Fatal(err)
	}
	if *again.Hub.DashboardGroups["living"].Order != 2 {
		t.Fatal("snapshot shares the order pointer with the cache")
	}
	if _, err := cache.RenameDashboardGroup("living", "lounge"); err != nil {
		t.Fatal(err)
	}
	renamed, _ := cache.LoadAppConfig()
	if o := renamed.Hub.DashboardGroups["lounge"].Order; o == nil || *o != 2 {
		t.Fatalf("rename must keep order, got %v", o)
	}
}
