package observations

import (
	"encoding/hex"
	"net"
	"testing"

	model "node-herder/models/observations"
)

// buildAD assembles BLE AD structures for a test packet.
func buildAD(structures ...[2][]byte) []byte {
	var out []byte
	for _, s := range structures {
		payload := s[1]
		out = append(out, byte(len(payload)+1), s[0][0])
		out = append(out, payload...)
	}
	return out
}

func mustAddress(t *testing.T, value string) net.HardwareAddr {
	t.Helper()
	address, err := net.ParseMAC(value)
	if err != nil {
		t.Fatalf("parse address %q: %v", value, err)
	}
	return address
}

// ibeaconData builds an Apple iBeacon payload with the given UUID, major and minor.
func ibeaconData(t *testing.T, uuid string, major, minor uint16) []byte {
	t.Helper()
	raw, err := hex.DecodeString(removeUUIDDashes(uuid))
	if err != nil || len(raw) != 16 {
		t.Fatalf("bad uuid %q", uuid)
	}
	body := append([]byte{ibeaconType, ibeaconLength}, raw...)
	body = append(body, byte(major>>8), byte(major), byte(minor>>8), byte(minor))
	// Company identifier is little-endian.
	payload := append([]byte{byte(companyApple), byte(companyApple >> 8)}, body...)
	return buildAD([2][]byte{{adManufacturerData}, payload})
}

func removeUUIDDashes(value string) string {
	out := make([]byte, 0, len(value))
	for i := 0; i < len(value); i++ {
		if value[i] != '-' {
			out = append(out, value[i])
		}
	}
	return string(out)
}

func TestResolveIdentityPrefersBeaconPayload(t *testing.T) {
	const uuid = "2f234454-cf6d-4a0f-adf2-f4911ba9ffa6"
	identity, _, ok := ResolveIdentity(Advertisement{
		// A rotating address must not override a payload identity.
		Address:     mustAddress(t, "4d:aa:bb:cc:dd:ee"),
		AddressType: AddressTypeRandom,
		Data:        ibeaconData(t, uuid, 1, 2),
	})
	if !ok {
		t.Fatal("ibeacon packet produced no identity")
	}
	want := model.KindIBeacon + ":" + uuid + ":1:2"
	if identity.Key != want {
		t.Fatalf("key = %q, want %q", identity.Key, want)
	}
	if identity.Kind != model.KindIBeacon || identity.Vendor != companyApple {
		t.Fatalf("kind/vendor = %q/%#x", identity.Kind, identity.Vendor)
	}
	if !identity.Stable {
		t.Fatal("ibeacon identity must be stable")
	}
}

func TestResolveIdentityClassifiesAddressType(t *testing.T) {
	for _, test := range []struct {
		name        string
		address     string
		addressType uint8
		wantStable  bool
	}{
		{"public", "aa:bb:cc:dd:ee:01", AddressTypePublic, true},
		{"public identity", "aa:bb:cc:dd:ee:02", AddressTypePublicID, true},
		{"static random", "c0:bb:cc:dd:ee:03", AddressTypeRandom, true},
		{"resolvable private", "40:bb:cc:dd:ee:04", AddressTypeRandom, false},
		{"non-resolvable private", "00:bb:cc:dd:ee:05", AddressTypeRandom, false},
		{"unknown type", "aa:bb:cc:dd:ee:06", 9, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			identity, _, ok := ResolveIdentity(Advertisement{
				Address:     mustAddress(t, test.address),
				AddressType: test.addressType,
			})
			if !ok {
				t.Fatal("packet produced no identity")
			}
			if identity.Stable != test.wantStable {
				t.Fatalf("stable = %v, want %v", identity.Stable, test.wantStable)
			}
			wantPrefix := "fp:"
			if test.wantStable {
				wantPrefix = "addr:"
			}
			if len(identity.Key) < len(wantPrefix) || identity.Key[:len(wantPrefix)] != wantPrefix {
				t.Fatalf("key = %q, want prefix %q", identity.Key, wantPrefix)
			}
		})
	}
}

func TestResolveIdentityRejectsUnusableAddress(t *testing.T) {
	short := mustAddress(t, "aa:bb:cc:dd:ee:ff")[:5]
	for _, address := range []net.HardwareAddr{nil, {}, short} {
		if _, _, ok := ResolveIdentity(Advertisement{Address: address, AddressType: AddressTypePublic}); ok {
			t.Fatalf("address %v produced an identity", address)
		}
	}
}

func TestResolveIdentityFingerprintIsStablePerAddress(t *testing.T) {
	first, _, _ := ResolveIdentity(Advertisement{Address: mustAddress(t, "40:11:22:33:44:55"), AddressType: AddressTypeRandom})
	repeat, _, _ := ResolveIdentity(Advertisement{Address: mustAddress(t, "40:11:22:33:44:55"), AddressType: AddressTypeRandom})
	rotated, _, _ := ResolveIdentity(Advertisement{Address: mustAddress(t, "40:11:22:33:44:56"), AddressType: AddressTypeRandom})
	if first.Key != repeat.Key {
		t.Fatalf("same address produced %q and %q", first.Key, repeat.Key)
	}
	if first.Key == rotated.Key {
		t.Fatal("rotated address reused the previous key")
	}
	if len(first.Key) != len("fp:")+fingerprintHexLength {
		t.Fatalf("fingerprint key %q is not short and fixed width", first.Key)
	}
}

func TestResolveIdentityReadsLocalNames(t *testing.T) {
	for _, test := range []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "complete name",
			data: buildAD([2][]byte{{adCompleteLocalName}, []byte("Kitchen tag")}),
			want: "Kitchen tag",
		},
		{
			name: "short name when complete absent",
			data: buildAD([2][]byte{{adShortLocalName}, []byte("Kit")}),
			want: "Kit",
		},
		{
			name: "complete preferred over short",
			data: buildAD(
				[2][]byte{{adShortLocalName}, []byte("Kit")},
				[2][]byte{{adCompleteLocalName}, []byte("Kitchen tag")},
			),
			want: "Kitchen tag",
		},
		{name: "absent", data: buildAD([2][]byte{{adFlags}, []byte{0x06}}), want: ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, name, ok := ResolveIdentity(Advertisement{
				Address:     mustAddress(t, "aa:bb:cc:dd:ee:07"),
				AddressType: AddressTypePublic,
				Data:        test.data,
			})
			if !ok || name != test.want {
				t.Fatalf("name = %q (ok=%v), want %q", name, ok, test.want)
			}
		})
	}
}

func TestResolveIdentityIgnoresMalformedData(t *testing.T) {
	for _, data := range [][]byte{
		{0x00},
		{0x05, adManufacturerData},
		{0x03, adManufacturerData, 0x4C},
		{0xFF, adManufacturerData},
		{0x02, adManufacturerData, 0x4C},
	} {
		identity, _, ok := ResolveIdentity(Advertisement{
			Address:     mustAddress(t, "aa:bb:cc:dd:ee:08"),
			AddressType: AddressTypePublic,
			Data:        data,
		})
		if !ok {
			t.Fatalf("data %v produced no identity", data)
		}
		if identity.Kind != model.KindAdvertisement || !identity.Stable {
			t.Fatalf("data %v produced %q stable=%v", data, identity.Kind, identity.Stable)
		}
	}
}

func TestResolveIdentityParsesEddystoneUID(t *testing.T) {
	// Service data: 16-bit UUID (little-endian) then frame type, tx power, namespace and
	// instance.
	namespace, _ := hex.DecodeString("00112233445566778899")
	instance, _ := hex.DecodeString("aabbccddeeff")
	body := append([]byte{0x00, 0xEC}, namespace...)
	body = append(body, instance...)
	payload := append([]byte{byte(serviceEddystone & 0xFF), byte(serviceEddystone >> 8)}, body...)
	data := buildAD([2][]byte{{adServiceData16}, payload})

	identity, _, ok := ResolveIdentity(Advertisement{
		Address:     mustAddress(t, "40:aa:bb:cc:dd:ef"),
		AddressType: AddressTypeRandom,
		Data:        data,
	})
	if !ok {
		t.Fatal("eddystone packet produced no identity")
	}
	want := model.KindEddystone + ":00112233445566778899:aabbccddeeff"
	if identity.Key != want || identity.Kind != model.KindEddystone || !identity.Stable {
		t.Fatalf("identity = %+v, want key %q", identity, want)
	}
}
