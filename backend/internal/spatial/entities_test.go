package spatial

import (
	"testing"

	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/models/settings"
)

func sourceDevice() *devices.Device {
	device := devices.NewDevice("0xabc123")
	device.FriendlyName = "Hall radar"
	device.PowerSource = "mains"
	return device
}

func TestDerivedIdentityRoundTrips(t *testing.T) {
	id := DeviceID("0xabc123")
	if id != "presence-0xabc123" || !isDerivedDeviceID(id) {
		t.Fatalf("device id = %q", id)
	}
	if isDerivedDeviceID("0xabc123") {
		t.Fatal("real device treated as derived")
	}
}

func TestZoneKeyIsStableShortAndDistinct(t *testing.T) {
	first := zoneKey("5c1f0d3e-1111-2222-3333-444455556666")
	if len(first) != 12 {
		t.Fatalf("zone key length = %d", len(first))
	}
	if first != zoneKey("5c1f0d3e-1111-2222-3333-444455556666") {
		t.Fatal("zone key is not stable")
	}
	if first == zoneKey("5c1f0d3e-1111-2222-3333-444455556667") {
		t.Fatal("distinct zones produced the same key")
	}
}

func TestBuildDeviceProjectsZonesReadOnly(t *testing.T) {
	zones := []settings.SpatialZone{
		{ID: "desk", Name: "Desk", Kind: "detection"},
		{ID: "door", Name: "Door", Kind: "entry"},
	}
	device := BuildDevice(sourceDevice(), zones)
	if device.Id != "presence-0xabc123" || device.Provider != providerName || device.ConnectionType != connectionType {
		t.Fatalf("device identity = %+v", device)
	}
	if device.FriendlyName != "Hall radar (presence)" {
		t.Fatalf("friendly name = %q", device.FriendlyName)
	}
	if device.Properties["source_device_id"] != "0xabc123" {
		t.Fatalf("properties = %+v", device.Properties)
	}
	for _, name := range []string{EntityOccupied, EntityTargetCount, EntityLastZone, EntityEmptySeconds} {
		if _, ok := device.Exposes[name]; !ok {
			t.Fatalf("missing global entity %q", name)
		}
	}
	occupied, count, dwell := ZoneEntityNames("desk")
	for _, name := range []string{occupied, count, dwell} {
		entity, ok := device.Exposes[name]
		if !ok {
			t.Fatalf("missing zone entity %q", name)
		}
		if entity.AccessMode != bridge.ReadAccessMode {
			t.Fatalf("entity %q is writable", name)
		}
		if entity.Attributes["zoneId"] != "desk" || entity.Attributes["zoneName"] != "Desk" {
			t.Fatalf("entity %q attributes = %+v", name, entity.Attributes)
		}
	}
	if len(device.Exposes) != 4+3*len(zones) {
		t.Fatalf("exposes = %d", len(device.Exposes))
	}
}

func TestRenameKeepsZoneEntityNames(t *testing.T) {
	original := BuildDevice(sourceDevice(), []settings.SpatialZone{{ID: "desk", Name: "Desk", Kind: "detection"}})
	renamed := BuildDevice(sourceDevice(), []settings.SpatialZone{{ID: "desk", Name: "Writing desk", Kind: "detection"}})
	occupied, count, dwell := ZoneEntityNames("desk")
	for _, name := range []string{occupied, count, dwell} {
		if _, ok := renamed.Exposes[name]; !ok {
			t.Fatalf("rename lost entity %q", name)
		}
		if renamed.Exposes[name].Attributes["zoneName"] != "Writing desk" {
			t.Fatalf("rename did not update the label for %q", name)
		}
	}
	if original.Exposes[occupied].Attributes["zoneName"] != "Desk" {
		t.Fatal("original inventory mutated")
	}
}

func TestPayloadCoversGlobalsAndZones(t *testing.T) {
	zones := []settings.SpatialZone{{ID: "desk", Name: "Desk", Kind: "detection"}}
	signals := Signals{
		Occupied: true, TargetCount: 1, LastZone: "Desk", EmptyFor: 0,
		Zones: map[string]ZoneSignal{"desk": {Occupied: true, Count: 1, Dwell: 4200 * 1000000}},
	}
	payload := buildPayload(zones, signals)
	if payload[EntityOccupied] != true || payload[EntityTargetCount] != 1 || payload[EntityLastZone] != "Desk" {
		t.Fatalf("globals = %+v", payload)
	}
	occupied, count, dwell := ZoneEntityNames("desk")
	if payload[occupied] != true || payload[count] != 1 || payload[dwell] != 4 {
		t.Fatalf("zone payload = %+v", payload)
	}
}
