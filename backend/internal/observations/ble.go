package observations

// This file decodes Bluetooth Low Energy advertising data into an Identity. It is medium
// logic, not transport logic: any BLE source (ESPHome proxy, USB dongle) reuses it by
// supplying the raw address, address type and AD structures.
import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"

	model "node-herder/models/observations"
)

// BLE address type values, as reported by every mainstream BLE stack (including the ESPHome
// Bluetooth proxy). A random address is further classified by the two most significant bits
// of its most significant byte; see randomAddressStable.
const (
	AddressTypePublic   uint8 = 0
	AddressTypeRandom   uint8 = 1
	AddressTypePublicID uint8 = 2
	AddressTypeRandomID uint8 = 3
)

// Bluetooth company identifiers used to recognise beacon payloads.
const (
	companyApple     uint16 = 0x004C
	serviceEddystone uint16 = 0xFEAA
)

// AD structure types (Assigned Numbers, "Common Data Types").
const (
	adFlags             uint8 = 0x01
	adCompleteLocalName uint8 = 0x09
	adShortLocalName    uint8 = 0x08
	adServiceData16     uint8 = 0x16
	adManufacturerData  uint8 = 0xFF
)

// iBeacon frame layout: company id, then 0x02 0x15, then the proximity payload.
const (
	ibeaconType     uint8 = 0x02
	ibeaconLength   uint8 = 0x15
	ibeaconMinBytes       = 2 + 16 + 2 + 2
)

// maxFingerprintBytes bounds how much of a rotating address feeds the fingerprint, so the
// key stays short and the hash input is unambiguous.
const fingerprintHexLength = 12

// Advertisement is a decoded BLE advertising packet: the medium-level facts an identity
// resolver needs, independent of which transport delivered them.
type Advertisement struct {
	// Address is the sender address in display order, six bytes for a public or random
	// address. A nil or non-six-byte address yields no identity.
	Address net.HardwareAddr
	// AddressType follows the BLE stack enumeration above.
	AddressType uint8
	RSSI        int
	// Data holds the raw AD structures exactly as received.
	Data []byte
}

// ResolveIdentity derives an identity from an advertisement. Payload identities take
// precedence over the address because they survive address rotation; a rotating random
// address yields a Stable false fingerprint so consumers cannot mistake it for an arrival.
// The returned name is the advertised local name, empty when absent.
func ResolveIdentity(adv Advertisement) (model.Identity, string, bool) {
	if len(adv.Address) != 6 || isZeroAddress(adv.Address) {
		return model.Identity{}, "", false
	}
	name := localName(adv.Data)
	if identity, ok := beaconIdentity(adv.Data); ok {
		return identity, name, true
	}
	stable := addressStable(adv.AddressType, adv.Address)
	key := "addr:" + adv.Address.String()
	if !stable {
		key = "fp:" + fingerprint(adv.Address)
	}
	return model.Identity{Key: key, Kind: model.KindAdvertisement, Stable: stable}, name, true
}

// isZeroAddress reports whether an address is unset. Some sources report an all-zero address
// when they cannot resolve one, and treating it as an identity would merge unrelated devices.
func isZeroAddress(address net.HardwareAddr) bool {
	for _, b := range address {
		if b != 0 {
			return false
		}
	}
	return true
}

// addressStable reports whether repeated sightings of this address can be treated as the
// same device. Public addresses are globally unique; random addresses are only stable when
// they use the static subtype (top two bits 0b11).
func addressStable(addressType uint8, address net.HardwareAddr) bool {
	switch addressType {
	case AddressTypePublic, AddressTypePublicID:
		return true
	case AddressTypeRandom, AddressTypeRandomID:
		return randomAddressStable(address)
	default:
		// An unknown address type cannot be trusted as stable; treat it as rotating so
		// consumers never count it as an arrival.
		return false
	}
}

// randomAddressStable classifies a random address by its most significant bits: 0b11 is a
// static address that persists across power cycles, 0b01 is a resolvable private address
// and 0b00 a non-resolvable private address, both of which rotate.
func randomAddressStable(address net.HardwareAddr) bool {
	return address[0]&0xC0 == 0xC0
}

// fingerprint is the short stable key for one rotating address. It changes when the address
// rotates, which is intended: the new address is a new ephemeral identity.
func fingerprint(address net.HardwareAddr) string {
	sum := sha256.Sum256(address[:min(len(address), 6)])
	return hex.EncodeToString(sum[:])[:fingerprintHexLength]
}

// beaconIdentity recognises beacon payloads that carry their own identity, which survives
// address rotation.
func beaconIdentity(data []byte) (model.Identity, bool) {
	var found model.Identity
	ok := false
	walkAD(data, func(adType uint8, payload []byte) {
		if ok {
			return
		}
		switch adType {
		case adManufacturerData:
			if identity, matched := ibeaconIdentity(payload); matched {
				found, ok = identity, true
			}
		case adServiceData16:
			if identity, matched := eddystoneIdentity(payload); matched {
				found, ok = identity, true
			}
		}
	})
	return found, ok
}

// ibeaconIdentity parses Apple iBeacon manufacturer data, whose first two bytes are the
// little-endian company identifier. The UUID, major and minor survive address rotation.
func ibeaconIdentity(payload []byte) (model.Identity, bool) {
	if len(payload) < 2 || uint16(payload[0])|uint16(payload[1])<<8 != companyApple {
		return model.Identity{}, false
	}
	body := payload[2:]
	if len(body) < ibeaconMinBytes || body[0] != ibeaconType || body[1] != ibeaconLength {
		return model.Identity{}, false
	}
	uuid := formatUUID(body[2:18])
	major := uint16(body[18])<<8 | uint16(body[19])
	minor := uint16(body[20])<<8 | uint16(body[21])
	return model.Identity{
		Key:    fmt.Sprintf("%s:%s:%d:%d", model.KindIBeacon, uuid, major, minor),
		Kind:   model.KindIBeacon,
		Vendor: companyApple,
		Stable: true,
	}, true
}

// eddystoneIdentity parses Eddystone service data. Only the UID frame carries a stable
// namespace and instance; URL, TLM and EID frames are ignored so they fall back to the
// address.
func eddystoneIdentity(payload []byte) (model.Identity, bool) {
	// A UID frame is the frame type, calibrated power, 10-byte namespace and 6-byte
	// instance.
	const uidFrame, uidLength = 0x00, 18
	if len(payload) < 2 || uint16(payload[0])|uint16(payload[1])<<8 != serviceEddystone {
		return model.Identity{}, false
	}
	body := payload[2:]
	if len(body) < uidLength || body[0] != uidFrame {
		return model.Identity{}, false
	}
	// Byte 1 is the calibrated transmit power; namespace and instance follow it.
	namespace := hex.EncodeToString(body[2:12])
	instance := hex.EncodeToString(body[12:18])
	return model.Identity{
		Key:    fmt.Sprintf("%s:%s:%s", model.KindEddystone, namespace, instance),
		Kind:   model.KindEddystone,
		Stable: true,
	}, true
}

// formatUUID renders a 16-byte value in canonical 8-4-4-4-12 lowercase form.
func formatUUID(value []byte) string {
	if len(value) != 16 {
		return ""
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16])
}

// localName returns the advertised local name, preferring the complete name over a
// shortened one. An empty result means the packet carried no name.
func localName(data []byte) string {
	var complete, short string
	walkAD(data, func(adType uint8, payload []byte) {
		if len(payload) == 0 {
			return
		}
		switch adType {
		case adCompleteLocalName:
			if complete == "" {
				complete = string(payload)
			}
		case adShortLocalName:
			if short == "" {
				short = string(payload)
			}
		}
	})
	if complete != "" {
		return complete
	}
	return short
}

// walkAD iterates BLE AD structures. It stops at the first malformed length so a truncated
// or hostile packet can never panic a caller or be read past its end.
func walkAD(data []byte, visit func(adType uint8, payload []byte)) {
	for offset := 0; offset < len(data); {
		length := int(data[offset])
		if length == 0 || offset+1+length > len(data) {
			return
		}
		if length > 1 {
			visit(data[offset+1], data[offset+2:offset+1+length])
		}
		offset += 1 + length
	}
}
