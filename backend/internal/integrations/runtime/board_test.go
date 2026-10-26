package runtime

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"testing"

	model "node-herder/models/integrations"
	"node-herder/repository"
	"node-herder/utils"
)

func seedBoard(t *testing.T, s *Service) {
	t.Helper()
	if err := s.repo.Update(func(tx *repository.IntegrationTx) error {
		for _, id := range []string{"a", "b", "c"} {
			instance := model.Instance{ID: id, Name: id, Enabled: true, Revision: 7, Generation: 3, CreatedAt: s.clock.Now(), UpdatedAt: s.clock.Now()}
			if err := tx.Put("instances", id, instance); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
func TestBoardAtomicLayoutAndLifecycleIsolation(t *testing.T) {
	s, _ := serviceFixture(t, providerFunc(func(_ *http.Request) (*http.Response, error) { t.Fatal("layout contacted provider"); return nil, nil }))
	seedBoard(t, s)
	initial, err := s.Board()
	if err != nil {
		t.Fatal(err)
	}
	if len(initial.Entries) != 3 || initial.Entries[0].InstanceID != "a" {
		t.Fatal(initial)
	}
	before, _ := s.Instance("a")
	proposal := []BoardEntry{{InstanceID: "c"}, {InstanceID: "a", Hidden: true}, {InstanceID: "b"}}
	saved, err := s.SaveBoard(proposal, initial.ETag)
	if err != nil {
		t.Fatal(err)
	}
	if saved.ETag == initial.ETag || saved.Entries[0].InstanceID != "c" || !saved.Entries[1].Hidden {
		t.Fatal(saved)
	}
	after, _ := s.Instance("a")
	if before.Revision != after.Revision || before.ETag != after.ETag || !before.UpdatedAt.Equal(after.UpdatedAt) {
		t.Fatal("layout changed lifecycle", before, after)
	}
	// Ordinary lifecycle updates must not invalidate a layout draft.
	if _, err = s.UpdateInstance("a", "Renamed", false, after.ETag); err != nil {
		t.Fatal(err)
	}
	unchanged, _ := s.Board()
	if unchanged.ETag != saved.ETag {
		t.Fatal("lifecycle changed board ETag")
	}
	repeated, err := s.SaveBoard(proposal, saved.ETag)
	if err != nil || repeated.ETag != saved.ETag {
		t.Fatal("idempotent save", err, repeated)
	}
	if _, err = s.SaveBoard(initial.Entries, initial.ETag); err == nil {
		t.Fatal("stale layout accepted")
	}
	current, _ := s.Instance("a")
	if err = s.DeleteInstance("a", current.ETag); err != nil {
		t.Fatal(err)
	}
	if _, err = s.SaveBoard(proposal, saved.ETag); err == nil {
		t.Fatal("deleted member accepted")
	}
}
func TestBoardInvalidProposalsDoNotPartiallyWrite(t *testing.T) {
	s, _ := serviceFixture(t, nil)
	seedBoard(t, s)
	before, _ := s.Board()
	for _, entries := range [][]BoardEntry{nil, {}, {{InstanceID: "a"}}, {{InstanceID: "a", Hidden: true}, {InstanceID: "a"}, {InstanceID: "c"}}, {{InstanceID: "a", Hidden: true}, {InstanceID: "b"}, {InstanceID: "missing"}}} {
		if _, err := s.SaveBoard(entries, before.ETag); err == nil {
			t.Fatal("invalid layout accepted", entries)
		}
		after, _ := s.Board()
		if after.ETag != before.ETag || after.Entries[0].Hidden {
			t.Fatal("invalid proposal partially wrote")
		}
	}
	if _, err := s.SaveBoard(before.Entries, ""); err == nil {
		t.Fatal("missing precondition accepted")
	}
	// A later failure in the same repository transaction rolls back earlier preference writes.
	rollback := errors.New("rollback")
	err := s.repo.Update(func(tx *repository.IntegrationTx) error {
		var instance model.Instance
		if _, err := tx.Get("instances", "a", &instance); err != nil {
			return err
		}
		instance.Board = &model.BoardPreference{Order: 0, Hidden: true, Revision: 1}
		if err := tx.Put("instances", "a", instance); err != nil {
			return err
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatal(err)
	}
	after, _ := s.Board()
	if after.ETag != before.ETag {
		t.Fatal("transaction did not roll back")
	}
}
func TestBoardPreferencesSurviveEncryptedRepositoryRestart(t *testing.T) {
	directory := t.TempDir()
	keys := repository.IntegrationKeyring{ActiveKeyID: "test", Keys: map[string]string{"test": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{5}, 32))}}
	repo, err := repository.OpenIntegrationRepository(directory, keys)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	s := &Service{repo: repo, clock: utils.NewRealClock()}
	seedBoard(t, s)
	board, _ := s.Board()
	board.Entries[0].Hidden = true
	saved, err := s.SaveBoard(board.Entries, board.ETag)
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := repository.OpenIntegrationRepository(directory, keys)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	s.repo = reopened
	got, err := s.Board()
	if err != nil || got.ETag != saved.ETag || !got.Entries[0].Hidden {
		t.Fatal(got, err)
	}
}
