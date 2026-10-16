package automations_test

import (
	"node-herder/internal/automations"
	"testing"
)

func TestConditionDeviceDependencies(t *testing.T) {
	for _, tc := range []struct {
		name      string
		condition automations.Condition
		property  string
		want      bool
	}{
		{"expose owns its dependency", &automations.ExposeCondition{Name: "illuminance"}, "illuminance", true},
		{"different expose is unrelated", &automations.ExposeCondition{Name: "illuminance"}, "presence", false},
		{"base has no device input", &automations.BaseCondition{}, "illuminance", false},
		{"time does not depend on a device field", &automations.TimeCondition{}, "illuminance", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.condition.DependsOn(tc.property); got != tc.want {
				t.Fatalf("DependsOn(%q)=%v, want %v", tc.property, got, tc.want)
			}
		})
	}
}
