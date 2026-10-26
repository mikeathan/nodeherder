package runtime

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"node-herder/internal/integrations/definitions"
	"node-herder/internal/integrations/flows/searchselect"
	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

// CreateSetup starts local shared state. It sends no provider request.
func (s *Service) CreateSetup(key, targetID, targetETag string) (SetupDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out SetupDTO
	if e := s.ready(); e != nil {
		return out, e
	}
	now := s.clock.Now()
	e := s.repo.Update(func(tx *repository.IntegrationTx) error {
		count := 0
		if e := tx.Each("setups", func(_ string, b []byte) error {
			var p model.Setup
			if e := repository.DecodeIntegrationRecord(b, &p); e != nil {
				return e
			}
			if p.Status != "completed" && p.Status != "expired" && now.Before(p.ExpiresAt) {
				count++
			}
			return nil
		}); e != nil {
			return e
		}
		if count >= 32 {
			return model.Fail("busy", "Finish or cancel another setup before starting one.")
		}
		r, c, e := s.definition(tx, key)
		if e != nil {
			return e
		}
		if r.DiskError != "" {
			return model.Fail("definition_unavailable", "Repair or reload this provider before setup.")
		}
		setupID, e := allocateRecordID(tx, "setups", uuid.NewRandom)
		if e != nil {
			return e
		}
		p := model.Setup{
			ID:                 setupID,
			DefinitionKey:      key,
			DefinitionYAML:     c.YAML,
			DefinitionRevision: r.Revision,
			Revision:           1,
			Generation:         1,
			Inputs:             map[string]string{},
			State:              map[string]any{},
			Choices:            []model.Choice{},
			Credentials:        map[string]model.Credential{},
			Status:             "input",
			CreatedAt:          now,
			ExpiresAt:          now.Add(30 * time.Minute),
			AbsoluteExpiry:     now.Add(24 * time.Hour),
		}
		if targetID != "" {
			var i model.Instance
			ok, e := tx.Get("instances", targetID, &i)
			if e != nil {
				return e
			}
			if !ok {
				return model.Fail("not_found", "Integration not found.")
			}
			if i.DefinitionKey != key {
				return model.Fail("invalid_input", "Use the existing provider to change the selection.")
			}
			if e = checkRevision(targetETag, revisionETag("instance", i.ID, i.Revision)); e != nil {
				return e
			}
			p.TargetInstanceID = i.ID
			p.TargetRevision = i.Revision
			p.Credentials = i.Credentials
		}
		if e = tx.Put("setups", p.ID, p); e != nil {
			return e
		}
		out = s.setupDTO(p)
		return nil
	})
	return out, e
}
func (s *Service) getSetup(tx *repository.IntegrationTx, id string) (model.Setup, error) {
	var p model.Setup
	ok, e := tx.Get("setups", id, &p)
	if e != nil {
		return p, e
	}
	if !ok {
		return p, model.Fail("not_found", "Setup not found.")
	}
	if p.Status == "expired" || !s.clock.Now().Before(p.ExpiresAt) {
		return p, model.Fail("setup_expired", "This setup expired. Start again to retrieve current results.")
	}
	return p, nil
}

// Setup reads shared progress or a completion receipt and enforces expiry.
func (s *Service) Setup(id string) (SetupDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out SetupDTO
	if e := s.ready(); e != nil {
		return out, e
	}
	e := s.repo.View(func(tx *repository.IntegrationTx) error {
		p, e := s.getSetup(tx, id)
		if e == nil {
			out = s.setupDTO(p)
			if diagnostic, ok := s.publicationFailures[p.OperationID]; ok && diagnostic.Generation == p.Generation && p.Status != "completed" {
				out.Error = diagnostic.Error
				out.Status = "error"
			}
		}
		return e
	})
	return out, e
}
func (s *Service) touch(p *model.Setup) {
	p.Revision++
	p.ExpiresAt = s.clock.Now().Add(30 * time.Minute)
	if p.ExpiresAt.After(p.AbsoluteExpiry) {
		p.ExpiresAt = p.AbsoluteExpiry
	}
}

// SearchSetup validates the configured search input and durably admits its named request.
func (s *Service) SearchSetup(id, query, expected string) (SetupDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out SetupDTO
	if e := s.ready(); e != nil {
		return out, e
	}
	if e := s.hasCapacity(); e != nil {
		return out, e
	}
	var admitted string
	e := s.repo.Update(func(tx *repository.IntegrationTx) error {
		p, e := s.getSetup(tx, id)
		if e != nil {
			return e
		}
		if e = checkRevision(expected, revisionETag("setup", id, p.Revision)); e != nil {
			return e
		}
		if p.InstanceID != "" {
			return model.Fail("revision_conflict", "This setup was already saved.")
		}
		c, e := s.compiler.CompileYAML(p.DefinitionYAML)
		if e != nil {
			return e
		}
		c.Revision = p.DefinitionRevision
		if e = s.invalidateSetup(tx, &p); e != nil {
			return e
		}
		normalized, err := searchselect.NormalizeInput(query, c.Definition.Inputs[c.Definition.Setup.Search.Input])
		if err != nil {
			return err
		}
		p.Inputs = map[string]string{c.Definition.Setup.Search.Input: normalized}
		p.State = map[string]any{}
		p.Choices = []model.Choice{}
		p.Status = "searching"
		p.SubmissionHash = ""
		p.ConnectionName = ""
		p.LastError = nil
		s.touch(&p)
		o, e := s.makeOperation(tx, c, p.Inputs, p.State, p.Credentials, c.Origin, c.Definition.Setup.Search.Request, "search", p.Generation)
		if e != nil {
			return e
		}
		o.SetupID = p.ID
		p.OperationID = o.ID
		if e = tx.Put("operations", o.ID, o); e != nil {
			return e
		}
		if e = tx.Put("setups", p.ID, p); e != nil {
			return e
		}
		out = s.setupDTO(p)
		admitted = o.ID
		return nil
	})
	if e == nil {
		s.enqueue(admitted)
	}
	return out, e
}
func (s *Service) invalidateSetup(tx *repository.IntegrationTx, p *model.Setup) error {
	p.Generation++
	if cancel := s.running[p.OperationID]; cancel != nil {
		cancel()
	}
	if p.OperationID != "" {
		if e := tx.Delete("operations", p.OperationID); e != nil {
			return e
		}
	}
	p.OperationID = ""
	return nil
}

// makeOperation pins all execution inputs and constructs exact POST bytes before admission.
// The caller must commit the returned operation with its owner before enqueueing; a retry
// must not rebuild a changed body under the same idempotency key.
func (s *Service) makeOperation(tx *repository.IntegrationTx, c *definitions.Compiled, inputs map[string]string, state map[string]any, credentials map[string]model.Credential, origin, requestID, kind string, generation uint64) (model.Operation, error) {
	now := s.clock.Now()
	if kind != "renewal" {
		count := 0
		if e := tx.Each("operations", func(_ string, b []byte) error {
			var existing model.Operation
			if e := repository.DecodeIntegrationRecord(b, &existing); e != nil {
				return e
			}
			if existing.Status != "done" && existing.Status != "failed" && now.Before(existing.ExpiresAt) {
				count++
			}
			return nil
		}); e != nil {
			return model.Operation{}, e
		}
		if count >= 64 {
			return model.Operation{}, model.Fail("busy", "Too many pending operations. Complete or cancel an existing lookup first.")
		}
	}
	id, err := allocateRecordID(tx, "operations", uuid.NewRandom)
	if err != nil {
		return model.Operation{}, err
	}
	o := model.Operation{
		ID:              id,
		Generation:      generation,
		ProfileRevision: c.Revision,
		DefinitionYAML:  c.YAML,
		DefinitionHash:  c.Hash,
		Inputs:          inputs,
		Origin:          origin,
		RequestID:       requestID,
		Kind:            kind,
		Status:          "queued",
		State:           state,
		CreatedAt:       now,
		ExpiresAt:       now.Add(24 * time.Hour),
		RetryAt:         now,
	}
	q := c.Definition.Requests[requestID]
	if q.Idempotency {
		o.IdempotencyKey = uuid.NewString()
	}
	cred := credential(c.Definition, credentials)
	if cred != nil {
		o.CredentialRevision = cred.Revision
	}
	request, body, e := operations.Build(context.Background(), operations.Attempt{
		Definition:     c.Definition,
		RequestID:      requestID,
		Inputs:         inputs,
		State:          state,
		ApprovedOrigin: origin,
		Credential:     cred,
		IdempotencyKey: o.IdempotencyKey,
	})
	o.Body = body
	if e == nil {
		o.RequestFingerprint = definitions.ContentHash([]byte(request.Method + "\x00" + request.URL.String()))
	}
	return o, e
}

// credential selects only the slot declared by the active definition. Other saved slots
// must never be borrowed as an implicit fallback or transferred to a new origin.
func credential(d model.Definition, credentials map[string]model.Credential) *model.Credential {
	if d.Auth.Mode == "none" {
		return nil
	}
	if c, ok := credentials[d.Auth.SecretSlot]; ok {
		return &c
	}
	return nil
}

// SaveSetup claims a selected candidate once, atomically with the instance and initial attempt.
func (s *Service) SaveSetup(id, selection, name, expected string) (SetupDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out SetupDTO
	if e := s.ready(); e != nil {
		return out, e
	}
	if e := s.hasCapacity(); e != nil {
		return out, e
	}
	name = strings.TrimSpace(name)
	if len(name) > 120 {
		return out, model.Fail("invalid_input", "Use a name of at most 120 characters.")
	}
	var admitted string
	e := s.repo.Update(func(tx *repository.IntegrationTx) error {
		p, e := s.getSetup(tx, id)
		if e != nil {
			return e
		}
		if name == "" {
			name = s.integrationName(p.DefinitionKey)
		}
		hash := definitions.ContentHash([]byte(selection + "\x00" + name))
		if p.InstanceID != "" || p.SubmissionHash != "" {
			if p.SubmissionHash != hash {
				return model.Fail("revision_conflict", "This setup was already saved with different choices.")
			}
			out = s.setupDTO(p)
			return nil
		}
		if e = checkRevision(expected, revisionETag("setup", id, p.Revision)); e != nil {
			return e
		}
		if p.Status != "choose" {
			return model.Fail("invalid_input", "Search and choose a result before saving.")
		}
		found := false
		for _, c := range p.Choices {
			if c.ID == selection {
				found = true
			}
		}
		if !found {
			return model.Fail("invalid_choice", "Choose a result from the current results.")
		}
		c, e := s.compiler.CompileYAML(p.DefinitionYAML)
		if e != nil {
			return e
		}
		c.Revision = p.DefinitionRevision
		inputs := p.Inputs
		inputs[c.Definition.Setup.Select.Input] = selection
		var existing string
		if e = tx.Each("instances", func(key string, b []byte) error {
			var i model.Instance
			if e := repository.DecodeIntegrationRecord(b, &i); e != nil {
				return e
			}
			if key != p.TargetInstanceID && i.DefinitionKey == p.DefinitionKey && i.Inputs[c.Definition.Setup.Select.Input] == selection {
				existing = key
			}
			return nil
		}); e != nil {
			return e
		}
		if existing != "" {
			return model.Fail("already_configured", "This selection already has an integration. Open it from Integrations.")
		}
		var i model.Instance
		if p.TargetInstanceID != "" {
			ok, e := tx.Get("instances", p.TargetInstanceID, &i)
			if e != nil {
				return e
			}
			if !ok || i.Revision != p.TargetRevision {
				return model.Fail("revision_conflict", "The integration changed while you were editing. Start its setup again.")
			}
			p.Status = "saving"
		} else {
			now := s.clock.Now()
			count := 0
			if e := tx.Each("instances", func(string, []byte) error { count++; return nil }); e != nil {
				return e
			}
			if count >= 50 {
				return model.Fail("busy", "This installation supports up to 50 integrations.")
			}
			instanceID, e := allocateRecordID(tx, "instances", uuid.NewRandom)
			if e != nil {
				return e
			}
			i = model.Instance{
				ID:              instanceID,
				Name:            name,
				DefinitionKey:   p.DefinitionKey,
				Revision:        1,
				Generation:      1,
				BindingID:       uuid.NewString(),
				BindingRevision: 1,
				Enabled:         true,
				SearchInput:     c.Definition.Setup.Search.Input,
				Inputs:          inputs,
				State:           p.State,
				Credentials:     p.Credentials,
				Origin:          c.Origin,
				Identities:      map[string]string{},
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			if e = tx.Put("instances", i.ID, i); e != nil {
				return e
			}
			p.InstanceID = i.ID
			p.Status = "completed"
		}
		o, e := s.makeOperation(tx, c, inputs, p.State, p.Credentials, c.Origin, c.Definition.Setup.Select.Request, "setup", i.Generation)
		if e != nil {
			return e
		}
		o.InstanceID = i.ID
		o.SetupID = p.ID
		if p.TargetInstanceID != "" {
			o.Generation = p.Generation
		}
		p.OperationID = o.ID
		p.SubmissionHash = hash
		p.ConnectionName = name
		p.Inputs = inputs
		s.touch(&p)
		if p.TargetInstanceID == "" {
			i.OperationID = o.ID
			now := s.clock.Now()
			i.LastRenewal = &now
			if e = tx.Put("instances", i.ID, i); e != nil {
				return e
			}
			p.Choices = []model.Choice{}
			p.Inputs = nil
			p.State = nil
			p.Credentials = nil
			p.DefinitionYAML = ""
			p.ConnectionName = ""
			p.ExpiresAt = s.clock.Now().Add(24 * time.Hour)
		}
		if e = tx.Put("operations", o.ID, o); e != nil {
			return e
		}
		if e = tx.Put("setups", p.ID, p); e != nil {
			return e
		}
		out = s.setupDTO(p)
		admitted = o.ID
		return nil
	})
	if e == nil && admitted != "" {
		s.enqueue(admitted)
	}
	return out, e
}

// CancelSetup removes local unfinished state; it cannot retract an admitted remote lookup.
func (s *Service) CancelSetup(id, expected string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e := s.ready(); e != nil {
		return e
	}
	return s.repo.Update(func(tx *repository.IntegrationTx) error {
		p, e := s.getSetup(tx, id)
		if e != nil {
			return e
		}
		if e = checkRevision(expected, revisionETag("setup", id, p.Revision)); e != nil {
			return e
		}
		if p.Status == "completed" {
			return tx.Delete("setups", id)
		}
		operationID := p.OperationID
		if e = s.invalidateSetup(tx, &p); e != nil {
			return e
		}
		if operationID != "" {
			if e = tx.Delete("operations", operationID); e != nil {
				return e
			}
		}
		return tx.Delete("setups", id)
	})
}

// integrationName derives a default from installed category metadata rather than a domain label.
func (s *Service) integrationName(key string) string {
	category := strings.SplitN(key, "/", 2)[0]
	for _, descriptor := range s.capabilities.Descriptors() {
		if descriptor.IntegrationType == category {
			return descriptor.Name
		}
	}
	return "Integration"
}
