package spatialhistory

import (
	"testing"
	"time"

	"node-herder/models/devices"
	"node-herder/utils"
)

func target(slot int, x, y float64) devices.SpatialTarget {
	return devices.SpatialTarget{ID: slot, Active: true, X: &devices.SpatialAxis{Value: x, Unit: "mm"}, Y: &devices.SpatialAxis{Value: y, Unit: "mm"}}
}

func TestRecorderRecordsMovementAndIgnoresStillnessAndDuplicates(t *testing.T) {
	store := newTestStore(t)
	recorder := NewRecorder(store, func(string) bool { return true }, utils.NewRealClock())
	base := time.Now().UTC().Truncate(time.Second)

	recorder.handle(job{device: "radar", at: base, targets: []rawTarget{{slot: 1, x: 1000, y: 2000, active: true}}})
	// Duplicate slot at nearly the same place must not create a second point.
	recorder.handle(job{device: "radar", at: base.Add(2 * time.Second), targets: []rawTarget{
		{slot: 1, x: 1040, y: 2000, active: true},
		{slot: 2, x: 1060, y: 2010, active: true},
	}})
	recorder.handle(job{device: "radar", at: base.Add(4 * time.Second), targets: []rawTarget{{slot: 1, x: 1450, y: 2000, active: true}}})
	recorder.handle(job{device: "radar", at: base.Add(6 * time.Second), targets: []rawTarget{{slot: 1, x: 1490, y: 2000, active: true}}})
	recorder.handle(job{device: "radar", at: base.Add(8 * time.Second), targets: nil})

	sessions, err := recorder.Service().Sessions("radar", base.Add(-time.Hour), base.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || sessions[0].Points != 2 {
		t.Fatalf("sessions = %+v", sessions)
	}
	points, err := recorder.Service().Points("radar", sessions[0].ID, base.Add(-time.Hour), base.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 2 || points[1].X != 1450 || points[1].Y != 2000 {
		t.Fatalf("points = %+v", points)
	}
	cells, err := recorder.Service().Heatmap("radar", base, base.Add(time.Hour), 250)
	if err != nil || len(cells) == 0 {
		t.Fatalf("cells = %+v err=%v", cells, err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRecorderFollowsHistoryGate(t *testing.T) {
	store := newTestStore(t)
	enabled := false
	recorder := NewRecorder(store, func(string) bool { return enabled }, utils.NewRealClock())
	base := time.Now().UTC().Truncate(time.Second)
	recorder.handle(job{device: "radar", at: base, targets: []rawTarget{{slot: 1, x: 1000, y: 2000, active: true}}})
	sessions, err := recorder.Service().Sessions("radar", base.Add(-time.Hour), base.Add(time.Hour))
	if err != nil || len(sessions) != 0 {
		t.Fatalf("disabled history recorded: %+v err=%v", sessions, err)
	}
	enabled = true
	recorder.handle(job{device: "radar", at: base.Add(time.Second), targets: []rawTarget{{slot: 1, x: 1000, y: 2000, active: true}}, active: true})
	recorder.handle(job{device: "radar", at: base.Add(2 * time.Second), targets: nil})
	sessions, err = recorder.Service().Sessions("radar", base.Add(-time.Hour), base.Add(time.Hour))
	if err != nil || len(sessions) != 1 || sessions[0].Points != 1 {
		t.Fatalf("enabled history missed: %+v err=%v", sessions, err)
	}
	_ = store.Close()
}
