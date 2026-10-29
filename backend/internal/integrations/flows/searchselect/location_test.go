package searchselect

// Checks location selection recipes and provider location choices.

import (
	"testing"

	model "node-herder/models/integrations"
)

func TestLocationChoicesIncludeContext(t *testing.T) {
	step := model.SearchStep{Options: model.ChoiceMapping{Items: "/result", Label: "/postcode", Latitude: "/latitude", Longitude: "/longitude", Context: []string{"/admin_district", "/country"}}}
	body := map[string]any{"result": []any{map[string]any{"postcode": "S8 9ED", "admin_district": "Sheffield", "country": "England", "latitude": 53.3, "longitude": -1.4}}}
	choices, err := Choices(body, step)
	if err != nil || len(choices) != 1 || choices[0].Label != "S8 9ED, Sheffield, England" || choices[0].ID != "53.3,-1.4" {
		t.Fatalf("unexpected choices: %#v, %v", choices, err)
	}
}
