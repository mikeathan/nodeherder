package runtime

import (
	"node-herder/internal/integrations/definitions"
	"time"

	"node-herder/internal/integrations/capabilities"
	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

func (s *Service) renewOperation(tx *repository.IntegrationTx, o *model.Operation, owner operationOwner, c *definitions.Compiled) error {
	i := owner.instance
	now := s.clock.Now()
	if c.Definition.Refresh.RenewalRequest == "" {
		return model.Fail("reconfigure_needed", "This provider does not have a renewal request.")
	}
	if i.LastRenewal != nil && now.Before(i.LastRenewal.Add(time.Duration(c.Definition.Policy.RenewalMinSeconds)*time.Second)) {
		return model.Fail("renewal_not_due", "A new lookup was requested recently. Keep the saved data and try again tomorrow.")
	}
	next, e := s.makeOperation(tx, c, i.Inputs, i.State, i.Credentials, i.Origin, c.Definition.Refresh.RenewalRequest, "renewal", i.Generation)
	if e != nil {
		return e
	}
	next.ID = o.ID
	next.InstanceID = i.ID
	next.ExpiresAt = o.ExpiresAt
	next.Renewed = true
	next.Status = "running"
	i.LastRenewal = &now
	i.OperationID = next.ID
	if e = tx.Put("instances", i.ID, i); e != nil {
		return e
	}
	*o = next
	return tx.Put("operations", o.ID, o)
}

func classifiedFailure(d model.Definition, r operations.Response) *model.Failure {
	if r.Status == 401 || r.Status == 403 || r.Status == 429 || r.Status == 503 {
		return providerFailure(r)
	}
	code, _ := operations.Pointer(r.Body, d.Errors.Code)
	for _, rule := range d.Errors.Rules {
		if rule.Status == r.Status && rule.Code == code {
			switch rule.Action {
			case "restart_setup":
				return model.Fail("restart_setup", "The provider no longer has this lookup. Repeat Setup to reconnect.")
			case "reconnect":
				return model.Fail("reconnect_required", "The provider needs updated authentication. Replace the token in Setup.")
			case "unsupported":
				return model.Fail("unsupported_setup", "The provider cannot retrieve this selection with the supported setup.")
			}
		}
	}
	return providerFailure(r)
}

func (s *Service) terminateOperation(tx *repository.IntegrationTx, id string, f *model.Failure) error {
	if id == "" {
		return nil
	}
	var o model.Operation
	ok, e := tx.Get("operations", id, &o)
	if e != nil || !ok {
		return e
	}
	if cancel := s.running[id]; cancel != nil {
		cancel()
	}
	o.Status = "failed"
	o.LastError = f
	clearAttempt(&o)
	return tx.Put("operations", id, o)
}

// renewAfter asks an optional domain policy, without importing any concrete capability.
func (s *Service) renewAfter(id string, err error) bool {
	handler, resolveErr := s.capabilities.Resolve(id)
	if resolveErr != nil {
		return false
	}
	policy, ok := handler.(capabilities.RenewalPolicy)
	return ok && policy.RenewAfter(err)
}
