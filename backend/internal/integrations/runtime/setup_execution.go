package runtime

import (
	"node-herder/internal/integrations/definitions"
	"node-herder/internal/integrations/flows/searchselect"
	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

func (s *Service) finishSearch(tx *repository.IntegrationTx, o *model.Operation, owner operationOwner, c *definitions.Compiled, r operations.Response) error {
	choices, e := searchselect.Choices(r.Body, c.Definition.Setup.Search)
	if e != nil {
		return s.failOperation(tx, o, owner, failure(e))
	}
	state, e := operations.Capture(r.Body, c.Definition.Requests[o.RequestID], o.State, true)
	if e != nil {
		return s.failOperation(tx, o, owner, failure(e))
	}
	p := owner.setup
	p.Choices = choices
	p.State = state
	p.Status = "choose"
	p.OperationID = ""
	p.LastError = nil
	s.touch(&p)
	o.Status = "done"
	clearAttempt(o)
	if e = tx.Put("setups", p.ID, p); e != nil {
		return e
	}
	return tx.Put("operations", o.ID, o)
}

// acceptLookupProgress saves only the lookup checkpoint for pending statuses. Completion
// permits normalization; failed/unknown statuses become sanitized durable failures.
func (s *Service) acceptLookupProgress(tx *repository.IntegrationTx, op *model.Operation, owner operationOwner, c *definitions.Compiled, response operations.Response) (bool, error) {
	if op.Kind == "refresh" || owner.test {
		return true, nil
	}

	progress, err := searchselect.Progress(response.Body, c.Definition.Setup.Status)
	if err != nil {
		return false, s.failOperation(tx, op, owner, failure(err))
	}
	switch progress {
	case "pending":
		state, e := operations.Capture(response.Body, c.Definition.Requests[op.RequestID], op.State, false)
		if e != nil {
			return false, s.failOperation(tx, op, owner, failure(e))
		}
		op.State = state
		op.RequestID = c.Definition.Setup.Status.Request
		op.Body = nil
		op.Status = "pending"
		op.RetryAt = s.nextCheck(c.Definition, response.RetryAt)
		op.LastError = nil
		if e := s.savePending(tx, *op, owner); e != nil {
			return false, e
		}
		return false, nil
	case "done":
		return true, nil

	default:
		return false, s.failOperation(tx, op, owner, model.Fail("invalid_response", "The provider returned an unknown lookup status."))
	}
}
