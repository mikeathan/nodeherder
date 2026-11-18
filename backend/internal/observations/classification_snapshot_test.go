package observations

// Tests for how classification reaches the snapshot. The distinction that matters here is between
// "no evidence yet" and "classified as learning": the first is the normal first-day state and must
// still produce a visible device, while the second is a real result. Conflating them would either
// hide everything on the first day or claim a class that does not exist.

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"node-herder/mocks"
	model "node-herder/models/observations"
	"node-herder/models/settings"
	"node-herder/utils"
)

// historyStub returns canned summaries per identity.
type historyStub struct {
	mutex sync.Mutex
	byKey map[string][]model.DailySummary
	err   error
	reads int
}

func (h *historyStub) Summaries(key string, _ int) ([]model.DailySummary, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.reads++
	if h.err != nil {
		return nil, h.err
	}
	return h.byKey[key], nil
}

func (h *historyStub) readCount() int {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	return h.reads
}

// decisionsStub satisfies the service's identity store.
type decisionsStub struct {
	decisions map[string]settings.ObservationIdentity
}

func (d decisionsStub) ObservationIdentities() map[string]settings.ObservationIdentity {
	return d.decisions
}

// recorderStub captures observations handed to the evidence recorder.
type recorderStub struct {
	mutex        sync.Mutex
	observations []model.Observation
}

func (r *recorderStub) Observe(observation model.Observation) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.observations = append(r.observations, observation)
}

func (r *recorderStub) count() int {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return len(r.observations)
}

// publishedCapture records what a service broadcast, so a test can inspect the result.
type publishedCapture struct {
	mutex     sync.Mutex
	snapshots []model.Snapshot
}

func (p *publishedCapture) publish(snapshot model.Snapshot) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.snapshots = append(p.snapshots, snapshot)
}

// last returns the most recent snapshot, or nil when nothing was published.
func (p *publishedCapture) last() *model.Snapshot {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if len(p.snapshots) == 0 {
		return nil
	}
	last := p.snapshots[len(p.snapshots)-1]
	return &last
}

// classifyService builds a service wired for classification tests.
func classifyService(t *testing.T, clock utils.Clock, history *historyStub, recorder *recorderStub) (*Service, *publishedCapture) {
	t.Helper()
	capture := &publishedCapture{}
	service := NewService(Deps{
		Store:            decisionsStub{},
		Clock:            clock,
		History:          history,
		Recorder:         recorder,
		Publish:          capture.publish,
		ClassifyInterval: time.Nanosecond,
	})
	return service, capture
}

func TestSnapshotAttachesACachedClass(t *testing.T) {
	clock := mocks.NewMockClock(func() time.Time { return at(9, 0) })
	history := &historyStub{byKey: map[string][]model.DailySummary{
		"addr:aa:bb:cc:dd:ee:01": {
			summaryOn(0, func(s *model.DailySummary) { s.ActiveHours = 24; s.HourMask = allDay() }),
		},
	}}
	service, capture := classifyService(t, clock, history, &recorderStub{})

	service.record(model.Observation{
		Source: "esp-1", Identity: model.Identity{Key: "addr:aa:bb:cc:dd:ee:01", Stable: true},
		RSSI: -55, ObservedAt: clock.Now(),
	})
	service.flush()

	snapshot := capture.last()
	if snapshot == nil {
		t.Fatal("no snapshot was published")
	}
	identity := snapshot.Identities[0]
	if identity.Classification == nil {
		t.Fatal("Classification = nil, want a class")
	}
	if identity.Classification.Class != ClassInfrastructure {
		t.Errorf("class = %q, want infrastructure", identity.Classification.Class)
	}
	if identity.Classification.Reason == "" {
		t.Errorf("Reason is empty, want the evidence behind the class")
	}
}

func TestSnapshotFallsBackToThePreFilterWithoutEvidence(t *testing.T) {
	clock := mocks.NewMockClock(func() time.Time { return at(9, 0) })
	history := &historyStub{byKey: map[string][]model.DailySummary{}}
	service, capture := classifyService(t, clock, history, &recorderStub{})

	// A rotating address: the pre-filter must answer immediately, because there is no history yet.
	service.record(model.Observation{
		Source: "esp-1", Identity: model.Identity{Key: "fp:9a3c1f", Stable: false},
		RSSI: -70, ObservedAt: clock.Now(),
	})
	service.flush()

	snapshot := capture.last()
	if snapshot == nil {
		t.Fatal("no snapshot was published; a first-day device must not be hidden")
	}
	classification := snapshot.Identities[0].Classification
	if classification == nil {
		t.Fatal("Classification = nil, want a pre-filter result")
	}
	if classification.Class != ProvisionalMobile {
		t.Errorf("class = %q, want %q", classification.Class, ProvisionalMobile)
	}
	if !classification.Provisional {
		t.Errorf("Provisional = false, want true so the UI cannot present it as settled")
	}
}

func TestSnapshotPreFilterKeepsAWeakStableDeviceVisible(t *testing.T) {
	clock := mocks.NewMockClock(func() time.Time { return at(9, 0) })
	history := &historyStub{byKey: map[string][]model.DailySummary{}}
	service, capture := classifyService(t, clock, history, &recorderStub{})

	// One strong device establishes the reference, then a weak one is judged against it.
	service.record(model.Observation{
		Source: "esp-1", Identity: model.Identity{Key: "addr:aa:bb:cc:dd:ee:01", Stable: true},
		RSSI: -50, ObservedAt: clock.Now(),
	})
	service.record(model.Observation{
		Source: "esp-1", Identity: model.Identity{Key: "addr:aa:bb:cc:dd:ee:02", Stable: true},
		RSSI: -92, ObservedAt: clock.Now(),
	})
	service.flush()

	snapshot := capture.last()
	if snapshot == nil {
		t.Fatal("no snapshot was published")
	}
	for _, identity := range snapshot.Identities {
		if identity.Key != "addr:aa:bb:cc:dd:ee:02" {
			continue
		}
		if !identity.Classification.ShowByDefault {
			t.Errorf("ShowByDefault = false for a weak stable device, want true")
		}
		if identity.Classification.Class != ProvisionalNeighbour {
			t.Errorf("class = %q, want %q", identity.Classification.Class, ProvisionalNeighbour)
		}
		return
	}
	t.Fatal("the weak identity was not in the snapshot")
}

func TestSnapshotKeepsItsClassWhenTheStoreFails(t *testing.T) {
	// A read failure must not clear a class that was previously known: showing nothing is a worse
	// answer than showing the last good evidence.
	clock := mocks.NewMockClock(func() time.Time { return at(9, 0) })
	history := &historyStub{byKey: map[string][]model.DailySummary{
		"addr:aa:bb:cc:dd:ee:01": {
			summaryOn(0, func(s *model.DailySummary) { s.ActiveHours = 24; s.HourMask = allDay() }),
		},
	}}
	service, capture := classifyService(t, clock, history, &recorderStub{})

	service.record(model.Observation{
		Source: "esp-1", Identity: model.Identity{Key: "addr:aa:bb:cc:dd:ee:01", Stable: true},
		RSSI: -55, ObservedAt: clock.Now(),
	})
	service.flush()

	history.mutex.Lock()
	history.err = errors.New("disk on fire")
	history.mutex.Unlock()

	service.flush()

	snapshot := capture.last()
	if snapshot == nil {
		t.Fatal("no snapshot was published")
	}
	if got := snapshot.Identities[0].Classification.Class; got != ClassInfrastructure {
		t.Errorf("class = %q after a read failure, want the last good class", got)
	}
}

func TestServiceFeedsTheRecorder(t *testing.T) {
	// Evidence must accumulate from live traffic, or classification would never have anything to
	// read. The recorder is fed from the aggregation worker, not from the transport, so a source
	// cannot bypass accumulation.
	clock := mocks.NewMockClock(func() time.Time { return at(9, 0) })
	recorder := &recorderStub{}
	service, _ := classifyService(t, clock, &historyStub{}, recorder)

	for i := 0; i < 3; i++ {
		service.record(model.Observation{
			Source: "esp-1", Identity: model.Identity{Key: "addr:aa:bb:cc:dd:ee:01", Stable: true},
			RSSI: -60, ObservedAt: clock.Now(),
		})
	}

	if got := recorder.count(); got != 3 {
		t.Errorf("recorder received %d observations, want 3", got)
	}
}

func TestServiceIgnoresAnIdentityWithoutAKey(t *testing.T) {
	clock := mocks.NewMockClock(func() time.Time { return at(9, 0) })
	recorder := &recorderStub{}
	service, _ := classifyService(t, clock, &historyStub{}, recorder)

	service.record(model.Observation{Source: "esp-1", RSSI: -60, ObservedAt: clock.Now()})

	if got := recorder.count(); got != 0 {
		t.Errorf("recorder received %d unattributable observations, want 0", got)
	}
}

func TestClassificationIsAlwaysAdvisory(t *testing.T) {
	// The precedence rule travels in the payload, not only in the UI: an operator's triage decision
	// outranks the class, so a consumer that presents the class as a fact is misusing it.
	clock := mocks.NewMockClock(func() time.Time { return at(9, 0) })
	history := &historyStub{byKey: map[string][]model.DailySummary{
		"addr:aa:bb:cc:dd:ee:01": {
			summaryOn(0, func(s *model.DailySummary) { s.ActiveHours = 24; s.HourMask = allDay() }),
		},
	}}
	service, capture := classifyService(t, clock, history, &recorderStub{})

	service.record(model.Observation{
		Source: "esp-1", Identity: model.Identity{Key: "addr:aa:bb:cc:dd:ee:01", Stable: true},
		RSSI: -55, ObservedAt: clock.Now(),
	})
	service.flush()

	snapshot := capture.last()
	if snapshot == nil {
		t.Fatal("no snapshot was published")
	}
	classification := snapshot.Identities[0].Classification
	if classification == nil {
		t.Fatal("Classification = nil")
	}
	if !classification.Advisory {
		t.Errorf("Advisory = false, want true for a settled class")
	}
}

func TestTriageDecisionOutranksTheClass(t *testing.T) {
	// An ignored identity is absent from the snapshot whatever its class says. The class is
	// evidence about behaviour; the decision is intent, and intent wins.
	clock := mocks.NewMockClock(func() time.Time { return at(9, 0) })
	history := &historyStub{byKey: map[string][]model.DailySummary{
		"addr:aa:bb:cc:dd:ee:01": {
			summaryOn(0, func(s *model.DailySummary) { s.ActiveHours = 24; s.HourMask = allDay() }),
		},
	}}
	service := NewService(Deps{
		Store: decisionsStub{decisions: map[string]settings.ObservationIdentity{
			"addr:aa:bb:cc:dd:ee:01": {Decision: settings.DecisionIgnored},
		}},
		Clock:            clock,
		History:          history,
		ClassifyInterval: time.Nanosecond,
	})

	service.record(model.Observation{
		Source: "esp-1", Identity: model.Identity{Key: "addr:aa:bb:cc:dd:ee:01", Stable: true},
		RSSI: -55, ObservedAt: clock.Now(),
	})

	// The snapshot is inspected directly rather than through a publish: an ignored identity leaves
	// the snapshot empty, and an empty snapshot is deliberately never broadcast, so waiting for a
	// publish would test the idle-hub rule instead of the precedence rule.
	service.mutex.Lock()
	snapshot := service.snapshotLocked(clock.Now())
	service.mutex.Unlock()

	if len(snapshot.Identities) != 0 {
		t.Errorf("got %d identities, want none: an ignored identity must not appear",
			len(snapshot.Identities))
	}
}

func TestShowByDefaultForSettledClasses(t *testing.T) {
	tests := []struct {
		name   string
		class  string
		stable bool
		want   bool
	}{
		{name: "infrastructure is hidden", class: ClassInfrastructure, stable: true, want: false},
		// A neighbour used to be hidden. That removed exactly the identifiable devices from the
		// radar while keeping the anonymous ones, which is the bug behind "I only get one dot".
		{name: "neighbour is shown", class: ClassNeighbour, stable: true, want: true},
		{name: "personal is shown", class: ClassPersonal, stable: false, want: true},
		{name: "transient is shown", class: ClassTransient, stable: false, want: true},
		{name: "learning is shown", class: ClassLearning, stable: true, want: true},
		{
			// A rotating address cannot be installed equipment even if the evidence looks like it.
			name: "rotating address is never hidden as infrastructure", class: ClassInfrastructure,
			stable: false, want: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := showByDefaultFor(test.class, test.stable); got != test.want {
				t.Errorf("showByDefaultFor(%q, %v) = %v, want %v", test.class, test.stable, got, test.want)
			}
		})
	}
}

/*
 * Regression guards for the review findings. Each of these was a real defect that no existing test
 * could see, which is the reason they are written at all: a passing suite that cannot observe a
 * class of bug gives false confidence.
 */

func TestAClassificationChangeTriggersAPublish(t *testing.T) {
	// The render signature omitted classification, so a class that changed without any other field
	// changing was never broadcast — the viewer would show the old class forever. This is the same
	// omission as D-4 (diagnostics) and was found the same way: by asking what can change while the
	// signature stays equal.
	base := model.ClassificationView{
		Class: "personal", ShowByDefault: true, Confidence: 0.5, HoursObserved: 30, Reason: "a",
	}
	changed := base
	changed.Class = "infrastructure"

	before := renderSignature(model.Snapshot{Identities: []model.IdentityView{
		{Key: "k", Classification: &base},
	}})
	after := renderSignature(model.Snapshot{Identities: []model.IdentityView{
		{Key: "k", Classification: &changed},
	}})

	if before == after {
		t.Error("renderSignature did not change when the class changed, so the update would not broadcast")
	}
}

func TestARefinedReasonTriggersAPublish(t *testing.T) {
	// The class can stay the same while its justification and evidence change — that is exactly the
	// moment a new day of evidence lands, and the one the operator is watching for.
	base := model.ClassificationView{Class: "personal", ShowByDefault: true, HoursObserved: 24, Reason: "a"}
	refined := base
	refined.HoursObserved = 48
	refined.Reason = "b"

	before := renderSignature(model.Snapshot{Identities: []model.IdentityView{{Key: "k", Classification: &base}}})
	after := renderSignature(model.Snapshot{Identities: []model.IdentityView{{Key: "k", Classification: &refined}}})

	if before == after {
		t.Error("renderSignature ignored a change in hours or reason")
	}
}

func TestAnAbsentClassificationDiffersFromAPresentOne(t *testing.T) {
	// "not classified" and "classified as learning" are different states; a signature that treated
	// them alike would suppress the transition between them.
	present := model.ClassificationView{Class: "learning", ShowByDefault: true}
	withClass := renderSignature(model.Snapshot{Identities: []model.IdentityView{
		{Key: "k", Classification: &present},
	}})
	withoutClass := renderSignature(model.Snapshot{Identities: []model.IdentityView{{Key: "k"}}})

	if withClass == withoutClass {
		t.Error("renderSignature treated an absent classification as identical to a present one")
	}
}

func TestTheClassificationCacheDoesNotGrowWithoutBound(t *testing.T) {
	// The cache was merged forward and never pruned, so it retained an entry for every identity ever
	// seen. Over a day of rotating addresses that is unbounded growth in a long-lived process.
	clock := mocks.NewMockClock(func() time.Time { return at(9, 0) })
	history := &historyStub{byKey: map[string][]model.DailySummary{}}
	service, _ := classifyService(t, clock, history, &recorderStub{})

	// Populate the cache directly: the store returns nothing here, and the growth rule is what is
	// under test rather than the classification itself.
	service.mutex.Lock()
	for index := 0; index < 100; index++ {
		service.classified[fmt.Sprintf("addr:gone%03d", index)] = model.ClassificationView{Class: "personal"}
	}
	service.aggregates["addr:live"] = &aggregate{
		identity: model.Identity{Key: "addr:live", Stable: true},
		lastSeen: clock.Now(),
		sources:  map[string]*int{},
	}
	service.mutex.Unlock()

	history.byKey["addr:live"] = []model.DailySummary{
		summaryOn(0, func(s *model.DailySummary) { s.ActiveHours = 24; s.HourMask = allDay() }),
	}
	service.flush()

	service.mutex.Lock()
	cached := len(service.classified)
	service.mutex.Unlock()

	if cached > 1 {
		t.Errorf("cache holds %d entries, want only identities still present", cached)
	}
}

func TestThePreFilterStrongestReadingIsComputedOnce(t *testing.T) {
	// The reference reading used to be derived by scanning every aggregate inside the per-identity
	// render path, which made one snapshot O(n^2). The behaviour must be identical; only the cost
	// changes, so this pins the value.
	clock := mocks.NewMockClock(func() time.Time { return at(9, 0) })
	service, capture := classifyService(t, clock, &historyStub{}, &recorderStub{})

	service.record(model.Observation{
		Source: "esp-1", Identity: model.Identity{Key: "addr:strong", Stable: true},
		RSSI: -50, ObservedAt: clock.Now(),
	})
	service.record(model.Observation{
		Source: "esp-1", Identity: model.Identity{Key: "addr:weak", Stable: true},
		RSSI: -95, ObservedAt: clock.Now(),
	})
	service.flush()

	snapshot := capture.last()
	if snapshot == nil {
		t.Fatal("no snapshot was published")
	}
	var found bool
	for _, identity := range snapshot.Identities {
		if identity.Key != "addr:weak" {
			continue
		}
		found = true
		if identity.Classification.Class != ProvisionalNeighbour {
			t.Errorf("class = %q, want %q: the reference reading should be -50",
				identity.Classification.Class, ProvisionalNeighbour)
		}
	}
	if !found {
		t.Fatal("the weak identity was not in the snapshot")
	}
}

func TestAReadFailurePreservesTheCachedClass(t *testing.T) {
	// The comment promised that a read failure would keep the last good class, but the cache was
	// replaced wholesale and the failure path only `continue`d — so the class was silently dropped
	// and the row fell back to the pre-filter. A comment that describes behaviour the code does not
	// have is worse than no comment: it stops the next reader from checking.
	clock := mocks.NewMockClock(func() time.Time { return at(9, 0) })
	history := &historyStub{byKey: map[string][]model.DailySummary{
		"addr:aa:bb:cc:dd:ee:01": {
			summaryOn(0, func(s *model.DailySummary) { s.ActiveHours = 24; s.HourMask = allDay() }),
		},
	}}
	service, capture := classifyService(t, clock, history, &recorderStub{})

	service.record(model.Observation{
		Source: "esp-1", Identity: model.Identity{Key: "addr:aa:bb:cc:dd:ee:01", Stable: true},
		RSSI: -55, ObservedAt: clock.Now(),
	})
	service.flush()

	if got := capture.last().Identities[0].Classification.Class; got != ClassInfrastructure {
		t.Fatalf("setup failed: class = %q, want infrastructure", got)
	}

	// Break the store, then classify again. The good class must survive.
	history.mutex.Lock()
	history.err = errors.New("disk on fire")
	history.mutex.Unlock()

	// Advance past the publish gap, or the second flush is suppressed as "nothing new to send" and
	// the snapshot inspected below would still be the first one — a test that passes without ever
	// reaching the code it claims to cover.
	clock.Advance(2 * time.Second)
	service.flush()

	service.mutex.Lock()
	cached, ok := service.classified["addr:aa:bb:cc:dd:ee:01"]
	service.mutex.Unlock()

	if !ok {
		t.Fatal("the cached class was dropped entirely after a read failure")
	}
	if cached.Class != ClassInfrastructure {
		t.Errorf("cached class = %q after a read failure, want the previous infrastructure", cached.Class)
	}
}

func TestWaitContextReturnsWhenTheWorkerStops(t *testing.T) {
	// Shutdown ordering depends on being able to observe the worker stopping, and on not hanging
	// when it does not. A failed Listen never cancels the service's context, so an unbounded wait
	// would hang the process instead of shutting it down.
	clock := mocks.NewMockClock(func() time.Time { return at(9, 0) })
	service, _ := classifyService(t, clock, &historyStub{}, &recorderStub{})

	ctx, cancel := context.WithCancel(context.Background())
	service.Start(ctx)
	cancel()

	if !service.WaitContext(context.Background()) {
		t.Error("WaitContext reported a timeout after the worker had stopped")
	}
}

func TestWaitContextGivesUpRatherThanHanging(t *testing.T) {
	clock := mocks.NewMockClock(func() time.Time { return at(9, 0) })
	service, _ := classifyService(t, clock, &historyStub{}, &recorderStub{})
	// Deliberately never cancelled, so the worker keeps running.
	service.Start(context.Background())

	expired, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if service.WaitContext(expired) {
		t.Error("WaitContext reported the worker stopped, but it is still running")
	}
}

func TestTwoSourcesUseTheStrongerReading(t *testing.T) {
	// With two proxies the nearer one hears a device better, so its reading describes the distance.
	// Averaging would give a number true of neither antenna, and a device moving towards one proxy
	// would appear to move by half as much.
	clock := mocks.NewMockClock(func() time.Time { return at(9, 0) })
	service, capture := classifyService(t, clock, &historyStub{}, &recorderStub{})

	key := "addr:aa:bb:cc:dd:ee:01"
	service.record(model.Observation{
		Source: "near", Identity: model.Identity{Key: key, Stable: true}, RSSI: -50, ObservedAt: clock.Now(),
	})
	service.record(model.Observation{
		Source: "far", Identity: model.Identity{Key: key, Stable: true}, RSSI: -80, ObservedAt: clock.Now(),
	})
	service.flush()

	identity := capture.last().Identities[0]
	if identity.RSSI == nil || *identity.RSSI != -50 {
		t.Errorf("RSSI = %v, want -50 from the near source", identity.RSSI)
	}
	if identity.RSSISmoothed == nil || *identity.RSSISmoothed != -50 {
		t.Errorf("RSSISmoothed = %v, want -50: the distance must come from the strong source", identity.RSSISmoothed)
	}
	if len(identity.Sources) != 2 {
		t.Errorf("got %d source readings, want both anchors visible", len(identity.Sources))
	}
}

func TestASourceChangeMovesTheReading(t *testing.T) {
	// A proxy that went quiet keeps its last reading, so judging by strength alone would pin the
	// distance at the closest point ever heard while a further antenna kept reporting.
	clock := mocks.NewMockClock(func() time.Time { return at(9, 0) })
	service, capture := classifyService(t, clock, &historyStub{}, &recorderStub{})

	key := "addr:aa:bb:cc:dd:ee:01"
	service.record(model.Observation{
		Source: "near", Identity: model.Identity{Key: key, Stable: true}, RSSI: -50, ObservedAt: clock.Now(),
	})
	service.flush()

	// Past the trend horizon, so the near source's reading is genuinely stale rather than merely
	// older — otherwise both are current and strength legitimately wins.
	clock.Advance(2 * time.Minute)
	service.record(model.Observation{
		Source: "far", Identity: model.Identity{Key: key, Stable: true}, RSSI: -80, ObservedAt: clock.Now(),
	})
	service.flush()

	identity := capture.last().Identities[0]
	if identity.RSSI == nil || *identity.RSSI != -80 {
		t.Errorf("RSSI = %v, want -80 once only the far source reports", identity.RSSI)
	}
}

func TestUnregisteringASourceDropsItsReadings(t *testing.T) {
	// A proxy restarting repeatedly must not leave a residue behind. The descriptor map was cleared
	// but each identity kept a reading window and a last value for that source, so a source that
	// cycled often accumulated state for every identity it had ever heard.
	clock := mocks.NewMockClock(func() time.Time { return at(9, 0) })
	service, capture := classifyService(t, clock, &historyStub{}, &recorderStub{})

	key := "addr:aa:bb:cc:dd:ee:01"
	service.record(model.Observation{
		Source: "departing", Identity: model.Identity{Key: key, Stable: true}, RSSI: -50, ObservedAt: clock.Now(),
	})
	service.flush()

	service.UnregisterSource("departing")

	service.mutex.Lock()
	current := service.aggregates[key]
	readings := len(current.readings)
	sources := len(current.sources)
	primary := current.primary
	rssi := current.rssi
	service.mutex.Unlock()

	if readings != 0 || sources != 0 {
		t.Errorf("kept %d reading windows and %d source values, want none", readings, sources)
	}
	if primary != "" {
		t.Errorf("primary = %q, want empty once its only source left", primary)
	}
	if rssi != nil {
		t.Errorf("rssi = %v, want nil: a departed source's last value is not a current distance", *rssi)
	}

	// The identity still exists so another source can pick it up within the stale window.
	service.clock = clock
	service.record(model.Observation{
		Source: "arriving", Identity: model.Identity{Key: key, Stable: true}, RSSI: -70, ObservedAt: clock.Now(),
	})
	clock.Advance(2 * time.Second)
	service.flush()

	identity := capture.last().Identities[0]
	if identity.RSSI == nil || *identity.RSSI != -70 {
		t.Errorf("RSSI = %v, want -70 from the new source", identity.RSSI)
	}
}

func TestAnEqualStrengthSourceDoesNotDisplaceTheIncumbent(t *testing.T) {
	// Two proxies hearing a device equally well must not take turns. The first to report becomes
	// primary, and a later equal reading cannot displace it — which is what keeps the displayed
	// distance from flickering between two anchors that are genuinely the same distance away.
	//
	// Verified by construction rather than by repeated runs: map iteration order is randomised, so a
	// test that merely ran this many times could pass by luck either way.
	clock := mocks.NewMockClock(func() time.Time { return at(9, 0) })
	service, _ := classifyService(t, clock, &historyStub{}, &recorderStub{})
	key := "addr:aa:bb:cc:dd:ee:01"

	service.record(model.Observation{
		Source: "first", Identity: model.Identity{Key: key, Stable: true}, RSSI: -60, ObservedAt: clock.Now(),
	})
	for attempt := 0; attempt < 5; attempt++ {
		clock.Advance(time.Second)
		service.record(model.Observation{
			Source: "second", Identity: model.Identity{Key: key, Stable: true}, RSSI: -60, ObservedAt: clock.Now(),
		})
	}

	service.mutex.Lock()
	primary := service.aggregates[key].primary
	service.mutex.Unlock()

	if primary != "first" {
		t.Errorf("primary = %q, want the incumbent on an equal reading", primary)
	}
}
