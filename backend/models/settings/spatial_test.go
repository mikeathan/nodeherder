package settings

import (
	"encoding/json"
	"math"
	"testing"
)

func spatialFixture() *SpatialWorkspace {
	return &SpatialWorkspace{
		Version: spatialWorkspaceVersion,
		Zones: []SpatialZone{{
			ID:       "desk",
			Name:     "Desk",
			Kind:     zoneKindDetection,
			Shape:    zoneShapeRectangle,
			Points:   []SpatialPoint{{X: -500, Y: 1000}, {X: 500, Y: 1000}, {X: 500, Y: 2000}, {X: -500, Y: 2000}},
			Entities: map[string]string{coordinateRoleBeginX: "a"},
		}},
		EntryEntities: map[string]string{entryRoleAssumed: "active"},
	}
}
func TestSpatialWorkspaceValidation(t *testing.T) {
	cases := []struct {
		name   string
		change func(*SpatialWorkspace)
	}{
		{"version", func(w *SpatialWorkspace) { w.Version = spatialWorkspaceVersion + 1 }},
		{"nan", func(w *SpatialWorkspace) { w.Sensor.X = math.NaN() }},
		{"fov", func(w *SpatialWorkspace) { w.Sensor.FOV = maxSpatialFOV + 1 }},
		{"degenerate", func(w *SpatialWorkspace) { w.Zones[0].Points[1] = w.Zones[0].Points[0] }},
		{"self crossing", func(w *SpatialWorkspace) {
			w.Zones[0].Shape = zoneShapePolygon
			w.Zones[0].Points[1], w.Zones[0].Points[2] = w.Zones[0].Points[2], w.Zones[0].Points[1]
		}},
		{"unknown role", func(w *SpatialWorkspace) { w.EntryEntities["other"] = "x" }},
		{"unknown coordinate", func(w *SpatialWorkspace) { w.Zones[0].Entities["unknown"] = "x" }},
		{"too many zones", func(w *SpatialWorkspace) {
			for i := 0; i < 5; i++ {
				w.Zones = append(w.Zones, w.Zones[0])
			}
		}},
	}
	if err := spatialFixture().Validate(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := spatialFixture()
			tc.change(w)
			if w.Validate() == nil {
				t.Fatal("invalid workspace accepted")
			}
		})
	}
}
func TestSpatialWorkspaceCloneAndLegacyRoundTrip(t *testing.T) {
	var legacy DeviceConfig
	if err := json.Unmarshal([]byte(`{"spatial":{"mappings":[],"scale":6000,"orientation":0}}`), &legacy); err != nil {
		t.Fatal(err)
	}
	if err := legacy.ValidateReportingSettings(); err != nil {
		t.Fatal(err)
	}
	original := spatialFixture()
	copied := cloneSpatialWorkspace(original)
	copied.Zones[0].Points[0].X = 99
	copied.EntryEntities["assumed"] = "other"
	copied.Zones[0].Entities["beginX"] = "other"
	if original.Zones[0].Points[0].X != -500 || original.EntryEntities["assumed"] != "active" || original.Zones[0].Entities["beginX"] != "a" {
		t.Fatal("clone shares mutable state")
	}
	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var restored SpatialWorkspace
	if err = json.Unmarshal(encoded, &restored); err != nil {
		t.Fatal(err)
	}
	if err = restored.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestSpatialPolygonZoneEntityRoles(t *testing.T) {
	polygon := func() *SpatialWorkspace {
		workspace := spatialFixture()
		workspace.Zones[0].Shape = zoneShapePolygon
		workspace.Zones[0].Points = []SpatialPoint{{X: 0, Y: 1000}, {X: 1000, Y: 1000}, {X: 500, Y: 2000}}
		workspace.Zones[0].Entities = map[string]string{zoneRolePolygon: "polygon_zone_1"}
		return workspace
	}
	if err := polygon().Validate(); err != nil {
		t.Fatalf("polygon zone with a polygon entity rejected: %v", err)
	}

	rectangleRoleOnPolygon := polygon()
	rectangleRoleOnPolygon.Zones[0].Entities = map[string]string{coordinateRoleBeginX: "begin_x"}
	if rectangleRoleOnPolygon.Validate() == nil {
		t.Fatal("polygon zone accepted a rectangle coordinate role")
	}

	polygonRoleOnRectangle := spatialFixture()
	polygonRoleOnRectangle.Zones[0].Entities = map[string]string{zoneRolePolygon: "polygon_zone_1"}
	if polygonRoleOnRectangle.Validate() == nil {
		t.Fatal("rectangle zone accepted a polygon role")
	}

	tooManyPolygonEntities := polygon()
	tooManyPolygonEntities.Zones[0].Entities[coordinateRoleEndX] = "end_x"
	if tooManyPolygonEntities.Validate() == nil {
		t.Fatal("polygon zone accepted more than one entity")
	}
}
