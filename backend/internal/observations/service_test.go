package observations

import (
	"context"

	"testing"
	"time"

	"node-herder/mocks"
	model "node-herder/models/observations"
	"node-herder/models/settings"
)

// newFakeClock returns the shared mock clock, so cadence tests never sleep and the
// repository keeps one controllable clock implementation.
func newFakeClock() *mocks.MockClock {
	start := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	return mocks.NewMockClock(func() time.Time { return start })
}

// recordingStore is a stub decision store.
type recordingStore struct {
	decisions map[string]settings.ObservationIdentity
}

func (s *recordingStore) ObservationIdentities() map[string]settings.ObservationIdentity {
	copied := make(map[string]settings.ObservationIdentity, len(s.decisions))
	for key, value := range s.decisions {
		copied[key] = value
	}
	return copied
}

func stableObservation(source, key string, rssi int, at time.Time) model.Observation {
	return model.Observation{
		Source:     source,
		Medium:     model.MediumBLEAdvertisement,
		Identity:   model.Identity{Key: key, Kind: model.KindIBeacon, Stable: true},
		RSSI:       rssi,
		ObservedAt: at,
	}
}

func TestObserveDropsOldestAndCountsWhenSaturated(t *testing.T) {
	clock := newFakeClock()
	dropped := 0
	service := NewService(Deps{
		Store:   &recordingStore{},
		Clock:   clock,
		Queue:   2,
		Dropped: func() { dropped++ },
	})
	// The worker is not started, so the queue stays full.
	for i := 0; i < 2; i++ {
		if err := service.ObserveAdvertisements(context.Background(), []model.Observation{stableObservation("a", "addr:01", -60, clock.Now())}); err != nil {
			t.Fatalf("observe: %v", err)
		}
	}
	if err := service.ObserveAdvertisements(context.Background(), []model.Observation{stableObservation("a", "addr:02", -61, clock.Now())}); err != nil {
		t.Fatalf("observe: %v", err)
	}
	if dropped != 1 {
		t.Fatalf("dropped = %d, want 1", dropped)
	}
	if len(service.queue) != 2 {
		t.Fatalf("queue length = %d, want the bounded capacity 2", len(service.queue))
	}
}

func TestFlushPublishesOnlyWhenStateChanges(t *testing.T) {
	clock := newFakeClock()
	var published []model.Snapshot
	service := NewService(Deps{
		Store:   &recordingStore{},
		Clock:   clock,
		Publish: func(s model.Snapshot) { published = append(published, s) },
	})
	service.RegisterSource("a", "Hall proxy", []model.Medium{model.MediumBLEAdvertisement}, []string{"kind"})

	// A source registration is itself a state change: the browser learns its facets before
	// any identity arrives.
	clock.Advance(time.Second)
	service.flush()
	if len(published) != 1 {
		t.Fatalf("published %d snapshots after registering a source, want 1", len(published))
	}
	if len(published[0].Identities) != 0 || len(published[0].Sources) != 1 {
		t.Fatalf("snapshot = %+v, want one descriptor and no identities", published[0])
	}
	if facets := published[0].Sources[0].Facets; len(facets) != 1 {
		t.Fatalf("facets = %v, want the declared facets", facets)
	}

	service.record(stableObservation("a", "addr:aa", -50, clock.Now()))
	clock.Advance(time.Second)
	service.flush()
	if len(published) != 2 {
		t.Fatalf("published %d snapshots after observations, want 2", len(published))
	}

	// Same state, later interval: still nothing.
	clock.Advance(time.Second)
	service.flush()
	if len(published) != 2 {
		t.Fatalf("published %d snapshots with unchanged state, want 2", len(published))
	}
}

func TestAggregatesAcrossSourcesAndKeepsBestRSSI(t *testing.T) {
	clock := newFakeClock()
	var published []model.Snapshot
	service := NewService(Deps{
		Store:   &recordingStore{},
		Clock:   clock,
		Publish: func(s model.Snapshot) { published = append(published, s) },
	})
	service.RegisterSource("a", "Hall", []model.Medium{model.MediumBLEAdvertisement}, []string{"kind", "vendor"})
	service.RegisterSource("b", "Landing", []model.Medium{model.MediumBLEAdvertisement}, []string{"kind"})

	service.record(stableObservation("a", "addr:aa", -70, clock.Now()))
	service.record(stableObservation("b", "addr:aa", -55, clock.Now()))
	clock.Advance(time.Second)
	service.flush()

	if len(published) != 1 || len(published[0].Identities) != 1 {
		t.Fatalf("published = %d, identities = %d", len(published), len(published[0].Identities))
	}
	identity := published[0].Identities[0]
	if identity.Packets != 2 {
		t.Fatalf("packets = %d, want 2", identity.Packets)
	}
	if identity.RSSI == nil || *identity.RSSI != -55 {
		t.Fatalf("rssi = %v, want the strongest -55", identity.RSSI)
	}
	if len(identity.Sources) != 2 {
		t.Fatalf("sources = %d, want one reading per source", len(identity.Sources))
	}
	if len(published[0].Sources) != 2 {
		t.Fatalf("source descriptors = %d, want 2", len(published[0].Sources))
	}
}

func TestIgnoredIdentitiesAreExcludedFromSnapshots(t *testing.T) {
	clock := newFakeClock()
	store := &recordingStore{decisions: map[string]settings.ObservationIdentity{
		"addr:ignored": {Decision: settings.DecisionIgnored},
		"addr:known":   {Decision: settings.DecisionKnown, Label: "Front door"},
	}}
	var published []model.Snapshot
	service := NewService(Deps{
		Store:   store,
		Clock:   clock,
		Publish: func(s model.Snapshot) { published = append(published, s) },
	})
	service.record(stableObservation("a", "addr:ignored", -60, clock.Now()))
	service.record(stableObservation("a", "addr:known", -65, clock.Now()))
	clock.Advance(time.Second)
	service.flush()

	if len(published) != 1 {
		t.Fatalf("published %d snapshots, want 1", len(published))
	}
	identities := published[0].Identities
	if len(identities) != 1 {
		t.Fatalf("identities = %d, want the ignored one excluded", len(identities))
	}
	if identities[0].Key != "addr:known" || identities[0].Label != "Front door" {
		t.Fatalf("identity = %+v, want known with its operator label", identities[0])
	}
	if identities[0].Decision != string(settings.DecisionKnown) {
		t.Fatalf("decision = %q", identities[0].Decision)
	}
}

func TestEphemeralIdentitiesAreNotCountedAsArrivals(t *testing.T) {
	clock := newFakeClock()
	var published []model.Snapshot
	service := NewService(Deps{
		Store:   &recordingStore{},
		Clock:   clock,
		Publish: func(s model.Snapshot) { published = append(published, s) },
	})
	stable := model.Observation{
		Source: "a", Medium: model.MediumBLEAdvertisement,
		Identity:   model.Identity{Key: "addr:aa", Kind: model.KindAdvertisement, Stable: true},
		RSSI:       -60,
		ObservedAt: clock.Now(),
	}
	ephemeral := model.Observation{
		Source: "a", Medium: model.MediumBLEAdvertisement,
		Identity:   model.Identity{Key: "fp:bb", Kind: model.KindAdvertisement, Stable: false},
		RSSI:       -70,
		ObservedAt: clock.Now(),
	}
	service.record(ephemeral)
	clock.Advance(time.Second)
	service.flush()
	if published[0].Arrivals != 0 {
		t.Fatalf("arrivals = %d, want 0 for a rotating identity", published[0].Arrivals)
	}
	service.record(stable)
	clock.Advance(time.Second)
	service.flush()
	if published[1].Arrivals != 1 {
		t.Fatalf("arrivals = %d, want 1 for a stable identity", published[1].Arrivals)
	}
}

func TestSnapshotIsBoundedAndReportsTruncation(t *testing.T) {
	clock := newFakeClock()
	var published []model.Snapshot
	service := NewService(Deps{
		Store:   &recordingStore{},
		Clock:   clock,
		Publish: func(s model.Snapshot) { published = append(published, s) },
	})
	for i := 0; i < model.MaxSnapshotIdentities+10; i++ {
		key := "addr:" + string(rune('a'+i%26)) + string(rune('a'+i/26)) + string(rune('a'+i%7))
		service.record(stableObservation("a", key, -50, clock.Now()))
	}
	clock.Advance(time.Second)
	service.flush()
	if len(published) != 1 {
		t.Fatalf("published %d snapshots", len(published))
	}
	if len(published[0].Identities) > model.MaxSnapshotIdentities {
		t.Fatalf("identities = %d, exceeds the cap", len(published[0].Identities))
	}
	if !published[0].Truncated {
		t.Fatal("truncation not reported")
	}
}

func TestStaleIdentitiesAreDropped(t *testing.T) {
	clock := newFakeClock()
	service := NewService(Deps{
		Store:      &recordingStore{},
		Clock:      clock,
		StaleAfter: time.Minute,
	})
	service.record(stableObservation("a", "addr:aa", -60, clock.Now()))
	clock.Advance(2 * time.Minute)
	service.flush()
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if len(service.aggregates) != 0 {
		t.Fatalf("aggregates = %d, want stale entries dropped", len(service.aggregates))
	}
}

func TestStartStopsOnContextCancel(t *testing.T) {
	clock := newFakeClock()
	ctx, cancel := context.WithCancel(context.Background())
	service := NewService(Deps{Store: &recordingStore{}, Clock: clock, PublishGap: time.Millisecond})
	service.Start(ctx)
	cancel()
	// A stopped worker must not keep consuming; the queue stays readable and bounded.
	if err := service.ObserveAdvertisements(context.Background(), []model.Observation{stableObservation("a", "addr:aa", -60, clock.Now())}); err != nil {
		t.Fatalf("observe after cancel: %v", err)
	}
}
