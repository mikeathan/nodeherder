package runtime

// Checks the optional origin projected from a private setup checkpoint.

import "testing"

func TestSelectedOrigin(t *testing.T) {
	origin, err := selectedOrigin("location_select", map[string]any{"latitude": 51.45, "longitude": -2.59})
	if err != nil || origin == nil || origin.Latitude != 51.45 || origin.Longitude != -2.59 {
		t.Fatalf("unexpected origin: %#v, %v", origin, err)
	}
	if origin, err = selectedOrigin("location_select", nil); err != nil || origin != nil {
		t.Fatalf("absent location must not be an origin: %#v, %v", origin, err)
	}
	if origin, err = selectedOrigin("search_select", map[string]any{"latitude": 51.45, "longitude": -2.59}); err != nil || origin != nil {
		t.Fatalf("non-location setup must not have an origin: %#v, %v", origin, err)
	}
	if _, err = selectedOrigin("location_select", map[string]any{"latitude": 91.0, "longitude": 0.0}); err == nil {
		t.Fatal("invalid origin accepted")
	}
}
