package runtime

// Applies provider responses, failures and renewal decisions to the current operation.

import (
	"time"

	"node-herder/internal/integrations/definitions"
	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

type operationPublication struct {
	next     bool
	deadline *time.Time
}

// publishResponse applies a provider reply inside the caller's transaction and lock.
func (s *Service) publishResponse(tx *repository.IntegrationTx, op *model.Operation, ownerRef *operationOwner, c *definitions.Compiled, response operations.Response, runErr error, result *operationPublication) error {
	current, e := s.owner(tx, *op)
	if e != nil {
		return e
	}
	*ownerRef = current
	owner := current
	if response.Status == 429 || response.Status >= 500 {
		f := providerFailure(response)
		if f.RetryAt == nil {
			deadline := s.clock.Now().Add(30 * time.Second)
			f.RetryAt = &deadline
		}
		result.deadline = f.RetryAt
		if e = saveProviderLimit(tx, op.Origin, *f.RetryAt); e != nil {
			return e
		}
		return s.failOperation(tx, op, owner, f)
	}
	if runErr != nil {
		return s.failOperation(tx, op, owner, failure(runErr))
	}
	if response.Status < 200 || response.Status >= 300 && response.Status != 304 {
		if op.Kind == "refresh" && !op.Renewed && canRenew(c.Definition, response) {
			if e = s.renewOperation(tx, op, owner, c); e != nil {
				return s.failOperation(tx, op, owner, failure(e))
			}
			result.next = true
			return nil
		}
		f := classifiedFailure(c.Definition, response)
		if f.RetryAt != nil {
			result.deadline = f.RetryAt
			if e = saveProviderLimit(tx, op.Origin, *f.RetryAt); e != nil {
				return e
			}
		}
		return s.failOperation(tx, op, owner, f)
	}
	return s.publishSuccessfulResponse(tx, op, ownerRef, c, response, result)
}

// publishSuccessfulResponse applies validated successful replies and optional renewal.
func (s *Service) publishSuccessfulResponse(tx *repository.IntegrationTx, op *model.Operation, ownerRef *operationOwner, c *definitions.Compiled, response operations.Response, result *operationPublication) error {
	owner := *ownerRef
	if op.Kind == "search" {
		return s.finishSearch(tx, op, owner, c, response)
	}
	if response.Status == 304 {
		return s.finishUnchanged(tx, op, owner)
	}

	complete, err := s.acceptLookupProgress(tx, op, owner, c, response)
	if err != nil || !complete {
		return err
	}

	r, e := s.normalizeOperationResult(tx, *op, owner, c, response)
	if e != nil {
		if s.renewAfter(c.Definition.Capability, e) && op.Kind == "refresh" && !op.Renewed && owner.instance.Snapshot != nil && c.Definition.Refresh.RenewalRequest != "" {
			if renewalErr := s.renewOperation(tx, op, owner, c); renewalErr == nil {
				result.next = true
				return nil
			}
		}
		return s.failOperation(tx, op, owner, failure(e))
	}
	state, e := operations.Capture(response.Body, c.Definition.Requests[op.RequestID], op.State, true)
	if e != nil {
		return s.failOperation(tx, op, owner, failure(e))
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
		clearAttempt(op)
		if e = tx.Put("setups", p.ID, p); e != nil {
			return e
		}
		return tx.Put("operations", op.ID, op)
	}
	if e = s.finishSnapshot(tx, op, owner, c, r, state, response.ETag); e != nil {
		return e
	}
	if op.Kind == "refresh" && !op.Renewed && r.Renew && c.Definition.Refresh.RenewalRequest != "" {
		var i model.Instance
		if _, e = tx.Get("instances", op.InstanceID, &i); e != nil {
			return e
		}
		owner.instance = i
		*ownerRef = owner
		if e = s.renewOperation(tx, op, owner, c); e == nil {
			result.next = true
		}
		return nil
	}
	return nil
}
