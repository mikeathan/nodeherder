package integrations

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"slices"
	"time"

	"node-herder/internal/integrations/capabilities"
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
				return owner, model.Fail("revision_conflict", "The connection changed during setup.")
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
		return owner, model.Fail("disabled", "Enable the connection before continuing its lookup.")
	}
	if !ok || i.Generation != o.Generation || i.OperationID != o.ID {
		return owner, model.Fail("operation_obsolete", "The connection changed or was disabled.")
	}
	owner.instance = i
	owner.credentials = i.Credentials
	return owner, nil
}
func (s *Service) execute(id string) {
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	var op model.Operation
	var owner operationOwner
	var c *CompiledDefinition
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
		c, e = CompileYAML(op.DefinitionYAML)
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
		s.mu.Unlock()
		return
	}
	s.running[id] = cancel
	s.mu.Unlock()
	defer func() { s.mu.Lock(); delete(s.running, id); s.mu.Unlock() }()
	for n := 0; n < 2; n++ {
		a := operations.Attempt{Definition: c.Definition, RequestID: op.RequestID, Inputs: op.Inputs, State: op.State, ApprovedOrigin: op.Origin, Credential: credential(c.Definition, owner.credentials), Body: op.Body, IdempotencyKey: op.IdempotencyKey, ETag: op.ETag}
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
		e = s.repo.Update(func(tx *repository.IntegrationTx) error {
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
				if e = tx.Put("limits", ContentHash([]byte(op.Origin)), struct {
					Origin   string    `json:"origin"`
					Deadline time.Time `json:"deadline"`
				}{op.Origin, *f.RetryAt}); e != nil {
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
					if e = tx.Put("limits", ContentHash([]byte(op.Origin)), struct {
						Origin   string    `json:"origin"`
						Deadline time.Time `json:"deadline"`
					}{op.Origin, *f.RetryAt}); e != nil {
						return e
					}
				}
				return s.failOperation(tx, &op, owner, f)
			}
			if op.Kind == "search" {
				return s.finishSearch(tx, &op, owner, c, response)
			}
			if response.Status == 304 {
				if owner.instance.Snapshot == nil {
					return s.failOperation(tx, &op, owner, model.Fail("invalid_response", "No saved collection data exists for this response."))
				}
				i := owner.instance
				i.Snapshot.LastCheckedAt = s.clock.Now().UTC()
				s.recordRun(&i, op, "done", "The schedule is unchanged; source verification time was preserved.")
				i.OperationID = ""
				i.LastError = nil
				i.Revision++
				i.UpdatedAt = s.clock.Now()
				op.Status = "done"
				clearAttempt(&op)
				if e = tx.Put("instances", i.ID, i); e != nil {
					return e
				}
				return tx.Put("operations", op.ID, op)
			}
			complete := op.Kind == "refresh" || owner.test
			if !complete {
				v, ok := Pointer(response.Body, c.Definition.Setup.Status.Pointer)
				status, is := v.(string)
				if !ok || !is {
					return s.failOperation(tx, &op, owner, model.Fail("invalid_response", "The provider omitted its lookup status."))
				}
				switch {
				case slices.Contains(c.Definition.Setup.Status.Pending, status):
					state, e := operations.Capture(response.Body, c.Definition.Requests[op.RequestID], op.State, false)
					if e != nil {
						return s.failOperation(tx, &op, owner, failure(e))
					}
					op.State = state
					op.RequestID = c.Definition.Setup.Status.Request
					op.Body = nil
					op.Status = "pending"
					op.RetryAt = s.nextCheck(c.Definition, response.RetryAt)
					op.LastError = nil
					if e = s.savePending(tx, op, owner); e != nil {
						return e
					}
					return nil
				case slices.Contains(c.Definition.Setup.Status.Done, status):
					complete = true
				case slices.Contains(c.Definition.Setup.Status.Failed, status):
					return s.failOperation(tx, &op, owner, model.Fail("lookup_failed", "The provider could not retrieve this address's collection dates."))
				default:
					return s.failOperation(tx, &op, owner, model.Fail("invalid_response", "The provider returned an unknown lookup status."))
				}
			}
			if complete {
				selected := op.Inputs[c.Definition.Setup.Select.Input]
				prior := owner.instance.Snapshot
				ids := owner.instance.ServiceIDs
				resource := owner.instance.ResourceID
				bindingRevision := owner.instance.BindingRevision
				if owner.test {
					prior = nil
					ids = nil
					resource = ""
				}
				if owner.candidate && owner.instance.Inputs[c.Definition.Setup.Select.Input] != selected {
					prior = nil
					ids = nil
					resource = ""
					bindingRevision++
				}
				r, e := capabilities.NormalizeWaste(response.Root, c.Definition.Mapping, capabilities.WasteContext{NewIdentity: snapshotIdentityAllocator(tx), InstanceID: op.InstanceID, ResourceID: resource, ProviderID: c.Definition.ID, ProviderVersion: "sha256:" + op.DefinitionHash, ProfileRevision: op.ProfileRevision, BindingRevision: bindingRevision, SelectedProperty: selected, Now: s.clock.Now(), Previous: prior, ServiceIDs: ids})
				if e != nil {
					if failure(e).Code == "no_collection_dates" && op.Kind == "refresh" && !op.Renewed && owner.instance.Snapshot != nil && c.Definition.Refresh.RenewalRequest != "" {
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
					p.Preview = &r.Snapshot
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
				if op.Kind == "refresh" && !op.Renewed && horizonExhausted(r.Snapshot, s.clock.Now()) && c.Definition.Refresh.RenewalRequest != "" {
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
func failure(e error) *model.Failure {
	var f *model.Failure
	if errors.As(e, &f) {
		return f
	}
	return model.Fail("storage_error", "The connection change could not be saved. Check persistent storage and try again.")
}
func providerFailure(r operations.Response) *model.Failure {
	if r.Status >= 500 {
		return &model.Failure{Code: "provider_unavailable", Message: "The provider is temporarily unavailable.", RetryAt: r.RetryAt}
	}
	f := model.Fail("provider_error", "The provider rejected the request. Check Setup or Configure operations.")
	switch r.Status {
	case 401, 403:
		f = model.Fail("reconnect_required", "The provider did not accept the credential. Replace it in Setup.")
	case 429:
		f = model.Fail("rate_limited", "The provider's request limit was reached. Try again after the indicated time.")
		f.RetryAt = r.RetryAt
	case 503:
		f = model.Fail("provider_unavailable", "The provider is temporarily unavailable.")
		f.RetryAt = r.RetryAt
	case 404:
		f = model.Fail("provider_not_found", "The provider did not find the configured request or saved lookup. Check the configuration.")
	}
	return f
}
func canRenew(d model.Definition, r operations.Response) bool {
	code, _ := Pointer(r.Body, d.Errors.Code)
	for _, rule := range d.Errors.Rules {
		if rule.Status == r.Status && rule.Code == code && rule.Action == "renewal" {
			return true
		}
	}
	return false
}
func (s *Service) nextCheck(d model.Definition, retry *time.Time) time.Time {
	t := s.clock.Now().Add(time.Duration(d.Policy.PollMinSeconds) * time.Second)
	if retry != nil && retry.After(t) {
		t = *retry
	}
	return t
}
func clearAttempt(o *model.Operation) {
	o.Body = nil
	o.Inputs = nil
	o.State = nil
	o.IdempotencyKey = ""
	o.ETag = ""
	o.DefinitionYAML = ""
}
func (s *Service) failOperation(tx *repository.IntegrationTx, o *model.Operation, owner operationOwner, f *model.Failure) error {
	o.LastError = f
	o.Status = "failed"
	now := s.clock.Now()
	if f.Code == "provider_unavailable" || f.Code == "rate_limited" || f.Code == "invalid_response" {
		o.Status = "retry_required"
		o.RetryAt = now.Add(5 * time.Second)
		if f.RetryAt != nil && f.RetryAt.After(o.RetryAt) {
			o.RetryAt = *f.RetryAt
		}
	}
	if o.Kind == "search" || owner.candidate || owner.test {
		p := owner.setup
		p.LastError = f
		p.Status = "error"
		p.Revision++
		if e := tx.Put("setups", p.ID, p); e != nil {
			return e
		}
	} else {
		i := owner.instance
		i.LastError = f
		s.recordRun(&i, *o, o.Status, f.Message)
		i.Revision++
		i.UpdatedAt = now
		if o.Status == "failed" {
			i.OperationID = ""
		}
		if e := tx.Put("instances", i.ID, i); e != nil {
			return e
		}
	}
	if o.Status == "failed" {
		clearAttempt(o)
	}
	return tx.Put("operations", o.ID, o)
}
func (s *Service) savePending(tx *repository.IntegrationTx, o model.Operation, owner operationOwner) error {
	if owner.candidate {
		p := owner.setup
		p.State = o.State
		p.Status = "pending"
		p.Revision++
		if e := tx.Put("setups", p.ID, p); e != nil {
			return e
		}
	} else {
		i := owner.instance
		i.State = o.State
		i.LastError = nil
		i.Revision++
		i.UpdatedAt = s.clock.Now()
		if e := tx.Put("instances", i.ID, i); e != nil {
			return e
		}
	}
	return tx.Put("operations", o.ID, o)
}
func (s *Service) finishSearch(tx *repository.IntegrationTx, o *model.Operation, owner operationOwner, c *CompiledDefinition, r operations.Response) error {
	step := c.Definition.Setup.Search
	if step.Expect != nil {
		v, ok := Pointer(r.Body, step.Expect.Pointer)
		if !ok || v != step.Expect.Value {
			return s.failOperation(tx, o, owner, model.Fail("unsupported_setup", "This provider needs a different type of address input for that postcode. This release supports address selection only."))
		}
	}
	v, ok := Pointer(r.Body, step.Options.Items)
	list, is := v.([]any)
	if !ok || !is || len(list) > 2000 {
		return s.failOperation(tx, o, owner, model.Fail("invalid_response", "The provider returned an invalid address list."))
	}
	choices := make([]model.Choice, 0, len(list))
	seen := map[string]bool{}
	for _, item := range list {
		id, _ := Pointer(item, step.Options.Value)
		label, _ := Pointer(item, step.Options.Label)
		idText, idOK := id.(string)
		labelText, labelOK := label.(string)
		if !idOK || !labelOK || idText == "" || labelText == "" || len(idText) > 2048 || len(labelText) > 512 || seen[idText] {
			return s.failOperation(tx, o, owner, model.Fail("invalid_response", "The provider returned invalid or duplicate addresses."))
		}
		seen[idText] = true
		choices = append(choices, model.Choice{ID: idText, Label: labelText})
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
func (s *Service) finishSnapshot(tx *repository.IntegrationTx, o *model.Operation, owner operationOwner, c *CompiledDefinition, result capabilities.WasteResult, state map[string]any, etag string) error {
	i := owner.instance
	if owner.candidate {
		if i.Inputs[c.Definition.Setup.Select.Input] != o.Inputs[c.Definition.Setup.Select.Input] {
			i.BindingID = uuid.NewString()
			i.BindingRevision++
		}
		i.SearchInput = c.Definition.Setup.Search.Input
		i.Inputs = o.Inputs
		if owner.setup.ConnectionName != "" {
			i.Name = owner.setup.ConnectionName
		}
		i.Credentials = owner.setup.Credentials
		i.Origin = c.Origin
		i.Generation++
		p := owner.setup
		p.InstanceID = i.ID
		p.Status = "completed"
		p.Inputs = nil
		p.State = nil
		p.Choices = []model.Choice{}
		p.Credentials = nil
		p.DefinitionYAML = ""
		p.ConnectionName = ""
		p.ExpiresAt = s.clock.Now().Add(24 * time.Hour)
		p.Revision++
		if e := tx.Put("setups", p.ID, p); e != nil {
			return e
		}
	}
	if e := reserveSnapshotIdentities(tx, result.Snapshot, i.ID); e != nil {
		return e
	}
	s.recordRun(&i, *o, "done", "Collection dates saved.")
	i.ResourceID = result.Snapshot.ResourceID
	i.Snapshot = &result.Snapshot
	i.ServiceIDs = result.ServiceIDs
	i.State = state
	i.LastError = nil
	i.OperationID = ""
	i.Revision++
	i.UpdatedAt = s.clock.Now()
	if o.Kind == "refresh" {
		i.ETag = etag
		i.CacheScope = o.CacheScope
	} else {
		i.ETag = ""
		i.CacheScope = ""
	}
	o.Status = "done"
	clearAttempt(o)
	if e := tx.Put("instances", i.ID, i); e != nil {
		return e
	}
	return tx.Put("operations", o.ID, o)
}
func horizonExhausted(snapshot model.WasteSchedule, now time.Time) bool {
	location, e := time.LoadLocation(snapshot.Timezone)
	if e != nil {
		return false
	}
	today := now.In(location).Format("2006-01-02")
	for _, service := range snapshot.Services {
		if service.DateState == "known" && len(service.Dates) > 0 && service.Dates[len(service.Dates)-1] < today {
			return true
		}
	}
	return false
}
func (s *Service) renewOperation(tx *repository.IntegrationTx, o *model.Operation, owner operationOwner, c *CompiledDefinition) error {
	i := owner.instance
	now := s.clock.Now()
	if c.Definition.Refresh.RenewalRequest == "" {
		return model.Fail("reconfigure_needed", "This provider does not have a renewal request.")
	}
	if i.LastRenewal != nil && now.Before(i.LastRenewal.Add(time.Duration(c.Definition.Policy.RenewalMinSeconds)*time.Second)) {
		return model.Fail("renewal_not_due", "A new lookup was requested recently. Keep the existing dates and try again tomorrow.")
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
	code, _ := Pointer(r.Body, d.Errors.Code)
	for _, rule := range d.Errors.Rules {
		if rule.Status == r.Status && rule.Code == code {
			switch rule.Action {
			case "restart_setup":
				return model.Fail("restart_setup", "The provider no longer has this lookup. Set up the address again.")
			case "reconnect":
				return model.Fail("reconnect_required", "The provider needs updated authentication. Replace the token in Setup.")
			case "unsupported":
				return model.Fail("unsupported_setup", "The provider cannot retrieve this address with the supported setup.")
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
