package spatialhistory

import (
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := NewStoreAt(filepath.Join(t.TempDir(), "spatial.db"))
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestStorePrunesExpiredChunks(t *testing.T) {
	store := newTestStore(t)
	old := time.Now().UTC().Add(-10 * 24 * time.Hour).Truncate(time.Second)
	chunk := Chunk{Session: "old", From: old.UnixMilli(), To: old.UnixMilli(), Points: []Point{{At: old.UnixMilli(), Slot: 1, X: 0, Y: 0}}}
	if err := store.SaveChunk("radar", chunk); err != nil {
		t.Fatal(err)
	}
	if err := store.Prune(time.Now().UTC().Add(-Retention)); err != nil {
		t.Fatal(err)
	}
	chunks, err := store.Chunks("radar", old.Add(-time.Hour), time.Now())
	if err != nil || len(chunks) != 0 {
		t.Fatalf("expired chunk kept: %+v err=%v", chunks, err)
	}
	_ = store.Close()
}
