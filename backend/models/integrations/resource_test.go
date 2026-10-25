package integrations

import (
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
