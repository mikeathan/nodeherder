package runtime

// Adds and removes an integration that needs no setup input.
//
// The authored setup flow exists to collect inputs, perform a search and record a selection, and
// every stage of it assumes a provider that answers over the network. A local provider has nothing
// to collect and nothing to search, so it needs one action rather than a wizard — but it produces
// exactly the same instance record, which is what makes it behave like any other integration
// afterwards.

import (
	"strings"

	"github.com/google/uuid"

	model "node-herder/models/integrations"
	"node-herder/repository"
)

// AddLocal creates an enabled instance for a local provider and checks its source.
//
// The instance is created enabled, matching a completed authored setup: adding an integration means
// wanting it to work. The check is not what makes data appear — the value is resolved on read
// (ADR-005) — but it is what lets a broken source be reported at the moment the operator added it,
// rather than looking like a working installation in a quiet room.
func (s *Service) AddLocal(key, name string) (InstanceDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out InstanceDTO
	if e := s.ready(); e != nil {
		return out, e
	}
	compiled, ok := s.localDefinition(key)
	if !ok {
		return out, model.Fail("not_found", "Provider not found.")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = compiled.Definition.Name
	}
	if len(name) > 120 {
		return out, model.Fail("invalid_input", "Enter a name of at most 120 characters.")
	}
	if e := s.hasCapacity(); e != nil {
		return out, e
	}

	e := s.repo.Update(func(tx *repository.IntegrationTx) error {
		if e := ensureInstanceCapacity(tx); e != nil {
			return e
		}
		// Refused rather than silently duplicated. Two instances of one local provider would read
		// the same radio and publish the same census twice, which is never what anyone means and
		// would be hard to diagnose from the board.
		if e := rejectDuplicateLocal(tx, key); e != nil {
			return e
		}
		id, e := allocateRecordID(tx, "instances", uuid.NewRandom)
		if e != nil {
			return e
		}
		now := s.clock.Now()
		instance := model.Instance{
			ID:              id,
			Name:            name,
			DefinitionKey:   key,
			Revision:        1,
			Generation:      1,
			BindingID:       uuid.NewString(),
			BindingRevision: 1,
			Enabled:         true,
			Origin:          compiled.Origin,
			Identities:      map[string]string{},
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		// publishLocal commits the instance, so it is not written again here.
		published, e := s.publishLocal(tx, &instance, compiled)
		if e != nil {
			return e
		}
		out = s.instanceDTO(published)
		return nil
	})
	return out, e
}

// LocalEnabled reports whether any enabled instance of a local provider exists.
//
// This is the question the observation service asks before subscribing: it is deliberately about
// enabled instances rather than about the device, because a device that can observe should stop
// doing so when no integration wants the data.
func (s *Service) LocalEnabled(key string) bool {
	if s.repo == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	enabled := false
	_ = s.repo.View(func(tx *repository.IntegrationTx) error {
		return tx.Each("instances", func(_ string, b []byte) error {
			var instance model.Instance
			if e := repository.DecodeIntegrationRecord(b, &instance); e != nil {
				return e
			}
			if instance.DefinitionKey == key && instance.Enabled {
				enabled = true
			}
			return nil
		})
	})
	return enabled
}

// rejectDuplicateLocal refuses a second instance of a provider that reads one shared source.
//
// The message names the existing instance rather than the type, because the type's display name is
// not available here and a hardcoded product name would be wrong for the next local provider.
func rejectDuplicateLocal(tx *repository.IntegrationTx, key string) error {
	existing := ""
	if e := tx.Each("instances", func(_ string, b []byte) error {
		var instance model.Instance
		if e := repository.DecodeIntegrationRecord(b, &instance); e != nil {
			return e
		}
		if instance.DefinitionKey == key {
			existing = instance.Name
		}
		return nil
	}); e != nil {
		return e
	}
	if existing != "" {
		return model.Fail("already_exists", "\""+existing+"\" already reads this source. Remove it before adding another.")
	}
	return nil
}
