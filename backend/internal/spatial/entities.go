package spatial

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/settings"
)

// ProviderName marks devices this package derives. Other layers use it to keep derived
// inventories out of provider semantics (binary preservation, enrollment, deletion).
const ProviderName = providerName

// IsDerivedDeviceID reports whether id belongs to a derived presence device.
func IsDerivedDeviceID(id string) bool { return isDerivedDeviceID(id) }

// SourceDeviceID returns the source device id for a derived device id.
func SourceDeviceID(id string) (string, bool) {
	if !isDerivedDeviceID(id) {
		return "", false
	}
	return strings.TrimPrefix(id, derivedPrefix), true
}

// providerName marks devices this package derives. They are read-only projections of a
// source device's mapping, never firmware truth.
const providerName = "presence"

// connectionType is the device-level label shown for derived devices.
const connectionType = "derived"

// Entity names shared by the engine, automations and the browser resolver.
const (
	EntityOccupied     = "occupied"
	EntityTargetCount  = "target_count"
	EntityLastZone     = "last_zone"
	EntityEmptySeconds = "empty_seconds"
)

// Derived device and zone entity identifiers are built here only, so the hub, the
// automation editor and the browser cannot invent different names.
const derivedPrefix = "presence-"

// DeviceID returns the stable identifier for the derived device of one source device.
func DeviceID(sourceID string) string {
	return derivedPrefix + sourceID
}

// isDerivedDeviceID reports whether id belongs to a derived presence device. The engine uses
// this to ignore its own devices as sources.
func isDerivedDeviceID(id string) bool {
	return strings.HasPrefix(id, derivedPrefix)
}

// zoneKey is the stable short key for a zone identifier. Renaming a zone keeps its entity
// names; changing the zone identifier would change them, which the settings model prevents.
func zoneKey(zoneID string) string {
	sum := sha256.Sum256([]byte(zoneID))
	return hex.EncodeToString(sum[:6])
}

// ZoneEntityNames returns the three entity names for one zone.
func ZoneEntityNames(zoneID string) (occupied, count, dwell string) {
	key := zoneKey(zoneID)
	return "zone_" + key + "_occupied", "zone_" + key + "_count", "zone_" + key + "_dwell_seconds"
}

// friendlyName is the device name shown in the device list.
func friendlyName(sourceName string) string {
	return fmt.Sprintf("%s (presence)", sourceName)
}

func readOnlyEntity(name, label string, dataType bridge.ExposeDataType, unit string, category bridge.ExposeCategory, attributes map[string]any) *devices.Entity {
	entity := devices.NewEntity(name)
	entity.Type = dataType
	entity.Unit = unit
	entity.AccessMode = bridge.ReadAccessMode
	entity.Category = category
	entity.Attributes = map[string]any{"label": label}
	for key, value := range attributes {
		entity.Attributes[key] = value
	}
	return entity
}

// BuildDevice projects a source device's saved zones into a read-only derived device. The
// result is a fresh inventory generation: providers replace it wholesale on mapping changes,
// which is why derived state must never be merged into the source device instead.
func BuildDevice(source *devices.Device, zones []settings.SpatialZone) *devices.Device {
	device := devices.NewDevice(DeviceID(source.Id))
	device.Provider = providerName
	device.ConnectionType = connectionType
	device.FriendlyName = friendlyName(source.FriendlyName)
	device.Description = "Derived presence from " + source.FriendlyName + ". Read-only; values follow the saved mapping and zones."
	device.PowerSource = source.PowerSource
	device.Properties = map[string]interface{}{
		"source_device_id": source.Id,
		"derived":          true,
	}

	device.Exposes[EntityOccupied] = readOnlyEntity(EntityOccupied, "Occupied", bridge.BinaryDataType, "", bridge.MeasurementCategory, map[string]any{"source": source.Id})
	device.Exposes[EntityTargetCount] = readOnlyEntity(EntityTargetCount, "Targets in zones", bridge.NumericDataType, "", bridge.MeasurementCategory, map[string]any{"source": source.Id})
	device.Exposes[EntityLastZone] = readOnlyEntity(EntityLastZone, "Last zone", bridge.TextDataType, "", bridge.DiagnosticCategory, map[string]any{"source": source.Id})
	device.Exposes[EntityEmptySeconds] = readOnlyEntity(EntityEmptySeconds, "Empty for", bridge.NumericDataType, "s", bridge.DiagnosticCategory, map[string]any{"source": source.Id})

	for _, zone := range zones {
		occupied, count, dwell := ZoneEntityNames(zone.ID)
		attributes := map[string]any{"source": source.Id, "zoneId": zone.ID, "zoneName": zone.Name, "zoneKind": zone.Kind}
		device.Exposes[occupied] = readOnlyEntity(occupied, zone.Name+" occupied", bridge.BinaryDataType, "", bridge.MeasurementCategory, attributes)
		device.Exposes[count] = readOnlyEntity(count, zone.Name+" targets", bridge.NumericDataType, "", bridge.MeasurementCategory, attributes)
		device.Exposes[dwell] = readOnlyEntity(dwell, zone.Name+" dwell", bridge.NumericDataType, "s", bridge.DiagnosticCategory, attributes)
	}
	return device
}

// buildPayload builds the report for one computed snapshot. It always writes a value; unknown
// state at startup comes from the registered device having no data yet, not from omission.
func buildPayload(zones []settings.SpatialZone, signals Signals) map[string]interface{} {
	payload := map[string]interface{}{
		EntityOccupied:     signals.Occupied,
		EntityTargetCount:  signals.TargetCount,
		EntityLastZone:     signals.LastZone,
		EntityEmptySeconds: int(signals.EmptyFor.Seconds()),
	}
	for _, zone := range zones {
		occupied, count, dwell := ZoneEntityNames(zone.ID)
		signal := signals.Zones[zone.ID]
		payload[occupied] = signal.Occupied
		payload[count] = signal.Count
		payload[dwell] = int(signal.Dwell.Seconds())
	}
	return payload
}
