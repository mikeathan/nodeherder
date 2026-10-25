package integrations

import (
	"github.com/google/uuid"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

// allocateRecordID checks before insertion; a UUID collision never overwrites an aggregate.
func allocateRecordID(tx *repository.IntegrationTx, bucket string, generate func() (uuid.UUID, error)) (string, error) {
	for attempt := 0; attempt < 3; attempt++ {
		id, e := generate()
		if e != nil {
			return "", e
		}
		var existing any
		found, e := tx.Get(bucket, id.String(), &existing)
		if e != nil {
			return "", e
		}
		if !found {
			return id.String(), nil
		}
	}
	return "", model.Fail("identity_collision", "Could not allocate a unique connection identity.")
}

type identityRecord struct {
	Kind  string `json:"kind"`
	Owner string `json:"owner"`
}

// New IDs are tentative until the snapshot publication transaction reserves them.
func snapshotIdentityAllocator(tx *repository.IntegrationTx) func(string) (string, error) {
	pending := map[string]bool{}
	return func(_ string) (string, error) {
		for attempt := 0; attempt < 3; attempt++ {
			id, e := uuid.NewRandom()
			if e != nil {
				return "", e
			}
			if pending[id.String()] {
				continue
			}
			var existing identityRecord
			found, e := tx.Get("identities", id.String(), &existing)
			if e != nil {
				return "", e
			}
			if !found {
				pending[id.String()] = true
				return id.String(), nil
			}
		}
		return "", model.Fail("identity_collision", "Could not allocate a unique resource identity.")
	}
}
func reserveSnapshotIdentities(tx *repository.IntegrationTx, snapshot model.WasteSchedule, owner string) error {
	identities := map[string]string{snapshot.ResourceID: "resource"}
	for _, service := range snapshot.Services {
		if _, duplicate := identities[service.ServiceID]; duplicate {
			return model.Fail("identity_collision", "A resource identity is duplicated.")
		}
		identities[service.ServiceID] = "service"
	}
	for id, kind := range identities {
		var existing identityRecord
		found, e := tx.Get("identities", id, &existing)
		if e != nil {
			return e
		}
		if found {
			if existing.Owner != owner || existing.Kind != kind {
				return model.Fail("identity_collision", "A resource identity belongs to another connection.")
			}
			continue
		}
		if e = tx.Put("identities", id, identityRecord{Kind: kind, Owner: owner}); e != nil {
			return e
		}
	}
	return nil
}
