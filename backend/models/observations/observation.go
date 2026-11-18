// Package observations defines the transport-agnostic record of a radio transmission that
// an enrolled device observed. It carries no transport details beyond the medium, so a
// Bluetooth proxy, a WiFi monitor and a USB dongle all produce comparable values.
package observations

import "time"

// Medium names the physical channel an observation arrived over. It is a closed set so
// consumers can switch exhaustively; unknown media are rejected at the trust boundary.
type Medium string

const (
	// MediumBLEAdvertisement is a Bluetooth Low Energy advertising packet.
	MediumBLEAdvertisement Medium = "ble.adv"
	// MediumWiFiProbe is reserved for a future source. Declaring it keeps the extension
	// path additive; no implementation exists yet.
	MediumWiFiProbe Medium = "wifi.probe"
)

// CapabilityWirelessObserve is declared on a device that can observe a radio medium. It is an
// abstract capability, never a provider name, so consumers stay source-agnostic.
const CapabilityWirelessObserve = "wireless.observe"

// Identity kinds reported in Identity.Kind. Consumers render labels from these values, so
// adding a kind is a presentation concern, not a new code path.
const (
	KindIBeacon       = "ibeacon"
	KindEddystone     = "eddystone"
	KindAdvertisement = "adv"
)

// Identity is the stable handle a viewer groups sightings by. Key is opaque to consumers.
// Stable is false for identities derived from an address that rotates, so consumers never
// present them as arrivals.
type Identity struct {
	Key    string `json:"key"`
	Kind   string `json:"kind"`
	Vendor uint16 `json:"vendor,omitempty"`
	Stable bool   `json:"stable"`
	Label  string `json:"label,omitempty"`
}

// Observation is one sensed transmission, normalised away from its transport.
type Observation struct {
	Source     string    `json:"source"`
	Medium     Medium    `json:"medium"`
	Identity   Identity  `json:"identity"`
	RSSI       int       `json:"rssi"`
	ObservedAt time.Time `json:"observedAt"`
}
