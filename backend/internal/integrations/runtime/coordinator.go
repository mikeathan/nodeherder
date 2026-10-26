package runtime

import (
	"context"
	"time"

	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

func (s *Service) worker() {
	defer s.wg.Done()
	for {
		select {
		case <-s.ctx.Done():
			return
		case id := <-s.queue:
			s.execute(id)
		}
	}
}

// execute performs at most the admitted request and one bounded renewal. Network I/O is
// outside the Service lock; every publication rechecks its owner inside the write transaction.
func (s *Service) execute(id string) {
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()
	op, owner, c, accepted := s.claimOperation(id, cancel)
	if !accepted {
		return
	}
	defer func() {
		s.mu.Lock()
		delete(s.running, id)
		s.mu.Unlock()
	}()
	// A renewal is the only extra iteration; pending lookup status never loops/sleeps here.
	for n := 0; n < 2; n++ {
		a := operations.Attempt{
			Definition:     c.Definition,
			RequestID:      op.RequestID,
			Inputs:         op.Inputs,
			State:          op.State,
			ApprovedOrigin: op.Origin,
			Credential:     credential(c.Definition, owner.credentials),
			Body:           op.Body,
			IdempotencyKey: op.IdempotencyKey,
			ETag:           op.ETag,
		}
		response, runErr := operations.Execute(ctx, s.http, a, s.clock.Now())
		if runErr == nil && response.Status == 304 && (a.ETag == "" || op.CacheScope == "" || owner.instance.Snapshot == nil || owner.instance.CacheScope != op.CacheScope) {
			a.ETag = ""
			response, runErr = operations.Execute(ctx, s.http, a, s.clock.Now())
			if runErr == nil && response.Status == 304 {
				runErr = model.Fail("invalid_response", "The provider returned an unchanged response without usable cached data.")
			}
		}
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			return
		}
		next := false
		var deadlineToPublish *time.Time
		e := s.repo.Update(func(tx *repository.IntegrationTx) error {
			current, e := s.owner(tx, op)
			if e != nil {
				return e
			}
			owner = current
			if response.Status == 429 || response.Status >= 500 {
				f := providerFailure(response)
				if f.RetryAt == nil {
					deadline := s.clock.Now().Add(30 * time.Second)
					f.RetryAt = &deadline
				}
				deadlineToPublish = f.RetryAt
				if e = saveProviderLimit(tx, op.Origin, *f.RetryAt); e != nil {
					return e
				}
				return s.failOperation(tx, &op, owner, f)
			}
			if runErr != nil {
				return s.failOperation(tx, &op, owner, failure(runErr))
			}
			if response.Status < 200 || response.Status >= 300 && response.Status != 304 {
				if op.Kind == "refresh" && !op.Renewed && canRenew(c.Definition, response) {
					if e = s.renewOperation(tx, &op, owner, c); e != nil {
						return s.failOperation(tx, &op, owner, failure(e))
					}
					next = true
					return nil
				}
				f := classifiedFailure(c.Definition, response)
				if f.RetryAt != nil {
					deadlineToPublish = f.RetryAt
					if e = saveProviderLimit(tx, op.Origin, *f.RetryAt); e != nil {
						return e
					}
				}
				return s.failOperation(tx, &op, owner, f)
			}
			if op.Kind == "search" {
				return s.finishSearch(tx, &op, owner, c, response)
			}
			if response.Status == 304 {
				return s.finishUnchanged(tx, &op, owner)
			}

			complete, err := s.acceptLookupProgress(tx, &op, owner, c, response)
			if err != nil || !complete {
				return err
			}

			r, e := s.normalizeOperationResult(tx, op, owner, c, response)
			if e != nil {
				if s.renewAfter(c.Definition.Capability, e) && op.Kind == "refresh" && !op.Renewed && owner.instance.Snapshot != nil && c.Definition.Refresh.RenewalRequest != "" {
					if renewalErr := s.renewOperation(tx, &op, owner, c); renewalErr == nil {
						next = true
						return nil
					}
				}
				return s.failOperation(tx, &op, owner, failure(e))
			}
			state, e := operations.Capture(response.Body, c.Definition.Requests[op.RequestID], op.State, true)
			if e != nil {
				return s.failOperation(tx, &op, owner, failure(e))
			}
			if owner.test {
				p := owner.setup
				p.Preview = &r.Resource
				p.Status = "completed"
				p.Revision++
				p.Inputs = nil
				p.State = nil
				p.Credentials = nil
				op.Status = "done"
				clearAttempt(&op)
				if e = tx.Put("setups", p.ID, p); e != nil {
					return e
				}
				return tx.Put("operations", op.ID, op)
			}
			if e = s.finishSnapshot(tx, &op, owner, c, r, state, response.ETag); e != nil {
				return e
			}
			if op.Kind == "refresh" && !op.Renewed && r.Renew && c.Definition.Refresh.RenewalRequest != "" {
				var i model.Instance
				if _, e = tx.Get("instances", op.InstanceID, &i); e != nil {
					return e
				}
				owner.instance = i
				if e = s.renewOperation(tx, &op, owner, c); e == nil {
					next = true
				}
				return nil
			}
			return nil
		})
		if e != nil {
			s.failedPublication(op, e)
		}
		if e == nil && deadlineToPublish != nil {
			s.providerDeadlines[op.Origin] = *deadlineToPublish
		}
		s.mu.Unlock()
		if e != nil {
			return
		}
		if !next {
			return
		}
	}
}
