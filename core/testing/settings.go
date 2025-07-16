package utils_test

import (
	"node-herder/mocks"
	"node-herder/models/settings"
	"node-herder/utils"
	"testing"
	"time"
)

func CreateDebouncer(id string) *settings.DeviceDebouncer {
	repo := mocks.NopSettingsrepo{}
	app := settings.NewAppConfig()
	cache := settings.NewDeviceConfigCache(&repo, app)
	return settings.NewDeviceDebouncer(id, cache, mocks.NewMockClock(func() time.Time { return time.Now() }))
}

func CreateDebouncerFromAppConfig(id string, appConfig *settings.AppConfig, clock utils.Clock) *settings.DeviceDebouncer {
	repo := mocks.NopSettingsrepo{}
	cache := settings.NewDeviceConfigCache(&repo, appConfig)
	return settings.NewDeviceDebouncer(id, cache, clock)
}

func CreateDashboardGroups() map[string]*settings.DashboardGroup {
	groups := map[string]*settings.DashboardGroup{}

	// Group 1
	group1 := settings.NewDashboardGroup("group1")
	group1.Name = "group1"
	group1.DeviceGroup = map[string]*settings.DeviceGroup{}

	g1d1 := settings.NewDeviceGroup("DeviceId1")
	g1d1.DeviceId = "DeviceId1"
	g1d1.Exposes = []string{"temperature", "humidity", "battery"}

	g1d2 := settings.NewDeviceGroup("DeviceId2")
	g1d2.DeviceId = "DeviceId2"
	g1d2.Exposes = []string{"alarm", "silence alarm", "battery"}

	g1d3 := settings.NewDeviceGroup("DeviceId3")
	g1d3.DeviceId = "DeviceId3"
	g1d3.Exposes = []string{"contact", "battery"}

	g1d4 := settings.NewDeviceGroup("DeviceId4")
	g1d4.DeviceId = "DeviceId4"
	g1d4.Exposes = []string{"presence", "illuminance", "battery"}

	group1.DeviceGroup["DeviceId1"] = g1d1
	group1.DeviceGroup["DeviceId2"] = g1d2
	group1.DeviceGroup["DeviceId3"] = g1d3
	group1.DeviceGroup["DeviceId4"] = g1d4

	// Group 2
	group2 := settings.NewDashboardGroup("group2")
	group2.Name = "group2"
	group2.DeviceGroup = map[string]*settings.DeviceGroup{}

	group2.DeviceGroup["DeviceId5"] = settings.NewDeviceGroup("DeviceId5")
	group2.DeviceGroup["DeviceId5"].DeviceId = "DeviceId5"
	group2.DeviceGroup["DeviceId5"].Exposes = []string{"temperature", "humidity", "battery"}

	group2.DeviceGroup["DeviceId6"] = settings.NewDeviceGroup("DeviceId6")
	group2.DeviceGroup["DeviceId6"].DeviceId = "DeviceId6"
	group2.DeviceGroup["DeviceId6"].Exposes = []string{"alarm", "silence alarm", "battery"}

	group2.DeviceGroup["DeviceId7"] = settings.NewDeviceGroup("DeviceId7")
	group2.DeviceGroup["DeviceId7"].DeviceId = "DeviceId7"
	group2.DeviceGroup["DeviceId7"].Exposes = []string{"contact", "battery"}

	// Assign groups to map
	groups["group1"] = group1
	groups["group2"] = group2
	return groups
}

func CompareDashboardGroups(t *testing.T, gotDashboardGroups map[string]*settings.DashboardGroup, wantDashboardGroups map[string]*settings.DashboardGroup) {

	if gotDashboardGroups == nil {
		t.Fatal("Expected dashboard groups, got nil")
	}

	if len(gotDashboardGroups) != len(wantDashboardGroups) {
		t.Fatalf("Expected dashboard groups %v', got '%v'", len(wantDashboardGroups), len(gotDashboardGroups))
	}

	for _, wantGroup := range wantDashboardGroups {
		gotDashboardGroup, ok := gotDashboardGroups[wantGroup.Name]
		if !ok {
			t.Fatalf("Expected dashboard group %v', got '%v'", wantGroup.Name, gotDashboardGroup)
		}
		if gotDashboardGroup.Name != wantGroup.Name {
			t.Fatalf("Expected dashboard group name %v', got '%v'", wantGroup.Name, gotDashboardGroup.Name)
		}
		for _, deviceGroup := range wantGroup.DeviceGroup {
			group, ok := gotDashboardGroup.DeviceGroup[deviceGroup.DeviceId]
			if !ok {
				t.Fatalf("Expected device group id %v', got '%v'", deviceGroup.DeviceId, group.DeviceId)
			}

			if group.DeviceId != deviceGroup.DeviceId {
				t.Fatalf("Expected device group id %v', got '%v'", deviceGroup.DeviceId, group.DeviceId)
			}

			for idx := range deviceGroup.Exposes {
				if group.Exposes[idx] != deviceGroup.Exposes[idx] {
					t.Fatalf("Expected expose %v', got '%v'", group.Exposes[idx], deviceGroup.Exposes[idx])
				}
			}
		}
	}
}
