package runtime

// Provides setup lifecycle operations and constructs durable provider attempts.

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"node-herder/internal/integrations/definitions"
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
	e := s.repo.Update(func(tx *repository.IntegrationTx) error {
		setup, err := s.createSetup(tx, key, targetID, targetETag)
		if err != nil {
			return err
		}
		out = s.setupDTO(setup)
		return nil
	})
	return out, e
}

// getSetup reads a live setup and rejects expired or missing setup records.
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

// touch advances setup revision and its bounded idle expiry.
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
		setup, operationID, err := s.admitSearch(tx, id, query, expected)
		if err != nil {
			return err
		}
		out = s.setupDTO(setup)
		admitted = operationID
		return nil
	})
	if e == nil {
		s.enqueue(admitted)
	}
	return out, e
}

// invalidateSetup cancels and terminates work tied to an earlier setup generation.
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
		if err := s.ensureOperationCapacity(tx); err != nil {
			return model.Operation{}, err
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
		setup, operationID, err := s.saveSelection(tx, id, selection, name, expected)
		if err != nil {
			return err
		}
		out = s.setupDTO(setup)
		admitted = operationID
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
