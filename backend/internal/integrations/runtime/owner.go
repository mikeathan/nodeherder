package runtime

import (
	"context"
	"node-herder/internal/integrations/definitions"

	model "node-herder/models/integrations"
	"node-herder/repository"
)

// owner identifies the generation that can publish an attempt, independently of browser lifetime.
type operationOwner struct {
	instance    model.Instance
	setup       model.Setup
	candidate   bool
	test        bool
	credentials map[string]model.Credential
}

func (s *Service) owner(tx *repository.IntegrationTx, o model.Operation) (operationOwner, error) {
	var owner operationOwner
	if o.Kind == "search" {
		p, e := s.getSetup(tx, o.SetupID)
		if e != nil {
			return owner, e
		}
		if p.Generation != o.Generation || p.OperationID != o.ID {
			return owner, model.Fail("operation_obsolete", "Setup changed.")
		}
		owner.setup = p
		owner.credentials = p.Credentials
		return owner, nil
	}
	if o.SetupID != "" {
		var p model.Setup
		ok, e := tx.Get("setups", o.SetupID, &p)
		if e != nil {
			return owner, e
		}
		if ok && p.TargetInstanceID != "" && p.Status != "completed" {
			if !s.clock.Now().Before(p.ExpiresAt) || p.Generation != o.Generation || p.OperationID != o.ID {
				return owner, model.Fail("operation_obsolete", "Setup changed or expired.")
			}
			var i model.Instance
			ok, e := tx.Get("instances", p.TargetInstanceID, &i)
			if e != nil {
				return owner, e
			}
			if !ok || i.Revision != p.TargetRevision {
				return owner, model.Fail("revision_conflict", "The integration changed during setup.")
			}
			owner.setup = p
			owner.instance = i
			owner.candidate = !p.TestOnly
			owner.test = p.TestOnly
			owner.credentials = p.Credentials
			return owner, nil
		}
	}
	var i model.Instance
	ok, e := tx.Get("instances", o.InstanceID, &i)
	if e != nil {
		return owner, e
	}
	if ok && !i.Enabled {
		return owner, model.Fail("disabled", "Enable the integration before continuing its lookup.")
	}
	if !ok || i.Generation != o.Generation || i.OperationID != o.ID {
		return owner, model.Fail("operation_obsolete", "The integration changed or was disabled.")
	}
	owner.instance = i
	owner.credentials = i.Credentials
	return owner, nil
}

// claimOperation moves queued work to running atomically before any network request. Admission,
// credential/generation checks and registration of its cancel function share the Service lock.
func (s *Service) claimOperation(id string, cancel context.CancelFunc) (op model.Operation, owner operationOwner, c *definitions.Compiled, accepted bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return op, owner, c, false
	}
	e := s.repo.Update(func(tx *repository.IntegrationTx) error {
		ok, e := tx.Get("operations", id, &op)
		if e != nil {
			return e
		}
		if !ok || op.Status != "queued" {
			return model.Fail("operation_obsolete", "Operation no longer queued.")
		}
		owner, e = s.owner(tx, op)
		if e != nil {
			return e
		}
		if !s.clock.Now().Before(op.ExpiresAt) {
			return model.Fail("operation_expired", "This operation expired. Start setup again.")
		}
		c, e = s.compiler.CompileYAML(op.DefinitionYAML)
		if e != nil {
			return e
		}
		c.Revision = op.ProfileRevision
		cred := credential(c.Definition, owner.credentials)
		if cred != nil && cred.Revision != op.CredentialRevision || cred == nil && op.CredentialRevision != 0 {
			return model.Fail("credential_changed", "Credentials changed. Restart this operation.")
		}
		if deadline := s.providerDeadlines[op.Origin]; s.clock.Now().Before(deadline) {
			return s.failOperation(tx, &op, owner, &model.Failure{Code: "rate_limited", Message: "The provider asked us to wait.", RetryAt: &deadline})
		}
		op.Status = "running"
		return tx.Put("operations", id, op)
	})
	if e != nil || op.Status != "running" {
		if e != nil && op.Status == "queued" {
			s.failedPublication(op, e)
		}
		return op, owner, c, false
	}
	s.running[id] = cancel
	return op, owner, c, true
}
