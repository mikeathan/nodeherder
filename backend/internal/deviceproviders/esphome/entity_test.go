package esphome

import (
	"math"
	"testing"

	"github.com/mycontroller-org/esphome_api/pkg/api"
)

func TestDescribeDiscoveredEntities(t *testing.T) {
	d, err := Describe(&api.ListEntitiesSensorResponse{Key: 42, Name: "Temperature", UnitOfMeasurement: "°C"})
	if err != nil || d.ID != "00000000-0000002a" || d.Kind != "sensor" || d.Writable {
		t.Fatalf("descriptor: %+v, %v", d, err)
	}
	options := []string{"one", "two"}
	d, err = Describe(&api.ListEntitiesSelectResponse{Key: 3, Options: options})
	if err != nil {
		t.Fatal(err)
	}
	options[0] = "changed"
	if d.Options[0] != "one" || !d.Writable {
		t.Fatalf("descriptor: %+v", d)
	}
	if _, err := Describe(&api.ListEntitiesNumberResponse{MinValue: float32(math.NaN()), MaxValue: 1, Step: 1}); err == nil {
		t.Fatal("NaN accepted")
	}
	if _, err := Describe(&api.ListEntitiesNumberResponse{MinValue: 2, MaxValue: 1, Step: 1}); err == nil {
		t.Fatal("inverted range accepted")
	}
	if _, err := Describe(&api.ListEntitiesSwitchResponse{Name: string([]byte{0xff})}); err == nil {
		t.Fatal("invalid UTF-8 accepted")
	}
	if d, err := Describe(&api.PingResponse{}); err != nil || d != nil {
		t.Fatalf("unknown message: %+v %v", d, err)
	}
}

func TestDescribeTextConstraintsAndSecrets(t *testing.T) {
	text, err := Describe(&api.ListEntitiesTextResponse{Key: 3, MaxLength: 32, Pattern: `[0-9:;]*`})
	if err != nil || text.Kind != "text" || !text.Writable {
		t.Fatalf("text: %+v %v", text, err)
	}
	secret, err := Describe(&api.ListEntitiesTextResponse{Key: 4, Mode: api.TextMode_TEXT_MODE_PASSWORD})
	if err != nil || secret.Kind != "secret_text" || secret.Writable {
		t.Fatalf("secret: %+v %v", secret, err)
	}
	unsupported, err := Describe(&api.ListEntitiesTextResponse{MaxLength: 33, Pattern: `(?=a)`})
	if err != nil || unsupported == nil || unsupported.Writable || unsupported.Kind != "text" {
		t.Fatalf("unsupported pattern must remain readable: %+v %v", unsupported, err)
	}
}

func TestDescribeLightAndUnknownEntity(t *testing.T) {
	light, err := Describe(&api.ListEntitiesLightResponse{Key: 1, Name: "Light"})
	if err != nil || light.Kind != "light" || !light.Writable {
		t.Fatalf("light: %+v %v", light, err)
	}
	light, err = Describe(&api.ListEntitiesLightResponse{Key: 1, SupportedColorModes: []api.ColorMode{api.ColorMode_COLOR_MODE_BRIGHTNESS}})
	if err != nil || !light.LimitedControl {
		t.Fatalf("light limitations: %+v %v", light, err)
	}
	camera, err := Describe(&api.ListEntitiesCameraResponse{Key: 2, Name: "Camera"})
	if err != nil || camera.Kind != "unsupported" || camera.Writable {
		t.Fatalf("camera: %+v %v", camera, err)
	}
}
