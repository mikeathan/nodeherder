package settings

// This file defines the versioned, per-device spatial workspace that the UI
// persists with device settings. It is visual configuration only: it never
// represents firmware truth, tracking history or a second device state store.

import (
	"fmt"
	"math"
	"strings"
)

const (
	spatialWorkspaceVersion = 1

	maxSpatialZones            = 8
	maxSpatialRoomObjects      = 40
	maxSpatialEntryEntities    = 5
	maxSpatialZoneEntities     = 5
	maxSpatialFirmwareEntities = 5
	maxDetectionZones          = 4
	maxExclusionZones          = 2
	maxEntryZones              = 2
	minSpatialPolygonPoints    = 3
	maxSpatialPolygonPoints    = 16
	spatialCoordinateLimitMM   = 20000
	maxSpatialRotation         = 360
	maxSpatialFOV              = 180
	maxSpatialNameRunes        = 80
	maxSpatialIDBytes          = 80
	maxSpatialEntityBytes      = 256
	// The shoelace sum is twice the area in square millimetres; 1 rejects
	// zero-area and near-degenerate polygons without affecting normal zones.
	minSpatialTwiceArea = 1
)

const (
	zoneKindDetection = "detection"
	zoneKindExclusion = "exclusion"
	zoneKindEntry     = "entry"

	zoneShapeRectangle = "rectangle"
	zoneShapePolygon   = "polygon"

	roomObjectWall      = "wall"
	roomObjectDoor      = "door"
	roomObjectFurniture = "furniture"

	entryRoleEnabled   = "enabled"
	entryRoleTimeout   = "timeout"
	entryRoleThreshold = "threshold"
	entryRoleAssumed   = "assumed"
	entryRoleRemaining = "remaining"

	// Firmware setting roles a workspace may map explicitly. They are ordinary writable device
	// entities; nothing is inferred from entity names.
	firmwareRoleInstallationAngle = "installationAngle"
	firmwareRoleInverseMounting   = "inverseMounting"
	firmwareRoleMaxDistance       = "maxDistance"
	firmwareRoleOccupancyOffDelay = "occupancyOffDelay"
	firmwareRolePolygonMode       = "polygonMode"

	// zoneRolePolygon maps a polygon overlay to one firmware text entity; it is only valid for
	// polygon shapes, exactly as the four coordinate roles are rectangle-only.
	zoneRolePolygon      = "polygon"
	zoneRoleOffDelay     = "offDelay"
	coordinateRoleBeginX = "beginX"
	coordinateRoleBeginY = "beginY"
	coordinateRoleEndX   = "endX"
	coordinateRoleEndY   = "endY"
)

// SpatialPoint is a sensor-relative or room-relative position in millimetres.
type SpatialPoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// SpatialZone is a named sensor-relative rectangle or simple polygon overlay.
type SpatialZone struct {
	OccupancyEntity string            `json:"occupancyEntity,omitempty"`
	Entities        map[string]string `json:"entities,omitempty"`
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Kind            string            `json:"kind"`
	Shape           string            `json:"shape"`
	Points          []SpatialPoint    `json:"points"`
}

// SpatialRoomObject is a visual room item; dimensions are millimetres and rotation is degrees.
type SpatialRoomObject struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Kind     string  `json:"kind"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	Rotation float64 `json:"rotation"`
}

// SpatialSensor places sensor-relative coordinates in the room; zero FOV hides the illustrative cone.
type SpatialSensor struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Rotation float64 `json:"rotation"`
	FOV      float64 `json:"fov"`
}

// SpatialWorkspace is versioned per-device visual configuration, never firmware truth or tracking history.
type SpatialWorkspace struct {
	Version       int                 `json:"version"`
	Zones         []SpatialZone       `json:"zones"`
	Outline       []SpatialPoint      `json:"outline"`
	Objects       []SpatialRoomObject `json:"objects"`
	Sensor        SpatialSensor       `json:"sensor"`
	EntryEntities map[string]string   `json:"entryEntities"`
	// FirmwareEntities maps optional firmware setting roles to writable entity names.
	FirmwareEntities map[string]string `json:"firmwareEntities,omitempty"`
}

// finiteWithin reports whether v is a finite value within +/- limit.
func finiteWithin(v, limit float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && math.Abs(v) <= limit
}

// spatialCross returns the cross product of AB and AC, whose sign describes
// which side of AB point C lies on.
func spatialCross(a, b, c SpatialPoint) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}

// spatialSegmentsCross reports whether two segments touch or cross. The
// non-strict orientation tests include collinear overlaps and endpoint touches;
// bounding boxes exclude disjoint segments on the same line.
func spatialSegmentsCross(a, b, c, d SpatialPoint) bool {
	return spatialCross(a, b, c)*spatialCross(a, b, d) <= 0 &&
		spatialCross(c, d, a)*spatialCross(c, d, b) <= 0 &&
		math.Max(a.X, b.X) >= math.Min(c.X, d.X) &&
		math.Max(c.X, d.X) >= math.Min(a.X, b.X) &&
		math.Max(a.Y, b.Y) >= math.Min(c.Y, d.Y) &&
		math.Max(c.Y, d.Y) >= math.Min(a.Y, b.Y)
}

// validSpatialPolygon bounds vertex count and coordinate range, rejects
// duplicate adjacent vertices, and requires a non-self-intersecting shape.
func validSpatialPolygon(points []SpatialPoint) bool {
	if len(points) < minSpatialPolygonPoints || len(points) > maxSpatialPolygonPoints {
		return false
	}
	area := 0.0
	for i, a := range points {
		b := points[(i+1)%len(points)]
		if !finiteWithin(a.X, spatialCoordinateLimitMM) || !finiteWithin(a.Y, spatialCoordinateLimitMM) || a == b {
			return false
		}
		// Shoelace terms accumulate twice the signed area, independent of winding.
		area += a.X*b.Y - b.X*a.Y
		// Adjacent edges share a valid endpoint; test only non-adjacent pairs.
		for j := i + 2; j < len(points); j++ {
			if i == 0 && j == len(points)-1 {
				continue
			}
			if spatialSegmentsCross(a, b, points[j], points[(j+1)%len(points)]) {
				return false
			}
		}
	}
	return math.Abs(area) > minSpatialTwiceArea
}

// validSpatialIDName checks a bounded, unique identifier and a nonempty display name.
func validSpatialIDName(id, name string, ids map[string]bool) bool {
	if id == "" || len(id) > maxSpatialIDBytes || ids[id] || strings.TrimSpace(name) == "" || len([]rune(name)) > maxSpatialNameRunes {
		return false
	}
	ids[id] = true
	return true
}

// validSpatialEntityMap checks bounded role names and entity identifiers.
func validSpatialEntityMap(entities map[string]string, validRole func(string) bool) bool {
	for role, id := range entities {
		if !validRole(role) || len(id) > maxSpatialEntityBytes || strings.ContainsRune(id, '\x00') {
			return false
		}
	}
	return true
}

// validZoneEntities accepts role names that match the zone shape: at most one polygon text
// entity for polygon zones, and the four rectangle coordinates for rectangle zones. A role
// from the other shape is rejected instead of being silently ignored.
func validZoneEntities(zone SpatialZone) bool {
	if len(zone.Entities) > maxSpatialZoneEntities {
		return false
	}
	if zone.Shape == zoneShapePolygon {
		if len(zone.Entities) > 2 {
			return false
		}
		return validSpatialEntityMap(zone.Entities, func(role string) bool {
			return role == zoneRolePolygon || role == zoneRoleOffDelay
		})
	}
	return validSpatialEntityMap(zone.Entities, func(role string) bool {
		return validCoordinateRole(role) || role == zoneRoleOffDelay
	})
}

// validCoordinateRole reports whether role is one of the four rectangle coordinates.
func validCoordinateRole(role string) bool {
	switch role {
	case coordinateRoleBeginX, coordinateRoleBeginY, coordinateRoleEndX, coordinateRoleEndY:
		return true
	default:
		return false
	}
}

// validFirmwareRole reports whether role belongs to the optional firmware setting mapping.
func validFirmwareRole(role string) bool {
	switch role {
	case firmwareRoleInstallationAngle, firmwareRoleInverseMounting, firmwareRoleMaxDistance,
		firmwareRoleOccupancyOffDelay, firmwareRolePolygonMode:
		return true
	default:
		return false
	}
}

// validEntryRole reports whether role belongs to the entry/assumed-presence mapping.
func validEntryRole(role string) bool {
	switch role {
	case entryRoleEnabled, entryRoleTimeout, entryRoleThreshold, entryRoleAssumed, entryRoleRemaining:
		return true
	default:
		return false
	}
}

// spatialZoneLimit returns the configured slot ceiling for a zone kind.
func spatialZoneLimit(kind string) (int, bool) {
	switch kind {
	case zoneKindDetection:
		return maxDetectionZones, true
	case zoneKindExclusion:
		return maxExclusionZones, true
	case zoneKindEntry:
		return maxEntryZones, true
	default:
		return 0, false
	}
}

// validSpatialZone validates one overlay including its kind-specific slot and shape.
func validSpatialZone(zone SpatialZone, counts map[string]int, ids map[string]bool) bool {
	limit, known := spatialZoneLimit(zone.Kind)
	if !known || counts[zone.Kind] >= limit {
		return false
	}
	counts[zone.Kind]++
	if len(zone.OccupancyEntity) > maxSpatialEntityBytes || strings.ContainsRune(zone.OccupancyEntity, '\x00') {
		return false
	}
	if !validZoneEntities(zone) {
		return false
	}
	if !validSpatialIDName(zone.ID, zone.Name, ids) || !validSpatialPolygon(zone.Points) {
		return false
	}
	if zone.Shape == zoneShapePolygon {
		return true
	}
	if zone.Shape != zoneShapeRectangle || len(zone.Points) != 4 {
		return false
	}
	p := zone.Points
	return p[0].Y == p[1].Y && p[1].X == p[2].X && p[2].Y == p[3].Y && p[3].X == p[0].X
}

// validSpatialObject validates one visual room item and its bounded dimensions.
func validSpatialObject(object SpatialRoomObject, ids map[string]bool) bool {
	if !validSpatialIDName(object.ID, object.Name, ids) {
		return false
	}
	switch object.Kind {
	case roomObjectWall, roomObjectDoor, roomObjectFurniture:
	default:
		return false
	}
	return finiteWithin(object.X, spatialCoordinateLimitMM) &&
		finiteWithin(object.Y, spatialCoordinateLimitMM) &&
		finiteWithin(object.Width, spatialCoordinateLimitMM) &&
		finiteWithin(object.Height, spatialCoordinateLimitMM) &&
		object.Width > 0 && object.Height > 0 &&
		finiteWithin(object.Rotation, maxSpatialRotation)
}

// Validate rejects malformed, unbounded and degenerate workspace settings before persistence.
// A nil workspace preserves compatibility with existing spatial configurations.
func (w *SpatialWorkspace) Validate() error {
	if w == nil {
		return nil
	}
	invalid := fmt.Errorf("invalid spatial workspace")
	if w.Version != spatialWorkspaceVersion || len(w.Zones) > maxSpatialZones || len(w.Objects) > maxSpatialRoomObjects || len(w.EntryEntities) > maxSpatialEntryEntities || len(w.FirmwareEntities) > maxSpatialFirmwareEntities {
		return invalid
	}
	if !finiteWithin(w.Sensor.X, spatialCoordinateLimitMM) ||
		!finiteWithin(w.Sensor.Y, spatialCoordinateLimitMM) ||
		!finiteWithin(w.Sensor.Rotation, maxSpatialRotation) ||
		!finiteWithin(w.Sensor.FOV, maxSpatialFOV) || w.Sensor.FOV < 0 {
		return invalid
	}
	if len(w.Outline) > 0 && !validSpatialPolygon(w.Outline) {
		return invalid
	}
	ids := make(map[string]bool, len(w.Zones)+len(w.Objects))
	counts := make(map[string]int, 3)
	for _, zone := range w.Zones {
		if !validSpatialZone(zone, counts, ids) {
			return invalid
		}
	}
	for _, object := range w.Objects {
		if !validSpatialObject(object, ids) {
			return invalid
		}
	}
	if !validSpatialEntityMap(w.EntryEntities, validEntryRole) {
		return invalid
	}
	if !validSpatialEntityMap(w.FirmwareEntities, validFirmwareRole) {
		return invalid
	}
	return nil
}

// cloneSpatialWorkspace returns an independent copy so edits and cache snapshots
// never share the maps or point slices stored in the original settings value.
func cloneSpatialWorkspace(w *SpatialWorkspace) *SpatialWorkspace {
	if w == nil {
		return nil
	}
	result := *w
	result.Outline = append([]SpatialPoint(nil), w.Outline...)
	result.Objects = append([]SpatialRoomObject(nil), w.Objects...)
	result.Zones = append([]SpatialZone(nil), w.Zones...)
	for i := range result.Zones {
		result.Zones[i].Entities = make(map[string]string, len(w.Zones[i].Entities))
		for key, value := range w.Zones[i].Entities {
			result.Zones[i].Entities[key] = value
		}
		result.Zones[i].Points = append([]SpatialPoint(nil), w.Zones[i].Points...)
	}
	result.EntryEntities = make(map[string]string, len(w.EntryEntities))
	for key, value := range w.EntryEntities {
		result.EntryEntities[key] = value
	}
	result.FirmwareEntities = make(map[string]string, len(w.FirmwareEntities))
	for key, value := range w.FirmwareEntities {
		result.FirmwareEntities[key] = value
	}
	return &result
}
