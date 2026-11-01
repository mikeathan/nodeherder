package esphome

import (
	"math"
	"testing"

	"github.com/mycontroller-org/esphome_api/pkg/api"
)

func TestEncodeCommand(t *testing.T) {
	button := Descriptor{Kind: "button", Key: 3, DeviceID: 2, Writable: true}
	msg, err := EncodeCommand(button, nil)
	if err != nil || msg.(*api.ButtonCommandRequest).DeviceId != 2 {
		t.Fatalf("button: %v %v", msg, err)
	}
	if _, err := EncodeCommand(button, true); err == nil {
		t.Fatal("button accepted value")
	}
	switchDesc := Descriptor{Kind: "switch", Key: 9, Writable: true}
	msg, err = EncodeCommand(switchDesc, true)
	if err != nil || !msg.(*api.SwitchCommandRequest).State {
		t.Fatalf("switch: %v %v", msg, err)
	}
	if _, err := EncodeCommand(switchDesc, "true"); err == nil {
		t.Fatal("switch accepted text")
	}
	number := Descriptor{Kind: "number", Writable: true, Min: 0, Max: 10, Step: 0.5}
	if _, err := EncodeCommand(number, 1.5); err != nil {
		t.Fatal(err)
	}
	for _, value := range []float64{-1, 11, 1.1, math.NaN(), math.Inf(1)} {
		if _, err := EncodeCommand(number, value); err == nil {
			t.Fatalf("accepted %v", value)
		}
	}
	selection := Descriptor{Kind: "select", Writable: true, Options: []string{"a", "b"}}
	if _, err := EncodeCommand(selection, "a"); err != nil {
		t.Fatal(err)
	}
	if _, err := EncodeCommand(selection, "c"); err == nil {
		t.Fatal("accepted unknown option")
	}
	selection.Writable = false
	if _, err := EncodeCommand(selection, "a"); err == nil {
		t.Fatal("accepted read-only command")
	}
}

func TestEncodeTextCommandEnforcesConstraints(t *testing.T) {
	d, err := Describe(&api.ListEntitiesTextResponse{Key: 3, MinLength: 1, MaxLength: 8, Pattern: `[0-9:;]+`})
	if err != nil {
		t.Fatal(err)
	}
	message, err := EncodeCommand(*d, "1:2;3:4")
	if err != nil || message.(*api.TextCommandRequest).State != "1:2;3:4" {
		t.Fatalf("command: %v %v", message, err)
	}
	for _, value := range []string{"", "bad", "123456789"} {
		if _, err := EncodeCommand(*d, value); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
}

func TestLightOnOffCommand(t *testing.T) {
	d, err := Describe(&api.ListEntitiesLightResponse{Key: 7})
	if err != nil {
		t.Fatal(err)
	}
	message, err := EncodeCommand(*d, true)
	if err != nil {
		t.Fatal(err)
	}
	command, ok := message.(*api.LightCommandRequest)
	if !ok || !command.HasState || !command.State || command.Key != 7 {
		t.Fatalf("command: %v", message)
	}
}
