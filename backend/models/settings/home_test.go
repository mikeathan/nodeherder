package settings_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"node-herder/models/settings"
)

func homeGroup(name string) *settings.DashboardGroup {
	return &settings.DashboardGroup{Name: name, DeviceGroup: map[string]*settings.DeviceGroup{
		"device": {DeviceId: "device", Exposes: []string{"temperature", "humidity"}},
	}}
}

func TestHomeMigrationIdentityOrderingAndProjection(t *testing.T) {
	groups := map[string]*settings.DashboardGroup{"Kitchen": homeGroup("Kitchen"), "Garden": homeGroup("Garden")}
	order := 0
	groups["Kitchen"].Order = &order
	layout := settings.HomeFromGroups(groups)
	if err := layout.Validate(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(layout, settings.HomeFromGroups(groups)) || layout.Blocks[0].Name != "Kitchen" {
		t.Fatal("migration must be deterministic and preserve explicit order")
	}
	if got := layout.Blocks[0].Cards[0]; got.Expose != "temperature" || !got.Panel {
		t.Fatal("migration changed expose order/panel opt-in", got)
	}
	clone := layout.Clone()
	clone.Blocks[0].Cards[0].Expose = "mutated"
	if layout.Blocks[0].Cards[0].Expose != "temperature" {
		t.Fatal("clone shares accepted cards")
	}
	layout.Blocks[0].Cards = append(layout.Blocks[0].Cards, settings.HomeCard{ID: "waste", Kind: "waste_service", Panel: true, InstanceID: "private-instance", ResourceID: "private-resource", ServiceID: "black"})
	projection, err := json.Marshal(layout.DeviceGroups())
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]*settings.DashboardGroup
	if err := json.Unmarshal(projection, &decoded); err != nil || len(decoded["Kitchen"].DeviceGroup) != 1 {
		t.Fatal("invalid legacy projection", string(projection), err)
	}
	if reflect.DeepEqual(layout.Blocks[0].Cards, settings.HomeFromGroups(layout.DeviceGroups()).Blocks[0].Cards) {
		t.Fatal("private integration reference entered device projection")
	}
}

func TestHomeValidationRejectsAmbiguityAndAliasing(t *testing.T) {
	valid := settings.HomeFromGroups(map[string]*settings.DashboardGroup{"Kitchen": homeGroup("Kitchen")})
	cases := map[string]func(*settings.HomeLayout){
		"version":          func(l *settings.HomeLayout) { l.Version = 2 },
		"null blocks":      func(l *settings.HomeLayout) { l.Blocks = nil },
		"null cards":       func(l *settings.HomeLayout) { l.Blocks[0].Cards = nil },
		"duplicate block":  func(l *settings.HomeLayout) { l.Blocks = append(l.Blocks, l.Blocks[0]) },
		"duplicate card":   func(l *settings.HomeLayout) { l.Blocks[0].Cards = append(l.Blocks[0].Cards, l.Blocks[0].Cards[0]) },
		"mixed references": func(l *settings.HomeLayout) { l.Blocks[0].Cards[0].InstanceID = "not-a-device" },
		"unknown renderer": func(l *settings.HomeLayout) { l.Blocks[0].Cards[0].Kind = "arbitrary-html" },
		"empty card block": func(l *settings.HomeLayout) {
			l.Blocks[0] = settings.HomeBlock{ID: "card", Kind: "card", Cards: []settings.HomeCard{}}
		},
		"case duplicate names": func(l *settings.HomeLayout) {
			l.Blocks = append(l.Blocks, settings.HomeBlock{ID: "second", Kind: "group", Name: "kitchen", Cards: []settings.HomeCard{}})
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			proposal := valid.Clone()
			mutate(&proposal)
			if proposal.Validate() == nil {
				t.Fatal("invalid layout accepted")
			}
		})
	}
	empty := settings.HomeLayout{Version: 1, Blocks: []settings.HomeBlock{}}
	if err := empty.Clone().Validate(); err != nil {
		t.Fatal("empty layout must remain an array", err)
	}
	card := valid.Blocks[0].Cards[0]
	other := card
	other.ID = "another-placement"
	other.Panel = false
	if other.SourceKey() != card.SourceKey() {
		t.Fatal("presentation choices changed source identity")
	}
}

func TestWeatherHomeLayoutReferences(t *testing.T) {
	layout := settings.HomeLayout{Version: 1, Blocks: []settings.HomeBlock{{ID: "block", Kind: "card", Cards: []settings.HomeCard{{ID: "weather", Kind: "weather", InstanceID: "connection", ResourceID: "forecast", Panel: true}}}}}
	if err := layout.Validate(); err != nil {
		t.Fatal(err)
	}
	layout.Blocks[0].Cards[0].DeviceID = "device"
	if layout.Validate() == nil {
		t.Fatal("mixed weather/service identity accepted")
	}
}

func TestHomePresenceCardReferencesWholeDevice(t *testing.T) {
	layout := settings.HomeFromGroups(map[string]*settings.DashboardGroup{"Kitchen": homeGroup("Kitchen")})
	layout.Blocks[0].Cards[0] = settings.HomeCard{ID: "presence", Kind: "presence", Panel: true, DeviceID: "device"}
	if err := layout.Validate(); err != nil {
		t.Fatalf("valid presence card rejected: %v", err)
	}
	invalid := []settings.HomeCard{
		{ID: "presence", Kind: "presence", Panel: true},
		{ID: "presence", Kind: "presence", Panel: true, DeviceID: "device", Expose: "temperature"},
		{ID: "presence", Kind: "presence", Panel: true, DeviceID: "device", InstanceID: "instance"},
	}
	for _, card := range invalid {
		candidate := layout.Clone()
		candidate.Blocks[0].Cards[0] = card
		if candidate.Validate() == nil {
			t.Fatalf("invalid presence card accepted: %+v", card)
		}
	}
}

func TestHomePresenceOverviewCardHasNoSourceReference(t *testing.T) {
	layout := settings.HomeFromGroups(map[string]*settings.DashboardGroup{"Kitchen": homeGroup("Kitchen")})
	layout.Blocks[0].Cards = append(layout.Blocks[0].Cards, settings.HomeCard{ID: "overview", Kind: "presence_overview", Panel: true})
	if err := layout.Validate(); err != nil {
		t.Fatalf("valid presence overview rejected: %v", err)
	}
	invalid := layout.Clone()
	invalid.Blocks[0].Cards[len(invalid.Blocks[0].Cards)-1].DeviceID = "radar"
	if invalid.Validate() == nil {
		t.Fatal("presence overview accepted a device reference")
	}
}

func TestHomeObservationCardHasNoSourceReference(t *testing.T) {
	// The radar reads every enrolled sensor, so the shortcut names no source. This failed in
	// practice: the kind fell through to the integration branch, which requires an instance and
	// resource, so saving a layout with this card returned "invalid or exceeds its limits" and the
	// operator saw the tile vanish.
	layout := settings.HomeFromGroups(map[string]*settings.DashboardGroup{"Kitchen": homeGroup("Kitchen")})
	layout.Blocks[0].Cards = append(layout.Blocks[0].Cards, settings.HomeCard{ID: "radar", Kind: "observation"})
	if err := layout.Validate(); err != nil {
		t.Fatalf("valid observation card rejected: %v", err)
	}

	// It is whole-hub, so any source reference on it is a malformed card rather than a narrower one.
	for _, mutate := range []func(*settings.HomeCard){
		func(card *settings.HomeCard) { card.DeviceID = "proxy" },
		func(card *settings.HomeCard) { card.Expose = "rssi" },
		func(card *settings.HomeCard) { card.InstanceID = "instance" },
		func(card *settings.HomeCard) { card.ResourceID = "resource" },
		func(card *settings.HomeCard) { card.ServiceID = "service" },
	} {
		invalid := layout.Clone()
		mutate(&invalid.Blocks[0].Cards[len(invalid.Blocks[0].Cards)-1])
		if invalid.Validate() == nil {
			t.Fatal("observation card accepted a source reference")
		}
	}
}
