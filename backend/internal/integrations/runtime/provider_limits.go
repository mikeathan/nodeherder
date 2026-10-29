package runtime

// Stores and removes provider-origin backoff deadlines shared by integration instances.

import (
	"time"

	"node-herder/internal/integrations/definitions"
	"node-herder/repository"
)

// providerLimit is the shared persisted backoff for an approved destination, independently of
// an instance's lifetime. Expired entries must be removed so repeated provider edits do not
// leave an ever-growing origin map in memory or encrypted runtime storage.
type providerLimit struct {
	Origin   string    `json:"origin"`
	Deadline time.Time `json:"deadline"`
}

// saveProviderLimit indexes backoff by a hash so the persisted key need not expose the origin.
func saveProviderLimit(tx *repository.IntegrationTx, origin string, deadline time.Time) error {
	return tx.Put("limits", definitions.ContentHash([]byte(origin)), providerLimit{Origin: origin, Deadline: deadline})
}

// activeProviderLimits runs inside a write transaction. Deletions are deferred until iteration
// finishes; the caller publishes the returned map only after that transaction commits.
func activeProviderLimits(tx *repository.IntegrationTx, now time.Time) (map[string]time.Time, error) {
	active := make(map[string]time.Time)
	expired := []string{}
	err := tx.Each("limits", func(id string, data []byte) error {
		var limit providerLimit
		if err := repository.DecodeIntegrationRecord(data, &limit); err != nil {
			return err
		}
		if now.Before(limit.Deadline) {
			active[limit.Origin] = limit.Deadline
		} else {
			expired = append(expired, id)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, id := range expired {
		if err := tx.Delete("limits", id); err != nil {
			return nil, err
		}
	}
	return active, nil
}
