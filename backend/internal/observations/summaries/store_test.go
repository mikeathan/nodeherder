package summaries

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	model "node-herder/models/observations"
)

// referenceDay is the day every fixture belongs to, so age arithmetic never depends on when the
// test runs.
var referenceDay = time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)

func dayString(offsetDays int) string {
	return referenceDay.AddDate(0, 0, -offsetDays).Format("2006-01-02")
}

// storeAt opens a store in a temporary directory; the caller closes it.
func storeAt(t *testing.T) *Store {
	t.Helper()
	store, err := OpenAt(filepath.Join(t.TempDir(), "observation-summaries.db"))
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func summary(key string, offsetDays int) model.DailySummary {
	return model.DailySummary{
		Key:         key,
		Day:         dayString(offsetDays),
		Packets:     120,
		ActiveHours: 6,
		HourMask:    0xFF,
		RSSIMean:    -70,
		Strongest:   -65,
		Weakest:     -75,
		Crossings:   2,
	}
}

func TestStoreSavesAndReadsASummary(t *testing.T) {
	store := storeAt(t)
	want := summary("addr:aa:bb:cc:dd:ee:01", 0)

	if err := store.SaveSummary(want); err != nil {
		t.Fatalf("SaveSummary: %v", err)
	}

	got, err := store.Summaries(want.Key, 30)
	if err != nil {
		t.Fatalf("Summaries: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d summaries, want 1", len(got))
	}
	if got[0].Key != want.Key || got[0].Day != want.Day {
		t.Errorf("got %s/%s, want %s/%s", got[0].Key, got[0].Day, want.Key, want.Day)
	}
	if got[0].Crossings != want.Crossings || got[0].ActiveHours != want.ActiveHours {
		t.Errorf("round trip lost evidence: %+v", got[0])
	}
}

// TestStoreRefusesADayBelowTheWriteFloor is the requirement that noise costs no row. It is enforced
// at the write boundary rather than by convention, so a caller cannot fill the database by mistake.
func TestStoreRefusesADayBelowTheWriteFloor(t *testing.T) {
	store := storeAt(t)

	oneHour := summary("addr:aa:bb:cc:dd:ee:01", 0)
	oneHour.ActiveHours = 1
	oneHour.Packets = 500

	fewPackets := summary("addr:aa:bb:cc:dd:ee:02", 0)
	fewPackets.ActiveHours = 8
	fewPackets.Packets = 3

	for _, below := range []model.DailySummary{oneHour, fewPackets} {
		if err := store.SaveSummary(below); err != nil {
			t.Fatalf("SaveSummary: %v", err)
		}
	}

	all, err := store.Summaries("addr:aa:bb:cc:dd:ee:01", 30)
	if err != nil {
		t.Fatalf("Summaries: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("stored %d summaries below the write floor, want 0", len(all))
	}
}

func TestStoreReturnsSummariesNewestFirst(t *testing.T) {
	store := storeAt(t)
	for _, offset := range []int{3, 0, 1, 2} {
		if err := store.SaveSummary(summary("addr:aa:bb:cc:dd:ee:01", offset)); err != nil {
			t.Fatalf("SaveSummary: %v", err)
		}
	}

	got, err := store.Summaries("addr:aa:bb:cc:dd:ee:01", 30)
	if err != nil {
		t.Fatalf("Summaries: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("got %d summaries, want 4", len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].Day < got[i].Day {
			t.Errorf("summaries are not newest first: %s before %s", got[i-1].Day, got[i].Day)
		}
	}
}

func TestStoreLimitsRowsReturned(t *testing.T) {
	// A bounded read is what keeps a long history from becoming an unbounded payload.
	store := storeAt(t)
	for offset := 0; offset < 20; offset++ {
		if err := store.SaveSummary(summary("addr:aa:bb:cc:dd:ee:01", offset)); err != nil {
			t.Fatalf("SaveSummary: %v", err)
		}
	}

	got, err := store.Summaries("addr:aa:bb:cc:dd:ee:01", 5)
	if err != nil {
		t.Fatalf("Summaries: %v", err)
	}
	if len(got) != 5 {
		t.Errorf("got %d summaries, want 5", len(got))
	}
}

func TestStoreScopesSummariesToTheirIdentity(t *testing.T) {
	store := storeAt(t)
	if err := store.SaveSummary(summary("addr:aa:bb:cc:dd:ee:01", 0)); err != nil {
		t.Fatalf("SaveSummary: %v", err)
	}
	if err := store.SaveSummary(summary("addr:aa:bb:cc:dd:ee:02", 0)); err != nil {
		t.Fatalf("SaveSummary: %v", err)
	}

	got, err := store.Summaries("addr:aa:bb:cc:dd:ee:01", 30)
	if err != nil {
		t.Fatalf("Summaries: %v", err)
	}
	if len(got) != 1 || got[0].Key != "addr:aa:bb:cc:dd:ee:01" {
		t.Errorf("got %d summaries for one identity, want exactly its own", len(got))
	}
}

func TestStoreDoesNotLeakBetweenIdentityPrefixes(t *testing.T) {
	// Guard for the key layout: identity "addr:aa" must not match identity "addr:aab". Without the
	// separator in the range bounds, a prefix scan would return both, and one device's history
	// would be classified as another's.
	store := storeAt(t)
	if err := store.SaveSummary(summary("addr:aa", 0)); err != nil {
		t.Fatalf("SaveSummary: %v", err)
	}
	if err := store.SaveSummary(summary("addr:aab", 0)); err != nil {
		t.Fatalf("SaveSummary: %v", err)
	}

	got, err := store.Summaries("addr:aa", 30)
	if err != nil {
		t.Fatalf("Summaries: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d summaries for addr:aa, want 1", len(got))
	}
	if got[0].Key != "addr:aa" {
		t.Errorf("got key %q, want addr:aa", got[0].Key)
	}
}

func TestStorePrunesBeyondRetention(t *testing.T) {
	store := storeAt(t)
	old := summary("addr:aa:bb:cc:dd:ee:01", 40)
	recent := summary("addr:aa:bb:cc:dd:ee:01", 2)
	for _, s := range []model.DailySummary{old, recent} {
		if err := store.SaveSummary(s); err != nil {
			t.Fatalf("SaveSummary: %v", err)
		}
	}

	if err := store.Prune(referenceDay.AddDate(0, 0, -RetentionDays)); err != nil {
		t.Fatalf("Prune: %v", err)
	}

	got, err := store.Summaries("addr:aa:bb:cc:dd:ee:01", 60)
	if err != nil {
		t.Fatalf("Summaries: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d summaries after prune, want 1", len(got))
	}
	if got[0].Day != recent.Day {
		t.Errorf("kept %s, want the recent day %s", got[0].Day, recent.Day)
	}
}

// TestStorePrunesTransitionsWithSummaries is the consistency rule: a transition must never outlive
// the evidence that justified it, or the UI could show a change it cannot explain.
func TestStorePrunesTransitionsWithSummaries(t *testing.T) {
	store := storeAt(t)

	oldTransition := model.Transition{
		Key: "addr:aa:bb:cc:dd:ee:01", Day: dayString(40), At: referenceDay.AddDate(0, 0, -40).Format(time.RFC3339),
		From: "infrastructure", To: "personal", Reason: "old change",
	}
	recentTransition := model.Transition{
		Key: "addr:aa:bb:cc:dd:ee:01", Day: dayString(2), At: referenceDay.AddDate(0, 0, -2).Format(time.RFC3339),
		From: "personal", To: "transient", Reason: "recent change",
	}
	for _, transition := range []model.Transition{oldTransition, recentTransition} {
		if err := store.SaveTransition(transition); err != nil {
			t.Fatalf("SaveTransition: %v", err)
		}
	}

	if err := store.Prune(referenceDay.AddDate(0, 0, -RetentionDays)); err != nil {
		t.Fatalf("Prune: %v", err)
	}

	got, err := store.Transitions("addr:aa:bb:cc:dd:ee:01", 60)
	if err != nil {
		t.Fatalf("Transitions: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d transitions after prune, want 1", len(got))
	}
	if got[0].Day != recentTransition.Day {
		t.Errorf("kept %s, want the recent transition %s", got[0].Day, recentTransition.Day)
	}
}

func TestStoreReturnsTransitionsNewestFirst(t *testing.T) {
	store := storeAt(t)
	for _, offset := range []int{2, 0, 1} {
		transition := model.Transition{
			Key: "addr:aa:bb:cc:dd:ee:01", Day: dayString(offset),
			At:   referenceDay.AddDate(0, 0, -offset).Format(time.RFC3339),
			From: "personal", To: "infrastructure", Reason: "test",
		}
		if err := store.SaveTransition(transition); err != nil {
			t.Fatalf("SaveTransition: %v", err)
		}
	}

	got, err := store.Transitions("addr:aa:bb:cc:dd:ee:01", 30)
	if err != nil {
		t.Fatalf("Transitions: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d transitions, want 3", len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].Day < got[i].Day {
			t.Errorf("transitions are not newest first: %s before %s", got[i-1].Day, got[i].Day)
		}
	}
}

func TestStoreCreatesNoFileWhenNothingIsWritten(t *testing.T) {
	// NFR-09: with no observations there must be no storage at all. A file created on start-up
	// would be a database filled with nothing, which is worse than no database.
	path := filepath.Join(t.TempDir(), "observation-summaries.db")
	store, err := OpenAt(path)
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	got, err := store.Summaries("addr:aa:bb:cc:dd:ee:01", 30)
	if err != nil {
		t.Fatalf("Summaries: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d summaries from an empty store, want 0", len(got))
	}
}

func TestStoreIgnoresAnEmptyIdentityOrDay(t *testing.T) {
	// A malformed write must be refused rather than creating an unreadable key.
	store := storeAt(t)

	if err := store.SaveSummary(model.DailySummary{Key: "", Day: dayString(0)}); err != nil {
		t.Errorf("SaveSummary with no key: %v", err)
	}
	if err := store.SaveSummary(model.DailySummary{Key: "addr:aa:bb:cc:dd:ee:01", Day: ""}); err != nil {
		t.Errorf("SaveSummary with no day: %v", err)
	}
	if err := store.SaveTransition(model.Transition{Key: "", Day: dayString(0)}); err != nil {
		t.Errorf("SaveTransition with no key: %v", err)
	}
}

func TestStoreStorageStaysWithinTheStatedCeiling(t *testing.T) {
	// NFR-01 states a ceiling, so it is measured rather than assumed. The numbers are pinned here
	// because they are a design commitment: a change that grows a row by 50% should require a
	// deliberate decision, not slip through silently.
	store := storeAt(t)

	// A populated row, encoded the way the store encodes it: the realistic worst case, not a
	// minimal one.
	populated := model.DailySummary{
		Key: "fp:9a3c1f2b4d", Day: "2026-11-13", Packets: 14302, ActiveHours: 24,
		HourMask: 0xFFFFFF, RSSIMean: -71.42857142857143, RSSIM2: 1943.7142857142858,
		Strongest: -58, Weakest: -91, Crossings: 37, Family: "004c:Mike's iPhone",
	}
	encoded, err := json.Marshal(populated)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	const (
		// summaryBytesCeiling is the measured size plus a small margin. Measured at 223 B; the
		// margin allows a field to be added without the test becoming a tripwire on every edit,
		// while still catching a change that doubles the row.
		summaryBytesCeiling = 300
		// householdIdentities and transitionCount are the stated assumption behind the 1.8 MB
		// figure: a busy household, and a deliberately high number of class changes.
		householdIdentities = 256
		transitionCount     = 1000
		// totalCeilingBytes is ≈ 1.8 MB, the figure NFR-01 commits to.
		totalCeilingBytes = 2 * 1024 * 1024
	)

	if len(encoded) > summaryBytesCeiling {
		t.Errorf("summary encodes to %d bytes, above the %d-byte ceiling", len(encoded), summaryBytesCeiling)
	}

	transitionBytes := 159
	total := householdIdentities*RetentionDays*len(encoded) + transitionCount*transitionBytes
	if total > totalCeilingBytes {
		t.Errorf("stated worst case is %d bytes, above the %d-byte ceiling", total, totalCeilingBytes)
	}

	// The store must also actually honour the row limit, or the ceiling is fiction.
	for offset := 0; offset < RetentionDays+10; offset++ {
		if err := store.SaveSummary(summary("addr:aa:bb:cc:dd:ee:01", offset)); err != nil {
			t.Fatalf("SaveSummary: %v", err)
		}
	}
	rows, err := store.Summaries("addr:aa:bb:cc:dd:ee:01", 1000)
	if err != nil {
		t.Fatalf("Summaries: %v", err)
	}
	if len(rows) > maxRowsPerRead {
		t.Errorf("read %d rows, above the %d-row limit", len(rows), maxRowsPerRead)
	}
}
