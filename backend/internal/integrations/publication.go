package integrations

import (
	model "node-herder/models/integrations"
	"node-herder/repository"
	"node-herder/utils"
	"time"
)

type publicationFailure struct {
	Generation uint64
	Error      *model.Failure
}

// failedPublication never publishes uncommitted provider data. If even a small failure checkpoint
// cannot be written, a bounded volatile diagnostic explains the failure until retry/restart.
func (s *Service) failedPublication(attempt model.Operation, cause error) {
	if attempt.ID == "" {
		return
	}
	f := failure(cause)
	err := s.repo.Update(func(tx *repository.IntegrationTx) error {
		var current model.Operation
		ok, e := tx.Get("operations", attempt.ID, &current)
		if e != nil || !ok {
			return e
		}
		if current.Generation != attempt.Generation || current.Status != "running" && current.Status != "queued" {
			return nil
		}
		current.LastError = f
		current.Status = "failed"
		if f.Code == "storage_error" {
			current.Status = "retry_required"
			current.RetryAt = s.clock.Now().Add(5 * time.Second)
		} else {
			clearAttempt(&current)
		}
		var p model.Setup
		hasSetup := false
		if current.SetupID != "" {
			hasSetup, e = tx.Get("setups", current.SetupID, &p)
			if e != nil {
				return e
			}
		}
		if hasSetup && p.Status != "completed" && p.OperationID == current.ID && p.Generation == current.Generation {
			p.LastError = f
			p.Status = "error"
			p.Revision++
			if e = tx.Put("setups", p.ID, p); e != nil {
				return e
			}
		} else if current.InstanceID != "" {
			var i model.Instance
			ok, e := tx.Get("instances", current.InstanceID, &i)
			if e != nil {
				return e
			}
			if ok && i.OperationID == current.ID && i.Generation == current.Generation {
				i.LastError = f
				i.Revision++
				if current.Status == "failed" {
					i.OperationID = ""
				}
				s.recordRun(&i, current, current.Status, f.Message)
				if e = tx.Put("instances", i.ID, i); e != nil {
					return e
				}
			}
		}
		return tx.Put("operations", current.ID, current)
	})
	if err != nil {
		if s.publicationFailures == nil {
			s.publicationFailures = map[string]publicationFailure{}
		}
		if len(s.publicationFailures) >= 128 {
			for id := range s.publicationFailures {
				delete(s.publicationFailures, id)
				break
			}
		}
		s.publicationFailures[attempt.ID] = publicationFailure{Generation: attempt.Generation, Error: model.Fail("storage_error", "The provider result could not be saved. Previous data was kept. Check persistent storage, then use Check status to retry the saved attempt.")}
	}
	utils.LogWarnf("Connection operation %s failed to publish: %s", attempt.ID, f.Code)
}
