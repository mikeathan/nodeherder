package runtime

// Expires setups and operations and prunes retained runtime records inside a transaction.

import (
	"time"

	model "node-herder/models/integrations"
	"node-herder/repository"
)

// Each cleanup phase runs inside the caller's transaction and service lock.
func (s *Service) expireSetups(tx *repository.IntegrationTx) error {
	expiredSetups := []model.Setup{}
	if e := tx.Each("setups", func(_ string, b []byte) error {
		var p model.Setup
		if e := repository.DecodeIntegrationRecord(b, &p); e != nil {
			return e
		}
		if !s.clock.Now().Before(p.ExpiresAt) {
			expiredSetups = append(expiredSetups, p)
		}
		return nil
	}); e != nil {
		return e
	}
	for _, p := range expiredSetups {
		if p.Status == "expired" || p.Status == "completed" {
			if e := tx.Delete("setups", p.ID); e != nil {
				return e
			}
			continue
		}
		if cancel := s.running[p.OperationID]; cancel != nil {
			cancel()
		}
		if p.OperationID != "" {
			if e := tx.Delete("operations", p.OperationID); e != nil {
				return e
			}
		}
		p = model.Setup{
			ID:            p.ID,
			DefinitionKey: p.DefinitionKey,
			Status:        "expired",
			Revision:      p.Revision + 1,
			Generation:    p.Generation + 1,
			Choices:       []model.Choice{},
			ExpiresAt:     s.clock.Now().Add(24 * time.Hour),
		}
		if e := tx.Put("setups", p.ID, p); e != nil {
			return e
		}
	}

	return nil
}

// expireOperations removes expired operations and clears matching instance references.
func (s *Service) expireOperations(tx *repository.IntegrationTx) error {
	expiredOps := []model.Operation{}
	if e := tx.Each("operations", func(_ string, b []byte) error {
		var o model.Operation
		if e := repository.DecodeIntegrationRecord(b, &o); e != nil {
			return e
		}
		if !s.clock.Now().Before(o.ExpiresAt) {
			expiredOps = append(expiredOps, o)
		}
		return nil
	}); e != nil {
		return e
	}
	for _, o := range expiredOps {
		if cancel := s.running[o.ID]; cancel != nil {
			cancel()
		}
		if o.InstanceID != "" {
			var i model.Instance
			ok, e := tx.Get("instances", o.InstanceID, &i)
			if e != nil {
				return e
			}
			if ok && i.OperationID == o.ID {
				i.OperationID = ""
				i.LastError = model.Fail("operation_expired", "The pending lookup expired. Repeat Setup to reconnect this integration.")
				i.Revision++
				if e = tx.Put("instances", i.ID, i); e != nil {
					return e
				}
			}
		}
		if e := tx.Delete("operations", o.ID); e != nil {
			return e
		}
	}

	return nil
}

// pruneRunHistories removes run summaries beyond the retention window.
func (s *Service) pruneRunHistories(tx *repository.IntegrationTx) error {
	histories := []model.Instance{}
	if e := tx.Each("instances", func(_ string, b []byte) error {
		var i model.Instance
		if e := repository.DecodeIntegrationRecord(b, &i); e != nil {
			return e
		}
		pruned := recentRuns(i.Runs, s.clock.Now())
		if len(pruned) != len(i.Runs) {
			i.Runs = pruned
			histories = append(histories, i)
		}
		return nil
	}); e != nil {
		return e
	}
	for _, i := range histories {
		if e := tx.Put("instances", i.ID, i); e != nil {
			return e
		}
	}

	return nil
}

// prunePreviousDefinitions removes expired rollback copies of provider definitions.
func (s *Service) prunePreviousDefinitions(tx *repository.IntegrationTx) error {
	previous := []model.DefinitionRecord{}
	if e := tx.Each("definitions", func(_ string, b []byte) error {
		var d model.DefinitionRecord
		if e := repository.DecodeIntegrationRecord(b, &d); e != nil {
			return e
		}
		if d.PreviousYAML != "" && !s.clock.Now().Before(d.PreviousExpiresAt) {
			d.PreviousYAML = ""
			previous = append(previous, d)
		}
		return nil
	}); e != nil {
		return e
	}
	for _, d := range previous {
		if e := tx.Put("definitions", d.Key, d); e != nil {
			return e
		}
	}
	return nil
}
