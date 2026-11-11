package esphome

// This file adapts ESPHome Bluetooth proxy advertising packets to the transport-agnostic
// observation model. It owns only ESPHome specifics: protobuf field layout and the session
// subscription. Identity resolution lives in internal/observations so any source reuses it.
import (
	"context"
	"fmt"

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

// observingSink adds observation to a device sink. It exists as a distinct type so that a
// driver without an observer produces a sink that does NOT satisfy Observer: the session
// asserts for the interface, and the assertion failing is what keeps a state-only device from
// sending an advertisement subscription.
type observingSink struct {
	*DeviceSink
	observer   Observer
	sourceID   string
	sourceName string
}

// ObserveAdvertisements forwards a batch to the shared aggregator.
func (s *observingSink) ObserveAdvertisements(ctx context.Context, batch []model.Observation) error {
	return s.observer.ObserveAdvertisements(ctx, batch)
}

// SourceDescriptor reports this device's identity and declared observation facets.
func (s *observingSink) SourceDescriptor() (string, string, []model.Medium, []string) {
	return s.sourceID, s.sourceName, bleMedia, bleFacets
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
