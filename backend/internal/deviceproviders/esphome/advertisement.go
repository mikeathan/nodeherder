package esphome

// This file adapts ESPHome Bluetooth proxy advertising packets to the transport-agnostic
// observation model. It owns only ESPHome specifics: protobuf field layout and the session
// subscription. Identity resolution lives in internal/observations so any source reuses it.
import (
	"context"
	"fmt"
	"time"

	"node-herder/internal/observations"
	model "node-herder/models/observations"

	"github.com/mycontroller-org/esphome_api/pkg/api"
)

// bleFacets are the filter dimensions this source can populate, declared to the viewer so it
// renders controls from data rather than from knowledge of this provider.
var bleFacets = []string{"kind", "vendor", "addressType"}

// bleMedia is the set of media this source can observe.
var bleMedia = []model.Medium{model.MediumBLEAdvertisement}

// Observer receives normalised observations from a session. It is optional: a driver that
// cannot observe a medium does not implement it, and Session.Run then sends no observation
// subscription. This mirrors the existing optional provider capabilities.
type Observer interface {
	// Observe delivers one batch. Implementations must not block for long and must return
	// on context cancellation.
	ObserveAdvertisements(ctx context.Context, batch []model.Observation) error
}

// ObservationSource identifies an observing session to the aggregator.
type ObservationSource interface {
	// SourceDescriptor reports the device id, display name and declared facets. The driver
	// supplies it once per session so the viewer learns the source before any packet arrives.
	SourceDescriptor() (id, name string, media []model.Medium, facets []string)
}

// sourceRegistry records an observing device so the viewer can render filter facets before
// the first advertisement arrives, and reports delivery counts so an empty viewer is
// diagnosable. It is optional, matching the other progressive-capability interfaces in this
// package: a sink without a registry simply announces nothing.
type sourceRegistry interface {
	RegisterSource(id, name string, media []model.Medium, facets []string)
	UnregisterSource(id string)
	// RecordSourceActivity reports raw advertisement entries received and how many produced a
	// usable identity. Raw counts can only be measured where packets are decoded, so the
	// transport reports them rather than the aggregator guessing.
	RecordSourceActivity(id string, advertisements, decoded uint64, at time.Time)
}

// observingSink adds observation to a device sink. It exists as a distinct type so that a
// driver without an observer produces a sink that does NOT satisfy Observer: the session
// asserts for the interface, and the assertion failing is what keeps a state-only device from
// sending an advertisement subscription.
type observingSink struct {
	*DeviceSink
	observer   Observer
	registry   sourceRegistry
	sourceID   string
	sourceName string
	announced  bool
}

// ObserveAdvertisements forwards a batch to the shared aggregator and reports delivery counts.
func (s *observingSink) ObserveAdvertisements(ctx context.Context, batch []model.Observation) error {
	// recordActivity runs before forwarding so counters reflect what arrived even when the
	// observer rejects the batch.
	s.recordActivity(len(batch), len(batch))
	return s.observer.ObserveAdvertisements(ctx, batch)
}

// RecordAdvertisementBatch reports the raw and decoded advertisement counts for one batch.
// The decoder knows the raw count because it discards unusable entries.
func (s *observingSink) RecordAdvertisementBatch(raw, decoded int) {
	s.recordActivity(raw, decoded)
}

// recordActivity accumulates counters for the viewer's diagnostics. It is a no-op when no
// registry is attached, so a state-only configuration pays nothing.
func (s *observingSink) recordActivity(raw, decoded int) {
	if s.registry == nil {
		return
	}
	s.registry.RecordSourceActivity(s.sourceID, uint64(raw), uint64(decoded), time.Now().UTC())
}

// SourceDescriptor reports this device's identity and declared observation facets.
func (s *observingSink) SourceDescriptor() (string, string, []model.Medium, []string) {
	return s.sourceID, s.sourceName, bleMedia, bleFacets
}

// RegisterSourceProvider attaches the aggregator that should learn this source's descriptor.
func (s *observingSink) RegisterSourceProvider(registry sourceRegistry) {
	s.registry = registry
}

// announceObservationSource declares the source to the aggregator exactly once. Repeating it
// is harmless but pointless, and the guard keeps a reconnecting session from duplicating the
// descriptor. A sink without a registry announces nothing.
func announceObservationSource(sink any) {
	observing, ok := sink.(*observingSink)
	if !ok || observing.registry == nil || observing.announced {
		return
	}
	id, name, media, facets := observing.SourceDescriptor()
	observing.registry.RegisterSource(id, name, media, facets)
	observing.announced = true
}

// withdrawObservationSource forgets a source that is no longer observing, so its stale facets
// do not linger in the viewer after a removal or reconfiguration.
func withdrawObservationSource(sink any) {
	observing, ok := sink.(*observingSink)
	if !ok || observing.registry == nil || !observing.announced {
		return
	}
	observing.registry.UnregisterSource(observing.sourceID)
	observing.announced = false
}

// subscribeAdvertisements asks the device to stream raw BLE advertisements. A device without
// the Bluetooth proxy component rejects the message, so callers treat a failure as "this
// device cannot observe" rather than as a session failure.
func subscribeAdvertisements(c nativeClient) error {
	if err := c.Send(&api.SubscribeBluetoothLEAdvertisementsRequest{}); err != nil {
		return fmt.Errorf("ESPHome advertisement subscription failed: %w", err)
	}
	return nil
}

// decodeAdvertisements converts one raw advertisement batch. Entries without a usable
// identity are skipped rather than failing the batch, because a malformed packet from a
// stranger must never break the observation stream.
func decodeAdvertisements(source string, response *api.BluetoothLERawAdvertisementsResponse) []model.Observation {
	if response == nil || len(response.Advertisements) == 0 {
		return nil
	}
	batch := make([]model.Observation, 0, len(response.Advertisements))
	for _, advertisement := range response.Advertisements {
		if advertisement == nil {
			continue
		}
		adv := observations.Advertisement{
			Address:     decodeBLEAddress(advertisement.Address),
			AddressType: uint8(advertisement.AddressType),
			RSSI:        int(advertisement.Rssi),
			Data:        advertisement.Data,
		}
		identity, name, ok := observations.ResolveIdentity(adv)
		if !ok {
			continue
		}
		payload := map[string]any{"addressType": adv.AddressType}
		if name != "" {
			payload["name"] = name
		}
		batch = append(batch, model.Observation{
			Source:   source,
			Medium:   model.MediumBLEAdvertisement,
			Identity: identity,
			RSSI:     adv.RSSI,
			Payload:  payload,
		})
	}
	return batch
}

// decodeBLEAddress expands the protobuf uint64 address into six bytes in display order.
// ESPHome packs the address so that the first displayed octet is the most significant byte,
// matching how a MAC string is converted to an integer.
func decodeBLEAddress(value uint64) []byte {
	address := make([]byte, 6)
	for i := 0; i < 6; i++ {
		address[i] = byte(value >> (8 * (5 - i)))
	}
	return address
}
