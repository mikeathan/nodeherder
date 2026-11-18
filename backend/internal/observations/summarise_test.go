package observations

// Recorder tests. The risky behaviour is time-boundary handling: a day rollover that either loses
// the previous day or double-counts it would silently corrupt the evidence classification reads.

import (
	"fmt"
	"sync"
	"testing"
	"time"

	model "node-herder/models/observations"
)

// fixedClock is a clock the test moves by hand, so day boundaries are exact.
type fixedClock struct {
	mutex sync.Mutex
	now   time.Time
}

func (c *fixedClock) Now() time.Time {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.now
}

func (c *fixedClock) advance(d time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.now = c.now.Add(d)
}

// recordingRepository captures what the recorder tried to store.
type recordingRepository struct {
	mutex     sync.Mutex
	summaries []model.DailySummary
	pruned    []time.Time
	closed    bool
}

func (r *recordingRepository) SaveSummary(summary model.DailySummary) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.summaries = append(r.summaries, summary)
	return nil
}

func (r *recordingRepository) SaveTransition(model.Transition) error { return nil }

func (r *recordingRepository) Summaries(string, int) ([]model.DailySummary, error) { return nil, nil }

func (r *recordingRepository) Transitions(string, int) ([]model.Transition, error) { return nil, nil }

func (r *recordingRepository) Prune(before time.Time) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.pruned = append(r.pruned, before)
	return nil
}

func (r *recordingRepository) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.closed = true
	return nil
}

func (r *recordingRepository) saved() []model.DailySummary {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return append([]model.DailySummary(nil), r.summaries...)
}

// waitFor polls a condition briefly, so a test can synchronise with the worker without sleeping a
// fixed amount and hoping.
func waitFor(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition was not met before the deadline")
}

// observationAt builds an observation for one identity at a given time.
func observationAt(key string, rssi int, when time.Time) model.Observation {
	return model.Observation{
		Source:     "esp-1",
		Medium:     model.MediumBLEAdvertisement,
		Identity:   model.Identity{Key: key, Kind: model.KindAdvertisement, Stable: true, Vendor: 0x004c},
		RSSI:       rssi,
		ObservedAt: when,
	}
}

// newTestRecorder builds a recorder wired to a fake clock and repository.
func newTestRecorder(t *testing.T, start time.Time) (*SummarisingRecorder, *recordingRepository, *fixedClock) {
	t.Helper()
	repo := &recordingRepository{}
	clock := &fixedClock{now: start}
	recorder := NewSummarisingRecorder(repo, clock)
	return recorder, repo, clock
}

func TestRecorderStoresADayOnShutdown(t *testing.T) {
	// A restart must not discard the evidence gathered so far, or every restart would reset a
	// device's history and make long-standing equipment look new.
	start := time.Date(2026, 11, 13, 9, 0, 0, 0, time.UTC)
	recorder, repo, _ := newTestRecorder(t, start)
	recorder.Start()

	for i := 0; i < 30; i++ {
		recorder.Observe(observationAt("addr:aa:bb:cc:dd:ee:01", -60, start.Add(time.Duration(i)*time.Minute)))
	}
	recorder.Observe(observationAt("addr:aa:bb:cc:dd:ee:01", -60, start.Add(2*time.Hour)))

	recorder.Close()

	saved := repo.saved()
	if len(saved) != 1 {
		t.Fatalf("saved %d summaries on shutdown, want 1", len(saved))
	}
	if saved[0].Day != "2026-11-13" {
		t.Errorf("Day = %q, want 2026-11-13", saved[0].Day)
	}
	if saved[0].Key != "addr:aa:bb:cc:dd:ee:01" {
		t.Errorf("Key = %q", saved[0].Key)
	}
	if !repo.closed {
		t.Errorf("repository was not closed")
	}
}

func TestRecorderSplitsDaysAtMidnight(t *testing.T) {
	// An observation arriving after midnight must complete the previous day and start a new one,
	// not append to yesterday. Otherwise one row would span two days and the hour mask would lie.
	start := time.Date(2026, 11, 13, 23, 0, 0, 0, time.UTC)
	recorder, repo, _ := newTestRecorder(t, start)
	recorder.Start()

	for i := 0; i < 25; i++ {
		recorder.Observe(observationAt("addr:aa:bb:cc:dd:ee:01", -60, start.Add(time.Duration(i)*time.Minute)))
	}
	recorder.Observe(observationAt("addr:aa:bb:cc:dd:ee:01", -60, start.Add(2*time.Hour)))
	for i := 0; i < 25; i++ {
		recorder.Observe(observationAt("addr:aa:bb:cc:dd:ee:01", -60, start.Add(2*time.Hour+time.Duration(i)*time.Minute)))
	}

	recorder.Close()

	saved := repo.saved()
	days := map[string]int{}
	for _, summary := range saved {
		days[summary.Day]++
	}
	if days["2026-11-13"] != 1 {
		t.Errorf("saved %d rows for 2026-11-13, want 1", days["2026-11-13"])
	}
	if days["2026-11-14"] != 1 {
		t.Errorf("saved %d rows for 2026-11-14, want 1", days["2026-11-14"])
	}
}

func TestRecorderKnowsItsFamily(t *testing.T) {
	// Rotating addresses from one device must share a family, or a phone becomes dozens of
	// separately-classified transient identities.
	start := time.Date(2026, 11, 13, 9, 0, 0, 0, time.UTC)
	recorder, repo, _ := newTestRecorder(t, start)
	recorder.Start()

	labelled := model.Observation{
		Medium: model.MediumBLEAdvertisement,
		Identity: model.Identity{
			Key: "fp:1111", Kind: model.KindAdvertisement, Stable: false,
			Vendor: 0x004c, Label: "Mike's iPhone",
		},
		RSSI: -60, ObservedAt: start,
	}
	unlabelled := labelled
	unlabelled.Identity = model.Identity{Key: "fp:2222", Kind: model.KindAdvertisement, Stable: false, Vendor: 0x004c}

	recorder.Observe(labelled)
	recorder.Observe(unlabelled)
	recorder.Close()

	saved := repo.saved()
	if len(saved) == 0 {
		t.Fatal("nothing was saved")
	}
	families := map[string]string{}
	for _, summary := range saved {
		families[summary.Key] = summary.Family
	}
	if families["fp:1111"] != "004c:Mike's iPhone" {
		t.Errorf("family = %q, want the vendor and label", families["fp:1111"])
	}
	// With no label the vendor alone is the family: coarser, but still groups the rotations.
	if families["fp:2222"] != "004c" {
		t.Errorf("family = %q, want the vendor alone", families["fp:2222"])
	}
}

func TestRecorderCountsDropsRatherThanBlocking(t *testing.T) {
	// A full queue must be visible, not silent, and must never stall the caller.
	start := time.Date(2026, 11, 13, 9, 0, 0, 0, time.UTC)
	recorder, _, _ := newTestRecorder(t, start)
	// Deliberately not started: nothing drains the queue, so it fills and rejects.

	for i := 0; i < summaryQueueSize+10; i++ {
		recorder.Observe(observationAt("addr:aa:bb:cc:dd:ee:01", -60, start))
	}

	if got := recorder.Dropped(); got != 10 {
		t.Errorf("Dropped = %d, want 10", got)
	}
}

func TestRecorderIgnoresObservationsWithoutAnIdentity(t *testing.T) {
	// An observation with no key cannot be attributed to a device, so storing it would create a row
	// nobody could look up.
	start := time.Date(2026, 11, 13, 9, 0, 0, 0, time.UTC)
	recorder, _, _ := newTestRecorder(t, start)

	recorder.Observe(model.Observation{RSSI: -60, ObservedAt: start})

	if got := recorder.Dropped(); got != 0 {
		t.Errorf("Dropped = %d, want 0: an unattributable observation is not a drop", got)
	}
}

func TestRecorderPromotesStaleDaysOnTick(t *testing.T) {
	// A device that stops advertising must still have its completed day stored; otherwise its
	// evidence would sit in memory until the next observation or a restart.
	start := time.Date(2026, 11, 13, 22, 0, 0, 0, time.UTC)
	recorder, repo, clock := newTestRecorder(t, start)
	recorder.Start()

	for i := 0; i < 30; i++ {
		recorder.Observe(observationAt("addr:aa:bb:cc:dd:ee:01", -70, start.Add(time.Duration(i)*time.Minute)))
	}

	// Wait for the worker to consume the queue, then tick by hand. Calling tick directly while the
	// worker still holds queued work would test the queue rather than the promotion.
	waitFor(t, func() bool {
		recorder.mutex.Lock()
		defer recorder.mutex.Unlock()
		accumulator, ok := recorder.accumulate["addr:aa:bb:cc:dd:ee:01"]
		return ok && accumulator.state.packets == 30
	})

	clock.advance(3 * time.Hour)
	recorder.tick()

	saved := repo.saved()
	if len(saved) != 1 {
		t.Fatalf("saved %d summaries on tick, want 1", len(saved))
	}
	if saved[0].Day != "2026-11-13" {
		t.Errorf("Day = %q, want 2026-11-13", saved[0].Day)
	}

	recorder.Close()
}

func TestRecorderBoundsTheAccumulatorMap(t *testing.T) {
	// A busy area produces thousands of rotating addresses, and rotating addresses are exactly what
	// would grow this map until midnight. The cap is what keeps memory finite; hitting it drops
	// rather than evicts, because an evicted accumulator would lose the only copy of that day.
	start := time.Date(2026, 11, 13, 9, 0, 0, 0, time.UTC)
	recorder, _, _ := newTestRecorder(t, start)
	recorder.Start()

	for i := 0; i < maxAccumulatedIdentities+50; i++ {
		recorder.Observe(observationAt(fmt.Sprintf("fp:%06x", i), -70, start))
	}

	waitFor(t, func() bool {
		recorder.mutex.Lock()
		defer recorder.mutex.Unlock()
		return recorder.dropped >= 50
	})

	recorder.mutex.Lock()
	size := len(recorder.accumulate)
	recorder.mutex.Unlock()

	if size > maxAccumulatedIdentities {
		t.Errorf("accumulator map holds %d identities, want at most %d", size, maxAccumulatedIdentities)
	}

	recorder.Close()
}
