package searchselect

// Checks location selection recipes and provider location choices.

import (
	"encoding/json"
	"math"
	"testing"

	model "node-herder/models/integrations"
)

func TestSelectedCoordinates(t *testing.T) {
	latitude, longitude, ok, err := SelectedCoordinates(map[string]any{"latitude": 51.45, "longitude": -2.59})
	if err != nil || !ok || latitude != 51.45 || longitude != -2.59 {
		t.Fatalf("unexpected coordinates: %v, %v, %v, %v", latitude, longitude, ok, err)
	}
	if _, _, ok, err = SelectedCoordinates(nil); err != nil || ok {
		t.Fatal("absent location must not be an origin")
	}
	_, _, ok, err = SelectedCoordinates(map[string]any{"latitude": 51.45, "longitude": json.Number("-2.59")})
	if err != nil || !ok {
		t.Fatal("json number coordinates rejected", err)
	}
	for _, state := range []map[string]any{
		{"latitude": 91.0, "longitude": 0.0},
		{"latitude": 0.0, "longitude": -181.0},
		{"latitude": math.NaN(), "longitude": 0.0},
		{"latitude": 51.45},
		{"latitude": "51.45", "longitude": "-2.59"},
	} {
		if _, _, _, err := SelectedCoordinates(state); err == nil {
			t.Fatalf("accepted invalid origin %#v", state)
		}
	}
}

func TestLocationChoicesIncludeContext(t *testing.T) {
	step := model.SearchStep{Options: model.ChoiceMapping{Items: "/result", Label: "/postcode", Latitude: "/latitude", Longitude: "/longitude", Context: []string{"/admin_district", "/country"}}}
	body := map[string]any{"result": []any{map[string]any{"postcode": "S8 9ED", "admin_district": "Sheffield", "country": "England", "latitude": 53.3, "longitude": -1.4}}}
	choices, err := Choices(body, step)
	if err != nil || len(choices) != 1 || choices[0].Label != "S8 9ED, Sheffield, England" || choices[0].ID != "53.3,-1.4" {
		t.Fatalf("unexpected choices: %#v, %v", choices, err)
	}
}
