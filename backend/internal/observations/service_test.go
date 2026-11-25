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

func TestAggregatesAcrossSources(t *testing.T) {
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
	if len(identity.Sources) != 2 {
		t.Fatalf("sources = %d, want one reading per source", len(identity.Sources))
	}
	if len(published[0].Sources) != 2 {
		t.Fatalf("source descriptors = %d, want 2", len(published[0].Sources))
	}
}

// RSSI must follow the newest reading so a device that moves away appears further. A lifetime
// maximum would freeze its distance at the closest point it ever reached.
func TestRSSITracksNewestReading(t *testing.T) {
	clock := newFakeClock()
	var published []model.Snapshot
	service := NewService(Deps{
		Store:   &recordingStore{},
		Clock:   clock,
		Publish: func(s model.Snapshot) { published = append(published, s) },
	})
	service.RegisterSource("a", "Hall", []model.Medium{model.MediumBLEAdvertisement}, nil)

	service.record(stableObservation("a", "addr:aa", -40, clock.Now()))
	clock.Advance(time.Second)
	service.flush()
	service.record(stableObservation("a", "addr:aa", -90, clock.Now()))
	clock.Advance(time.Second)
	service.flush()

	last := published[len(published)-1].Identities[0]
	if last.RSSI == nil || *last.RSSI != -90 {
		t.Fatalf("rssi = %v, want the newest reading -90", last.RSSI)
	}
}

// A source that receives advertisements which decode to nothing changes no identity, but the
// counters must still reach the browser: that is the state that diagnoses a configuration
// fault. Without this the viewer would freeze at "0 advertisements".
func TestDiagnosticsAloneTriggerAPublish(t *testing.T) {
	clock := newFakeClock()
	var published []model.Snapshot
	service := NewService(Deps{
		Store:   &recordingStore{},
		Clock:   clock,
		Publish: func(s model.Snapshot) { published = append(published, s) },
	})
	service.RegisterSource("a", "Hall", []model.Medium{model.MediumBLEAdvertisement}, nil)
	clock.Advance(time.Second)
	service.flush()
	before := len(published)

	// Fifty advertisements arrived, none produced an identity.
	service.RecordSourceActivity("a", 50, 0, clock.Now())
	clock.Advance(time.Second)
	service.flush()

	if len(published) == before {
		t.Fatal("diagnostics changed but no snapshot published")
	}
	if got := published[len(published)-1].Sources[0].Diagnostics.Advertisements; got != 50 {
		t.Fatalf("advertisements = %d, want 50", got)
	}
}

// Continuous per-packet telemetry must not publish a full snapshot on every 250 ms tick.
// Structural changes still publish at the fast cadence. See spec 069.
func TestTelemetryChurnPublishesAtTelemetryCadence(t *testing.T) {
	clock := newFakeClock()
	var published []model.Snapshot
	service := NewService(Deps{
		Store:        &recordingStore{},
		Clock:        clock,
		Publish:      func(s model.Snapshot) { published = append(published, s) },
		TelemetryGap: 5 * time.Second,
	})
	service.RegisterSource("a", "Hall", []model.Medium{model.MediumBLEAdvertisement}, nil)
	service.record(stableObservation("a", "addr:aa", -60, clock.Now()))
	clock.Advance(time.Second)
	service.flush()
	if len(published) != 1 {
		t.Fatalf("published %d snapshots after the structural change, want 1", len(published))
	}

	// Eight telemetry ticks inside one telemetry gap: only fresh RSSI and counters change.
	for i := 0; i < 8; i++ {
		service.RecordSourceActivity("a", 1, 1, clock.Now())
		service.record(stableObservation("a", "addr:aa", -60-i, clock.Now()))
		clock.Advance(defaultPublishGap)
		service.flush()
	}
	if len(published) != 1 {
		t.Fatalf("telemetry churn published %d snapshots inside one telemetry gap, want 1", len(published))
	}

	// Once the gap elapses the newest exact values reach the viewer in one snapshot.
	clock.Advance(5 * time.Second)
	service.flush()
	if len(published) != 2 {
		t.Fatalf("published %d snapshots after the telemetry gap, want 2", len(published))
	}
	last := published[len(published)-1]
	if got := last.Sources[0].Diagnostics.Advertisements; got != 8 {
		t.Fatalf("advertisements = %d, want the exact count 8", got)
	}
	if last.Identities[0].RSSI == nil || *last.Identities[0].RSSI != -67 {
		t.Fatalf("rssi = %v, want the exact newest reading -67", last.Identities[0].RSSI)
	}
}

// A structural change must not wait for the slow telemetry cadence.
func TestStructuralChangePublishesAtPublishCadence(t *testing.T) {
	clock := newFakeClock()
	var published []model.Snapshot
	service := NewService(Deps{
		Store:        &recordingStore{},
		Clock:        clock,
		Publish:      func(s model.Snapshot) { published = append(published, s) },
		TelemetryGap: time.Hour,
	})
	service.RegisterSource("a", "Hall", nil, nil)
	clock.Advance(time.Second)
	service.flush()
	if len(published) != 1 {
		t.Fatalf("published %d snapshots after registering a source, want 1", len(published))
	}

	// A new identity is structural, so it publishes at publishGap despite the long
	// telemetry gap; telemetry-only churn on that same identity stays throttled.
	service.record(stableObservation("a", "addr:aa", -60, clock.Now()))
	clock.Advance(defaultPublishGap)
	service.flush()
	if len(published) != 2 {
		t.Fatalf("structural change published %d snapshots, want 2", len(published))
	}
	service.RecordSourceActivity("a", 1, 1, clock.Now())
	clock.Advance(defaultPublishGap)
	service.flush()
	if len(published) != 2 {
		t.Fatalf("telemetry after a structural change published %d snapshots, want 2", len(published))
	}
}

// A structural change that misses publishGap must not fall back to telemetryGap on the
// next tick. lastStructural is only updated on publish, so the retry stays structural.
func TestStructuralChangeRetriesAfterMissedPublishGap(t *testing.T) {
	clock := newFakeClock()
	var published []model.Snapshot
	service := NewService(Deps{
		Store:        &recordingStore{},
		Clock:        clock,
		Publish:      func(s model.Snapshot) { published = append(published, s) },
		PublishGap:   time.Second,
		TelemetryGap: time.Hour,
	})
	service.RegisterSource("a", "Hall", nil, nil)
	clock.Advance(time.Second)
	service.flush()
	if len(published) != 1 {
		t.Fatalf("published %d snapshots after registering a source, want 1", len(published))
	}

	// Same instant: changed but inside publishGap, so nothing is published.
	service.record(stableObservation("a", "addr:aa", -60, clock.Now()))
	service.flush()
	if len(published) != 1 {
		t.Fatalf("published %d snapshots inside publishGap, want 1", len(published))
	}

	// Telemetry churn in between must not turn the pending structural change into a
	// telemetry-only one. At publishGap it publishes, well before telemetryGap.
	service.RecordSourceActivity("a", 1, 1, clock.Now())
	clock.Advance(time.Second)
	service.flush()
	if len(published) != 2 {
		t.Fatalf("structural change published %d snapshots at publishGap, want 2", len(published))
	}
	if last := published[len(published)-1]; len(last.Identities) != 1 {
		t.Fatalf("published snapshot has %d identities, want the structural change", len(last.Identities))
	}
}

func TestTelemetryGapIsConfigurable(t *testing.T) {
	clock := newFakeClock()
	var published []model.Snapshot
	service := NewService(Deps{
		Store:        &recordingStore{},
		Clock:        clock,
		Publish:      func(s model.Snapshot) { published = append(published, s) },
		TelemetryGap: 500 * time.Millisecond,
	})
	service.RegisterSource("a", "Hall", nil, nil)
	clock.Advance(time.Second)
	service.flush()

	service.RecordSourceActivity("a", 1, 0, clock.Now())
	clock.Advance(500 * time.Millisecond)
	service.flush()
	if len(published) != 2 {
		t.Fatalf("published %d snapshots with a 500ms telemetry gap, want 2", len(published))
	}
}

// With telemetry counters removed, source IDs must not be able to alias identity keys.
// A NUL separator keeps the two sections distinct; this guards that invariance.
func TestStructuralSignatureSeparatesSourcesFromIdentities(t *testing.T) {
	twoSources := model.Snapshot{Sources: []model.SourceDescriptor{{ID: "a"}, {ID: "b"}}}
	oneSourceAndIdentity := model.Snapshot{
		Sources:    []model.SourceDescriptor{{ID: "a"}},
		Identities: []model.IdentityView{{Key: "b"}},
	}
	if structuralSignature(twoSources) == structuralSignature(oneSourceAndIdentity) {
		t.Fatal("source IDs and identity keys collided in the structural change key")
	}
}

func TestTelemetryGapNeverBelowPublishGap(t *testing.T) {
	service := NewService(Deps{
		Store:        &recordingStore{},
		Clock:        newFakeClock(),
		PublishGap:   2 * time.Second,
		TelemetryGap: time.Millisecond,
	})
	if service.telemetryGap != service.publishGap {
		t.Fatalf("telemetryGap = %s, want the clamped %s", service.telemetryGap, service.publishGap)
	}
}

// Repeated identical counters must not keep rebroadcasting; change detection still applies.
func TestUnchangedDiagnosticsDoNotRepublish(t *testing.T) {
	clock := newFakeClock()
	var published []model.Snapshot
	service := NewService(Deps{
		Store:   &recordingStore{},
		Clock:   clock,
		Publish: func(s model.Snapshot) { published = append(published, s) },
	})
	service.RegisterSource("a", "Hall", []model.Medium{model.MediumBLEAdvertisement}, nil)
	service.RecordSourceActivity("a", 5, 5, clock.Now())
	clock.Advance(time.Second)
	service.flush()
	after := len(published)

	clock.Advance(time.Second)
	service.flush()
	if len(published) != after {
		t.Fatalf("published %d snapshots with unchanged counters, want %d", len(published), after)
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

func TestSourceDiagnosticsReachTheSnapshot(t *testing.T) {
	clock := newFakeClock()
	var published []model.Snapshot
	service := NewService(Deps{
		Store:   &recordingStore{},
		Clock:   clock,
		Publish: func(s model.Snapshot) { published = append(published, s) },
	})
	service.RegisterSource("a", "Hall proxy", []model.Medium{model.MediumBLEAdvertisement}, []string{"kind"})

	// A source that connects but delivers nothing must still be visible with zero counters,
	// which is what distinguishes it from a source that was never enrolled.
	clock.Advance(time.Second)
	service.flush()
	if len(published) != 1 || len(published[0].Sources) != 1 {
		t.Fatalf("published = %+v", published)
	}
	if got := published[0].Sources[0].Diagnostics; got.Advertisements != 0 || got.LastAdvertisement != nil {
		t.Fatalf("diagnostics = %+v, want zeros and no timestamp", got)
	}

	service.RecordSourceActivity("a", 5, 3, clock.Now())
	service.record(stableObservation("a", "addr:aa", -60, clock.Now()))
	clock.Advance(time.Second)
	service.flush()

	last := published[len(published)-1]
	got := last.Sources[0].Diagnostics
	if got.Advertisements != 5 || got.Decoded != 3 {
		t.Fatalf("diagnostics = %+v, want 5 advertisements and 3 decoded", got)
	}
	if got.LastAdvertisement == nil {
		t.Fatal("lastAdvertisement not recorded")
	}
}

func TestRecordSourceActivityIgnoresUnknownSources(t *testing.T) {
	clock := newFakeClock()
	service := NewService(Deps{Store: &recordingStore{}, Clock: clock})
	// Reporting before registration must not create a phantom descriptor.
	service.RecordSourceActivity("ghost", 10, 10, clock.Now())
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if len(service.sources) != 0 {
		t.Fatalf("sources = %d, want none", len(service.sources))
	}
}

func TestDiagnosticsSurviveReRegister(t *testing.T) {
	clock := newFakeClock()
	service := NewService(Deps{Store: &recordingStore{}, Clock: clock})
	service.RegisterSource("a", "Hall", []model.Medium{model.MediumBLEAdvertisement}, nil)
	service.RecordSourceActivity("a", 4, 4, clock.Now())
	// A reconnect re-registers the same id; counters must not reset, or the viewer looks
	// like it has never received anything after every session restart.
	service.RegisterSource("a", "Hall renamed", []model.Medium{model.MediumBLEAdvertisement}, []string{"kind"})
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if service.sources["a"].diagnostics.Advertisements != 4 {
		t.Fatalf("diagnostics reset on re-register: %+v", service.sources["a"].diagnostics)
	}
	if service.sources["a"].name != "Hall renamed" {
		t.Fatal("descriptor not updated on re-register")
	}
}

func TestDropsAreAttributedToTheirSource(t *testing.T) {
	clock := newFakeClock()
	service := NewService(Deps{
		Store: &recordingStore{},
		Clock: clock,
		Queue: 1,
	})
	service.RegisterSource("a", "Hall", []model.Medium{model.MediumBLEAdvertisement}, nil)
	// Fill the queue without a worker, then overflow it.
	_ = service.ObserveAdvertisements(context.Background(), []model.Observation{stableObservation("a", "addr:01", -60, clock.Now())})
	_ = service.ObserveAdvertisements(context.Background(), []model.Observation{stableObservation("a", "addr:02", -61, clock.Now())})

	service.mutex.Lock()
	defer service.mutex.Unlock()
	if service.sources["a"].diagnostics.Dropped != 1 {
		t.Fatalf("dropped = %d, want 1", service.sources["a"].diagnostics.Dropped)
	}
}

func TestUnregisterSourceRemovesDescriptor(t *testing.T) {
	clock := newFakeClock()
	var published []model.Snapshot
	service := NewService(Deps{
		Store:   &recordingStore{},
		Clock:   clock,
		Publish: func(s model.Snapshot) { published = append(published, s) },
	})
	service.RegisterSource("a", "Hall", []model.Medium{model.MediumBLEAdvertisement}, nil)
	service.UnregisterSource("a")
	clock.Advance(time.Second)
	service.flush()
	for _, snapshot := range published {
		if len(snapshot.Sources) != 0 {
			t.Fatalf("withdrawn source still present: %+v", snapshot.Sources)
		}
	}
}

// firstSeen must be the first observation of an identity, not its latest: the viewer needs it
// to distinguish a device that just appeared from one that has been present for hours.
func TestFirstSeenIsStableAcrossObservations(t *testing.T) {
	clock := newFakeClock()
	var published []model.Snapshot
	service := NewService(Deps{
		Store:   &recordingStore{},
		Clock:   clock,
		Publish: func(s model.Snapshot) { published = append(published, s) },
	})
	service.RegisterSource("a", "Hall", []model.Medium{model.MediumBLEAdvertisement}, nil)

	first := clock.Now()
	service.record(stableObservation("a", "addr:aa", -60, first))
	clock.Advance(time.Minute)
	service.record(stableObservation("a", "addr:aa", -61, clock.Now()))
	clock.Advance(time.Second)
	service.flush()

	view := published[len(published)-1].Identities[0]
	if !view.FirstSeen.Equal(first) {
		t.Fatalf("firstSeen = %v, want the first observation %v", view.FirstSeen, first)
	}
	if !view.LastSeen.After(view.FirstSeen) {
		t.Fatalf("lastSeen %v should follow firstSeen %v", view.LastSeen, view.FirstSeen)
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

// TestDisabledGateDiscardsAdvertisements checks that disabling observation actually stops it.
//
// The failure this guards against is a toggle that only hides the view: the queue would fill,
// the recorder would keep accumulating evidence and the device would keep spending radio time on
// data nobody wants. Asserting on the queue proves the work stopped rather than being filtered later.
func TestDisabledGateDiscardsAdvertisements(t *testing.T) {
	clock := newFakeClock()
	enabled := false
	service := NewService(Deps{
		Store:  &recordingStore{},
		Clock:  clock,
		Queue:  4,
		Active: func() bool { return enabled },
	})
	observation := []model.Observation{stableObservation("a", "addr:01", -60, clock.Now())}

	if err := service.ObserveAdvertisements(context.Background(), observation); err != nil {
		t.Fatalf("observe while disabled: %v", err)
	}
	if len(service.queue) != 0 {
		t.Fatalf("queue length = %d, want 0 while observation is disabled", len(service.queue))
	}

	enabled = true
	if err := service.ObserveAdvertisements(context.Background(), observation); err != nil {
		t.Fatalf("observe while enabled: %v", err)
	}
	if len(service.queue) != 1 {
		t.Fatalf("queue length = %d, want 1 once observation is enabled", len(service.queue))
	}
}

// TestNilGateObservesAlways pins the default: an installation with no observation integration
// behaves exactly as it did before the gate existed, so upgrading cannot silently stop observation.
func TestNilGateObservesAlways(t *testing.T) {
	clock := newFakeClock()
	service := NewService(Deps{Store: &recordingStore{}, Clock: clock, Queue: 4})
	if err := service.ObserveAdvertisements(context.Background(), []model.Observation{stableObservation("a", "addr:01", -60, clock.Now())}); err != nil {
		t.Fatalf("observe: %v", err)
	}
	if len(service.queue) != 1 {
		t.Fatalf("queue length = %d, want 1 with no gate", len(service.queue))
	}
	if !service.Snapshot().Observing {
		t.Fatal("a nil gate must report observing")
	}
}

// TestDisabledGateStopsCounters is the regression for the page that read "0 devices heard" beside
// "571 heard · 571 decoded".
//
// The two numbers came from different paths, and only one honoured the gate. A sensor streaming
// before the integration was added advanced the counters while every identity was discarded, so the
// page argued with itself and neither number could be acted on. A counter that measures what the
// radio delivered but nobody kept is not a diagnostic — it is noise with a timestamp.
func TestDisabledGateStopsCounters(t *testing.T) {
	clock := newFakeClock()
	enabled := false
	service := NewService(Deps{
		Store:  &recordingStore{},
		Clock:  clock,
		Queue:  4,
		Active: func() bool { return enabled },
	})
	service.RegisterSource("ble-1", "Hallway", nil, nil)

	service.RecordSourceActivity("ble-1", 571, 571, clock.Now())
	if got := service.Snapshot().Sources[0].Diagnostics.Advertisements; got != 0 {
		t.Fatalf("advertisements = %d while disabled, want 0", got)
	}

	enabled = true
	service.RecordSourceActivity("ble-1", 2, 2, clock.Now())
	if got := service.Snapshot().Sources[0].Diagnostics.Advertisements; got != 2 {
		t.Fatalf("advertisements = %d after enabling, want 2", got)
	}
}

// TestClosingTheGateClearsTheCensus is why the Home tile can stop showing a census.
//
// Removing or disabling the only integration used to stop the gate and nothing else: the browser
// kept its last snapshot and rendered a room that was no longer being listened to. The service now
// publishes one clearing snapshot, and the gate flag is in the signature so that publish is not
// suppressed as an unchanged duplicate.
func TestClosingTheGateClearsTheCensus(t *testing.T) {
	clock := newFakeClock()
	enabled := true
	published := []model.Snapshot{}
	service := NewService(Deps{
		Store:   &recordingStore{},
		Clock:   clock,
		Queue:   8,
		Publish: func(snapshot model.Snapshot) { published = append(published, snapshot) },
		Active:  func() bool { return enabled },
	})

	service.RegisterSource("ble-1", "Hallway", nil, nil)
	service.record(stableObservation("ble-1", "addr:01", -60, clock.Now()))
	clock.Advance(300 * time.Millisecond)
	service.flush()
	if len(service.Snapshot().Identities) != 1 {
		t.Fatal("the fixture did not observe anything")
	}

	enabled = false
	clock.Advance(300 * time.Millisecond)
	service.flush()

	empty := service.Snapshot()
	if empty.Observing {
		t.Fatal("the snapshot still reports observing after the gate closed")
	}
	if len(empty.Identities) != 0 {
		t.Fatalf("the observed identities survived the gate closing: %d", len(empty.Identities))
	}
	// Sources survive on purpose: a descriptor names a device and the facets it can populate, which
	// is a fact about the device rather than about what was observed, and only the live session can
	// announce it — once. Clearing it here left a disabled-then-re-enabled integration with no
	// sources until the ESPHome session reconnected, so the filters vanished and a silent source
	// became indistinguishable from an absent one. The counters do not lie, because
	// RecordSourceActivity honours the same gate.
	if len(empty.Sources) != 1 {
		t.Fatalf("sources = %d after the gate closed, want the 1 descriptor kept", len(empty.Sources))
	}
	if got := empty.Sources[0].Diagnostics.Advertisements; got != 0 {
		t.Fatalf("advertisements = %d, want 0: counters must not advance while not observing", got)
	}
	// The clearing snapshot must have been broadcast, not folded away as a duplicate.
	if len(published) == 0 || published[len(published)-1].Observing {
		t.Fatalf("published snapshots = %d, last observing = %v", len(published), published[len(published)-1].Observing)
	}
}

// TestReEnablingTheGateRestoresObservation is the regression for a disable/enable cycle stranding
// the viewer with no sources.
//
// The session announces its descriptor once per connection, guarded by its own flag, and disabling
// the integration deliberately does not close the session. So anything the gate's close discards
// cannot come back until the device reconnects — which may be hours.
func TestReEnablingTheGateRestoresObservation(t *testing.T) {
	clock := newFakeClock()
	enabled := true
	service := NewService(Deps{
		Store:  &recordingStore{},
		Clock:  clock,
		Queue:  8,
		Active: func() bool { return enabled },
	})
	service.RegisterSource("ble-1", "Hallway", nil, nil)

	enabled = false
	clock.Advance(300 * time.Millisecond)
	service.flush()

	enabled = true
	clock.Advance(300 * time.Millisecond)
	service.flush()

	snapshot := service.Snapshot()
	if !snapshot.Observing {
		t.Fatal("the gate did not reopen")
	}
	if len(snapshot.Sources) != 1 {
		t.Fatalf("sources = %d after re-enabling, want 1: the descriptor was lost and cannot be re-announced", len(snapshot.Sources))
	}
	// Recording resumes against the surviving descriptor.
	service.RecordSourceActivity("ble-1", 5, 5, clock.Now())
	if got := service.Snapshot().Sources[0].Diagnostics.Advertisements; got != 5 {
		t.Fatalf("advertisements = %d after re-enabling, want 5", got)
	}
}

// TestZeroRSSIIsNotAZeroReading pins that an unmeasurable signal is not plotted as zero dBm.
//
// ESPHome reports RSSI 0 when it could not measure. Every other consumer of this value already
// rejects it; accepting it here made the median 0 and drew the device on top of the radar.
func TestZeroRSSIIsNotAZeroReading(t *testing.T) {
	clock := newFakeClock()
	service := NewService(Deps{Store: &recordingStore{}, Clock: clock, Queue: 8})
	service.RegisterSource("ble-1", "Hallway", nil, nil)

	service.record(stableObservation("ble-1", "addr:01", -60, clock.Now()))
	service.record(stableObservation("ble-1", "addr:01", 0, clock.Now()))

	snapshot := service.Snapshot()
	if len(snapshot.Identities) != 1 {
		t.Fatal("the fixture did not observe anything")
	}
	if rssi := snapshot.Identities[0].RSSI; rssi != nil && *rssi == 0 {
		t.Fatal("a zero RSSI was stored as a real reading")
	}
	if rssi := snapshot.Identities[0].RSSI; rssi == nil || *rssi != -60 {
		t.Fatalf("RSSI = %v, want the last measurable reading -60", rssi)
	}
	// The packet still counts: the device was heard, the signal just could not be measured.
	if snapshot.Identities[0].Packets != 2 {
		t.Fatalf("packets = %d, want 2", snapshot.Identities[0].Packets)
	}
}

// TestSilentRoomIsNotAnAbsentObserver keeps the two empty states distinguishable.
//
// An empty identity list is the normal state of a working installation. Without the gate flag, a
// consumer would have to guess why the list is empty, and the two answers have opposite fixes.
func TestSilentRoomIsNotAnAbsentObserver(t *testing.T) {
	clock := newFakeClock()
	service := NewService(Deps{Store: &recordingStore{}, Clock: clock, Queue: 4})
	service.RegisterSource("ble-1", "Hallway", nil, nil)
	clock.Advance(300 * time.Millisecond)
	service.flush()

	snapshot := service.Snapshot()
	if len(snapshot.Identities) != 0 {
		t.Fatal("the fixture heard a device")
	}
	if !snapshot.Observing {
		t.Fatal("a silent room was reported as not being observed")
	}
}
