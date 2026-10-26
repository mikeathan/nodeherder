package integrations

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestProject(t *testing.T) {
	source := time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC)
	expiry := source.Add(24 * time.Hour)
	s := WasteSchedule{Revision: 7, SourceRetrievedAt: &source, FreshUntil: &expiry}
	if s.Project(source).Freshness != "fresh" || s.Project(expiry).Freshness != "stale" {
		t.Fatal("freshness")
	}
	if s.Revision != 7 || !s.SourceRetrievedAt.Equal(source) {
		t.Fatal("mutated source")
	}
	s.FreshUntil = nil
	if s.Project(source).Freshness != "unknown" {
		t.Fatal("guessed expiry")
	}
}

func TestProjectCollectionOrder(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	snapshot := WasteSchedule{Timezone: "Europe/London", Revision: 7, Services: []WasteService{
		{ServiceID: "blue", Dates: []string{"2026-11-17"}},
		{ServiceID: "empty", DateState: "confirmed_empty", Dates: []string{}},
		{ServiceID: "black", Dates: []string{"2026-10-27"}},
		{ServiceID: "brown", Dates: []string{"2026-11-03"}},
		{ServiceID: "same-day", Dates: []string{"2026-11-03"}},
		{ServiceID: "exhausted", Dates: []string{"2026-10-25"}},
	}}
	before, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	projected := snapshot.Project(now)
	var ids []string
	for _, service := range projected.Services {
		ids = append(ids, service.ServiceID)
	}
	if !reflect.DeepEqual(ids, []string{"black", "brown", "same-day", "blue", "empty", "exhausted"}) {
		t.Fatal(ids)
	}
	after, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) || projected.Revision != snapshot.Revision {
		t.Fatal("projection changed saved snapshot/identity")
	}
}

func TestProjectReordersAtUKMidnight(t *testing.T) {
	snapshot := WasteSchedule{Timezone: "Europe/London", Revision: 9, Services: []WasteService{
		{ServiceID: "tomorrow", Dates: []string{"2026-10-25"}},
		{ServiceID: "today", Dates: []string{"2026-10-24"}},
		{ServiceID: "later", Dates: []string{"2026-10-26"}},
	}}
	for _, tc := range []struct {
		name string
		now  time.Time
		want []string
	}{
		{"before midnight BST", time.Date(2026, 10, 24, 22, 30, 0, 0, time.UTC), []string{"today", "tomorrow", "later"}},
		{"after midnight BST", time.Date(2026, 10, 24, 23, 30, 0, 0, time.UTC), []string{"tomorrow", "later", "today"}},
		{"after clock change GMT", time.Date(2026, 10, 26, 0, 30, 0, 0, time.UTC), []string{"later", "tomorrow", "today"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			projected := snapshot.Project(tc.now)
			var ids []string
			for _, service := range projected.Services {
				ids = append(ids, service.ServiceID)
			}
			if !reflect.DeepEqual(ids, tc.want) || projected.Revision != 9 {
				t.Fatal(ids, projected.Revision)
			}
		})
	}
	if snapshot.Services[0].ServiceID != "tomorrow" || snapshot.Services[1].Dates[0] != "2026-10-24" {
		t.Fatal("projection mutated source or removed historical dates")
	}
}
