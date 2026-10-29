package runtime

// Allocates and reserves integration resource identities within repository transactions.

import (
	"github.com/google/uuid"

	"node-herder/internal/integrations/capabilities"
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
	return "", model.Fail("identity_collision", "Could not allocate a unique integration identity.")
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

// reserveResourceIdentities records tentative IDs in the publication transaction, preventing
// a collision from overwriting another resource while avoiding reservations on failed normalization.
func reserveResourceIdentities(tx *repository.IntegrationTx, resource model.Resource, reservations []capabilities.Identity, owner string) error {
	identities := make(map[string]string, len(reservations))
	if len(reservations) == 0 || len(reservations) > 1024 {
		return model.Fail("identity_collision", "Invalid resource identity reservations.")
	}
	for _, identity := range reservations {
		if _, err := uuid.Parse(identity.ID); err != nil || identity.Kind == "" {
			return model.Fail("identity_collision", "Invalid resource identity.")
		}
		if _, duplicate := identities[identity.ID]; duplicate {
			return model.Fail("identity_collision", "A resource identity is duplicated.")
		}
		identities[identity.ID] = identity.Kind
	}
	if identities[resource.ResourceID] != "resource" {
		return model.Fail("identity_collision", "The resource identity was not reserved.")
	}
	for id, kind := range identities {
		var existing identityRecord
		found, e := tx.Get("identities", id, &existing)
		if e != nil {
			return e
		}
		if found {
			if existing.Owner != owner || existing.Kind != kind {
				return model.Fail("identity_collision", "A resource identity belongs to another integration.")
			}
			continue
		}
		if e = tx.Put("identities", id, identityRecord{Kind: kind, Owner: owner}); e != nil {
			return e
		}
	}
	return nil
}
