package esphome

import (
	"context"
	"net"
	"testing"

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
