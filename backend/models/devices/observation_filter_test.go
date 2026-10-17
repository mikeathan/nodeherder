package devices_test

import (
	"math"
	"node-herder/models/bridge"
	"node-herder/models/devices"
	"testing"
)

func TestObservationFilterJitterAndCumulativeDrift(t *testing.T) {
	e := devices.NewEntity("linkquality")
	e.Type, e.AccessMode = bridge.NumericDataType, bridge.ReadAccessMode
	f := devices.NewObservationFilter()
	f.Prime(e, float64(134))
	for _, value := range []float64{138, 134, 131, 134, 138, 134} {
		if f.Accept(e, value, 10) {
			t.Fatalf("jitter %v published", value)
		}
	}
	if !f.Accept(e, float64(144), 10) {
		t.Fatal("exact threshold must publish")
	}
	if f.Accept(e, float64(145), 10) {
		t.Fatal("baseline must advance only on publication")
	}
	if !f.Accept(e, float64(154), 10) {
		t.Fatal("cumulative drift must publish")
	}
}

func TestObservationFilterPreservesEventsAndConfirmations(t *testing.T) {
	for _, tc := range []struct {
		name   string
		kind   bridge.ExposeDataType
		access bridge.ExposeAccessMode
		value  any
	}{
		{"presence", bridge.BinaryDataType, bridge.ReadAccessMode, true},
		{"action", bridge.NumericDataType, bridge.ReadAccessMode, float64(1)},
		{"brightness", bridge.NumericDataType, bridge.ReadWriteAccessMode, float64(1)},
		{"mode", bridge.EnumDataType, bridge.ReadAccessMode, "auto"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := devices.NewEntity(tc.name)
			e.Type, e.AccessMode = tc.kind, tc.access
			f := devices.NewObservationFilter()
			f.Prime(e, tc.value)
			if !f.Accept(e, tc.value, 100) || !f.Accept(e, tc.value, 100) {
				t.Fatal("event/confirmation suppressed")
			}
		})
	}
}

func TestObservationFilterMissingInvalidAndDisabled(t *testing.T) {
	e := devices.NewEntity("target_distance")
	e.Type, e.AccessMode = bridge.NumericDataType, bridge.ReadAccessMode
	f := devices.NewObservationFilter()
	if !f.Accept(e, float64(2), .1) {
		t.Fatal("first value lost")
	}
	if f.Accept(e, float64(2.03), .1) {
		t.Fatal("small fluctuation published")
	}
	if !f.Accept(e, float64(2.04), 0) {
		t.Fatal("zero must disable filtering")
	}
	for _, value := range []any{nil, "unknown", math.NaN(), math.Inf(1)} {
		if !f.Accept(e, value, .1) {
			t.Fatalf("non-numeric evidence %v must not be hidden", value)
		}
	}
	if !f.Accept(e, float64(2.04), .1) {
		t.Fatal("recovery after missing value must publish")
	}
}

func TestObservationFilterUnknownSnapshotClearsBaseline(t *testing.T) {
	e := devices.NewEntity("distance")
	e.Type, e.AccessMode = bridge.NumericDataType, bridge.ReadAccessMode
	f := devices.NewObservationFilter()
	f.Prime(e, float64(2))
	f.Prime(e, nil)
	if f.HasBaseline(e.Name) || !f.Accept(e, float64(2), .2) {
		t.Fatal("unknown snapshot hid recovery to previous numeric reading")
	}
}
