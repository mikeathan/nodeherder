package services

import (
	"math"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/settings"
	"node-herder/utils"
)

// BuildSpatialTargets converts saved mapping rows and accepted device state into frame
// targets. A row is kept only when it names three distinct entities and its active flag is
// readable. Axes are attached only for an active slot: a cleared slot must not ship a stale
// position, and a consumer must not have to guess whether coordinates still apply. Values
// are never converted, clamped or rounded: the frame carries the entity's own unit, and the
// consumer owns display conversion.
//
// The lookup must read accepted state. This function performs no locking of its own and
// must be called on the device's lane, like the rest of the lifetime update path.
func BuildSpatialTargets(mappings []settings.SpatialMapping, lookup func(string) (*devices.Entity, bool)) []devices.SpatialTarget {
	targets := make([]devices.SpatialTarget, 0, len(mappings))
	for index, mapping := range mappings {
		if mapping.X == "" || mapping.Y == "" || mapping.Active == "" {
			continue
		}
		if mapping.X == mapping.Y || mapping.Active == mapping.X || mapping.Active == mapping.Y {
			continue
		}
		active, ok := spatialActiveFlag(lookup, mapping.Active)
		if !ok {
			continue
		}
		target := devices.SpatialTarget{ID: index + 1, Active: active}
		if active {
			target.X = spatialAxis(lookup, mapping.X)
			target.Y = spatialAxis(lookup, mapping.Y)
		}
		targets = append(targets, target)
	}
	return targets
}

// spatialAxis returns a usable reading for one mapped coordinate, or nil when the entity is
// missing, not numeric, declares an unknown length unit, or holds a non-finite value.
func spatialAxis(lookup func(string) (*devices.Entity, bool), name string) *devices.SpatialAxis {
	entity, ok := lookup(name)
	if !ok || entity == nil || entity.Type != bridge.NumericDataType || !devices.IsSpatialLengthUnit(entity.Unit) {
		return nil
	}
	value, ok := utils.ConvertToFloat64(entity.Data.Value())
	if !ok || math.IsNaN(value) || math.IsInf(value, 0) {
		return nil
	}
	return &devices.SpatialAxis{Value: value, Unit: entity.Unit}
}

// spatialActiveFlag reads one mapped active flag. A missing, non-binary or unreadable entity
// makes the whole row unusable, because a position without a trustworthy active state could be
// shown as a live target after the sensor stopped tracking it.
func spatialActiveFlag(lookup func(string) (*devices.Entity, bool), name string) (bool, bool) {
	entity, ok := lookup(name)
	if !ok || entity == nil || entity.Type != bridge.BinaryDataType {
		return false, false
	}
	return utils.ConvertToBool(entity.Data.Value())
}
