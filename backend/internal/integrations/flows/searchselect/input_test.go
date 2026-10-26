package searchselect

import (
	"testing"

	model "node-herder/models/integrations"
)

func TestNormalizeInput(t *testing.T) {
	for _, test := range []struct {
		kind, value, want string
		valid             bool
	}{
		{"text", "  Station One  ", "Station One", true},
		{"postcode", " s8 9ed ", "S8 9ED", true},
		{"postcode", "not a postcode", "", false},
		{"text", "  ", "", false},
		{"select", "selection", "", false},
	} {
		t.Run(test.kind+"/"+test.value, func(t *testing.T) {
			got, err := NormalizeInput(test.value, model.Input{Type: test.kind})
			if (err == nil) != test.valid || got != test.want {
				t.Fatal(got, err)
			}
		})
	}
}
