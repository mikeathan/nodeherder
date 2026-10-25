package integrations

import (
	"node-herder/repository"
	"testing"
	"time"
)

func TestProviderLimitsExpireFromStorageAndMemory(t *testing.T) {
	s, _ := serviceFixture(t, providerFunc(standardProvider))
	now := s.clock.Now()
	future := now.Add(time.Hour)
	if err := s.repo.Update(func(tx *repository.IntegrationTx) error {
		if err := saveProviderLimit(tx, "https://expired.example", now.Add(-time.Second)); err != nil {
			return err
		}
		return saveProviderLimit(tx, "https://active.example", future)
	}); err != nil {
		t.Fatal(err)
	}
	s.providerDeadlines["https://expired.example"] = now.Add(-time.Second)
	if err := s.Cleanup(); err != nil {
		t.Fatal(err)
	}
	if len(s.providerDeadlines) != 1 || !s.providerDeadlines["https://active.example"].Equal(future) {
		t.Fatal("expired origin retained", s.providerDeadlines)
	}
	if err := s.repo.View(func(tx *repository.IntegrationTx) error {
		count := 0
		err := tx.Each("limits", func(_ string, _ []byte) error { count++; return nil })
		if count != 1 {
			t.Errorf("stored limits = %d, want 1", count)
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
}
