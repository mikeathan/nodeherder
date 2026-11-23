package spatialhistory

import (
	"testing"
	"time"
)

func TestSessionsIncludeAnOngoingChunkStartedBeforeTheRange(t *testing.T) {
	store := newTestStore(t)
	base := time.Now().UTC().Truncate(time.Second)
	chunk := Chunk{Session: "ongoing", From: base.UnixMilli(), To: base.Add(4 * time.Minute).UnixMilli(), Points: []Point{
		{At: base.UnixMilli(), Slot: 1, X: 0, Y: 0},
		{At: base.Add(4 * time.Minute).UnixMilli(), Slot: 1, X: 400, Y: 0},
	}}
	if err := store.SaveChunk("radar", chunk); err != nil {
		t.Fatal(err)
	}
	service := NewService(store)
	sessions, err := service.Sessions("radar", base.Add(time.Minute), base.Add(10*time.Minute))
	if err != nil || len(sessions) != 1 || sessions[0].Points != 1 {
		t.Fatalf("overlap sessions = %+v err=%v", sessions, err)
	}
	_ = store.Close()
}

func TestSessionsReportEveryRecordedSampleInsideTheRange(t *testing.T) {
	store := newTestStore(t)
	base := time.Now().UTC().Truncate(time.Second)
	// Two chunks of the same session, separated by a pause. Every stored position counts as
	// activity: marking only where movement resumed left continuously moving recordings looking
	// empty (18 real samples produced 10 marks). A third chunk belongs to another session and must
	// not leak into this session's samples.
	first := Chunk{Session: "day", From: base.UnixMilli(), To: base.Add(5 * time.Second).UnixMilli(), Points: []Point{
		{At: base.UnixMilli(), Slot: 1, X: 0, Y: 0},
		{At: base.Add(5 * time.Second).UnixMilli(), Slot: 1, X: 400, Y: 0},
	}}
	second := Chunk{Session: "day", From: base.Add(10 * time.Minute).UnixMilli(), To: base.Add(10 * time.Minute).UnixMilli(), Points: []Point{
		{At: base.Add(10 * time.Minute).UnixMilli(), Slot: 1, X: 900, Y: 0},
	}}
	other := Chunk{Session: "other", From: base.Add(20 * time.Minute).UnixMilli(), To: base.Add(20 * time.Minute).Add(5 * time.Second).UnixMilli(), Points: []Point{
		{At: base.Add(20 * time.Minute).UnixMilli(), Slot: 1, X: 0, Y: 0},
		{At: base.Add(20 * time.Minute).Add(5 * time.Second).UnixMilli(), Slot: 1, X: 0, Y: 0},
	}}
	for _, chunk := range []Chunk{first, second, other} {
		if err := store.SaveChunk("radar", chunk); err != nil {
			t.Fatal(err)
		}
	}
	service := NewService(store)
	sessions, err := service.Sessions("radar", base.Add(-time.Minute), base.Add(30*time.Minute))
	if err != nil || len(sessions) != 2 {
		t.Fatalf("sessions = %+v err=%v", sessions, err)
	}
	want := []int64{base.UnixMilli(), base.Add(5 * time.Second).UnixMilli(), base.Add(10 * time.Minute).UnixMilli()}
	if got := sessions[0].Samples; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("day samples = %v want %v", got, want)
	}
	// Every sample of the other session is reported too, not only the first of a burst.
	if got := sessions[1].Samples; len(got) != 2 {
		t.Fatalf("other samples = %v", got)
	}
	_ = store.Close()
}

func TestSessionsIgnoreMovementsOutsideTheRequestedRange(t *testing.T) {
	store := newTestStore(t)
	base := time.Now().UTC().Truncate(time.Second)
	chunk := Chunk{Session: "day", From: base.UnixMilli(), To: base.Add(4 * time.Minute).UnixMilli(), Points: []Point{
		{At: base.UnixMilli(), Slot: 1, X: 0, Y: 0},
		{At: base.Add(2 * time.Minute).UnixMilli(), Slot: 1, X: 400, Y: 0},
		{At: base.Add(4 * time.Minute).UnixMilli(), Slot: 1, X: 900, Y: 0},
	}}
	if err := store.SaveChunk("radar", chunk); err != nil {
		t.Fatal(err)
	}
	service := NewService(store)
	sessions, err := service.Sessions("radar", base.Add(time.Minute), base.Add(3*time.Minute))
	if err != nil || len(sessions) != 1 {
		t.Fatalf("sessions = %+v err=%v", sessions, err)
	}
	if got := sessions[0].Samples; len(got) != 1 || got[0] != base.Add(2*time.Minute).UnixMilli() {
		t.Fatalf("samples = %v", got)
	}
	_ = store.Close()
}

func TestSamplesFromCapsDenseActivityAndKeepsItsShape(t *testing.T) {
	times := make([]int64, 0, MaxSamplesPerSession*2)
	for index := 0; index < MaxSamplesPerSession*2; index++ {
		times = append(times, int64(index)*1_000)
	}
	samples := samplesFrom(times)
	if len(samples) != MaxSamplesPerSession {
		t.Fatalf("samples = %d", len(samples))
	}
	// Evenly spread and closing on the true last sample: the reduction must not truncate to the
	// opening moments or drop the recording's final position.
	if samples[0] != times[0] || samples[len(samples)-1] != times[len(times)-1] {
		t.Fatalf("samples do not keep both endpoints: first=%d last=%d want first=%d last=%d",
			samples[0], samples[len(samples)-1], times[0], times[len(times)-1])
	}
	if got := samplesFrom(nil); len(got) != 0 {
		t.Fatalf("empty samples = %v", got)
	}
	short := samplesFrom([]int64{5, 9})
	if len(short) != 2 || short[0] != 5 || short[1] != 9 {
		t.Fatalf("short samples = %v", short)
	}
}

// A recording whose chunk started before the window is still listed by Sessions (it has points
// inside the range). Points must return the same subset; when it did not, the player listed a
// recording, loaded nothing for it and disabled Play, which read as a broken recording.
func TestPointsReturnWhatSessionsListedForAStraddlingChunk(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	base := time.Now().UTC().Truncate(time.Second)
	chunk := Chunk{Session: "straddle", From: base.UnixMilli(), To: base.Add(25 * time.Minute).UnixMilli(), Points: []Point{
		{At: base.UnixMilli(), Slot: 1, X: 0, Y: 0},
		{At: base.Add(15 * time.Minute).UnixMilli(), Slot: 1, X: 400, Y: 0},
		{At: base.Add(25 * time.Minute).UnixMilli(), Slot: 1, X: 900, Y: 0},
	}}
	if err := store.SaveChunk("radar", chunk); err != nil {
		t.Fatal(err)
	}
	service := NewService(store)
	from, to := base.Add(5*time.Minute), base.Add(40*time.Minute)
	sessions, err := service.Sessions("radar", from, to)
	if err != nil || len(sessions) != 1 {
		t.Fatalf("sessions = %+v err=%v", sessions, err)
	}
	points, err := service.Points("radar", "straddle", from, to)
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != sessions[0].Points {
		t.Fatalf("listed %d points but loaded %d", sessions[0].Points, len(points))
	}
}

// The heatmap is drawn beside the recordings and points, so it must count the same straddling
// chunk they do. It read Chunks(device, from, to) without the lookback, so a chunk that started
// just before the window was listed with points but contributed no cells.
func TestHeatmapCountsAChunkThatStartedBeforeTheRange(t *testing.T) {
	store := newTestStore(t)
	defer store.Close()
	base := time.Now().UTC().Truncate(time.Second)
	chunk := Chunk{Session: "straddle", From: base.UnixMilli(), To: base.Add(25 * time.Minute).UnixMilli(), Points: []Point{
		{At: base.UnixMilli(), Slot: 1, X: 0, Y: 0},
		{At: base.Add(15 * time.Minute).UnixMilli(), Slot: 1, X: 400, Y: 0},
		{At: base.Add(25 * time.Minute).UnixMilli(), Slot: 1, X: 900, Y: 0},
	}}
	if err := store.SaveChunk("radar", chunk); err != nil {
		t.Fatal(err)
	}
	service := NewService(store)
	from, to := base.Add(5*time.Minute), base.Add(40*time.Minute)
	cells, err := service.Heatmap("radar", from, to, 250)
	if err != nil {
		t.Fatal(err)
	}
	// Two of the three points fall inside the window; neither may be dropped.
	total := 0
	for _, cell := range cells {
		total += cell.Count
	}
	if total != 2 {
		t.Fatalf("heatmap counted %d points, want 2 (cells = %+v)", total, cells)
	}
}
