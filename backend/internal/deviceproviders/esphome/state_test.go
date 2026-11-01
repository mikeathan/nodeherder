package esphome

import (
	"math"
	"testing"

	"github.com/mycontroller-org/esphome_api/pkg/api"
)

func TestDecodeState(t *testing.T) {
	d, err := Describe(&api.ListEntitiesSensorResponse{Key: 4, DeviceId: 2})
	if err != nil {
		t.Fatal(err)
	}
	for _, state := range []float32{float32(math.NaN()), float32(math.Inf(1))} {
		got, err := DecodeState(*d, &api.SensorStateResponse{Key: 4, DeviceId: 2, State: state})
		if err != nil || got.Available || got.Value != nil {
			t.Fatalf("non-finite value: %+v %v", got, err)
		}
	}
	got, err := DecodeState(*d, &api.SensorStateResponse{Key: 4, DeviceId: 2, State: 0})
	if err != nil || !got.Available || got.Value != float64(0) {
		t.Fatalf("zero: %+v %v", got, err)
	}
	got, err = DecodeState(*d, &api.SensorStateResponse{Key: 4, DeviceId: 2, MissingState: true})
	if err != nil || got.Available || got.Value != nil {
		t.Fatalf("missing: %+v %v", got, err)
	}
	if _, err := DecodeState(*d, &api.SensorStateResponse{Key: 4, DeviceId: 1}); err == nil {
		t.Fatal("accepted wrong subdevice")
	}
	if _, err := DecodeState(*d, &api.SwitchStateResponse{Key: 4, DeviceId: 2}); err == nil {
		t.Fatal("accepted wrong type")
	}
}

func TestSecretTextStateNeverExposesValue(t *testing.T) {
	d, err := Describe(&api.ListEntitiesTextResponse{Key: 4, Mode: api.TextMode_TEXT_MODE_PASSWORD})
	if err != nil {
		t.Fatal(err)
	}
	observation, err := DecodeState(*d, &api.TextStateResponse{Key: 4, State: "secret"})
	if err != nil || observation.Available || observation.Value != nil {
		t.Fatalf("secret exposed: %+v %v", observation, err)
	}
}
