package esphome

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mycontroller-org/esphome_api/pkg/api"
	types "github.com/mycontroller-org/esphome_api/pkg/types"
	"node-herder/internal/ws"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"node-herder/testing/hubharness"
)

type testPipeline struct {
	device        *devices.Device
	updates       []map[string]interface{}
	offline       bool
	registrations int
}

func TestDeviceEntityCategoryDefaults(t *testing.T) {
	for _, tc := range []struct{ kind, category, want string }{
		{"switch", "", bridge.MeasurementCategory}, {"light", "", bridge.MeasurementCategory},
		{"number", "", bridge.ConfigCategory}, {"switch", bridge.DiagnosticCategory, bridge.DiagnosticCategory},
	} {
		d := Descriptor{ID: "00000000-00000001", Key: 1, Kind: tc.kind, Writable: true, Category: tc.category, Max: 1, Step: 1}
		entity, err := deviceEntity(d)
		if err != nil || entity.Category != tc.want {
			t.Fatalf("%s: %+v %v", tc.kind, entity, err)
		}
	}
}

func TestDeviceSinkValidatesPhysicalIdentity(t *testing.T) {
	sink, err := NewDeviceSink(&testPipeline{}, uuid.NewString(), "sensor")
	if err != nil {
		t.Fatal(err)
	}
	var identity string
	sink.identity = func(_ context.Context, value string) error { identity = value; return nil }
	for _, value := range []string{"", "00:00:00:00:00:00", "ff:ff:ff:ff:ff:ff", "01:00:00:00:00:01", "invalid"} {
		if err := sink.Identify(context.Background(), &types.DeviceInfo{MacAddress: value}); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
	if err := sink.Identify(context.Background(), &types.DeviceInfo{MacAddress: "02-AA-BB-CC-DD-EE"}); err != nil || identity != "02:aa:bb:cc:dd:ee" {
		t.Fatalf("identity %q: %v", identity, err)
	}
	sink.automaticName = true
	if err := sink.Identify(context.Background(), &types.DeviceInfo{MacAddress: "02:aa:bb:cc:dd:ee", FriendlyName: "  Radar  "}); err != nil || sink.name != "Radar" {
		t.Fatalf("automatic name %q: %v", sink.name, err)
	}
	sink.automaticName = false
	if err := sink.Identify(context.Background(), &types.DeviceInfo{MacAddress: "02:aa:bb:cc:dd:ee", FriendlyName: "Other"}); err != nil || sink.name != "Radar" {
		t.Fatalf("overrode configured name %q: %v", sink.name, err)
	}
}

func TestDeviceSinkRejectsChangedInventory(t *testing.T) {
	sink, err := NewDeviceSink(&testPipeline{}, uuid.NewString(), "sensor")
	if err != nil {
		t.Fatal(err)
	}
	conflict := errors.New("inventory conflict")
	sink.inventory = func(_ context.Context, kinds map[string]string) error {
		if kinds["00000000-00000001"] != "sensor" {
			t.Fatalf("unexpected inventory: %v", kinds)
		}
		return conflict
	}
	descriptor, err := Describe(&api.ListEntitiesSensorResponse{Key: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := sink.Inventory(context.Background(), []Descriptor{*descriptor}); !errors.Is(err, conflict) {
		t.Fatalf("inventory error: %v", err)
	}
	if sink.pending != nil {
		t.Fatal("rejected inventory was published")
	}
}

func TestDeviceSinkUsesExistingHubPipeline(t *testing.T) {
	h := hubharness.NewFromBridge(t, []byte(`[]`), 0)
	id := uuid.NewString()
	sink, err := NewDeviceSink(h.Hub, id, "ESPHome test device")
	if err != nil {
		t.Fatal(err)
	}
	descriptor, err := Describe(&api.ListEntitiesSensorResponse{Key: 1, Name: "Temperature", UnitOfMeasurement: "°C"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sink.Inventory(ctx, []Descriptor{*descriptor}); err != nil {
		t.Fatal(err)
	}
	if err := sink.Ready(ctx); err != nil {
		t.Fatal(err)
	}
	if err := sink.Observe(ctx, Observation{EntityID: descriptor.ID, Available: true, Value: float64(22)}); err != nil {
		t.Fatal(err)
	}
	device, err := h.Store.FindDeviceById(id)
	if err != nil || device.Exposes[descriptor.ID].Data.Value() != float64(22) {
		t.Fatalf("device state: %v %v", device, err)
	}
	if err := sink.Status(ctx, Offline); err != nil {
		t.Fatal(err)
	}
	device, err = h.Store.FindDeviceById(id)
	if err != nil || device.GetAvailability() != devices.OfflineAvailability {
		t.Fatalf("offline state: %v %v", device, err)
	}
}

func TestDeviceSinkBroadcastsChangedReadings(t *testing.T) {
	h := hubharness.NewFromBridge(t, []byte(`[]`), 0)
	id := uuid.NewString()
	sink, err := NewDeviceSink(h.Hub, id, "Radar")
	if err != nil {
		t.Fatal(err)
	}
	descriptor, err := Describe(&api.ListEntitiesSensorResponse{Key: 1, Name: "Distance", UnitOfMeasurement: "mm"})
	if err != nil {
		t.Fatal(err)
	}
	updates := make(chan *devices.UpdatePackage, 2)
	h.EventHub.OnBroadcast(func(event string, payload interface{}) {
		if event == ws.DeviceUpdated {
			if update, ok := payload.(*devices.UpdatePackage); ok {
				updates <- update
			}
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sink.Inventory(ctx, []Descriptor{*descriptor}); err != nil {
		t.Fatal(err)
	}
	if err := sink.Ready(ctx); err != nil {
		t.Fatal(err)
	}
	for _, value := range []float64{100, 200} {
		if err := sink.Observe(ctx, Observation{EntityID: descriptor.ID, Available: true, Value: value}); err != nil {
			t.Fatal(err)
		}
		select {
		case update := <-updates:
			if update.Id != id || update.Data[descriptor.ID] != value {
				t.Fatalf("update: %+v", update)
			}
		default:
			t.Fatalf("missing live update for %v", value)
		}
	}
}

func (p *testPipeline) RegisterProviderDevice(_ context.Context, d *devices.Device) error {
	p.device = d
	p.registrations++
	return nil
}
func (p *testPipeline) UpdateProviderDevice(_ context.Context, _ string, value map[string]interface{}) error {
	p.updates = append(p.updates, value)
	return nil
}
func (p *testPipeline) ProviderDeviceOffline(_ context.Context, _ string) error {
	p.offline = true
	return nil
}
func (p *testPipeline) RemoveProviderDevice(context.Context, string) error {
	p.device = nil
	return nil
}

func TestDeviceSinkInstallsInventoryAndReports(t *testing.T) {
	pipeline := &testPipeline{}
	sink, err := NewDeviceSink(pipeline, uuid.NewString(), "Radar")
	if err != nil {
		t.Fatal(err)
	}
	sensor, err := Describe(&api.ListEntitiesSensorResponse{Key: 1, Name: "Target X", UnitOfMeasurement: "mm"})
	if err != nil {
		t.Fatal(err)
	}
	number, err := Describe(&api.ListEntitiesNumberResponse{Key: 2, Name: "Distance", MinValue: 0, MaxValue: 10, Step: 0.5})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := sink.Inventory(ctx, []Descriptor{*sensor, *number}); err != nil {
		t.Fatal(err)
	}
	if pipeline.device != nil {
		t.Fatal("inventory published before ready")
	}
	if err := sink.Ready(ctx); err != nil {
		t.Fatal(err)
	}
	if pipeline.device.Provider != "esphome" || len(pipeline.device.Exposes) != 2 {
		t.Fatalf("device: %+v", pipeline.device)
	}
	expose := pipeline.device.Exposes[number.ID]
	if expose.AccessMode != bridge.ReadWriteAccessMode || expose.Attributes["label"] != "Distance" || expose.Attributes["step"] != float64(0.5) {
		t.Fatalf("expose: %+v", expose)
	}
	if err := sink.Observe(ctx, Observation{EntityID: sensor.ID, Available: true, Value: float64(0)}); err != nil {
		t.Fatal(err)
	}
	if err := sink.Observe(ctx, Observation{EntityID: sensor.ID, Available: false, Value: float64(7)}); err != nil {
		t.Fatal(err)
	}
	if pipeline.updates[0][sensor.ID] != float64(0) || pipeline.updates[1][sensor.ID] != nil {
		t.Fatal("missing and zero states conflated")
	}
	if err := sink.Status(ctx, Offline); err != nil || !pipeline.offline {
		t.Fatalf("offline: %v", err)
	}
}

func TestDeviceSinkRegistersOneDeviceForManyEntities(t *testing.T) {
	const entityCount = 74
	descriptors := make([]Descriptor, 0, entityCount)
	for key := 1; key <= entityCount; key++ {
		descriptors = append(descriptors, Descriptor{
			ID:       fmt.Sprintf("00000000-%08x", key),
			Key:      uint32(key),
			Kind:     kindSensor,
			Name:     fmt.Sprintf("Target %d", key),
			Category: bridge.MeasurementCategory,
		})
	}
	pipeline := &testPipeline{}
	sink, err := NewDeviceSink(pipeline, uuid.NewString(), "Everything Presence Lite")
	if err != nil {
		t.Fatal(err)
	}
	if err := sink.Inventory(context.Background(), descriptors); err != nil {
		t.Fatal(err)
	}
	if err := sink.Ready(context.Background()); err != nil {
		t.Fatal(err)
	}
	if pipeline.registrations != 1 || len(pipeline.device.Exposes) != entityCount {
		t.Fatalf("registrations=%d exposes=%d", pipeline.registrations, len(pipeline.device.Exposes))
	}
	for id, entity := range pipeline.device.Exposes {
		if entity.Attributes[attributeLabel] == nil || entity.Attributes[attributeLabel] == id {
			t.Fatalf("entity %s lost its discovered label", id)
		}
	}
}

func TestDeviceSinkPublishesOnlyBoundedPublicMetadata(t *testing.T) {
	pipeline := &testPipeline{}
	sink, err := NewDeviceSink(pipeline, uuid.NewString(), "sensor")
	if err != nil {
		t.Fatal(err)
	}
	sink.identity = func(context.Context, string) error { return nil }
	info := &types.DeviceInfo{MacAddress: "02:00:00:00:00:01", Model: "Radar", Manufacturer: "Maker", EsphomeVersion: "2026.10", ProjectName: "Presence", ProjectVersion: "1.5"}
	if err := sink.Identify(context.Background(), info); err != nil {
		t.Fatal(err)
	}
	if err := sink.Inventory(context.Background(), []Descriptor{{ID: "00000000-00000001", Key: 1, Kind: "binary_sensor", DeviceClass: "occupancy", Name: "Presence"}}); err != nil {
		t.Fatal(err)
	}
	if err := sink.Ready(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(pipeline.device.Properties) != 5 || pipeline.device.Properties["model"] != "Radar" {
		t.Fatal(pipeline.device.Properties)
	}
	if pipeline.device.Exposes["00000000-00000001"].Attributes["device_class"] != "occupancy" {
		t.Fatal("class lost")
	}
}

func TestDeviceSinkOmitsInvalidPublicMetadata(t *testing.T) {
	sink, err := NewDeviceSink(&testPipeline{}, uuid.NewString(), "sensor")
	if err != nil {
		t.Fatal(err)
	}
	sink.identity = func(context.Context, string) error { return nil }
	if err := sink.Identify(context.Background(), &types.DeviceInfo{MacAddress: "02:00:00:00:00:01", Manufacturer: strings.Repeat("x", 257), ProjectName: "bad\x00name", Model: "Valid"}); err != nil {
		t.Fatal(err)
	}
	if len(sink.metadata) != 1 || sink.metadata["model"] != "Valid" {
		t.Fatal(sink.metadata)
	}
}
