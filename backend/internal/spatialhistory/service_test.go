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
