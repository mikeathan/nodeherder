package runtime

// Handles manual refresh, status checks, instance changes and credential updates.

import (
	"encoding/json"
	"strings"
	"time"

	"node-herder/internal/integrations/definitions"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

// Refresh durably admits a manual schedule read, coalescing any existing attempt.
func (s *Service) Refresh(id, expected string) (OperationDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out OperationDTO
	if e := s.ready(); e != nil {
		return out, e
	}
	if e := s.hasCapacity(); e != nil {
		return out, e
	}
	var admitted string
	e := s.repo.Update(func(tx *repository.IntegrationTx) error {
		var i model.Instance
		ok, e := tx.Get("instances", id, &i)
		if e != nil {
			return e
		}
		if !ok {
			return model.Fail("not_found", "Integration not found.")
		}
		if e = checkRevision(expected, revisionETag("instance", id, i.Revision)); e != nil {
			return e
		}
		if !i.Enabled {
			return model.Fail("disabled", "Enable the integration before refreshing.")
		}
		if i.OperationID != "" {
			var op model.Operation
			ok, e := tx.Get("operations", i.OperationID, &op)
			if e != nil {
				return e
			}
			if ok && op.Status != "done" && op.Status != "failed" {
				out = operationDTO(op)
				return nil
			}
		}
		r, c, e := s.definition(tx, i.DefinitionKey)
		if e != nil {
			return e
		}
		if r.DiskError != "" {
			return model.Fail("definition_unavailable", "Repair or reload the provider configuration first.")
		}
		if i.Origin != c.Origin {
			return model.Fail("reconnect_required", "Review the changed provider destination in Configure operations.")
		}
		if until := s.providerDeadlines[c.Origin]; s.clock.Now().Before(until) {
			return &model.Failure{Code: "rate_limited", Message: "The provider asked us to wait before another request.", RetryAt: &until}
		}
		o, e := s.makeOperation(tx, c, i.Inputs, i.State, i.Credentials, i.Origin, c.Definition.Refresh.Request, "refresh", i.Generation)
		if e != nil {
			return e
		}
		o.InstanceID = i.ID
		o.CacheScope = definitions.ContentHash([]byte(c.Hash + "\x00" + i.BindingID + "\x00" + c.Definition.Refresh.Request + "\x00" + jsonNumber(o.CredentialRevision) + "\x00" + jsonNumber(i.Generation) + "\x00" + o.RequestFingerprint))
		if i.CacheScope == o.CacheScope && i.Snapshot != nil {
			o.ETag = i.ETag
		}
		i.OperationID = o.ID
		i.LastError = nil
		i.Revision++
		i.UpdatedAt = s.clock.Now()
		if e = tx.Put("operations", o.ID, o); e != nil {
			return e
		}
		if e = tx.Put("instances", i.ID, i); e != nil {
			return e
		}
		out = operationDTO(o)
		admitted = o.ID
		return nil
	})
	if e == nil && admitted != "" {
		s.enqueue(admitted)
	}
	return out, e
}
func jsonNumber(n uint64) string { b, _ := json.Marshal(n); return string(b) }

// CheckOperation is the only continuation trigger in M1. It never allocates a replacement POST key.
func (s *Service) CheckOperation(id, expected string) (OperationDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out OperationDTO
	if e := s.ready(); e != nil {
		return out, e
	}
	if e := s.hasCapacity(); e != nil {
		return out, e
	}
	admit := false
	e := s.repo.Update(func(tx *repository.IntegrationTx) error {
		var o model.Operation
		ok, e := tx.Get("operations", id, &o)
		if e != nil {
			return e
		}
		if !ok {
			return model.Fail("not_found", "Operation not found.")
		}
		if o.Status == "done" || o.Status == "failed" {
			out = operationDTO(o)
			return nil
		}
		owner, e := s.owner(tx, o)
		if e != nil {
			return e
		}
		actual := revisionETag("instance", owner.instance.ID, owner.instance.Revision)
		if o.Kind == "search" || owner.candidate || owner.test {
			actual = revisionETag("setup", owner.setup.ID, owner.setup.Revision)
		}
		if e = checkRevision(expected, actual); e != nil {
			return e
		}
		if s.running[id] != nil || o.Status == "queued" {
			out = operationDTO(o)
			return nil
		}
		now := s.clock.Now()
		if !now.Before(o.ExpiresAt) {
			return model.Fail("operation_expired", "The saved lookup expired. Repeat Setup to reconnect.")
		}
		if now.Before(o.RetryAt) {
			return &model.Failure{Code: "retry_later", Message: "Wait before checking this lookup again.", RetryAt: &o.RetryAt}
		}
		c, e := s.compiler.CompileYAML(o.DefinitionYAML)
		if e != nil {
			return e
		}
		if deadline := s.providerDeadlines[o.Origin]; now.Before(deadline) {
			return &model.Failure{Code: "rate_limited", Message: "The provider asked us to wait.", RetryAt: &deadline}
		}
		q := c.Definition.Requests[o.RequestID]
		if q.Method == "POST" && (c.Definition.Policy.IdempotencyWindowSeconds == 0 || !bool(q.Idempotency) || o.IdempotencyKey == "" || now.After(o.CreatedAt.Add(time.Duration(c.Definition.Policy.IdempotencyWindowSeconds)*time.Second))) {
			return model.Fail("reconciliation_required", "The previous submission cannot be retried safely. Start a reviewed setup again.")
		}
		o.Status = "queued"
		o.LastError = nil
		if e = tx.Put("operations", o.ID, o); e != nil {
			return e
		}
		out = operationDTO(o)
		admit = true
		return nil
	})
	if e == nil && admit {
		s.enqueue(id)
	}
	return out, e
}

// UpdateInstance changes name/enabled state with generation fencing; enable never starts I/O.
func (s *Service) UpdateInstance(id, name string, enabled bool, expected string) (InstanceDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out InstanceDTO
	if e := s.ready(); e != nil {
		return out, e
	}
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 120 {
		return out, model.Fail("invalid_input", "Enter a name of at most 120 characters.")
	}
	e := s.repo.Update(func(tx *repository.IntegrationTx) error {
		var i model.Instance
		ok, e := tx.Get("instances", id, &i)
		if e != nil {
			return e
		}
		if !ok {
			return model.Fail("not_found", "Integration not found.")
		}
		if e = checkRevision(expected, revisionETag("instance", id, i.Revision)); e != nil {
			return e
		}
		if i.Enabled != enabled {
			i.Generation++
			i.ETag = ""
			i.CacheScope = ""
			if cancel := s.running[i.OperationID]; cancel != nil {
				cancel()
			}
			if i.OperationID != "" {
				var op model.Operation
				ok, e := tx.Get("operations", i.OperationID, &op)
				if e != nil {
					return e
				}
				if ok && op.Status != "done" && op.Status != "failed" {
					op.Generation = i.Generation
					op.ETag = ""
					op.CacheScope = ""
					if enabled {
						op.Status = "retry_required"
					} else {
						op.Status = "paused"
					}
					if e = tx.Put("operations", op.ID, op); e != nil {
						return e
					}
				}
			}
		}
		i.Name = name
		i.Enabled = enabled
		i.Revision++
		i.UpdatedAt = s.clock.Now()
		if e = tx.Put("instances", id, i); e != nil {
			return e
		}
		out = s.instanceDTO(i)
		return nil
	})
	return out, e
}

// DeleteInstance removes the binding, snapshots, attempts and identities atomically.
func (s *Service) DeleteInstance(id, expected string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e := s.ready(); e != nil {
		return e
	}
	return s.repo.Update(func(tx *repository.IntegrationTx) error {
		var i model.Instance
		ok, e := tx.Get("instances", id, &i)
		if e != nil {
			return e
		}
		if !ok {
			return model.Fail("not_found", "Integration not found.")
		}
		if e = checkRevision(expected, revisionETag("instance", id, i.Revision)); e != nil {
			return e
		}
		if cancel := s.running[i.OperationID]; cancel != nil {
			cancel()
		}
		if e = tx.Delete("instances", id); e != nil {
			return e
		}
		for _, bucket := range []string{"setups", "operations", "identities"} {
			keys := []string{}
			if e = tx.Each(bucket, func(key string, b []byte) error {
				if bucket == "identities" {
					var identity identityRecord
					if e := repository.DecodeIntegrationRecord(b, &identity); e != nil {
						return e
					}
					if identity.Owner == id {
						keys = append(keys, key)
					}
				} else if bucket == "operations" {
					var op model.Operation
					if e := repository.DecodeIntegrationRecord(b, &op); e != nil {
						return e
					}
					if op.InstanceID == id {
						keys = append(keys, key)
					}
				} else {
					var p model.Setup
					if e := repository.DecodeIntegrationRecord(b, &p); e != nil {
						return e
					}
					if p.InstanceID == id || p.TargetInstanceID == id {
						keys = append(keys, key)
					}
				}
				return nil
			}); e != nil {
				return e
			}
			for _, key := range keys {
				if cancel := s.running[key]; cancel != nil {
					cancel()
				}
				if e = tx.Delete(bucket, key); e != nil {
					return e
				}
			}
		}
		return nil
	})
}

// SetCredential writes or removes a private slot and returns only a redacted setup/instance DTO.
func (s *Service) SetCredential(kind, id, slot, value, expected string, remove bool) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e := s.ready(); e != nil {
		return nil, e
	}
	if !remove && (value == "" || len(value) > 8192 || strings.ContainsAny(value, "\r\n")) {
		return nil, model.Fail("invalid_credential", "Enter a token of at most 8 KiB without line breaks.")
	}
	var out any
	e := s.repo.Update(func(tx *repository.IntegrationTx) error {
		if kind == "setup" {
			setup, err := s.setSetupCredential(tx, id, slot, value, expected)
			if err != nil {
				return err
			}
			out = setup
			return nil
		}
		var i model.Instance
		ok, e := tx.Get("instances", id, &i)
		if e != nil {
			return e
		}
		if !ok {
			return model.Fail("not_found", "Integration not found.")
		}
		if e = checkRevision(expected, revisionETag("instance", id, i.Revision)); e != nil {
			return e
		}
		_, c, e := s.definition(tx, i.DefinitionKey)
		if e != nil {
			return e
		}
		_, storedSlot := i.Credentials[slot]
		if (!remove && (c.Definition.Auth.Mode == "none" || slot != c.Definition.Auth.SecretSlot)) || (remove && !storedSlot) {
			return model.Fail("invalid_credential", "This provider does not declare that credential slot.")
		}
		old := i.Credentials[slot]
		if remove {
			delete(i.Credentials, slot)
		} else {
			i.Credentials[slot] = model.Credential{Value: value, Origin: c.Origin, Revision: old.Revision + 1}
		}
		if cancel := s.running[i.OperationID]; cancel != nil {
			cancel()
		}
		if e = s.terminateOperation(tx, i.OperationID, model.Fail("credential_changed", "The credential changed. Start a new refresh or setup.")); e != nil {
			return e
		}
		i.Generation++
		i.Revision++
		i.OperationID = ""
		i.ETag = ""
		i.CacheScope = ""
		i.UpdatedAt = s.clock.Now()
		i.LastError = nil
		if remove && c.Definition.Auth.RequiresCredential() {
			i.LastError = model.Fail("reconnect_required", "This provider requires a credential.")
		}
		if e = tx.Put("instances", id, i); e != nil {
			return e
		}
		out = s.instanceDTO(i)
		return nil
	})
	return out, e
}

// setSetupCredential updates a setup credential and invalidates choices tied to the previous value.
func (s *Service) setSetupCredential(tx *repository.IntegrationTx, id, slot, value, expected string) (SetupDTO, error) {
	p, err := s.getSetup(tx, id)
	if err != nil {
		return SetupDTO{}, err
	}
	if p.Status == "completed" {
		return SetupDTO{}, model.Fail("revision_conflict", "This setup is already saved.")
	}
	if err = checkRevision(expected, revisionETag("setup", id, p.Revision)); err != nil {
		return SetupDTO{}, err
	}
	c, err := s.compiler.CompileYAML(p.DefinitionYAML)
	if err != nil {
		return SetupDTO{}, err
	}
	if c.Definition.Auth.Mode == "none" || slot != c.Definition.Auth.SecretSlot {
		return SetupDTO{}, model.Fail("invalid_credential", "This provider does not declare that credential slot.")
	}
	old := p.Credentials[slot]
	if err = s.invalidateSetup(tx, &p); err != nil {
		return SetupDTO{}, err
	}
	p.Credentials[slot] = model.Credential{Value: value, Origin: c.Origin, Revision: old.Revision + 1}
	p.Status = "input"
	p.Choices = []model.Choice{}
	p.State = map[string]any{}
	s.touch(&p)
	if err = tx.Put("setups", id, p); err != nil {
		return SetupDTO{}, err
	}
	return s.setupDTO(p), nil
}
