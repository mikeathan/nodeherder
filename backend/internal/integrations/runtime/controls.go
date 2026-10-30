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
			return destinationReviewFailure("Review the changed provider destination in Configure operations.")
		}
		if until := s.providerDeadlines[c.Origin]; s.clock.Now().Before(until) {
			return &model.Failure{Code: "rate_limited", Message: "The provider asked us to wait before another request.", RetryAt: &until}
		}
		o, e := s.admitRefresh(tx, &i, c, "manual")
		if e != nil {
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

// admitRefresh builds and commits one refresh operation for an instance. The caller holds the
// service lock and repository transaction and has already checked lifecycle/revision policy.
func (s *Service) admitRefresh(tx *repository.IntegrationTx, i *model.Instance, c *definitions.Compiled, trigger string) (model.Operation, error) {
	now := s.clock.Now()
	o, e := s.makeOperation(tx, c, i.Inputs, i.State, i.Credentials, i.Origin, c.Definition.Refresh.Request, "refresh", i.Generation)
	if e != nil {
		return o, e
	}
	o.InstanceID = i.ID
	o.Trigger = trigger
	o.CacheScope = definitions.ContentHash([]byte(c.Hash + "\x00" + i.BindingID + "\x00" + c.Definition.Refresh.Request + "\x00" + jsonNumber(o.CredentialRevision) + "\x00" + jsonNumber(i.Generation) + "\x00" + o.RequestFingerprint))
	if i.CacheScope == o.CacheScope && i.Snapshot != nil {
		o.ETag = i.ETag
	}
	i.OperationID = o.ID
	i.LastError = nil
	i.Revision++
	i.UpdatedAt = now
	if i.Schedule != nil {
		// A new attempt supersedes any retry backoff from the previous operation, even while
		// automatic scheduling is off; an interruption must resume at this operation's RetryAt.
		i.Schedule.BackoffUntil = nil
		if i.Schedule.Enabled {
			i.Schedule.LastAttemptAt = &now
			i.Schedule.NextDueAt = s.nextRefreshDue(now, i.Schedule.IntervalSeconds, i.ID)
		}
	}
	if e = tx.Put("operations", o.ID, o); e != nil {
		return o, e
	}
	if e = tx.Put("instances", i.ID, *i); e != nil {
		return o, e
	}
	return o, nil
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
		if e = s.admitContinuation(tx, &o, now); e != nil {
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

// admitContinuation re-queues an existing operation after expiry, retry, provider-backoff and
// ambiguous-POST checks. The caller holds the service lock and repository transaction.
func (s *Service) admitContinuation(tx *repository.IntegrationTx, o *model.Operation, now time.Time) error {
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
	return tx.Put("operations", o.ID, *o)
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
		var resumeAt time.Time
		outstanding := false
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
					outstanding = true
					op.Generation = i.Generation
					op.ETag = ""
					op.CacheScope = ""
					if enabled {
						op.Status = "retry_required"
						resumeAt = op.RetryAt
					} else {
						op.Status = "paused"
					}
					if e = tx.Put("operations", op.ID, op); e != nil {
						return e
					}
				}
			}
			if enabled && i.Schedule != nil && i.Schedule.Enabled && i.Schedule.IntervalSeconds > 0 {
				switch {
				case !outstanding:
					i.Schedule.NextDueAt = s.nextRefreshDue(s.clock.Now(), i.Schedule.IntervalSeconds, i.ID)
				case i.Schedule.NextDueAt.IsZero() && !resumeAt.IsZero():
					i.Schedule.NextDueAt = resumeAt
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

// UpdateSchedule replaces the automatic-refresh choice with revision fencing. Saving a schedule
// performs no provider work; enabling validates the interval against the active definition.
func (s *Service) UpdateSchedule(id string, enabled bool, intervalSeconds int64, expected string) (InstanceDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out InstanceDTO
	if e := s.ready(); e != nil {
		return out, e
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
		minimum, e := s.scheduleMinimum(tx, i, enabled)
		if e != nil {
			return e
		}
		if e = validateScheduleInterval(intervalSeconds, minimum, enabled); e != nil {
			return e
		}
		preserveDue, e := s.hasOutstandingOperation(tx, i)
		if e != nil {
			return e
		}
		now := s.clock.Now()
		s.replaceSchedule(&i, enabled, intervalSeconds, now, preserveDue)
		i.Revision++
		i.UpdatedAt = now
		if e = tx.Put("instances", id, i); e != nil {
			return e
		}
		out = s.instanceDTO(i)
		return nil
	})
	return out, e
}

// scheduleMinimum returns the effective lower bound: the runtime floor or the active provider's
// advisory minimum. Disabling a schedule does not require a readable definition.
func (s *Service) scheduleMinimum(tx *repository.IntegrationTx, i model.Instance, enabled bool) (int64, error) {
	minimum := int64(model.MinimumRefreshSeconds)
	if !enabled {
		return minimum, nil
	}
	r, c, e := s.definition(tx, i.DefinitionKey)
	if e != nil {
		return minimum, e
	}
	if r.DiskError != "" {
		return minimum, model.Fail("definition_unavailable", "Repair or reload the provider configuration first.")
	}
	if c.Definition.Policy.RefreshMinSeconds > int(minimum) {
		minimum = int64(c.Definition.Policy.RefreshMinSeconds)
	}
	return minimum, nil
}

// validateScheduleInterval rejects values that cannot be represented or fall outside the hard and
// provider-advisory bounds. A disabled schedule may keep a zero interval for manual-only use.
func validateScheduleInterval(intervalSeconds, minimum int64, enabled bool) error {
	hardMinimum := int64(model.MinimumRefreshSeconds)
	hardMaximum := int64(model.MaximumRefreshSeconds)
	invalid := (intervalSeconds != 0 && (intervalSeconds < hardMinimum || intervalSeconds > hardMaximum)) ||
		(enabled && intervalSeconds < minimum)
	if !invalid {
		return nil
	}
	return &model.Failure{
		Code:    "invalid_schedule",
		Message: "Choose an automatic refresh interval within the supported range.",
		Fields:  []model.FieldError{{Path: "schedule_interval_seconds", Message: "Use an interval between 60 seconds and 30 days."}},
	}
}

// hasOutstandingOperation reports whether the instance references a non-terminal operation whose
// retry/poll deadline must survive a cadence edit. Terminal references are treated as absent.
func (s *Service) hasOutstandingOperation(tx *repository.IntegrationTx, i model.Instance) (bool, error) {
	if i.OperationID == "" {
		return false, nil
	}
	var o model.Operation
	ok, e := tx.Get("operations", i.OperationID, &o)
	if e != nil {
		return false, e
	}
	return ok && o.Status != "done" && o.Status != "failed", nil
}

// replaceSchedule preserves observation history and, while an operation is outstanding, its existing
// retry/poll deadline. A cadence edit must not postpone a retry past the operation's expiry.
func (s *Service) replaceSchedule(i *model.Instance, enabled bool, intervalSeconds int64, now time.Time, preserveDue bool) {
	previous := i.Schedule
	schedule := model.RefreshSchedule{Enabled: enabled, IntervalSeconds: intervalSeconds}
	if previous != nil {
		schedule.LastAttemptAt = previous.LastAttemptAt
		schedule.LastSuccessAt = previous.LastSuccessAt
		schedule.LastFailureAt = previous.LastFailureAt
		schedule.BackoffUntil = previous.BackoffUntil
		schedule.ConsecutiveFailures = previous.ConsecutiveFailures
	}
	if preserveDue && previous != nil && !previous.NextDueAt.IsZero() {
		schedule.NextDueAt = previous.NextDueAt
	} else if enabled {
		schedule.NextDueAt = s.nextRefreshDue(now, intervalSeconds, i.ID)
	}
	i.Schedule = &schedule
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
