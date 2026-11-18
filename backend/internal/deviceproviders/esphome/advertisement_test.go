package esphome

import (
	"context"
	"net"
	"testing"
	"time"

	"node-herder/internal/observations"
	model "node-herder/models/observations"

	"github.com/mycontroller-org/esphome_api/pkg/api"
	"google.golang.org/protobuf/proto"
)

// recordingObserver captures advertisement batches for assertions.
type recordingObserver struct {
	batches [][]model.Observation
}

func (o *recordingObserver) ObserveAdvertisements(_ context.Context, batch []model.Observation) error {
	o.batches = append(o.batches, batch)
	return nil
}

// observingTestSink wraps a device sink so the session's interface assertion succeeds.
func observingTestSink(observer Observer) *observingSink {
	return &observingSink{
		DeviceSink: &DeviceSink{},
		observer:   observer,
		sourceID:   "enrollment-1",
		sourceName: "Hall proxy",
	}
}

// recordingRegistry captures source registration so the session's wiring is observable.
type recordingRegistry struct {
	registered     []string
	unregistered   []string
	media          []model.Medium
	facets         []string
	advertisements uint64
	decoded        uint64
	lastAt         *time.Time
}

func (r *recordingRegistry) RegisterSource(id, name string, media []model.Medium, facets []string) {
	r.registered = append(r.registered, id+"|"+name)
	r.media = media
	r.facets = facets
}

func (r *recordingRegistry) UnregisterSource(id string) { r.unregistered = append(r.unregistered, id) }

// RecordSourceActivity records delivery counts reported by the sink.
func (r *recordingRegistry) RecordSourceActivity(id string, advertisements, decoded uint64, at time.Time) {
	r.advertisements += advertisements
	r.decoded += decoded
	if r.lastAt == nil || at.After(*r.lastAt) {
		observed := at
		r.lastAt = &observed
	}
}

// observingRegistrySink is an observing sink whose observer also records descriptors.
func observingRegistrySink(registry *recordingRegistry) *observingSink {
	return &observingSink{
		DeviceSink: &DeviceSink{},
		observer:   &recordingObserver{},
		registry:   registry,
		sourceID:   "enrollment-1",
		sourceName: "Hall proxy",
	}
}

func TestObservationSourceRegistersOnceWhileObserving(t *testing.T) {
	registry := &recordingRegistry{}
	sink := observingRegistrySink(registry)

	announceObservationSource(sink)

	if len(registry.registered) != 1 || registry.registered[0] != "enrollment-1|Hall proxy" {
		t.Fatalf("registered = %v, want one descriptor", registry.registered)
	}
	// Facets must reach the registry, otherwise the viewer renders no filters at all.
	if len(registry.facets) == 0 {
		t.Fatal("no facets registered; the viewer would show no filter controls")
	}
	if len(registry.media) != 1 || registry.media[0] != model.MediumBLEAdvertisement {
		t.Fatalf("media = %v", registry.media)
	}

	// A second call must not duplicate the descriptor.
	announceObservationSource(sink)
	if len(registry.registered) != 1 {
		t.Fatalf("registered %d times, want idempotent registration", len(registry.registered))
	}
}

func TestNonObservingSinkIsNeverRegistered(t *testing.T) {
	registry := &recordingRegistry{}
	// A plain sink has no registry and no observer, so nothing is announced.
	if _, ok := any(&DeviceSink{}).(sourceRegistry); ok {
		t.Fatal("plain sink satisfied sourceRegistry")
	}
	sink := &observingSink{DeviceSink: &DeviceSink{}, observer: &recordingObserver{}, sourceID: "x", sourceName: "y"}
	sink.registry = nil
	announceObservationSource(sink)
	if len(registry.registered) != 0 {
		t.Fatalf("nil registry registered %v", registry.registered)
	}
}

func TestObservationCapabilityIsStructural(t *testing.T) {
	// A plain sink must not satisfy the observation interfaces: that is what keeps a
	// state-only device from sending an advertisement subscription.
	plain := &DeviceSink{}
	if _, ok := any(plain).(Observer); ok {
		t.Fatal("plain sink satisfied Observer")
	}
	if _, ok := any(plain).(ObservationSource); ok {
		t.Fatal("plain sink satisfied ObservationSource")
	}

	observing := observingTestSink(&recordingObserver{})
	if _, ok := any(observing).(Observer); !ok {
		t.Fatal("observing sink did not satisfy Observer")
	}
	if _, ok := any(observing).(ObservationSource); !ok {
		t.Fatal("observing sink did not satisfy ObservationSource")
	}
}

func TestSourceDescriptorDeclaresMediaAndFacets(t *testing.T) {
	id, name, media, facets := observingTestSink(&recordingObserver{}).SourceDescriptor()
	if id != "enrollment-1" || name != "Hall proxy" {
		t.Fatalf("id/name = %q/%q", id, name)
	}
	if len(media) != 1 || media[0] != model.MediumBLEAdvertisement {
		t.Fatalf("media = %v", media)
	}
	if len(facets) == 0 {
		t.Fatal("no facets declared; the viewer would render no filters")
	}
}

func TestDecodeBLEAddressUsesDisplayOrder(t *testing.T) {
	// ESPHome packs the address as an integer whose most significant byte is the first
	// displayed octet.
	address := net.HardwareAddr(decodeBLEAddress(0x0000AABBCCDDEEFF))
	want := net.HardwareAddr{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
	if address.String() != want.String() {
		t.Fatalf("address = %s, want %s", address.String(), want.String())
	}
	if len(address) != 6 {
		t.Fatalf("address length = %d", len(address))
	}
}

func TestDecodeAdvertisementsSkipsUnusableEntries(t *testing.T) {
	// A nil entry and a zero address must be skipped, not fail the batch.
	batch := decodeAdvertisements("source-1", &api.BluetoothLERawAdvertisementsResponse{
		Advertisements: []*api.BluetoothLERawAdvertisement{
			nil,
			{Address: 0, Rssi: -70},
			{Address: 0x0000AABBCCDDEEFF, Rssi: -62, AddressType: uint32(observations.AddressTypePublic)},
		},
	})
	if len(batch) != 1 {
		t.Fatalf("batch length = %d, want the single usable advertisement", len(batch))
	}
	observation := batch[0]
	if observation.Source != "source-1" || observation.Medium != model.MediumBLEAdvertisement {
		t.Fatalf("observation = %+v", observation)
	}
	if observation.RSSI != -62 {
		t.Fatalf("rssi = %d, want -62", observation.RSSI)
	}
	if observation.Identity.Key != "addr:aa:bb:cc:dd:ee:ff" || !observation.Identity.Stable {
		t.Fatalf("identity = %+v", observation.Identity)
	}
}

// TestDecodeAdvertisementsStampsObservedAt pins the timestamp, which used to be zero.
//
// The aggregator substitutes its own clock when it records, so a zero value looked harmless. The
// recorder does not: it files evidence by this field, so every reading landed on day 0001-01-01 and
// hour 0. The accumulator then never promoted, the write floor rejected the row, no summary was
// ever stored, and every device stayed "learning" indefinitely with no error anywhere. Nothing
// caught it because every recorder test sets the field itself.
func TestDecodeAdvertisementsStampsObservedAt(t *testing.T) {
	before := time.Now().UTC()
	batch := decodeAdvertisements("source-1", &api.BluetoothLERawAdvertisementsResponse{
		Advertisements: []*api.BluetoothLERawAdvertisement{
			{Address: 0x0000AABBCCDDEEFF, Rssi: -62, AddressType: uint32(observations.AddressTypePublic)},
		},
	})
	after := time.Now().UTC()
	if len(batch) != 1 {
		t.Fatalf("batch length = %d", len(batch))
	}
	stamped := batch[0].ObservedAt
	if stamped.IsZero() {
		t.Fatal("ObservedAt is zero, so evidence would be filed under year 1")
	}
	if stamped.Before(before) || stamped.After(after) {
		t.Fatalf("ObservedAt = %v, want a value between %v and %v", stamped, before, after)
	}
}

func TestDecodeAdvertisementsKeepsAdvertisedName(t *testing.T) {
	// An advertised local name must become the identity label, otherwise the viewer can only
	// show an opaque key even when the device names itself.
	// AD structure: length, type 0x09 (complete local name), then the name bytes.
	const completeLocalName = 0x09
	name := append([]byte{byte(len("Kitchen tag") + 1), completeLocalName}, []byte("Kitchen tag")...)
	batch := decodeAdvertisements("source-1", &api.BluetoothLERawAdvertisementsResponse{
		Advertisements: []*api.BluetoothLERawAdvertisement{
			{Address: 0x0000AABBCCDDEEFF, Rssi: -62, AddressType: uint32(observations.AddressTypePublic), Data: name},
		},
	})
	if len(batch) != 1 {
		t.Fatalf("batch length = %d", len(batch))
	}
	if batch[0].Identity.Label != "Kitchen tag" {
		t.Fatalf("label = %q, want the advertised name", batch[0].Identity.Label)
	}
}

// Forwarding a batch must not count it again: the session already reports the raw and decoded
// counts, and the forwarded slice has had unusable entries removed. Counting in both places
// doubled every counter.
func TestForwardingDoesNotDoubleCount(t *testing.T) {
	registry := &recordingRegistry{}
	sink := observingRegistrySink(registry)

	const raw, decoded = 10, 8
	sink.RecordAdvertisementBatch(raw, decoded)
	_ = sink.ObserveAdvertisements(context.Background(), make([]model.Observation, decoded))

	if registry.advertisements != raw {
		t.Fatalf("advertisements = %d, want %d", registry.advertisements, raw)
	}
	if registry.decoded != decoded {
		t.Fatalf("decoded = %d, want %d", registry.decoded, decoded)
	}
}

// The supervisor decorates the sink before handing it to the session. Embedding promotes only
// the interface it declares, so capability assertions on the wrapper used to fail and
// observation was silently dead: the device connected and no advertisement was ever
// subscribed. This is the regression guard for that.
func TestSupervisorWrapperStillExposesObservation(t *testing.T) {
	observing := observingRegistrySink(&recordingRegistry{})
	wrapped := &supervisedSink{SupervisorSink: observing}

	// The wrapper itself must not satisfy the capability; that is what made the bug silent.
	if _, ok := any(wrapped).(Observer); ok {
		t.Fatal("wrapper unexpectedly implements Observer; the unwrap path is untested")
	}
	// Resolving through the capability helper must find it on the decorated sink.
	if _, ok := capability[Observer](wrapped); !ok {
		t.Fatal("Observer lost through the supervisor wrapper")
	}
	if _, ok := capability[ObservationSource](wrapped); !ok {
		t.Fatal("ObservationSource lost through the supervisor wrapper")
	}
	if _, ok := capability[interface{ RecordAdvertisementBatch(int, int) }](wrapped); !ok {
		t.Fatal("advertisement reporter lost through the supervisor wrapper")
	}
}

// A plain sink must still report no capability through the helper.
func TestCapabilityHelperRejectsPlainSink(t *testing.T) {
	if _, ok := capability[Observer](&DeviceSink{}); ok {
		t.Fatal("plain sink resolved an Observer")
	}
}

// The session announces through the supervisor wrapper, not the bare sink. A direct assertion
// for *observingSink failed there, so the source was never registered: identities arrived but
// the viewer reported "0 sources" and rendered no filter facets. This is that regression.
func TestSourceAnnouncesThroughTheSupervisorWrapper(t *testing.T) {
	registry := &recordingRegistry{}
	wrapped := &supervisedSink{SupervisorSink: observingRegistrySink(registry)}

	announceObservationSource(wrapped)

	if len(registry.registered) != 1 {
		t.Fatalf("registered = %v, want the source announced through the wrapper", registry.registered)
	}
	if len(registry.facets) == 0 {
		t.Fatal("no facets registered; the viewer would render no filter controls")
	}

	// Withdrawing through the wrapper must reach the same sink.
	withdrawObservationSource(wrapped)
	if len(registry.unregistered) != 1 {
		t.Fatalf("unregistered = %v, want the withdrawal to reach the wrapped sink", registry.unregistered)
	}
}

func TestAnnounceIgnoresNonObservingSinks(t *testing.T) {
	registry := &recordingRegistry{}
	// A plain sink has no registry and no observer, so nothing is announced.
	announceObservationSource(&DeviceSink{})
	if len(registry.registered) != 0 {
		t.Fatalf("nil registry registered %v", registry.registered)
	}
	// A wrapper around a plain sink must also stay silent.
	announceObservationSource(&supervisedSink{SupervisorSink: &DeviceSink{}})
	if len(registry.registered) != 0 {
		t.Fatalf("plain wrapped sink registered %v", registry.registered)
	}
}

func TestDecodeAdvertisementsHandlesEmptyResponses(t *testing.T) {
	if got := decodeAdvertisements("s", nil); got != nil {
		t.Fatalf("nil response produced %v", got)
	}
	if got := decodeAdvertisements("s", &api.BluetoothLERawAdvertisementsResponse{}); got != nil {
		t.Fatalf("empty response produced %v", got)
	}
}

func TestSubscribeAdvertisementsSendsTheRequest(t *testing.T) {
	sent := make(chan any, 1)
	client := &stubSubscribeClient{sent: sent}
	if err := subscribeAdvertisements(client); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	select {
	case message := <-sent:
		if _, ok := message.(*api.SubscribeBluetoothLEAdvertisementsRequest); !ok {
			t.Fatalf("sent %T", message)
		}
	default:
		t.Fatal("no subscription message was sent")
	}
}

// stubSubscribeClient records the message passed to Send. The embedded interface supplies the
// unused methods, so the stub only overrides what the test exercises.
type stubSubscribeClient struct {
	nativeClient
	sent chan any
}

func (c *stubSubscribeClient) Send(message proto.Message) error {
	c.sent <- message
	return nil
}
