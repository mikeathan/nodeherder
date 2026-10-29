package runtime

// Runs admitted operations in workers and coordinates request execution and publication.

import (
	"context"
	"time"

	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

// worker consumes admitted operations until the service context is cancelled.
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
		// A 304 is useful only when the saved snapshot belongs to this exact cache scope.
		// Otherwise retry without the conditional header so old data cannot be reused.
		if runErr == nil && response.Status == 304 && (a.ETag == "" || op.CacheScope == "" || owner.instance.Snapshot == nil || owner.instance.CacheScope != op.CacheScope) {
			a.ETag = ""
			response, runErr = operations.Execute(ctx, s.http, a, s.clock.Now())
			if runErr == nil && response.Status == 304 {
				runErr = model.Fail("invalid_response", "The provider returned an unchanged response without usable cached data.")
			}
		}
		// The request ran without the lock; publication must recheck ownership in the transaction.
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			return
		}
		result := operationPublication{}
		e := s.repo.Update(func(tx *repository.IntegrationTx) error {
			return s.publishResponse(tx, &op, &owner, c, response, runErr, &result)
		})
		if e != nil {
			s.failedPublication(op, e)
		}
		if e == nil && result.deadline != nil {
			s.providerDeadlines[op.Origin] = *result.deadline
		}
		s.mu.Unlock()
		if e != nil {
			return
		}
		if !result.next {
			return
		}
	}
}
