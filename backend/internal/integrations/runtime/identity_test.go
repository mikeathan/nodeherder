package runtime

// Checks allocation and reservation of integration identities.

import (
	"testing"

	"github.com/google/uuid"

	"node-herder/repository"
)

func TestAllocateRecordIDCollision(t *testing.T) {
	s, _ := serviceFixture(t, providerFunc(standardProvider))
	fixed := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	attempts := 0
	e := s.repo.Update(func(tx *repository.IntegrationTx) error {
		if e := tx.Put("identities", fixed.String(), identityRecord{Kind: "resource", Owner: "original"}); e != nil {
			return e
		}
		_, e := allocateRecordID(tx, "identities", func() (uuid.UUID, error) { attempts++; return fixed, nil })
		if e == nil {
			t.Fatal("collision accepted")
		}
		return nil
	})
	if e != nil || attempts != 3 {
		t.Fatal(attempts, e)
	}
}
