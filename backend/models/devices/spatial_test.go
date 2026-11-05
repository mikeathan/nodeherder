package devices

import (
	"encoding/json"
	"testing"
	"time"
)

func TestIsSpatialLengthUnit(t *testing.T) {
	for _, testCase := range []struct {
		unit string
		want bool
	}{
		{"mm", true},
		{" MM ", true},
		{"cm", true},
		{"m", true},
		{"m/s", false},
		{"°", false},
		{"", false},
		{"ft", false},
	} {
		if got := IsSpatialLengthUnit(testCase.unit); got != testCase.want {
			t.Errorf("IsSpatialLengthUnit(%q) = %t, want %t", testCase.unit, got, testCase.want)
		}
	}
}

func TestSpatialFrameEncodingMatchesContract(t *testing.T) {
	frame := SpatialFrame{
		Id: "device",
		At: time.Date(2026, 11, 3, 10, 15, 22, 480000000, time.UTC),
		Targets: []SpatialTarget{
			{ID: 1, Active: true, X: &SpatialAxis{Value: 1234, Unit: "mm"}, Y: &SpatialAxis{Value: 2870, Unit: "mm"}},
			{ID: 2, Active: false},
		},
	}
	encoded, err := json.Marshal(frame)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["id"] != "device" || decoded["at"] != "2026-11-03T10:15:22.48Z" {
		t.Fatalf("frame envelope = %v", decoded)
	}
	targets, ok := decoded["targets"].([]any)
	if !ok || len(targets) != 2 {
		t.Fatalf("targets = %v", decoded["targets"])
	}
	first := targets[0].(map[string]any)
	if first["id"] != float64(1) || first["active"] != true {
		t.Fatalf("first target = %v", first)
	}
	axis, ok := first["x"].(map[string]any)
	if !ok || axis["value"] != float64(1234) || axis["unit"] != "mm" {
		t.Fatalf("first axis = %v", first["x"])
	}
	second := targets[1].(map[string]any)
	if _, present := second["x"]; present {
		t.Fatalf("cleared slot shipped an axis: %v", second)
	}
	if second["active"] != false {
		t.Fatalf("cleared slot lost its active flag: %v", second)
	}
}

func TestSpatialFrameEmptyTargetsEncodeAsArray(t *testing.T) {
	encoded, err := json.Marshal(SpatialFrame{Id: "device", Targets: []SpatialTarget{}})
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if targets, ok := decoded["targets"].([]any); !ok || len(targets) != 0 {
		t.Fatalf("empty frame targets = %v (want an empty array, never null)", decoded["targets"])
	}
}
