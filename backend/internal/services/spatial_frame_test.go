package services_test

import (
	"math"
	"testing"

	"node-herder/internal/services"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/settings"
)

func spatialTestEntity(name string, dataType bridge.ExposeDataType, unit string, value any) *devices.Entity {
	entity := devices.NewEntity(name)
	entity.Type = dataType
	entity.Unit = unit
	entity.AccessMode = bridge.ReadAccessMode
	entity.Data.SetValue(value)
	return entity
}

func spatialLookup(entities ...*devices.Entity) func(string) (*devices.Entity, bool) {
	byName := make(map[string]*devices.Entity, len(entities))
	for _, entity := range entities {
		byName[entity.Name] = entity
	}
	return func(name string) (*devices.Entity, bool) {
		entity, ok := byName[name]
		return entity, ok
	}
}

func TestBuildSpatialTargetsReportsActiveSlot(t *testing.T) {
	lookup := spatialLookup(
		spatialTestEntity("target_x", bridge.NumericDataType, "mm", float64(1200)),
		spatialTestEntity("target_y", bridge.NumericDataType, "mm", float64(3400)),
		spatialTestEntity("target_active", bridge.BinaryDataType, "", true),
	)
	targets := services.BuildSpatialTargets([]settings.SpatialMapping{{X: "target_x", Y: "target_y", Active: "target_active"}}, lookup)
	if len(targets) != 1 {
		t.Fatalf("targets = %v", targets)
	}
	target := targets[0]
	if target.ID != 1 || !target.Active {
		t.Fatalf("target identity = %+v", target)
	}
	if target.X == nil || target.X.Value != 1200 || target.X.Unit != "mm" {
		t.Fatalf("x axis = %+v", target.X)
	}
	if target.Y == nil || target.Y.Value != 3400 {
		t.Fatalf("y axis = %+v", target.Y)
	}
}

func TestBuildSpatialTargetsKeepsRowOrderAsSlot(t *testing.T) {
	lookup := spatialLookup(
		spatialTestEntity("x1", bridge.NumericDataType, "mm", float64(1)),
		spatialTestEntity("y1", bridge.NumericDataType, "mm", float64(2)),
		spatialTestEntity("a1", bridge.BinaryDataType, "", true),
		spatialTestEntity("x3", bridge.NumericDataType, "cm", float64(3)),
		spatialTestEntity("y3", bridge.NumericDataType, "cm", float64(4)),
		spatialTestEntity("a3", bridge.BinaryDataType, "", true),
	)
	targets := services.BuildSpatialTargets([]settings.SpatialMapping{
		{X: "x1", Y: "y1", Active: "a1"},
		{X: "", Y: "", Active: ""},
		{X: "x3", Y: "y3", Active: "a3"},
	}, lookup)
	if len(targets) != 2 || targets[0].ID != 1 || targets[1].ID != 3 {
		t.Fatalf("slot identifiers = %+v", targets)
	}
	if targets[1].X.Unit != "cm" || targets[1].X.Value != 3 {
		t.Fatalf("units were converted: %+v", targets[1].X)
	}
}

func TestBuildSpatialTargetsOmitsUnusableRowsAndAxes(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		mapping  settings.SpatialMapping
		entities []*devices.Entity
		want     int
		wantAxis bool
	}{
		{
			name:    "incomplete mapping",
			mapping: settings.SpatialMapping{X: "x", Y: "y"},
			entities: []*devices.Entity{
				spatialTestEntity("x", bridge.NumericDataType, "mm", float64(1)),
				spatialTestEntity("y", bridge.NumericDataType, "mm", float64(2)),
			},
		},
		{
			name:    "same axis twice",
			mapping: settings.SpatialMapping{X: "x", Y: "x", Active: "a"},
			entities: []*devices.Entity{
				spatialTestEntity("x", bridge.NumericDataType, "mm", float64(1)),
				spatialTestEntity("a", bridge.BinaryDataType, "", true),
			},
		},
		{
			name:    "missing active entity",
			mapping: settings.SpatialMapping{X: "x", Y: "y", Active: "missing"},
			entities: []*devices.Entity{
				spatialTestEntity("x", bridge.NumericDataType, "mm", float64(1)),
				spatialTestEntity("y", bridge.NumericDataType, "mm", float64(2)),
			},
		},
		{
			name:    "active entity is not binary",
			mapping: settings.SpatialMapping{X: "x", Y: "y", Active: "a"},
			entities: []*devices.Entity{
				spatialTestEntity("x", bridge.NumericDataType, "mm", float64(1)),
				spatialTestEntity("y", bridge.NumericDataType, "mm", float64(2)),
				spatialTestEntity("a", bridge.NumericDataType, "", float64(1)),
			},
		},
		{
			name:    "cleared slot ships no axes",
			mapping: settings.SpatialMapping{X: "x", Y: "y", Active: "a"},
			entities: []*devices.Entity{
				spatialTestEntity("x", bridge.NumericDataType, "mm", float64(1)),
				spatialTestEntity("y", bridge.NumericDataType, "mm", float64(2)),
				spatialTestEntity("a", bridge.BinaryDataType, "", false),
			},
			want: 1,
		},
		{
			name:    "unknown length unit",
			mapping: settings.SpatialMapping{X: "x", Y: "y", Active: "a"},
			entities: []*devices.Entity{
				spatialTestEntity("x", bridge.NumericDataType, "ft", float64(1)),
				spatialTestEntity("y", bridge.NumericDataType, "mm", float64(2)),
				spatialTestEntity("a", bridge.BinaryDataType, "", true),
			},
			want: 1,
		},
		{
			name:    "non-finite coordinate",
			mapping: settings.SpatialMapping{X: "x", Y: "y", Active: "a"},
			entities: []*devices.Entity{
				spatialTestEntity("x", bridge.NumericDataType, "mm", math.NaN()),
				spatialTestEntity("y", bridge.NumericDataType, "mm", float64(2)),
				spatialTestEntity("a", bridge.BinaryDataType, "", true),
			},
			want: 1,
		},
		{
			name:    "usable slot",
			mapping: settings.SpatialMapping{X: "x", Y: "y", Active: "a"},
			entities: []*devices.Entity{
				spatialTestEntity("x", bridge.NumericDataType, "mm", float64(1)),
				spatialTestEntity("y", bridge.NumericDataType, "mm", float64(2)),
				spatialTestEntity("a", bridge.BinaryDataType, "", true),
			},
			want:     1,
			wantAxis: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			targets := services.BuildSpatialTargets([]settings.SpatialMapping{testCase.mapping}, spatialLookup(testCase.entities...))
			if len(targets) != testCase.want {
				t.Fatalf("targets = %+v, want %d", targets, testCase.want)
			}
			if testCase.want == 0 {
				return
			}
			if hasAxis := targets[0].X != nil && targets[0].Y != nil; hasAxis != testCase.wantAxis {
				t.Fatalf("axes present = %t, want %t (%+v)", hasAxis, testCase.wantAxis, targets[0])
			}
		})
	}
}
