package searchselect

// Checks lookup progress and provider search choice handling.

import (
	"testing"

	model "node-herder/models/integrations"
)

func TestProgressUsesConfiguredStatuses(t *testing.T) {
	step := model.StatusStep{Pointer: "/phase", Pending: []string{"waiting"}, Done: []string{"finished"}, Failed: []string{"broken"}}
	for _, test := range []struct {
		phase, want string
		valid       bool
	}{{"waiting", "pending", true}, {"finished", "done", true}, {"broken", "", false}, {"unknown", "", false}} {
		got, err := Progress(map[string]any{"phase": test.phase}, step)
		if got != test.want || (err == nil) != test.valid {
			t.Fatal(test, got, err)
		}
	}
	if _, err := Progress(map[string]any{}, step); err == nil {
		t.Fatal("missing status accepted")
	}
}

func TestChoicesAreOpaqueAndRejectAmbiguity(t *testing.T) {
	step := model.SearchStep{Options: model.ChoiceMapping{Items: "/stations", Value: "/code", Label: "/title"}}
	item := map[string]any{"code": "station:42", "title": "Station One"}
	choices, err := Choices(map[string]any{"stations": []any{item}}, step)
	if err != nil || len(choices) != 1 || choices[0].ID != "station:42" {
		t.Fatal(choices, err)
	}
	if _, err = Choices(map[string]any{"stations": []any{item, item}}, step); err == nil {
		t.Fatal("duplicate selections accepted")
	}
	step.Expect = &model.Discriminator{Pointer: "/input", Value: "station"}
	if _, err = Choices(map[string]any{"stations": []any{item}}, step); err == nil {
		t.Fatal("unsupported journey accepted")
	}
}
