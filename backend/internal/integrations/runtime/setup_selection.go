package runtime

// Validates a setup selection and admits the operation that creates or updates an instance.

import (
	"maps"
	"time"

	"github.com/google/uuid"

	"node-herder/internal/integrations/definitions"
	"node-herder/internal/integrations/flows/searchselect"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

// saveSelection runs under the caller's service lock and single repository transaction.
// It returns an existing receipt for an identical replay, or a newly admitted operation ID.
func (s *Service) saveSelection(tx *repository.IntegrationTx, id, selection, name, expected string) (model.Setup, string, error) {
	setup, err := s.getSetup(tx, id)
	if err != nil {
		return setup, "", err
	}
	if name == "" {
		name = s.integrationName(setup.DefinitionKey)
	}
	submissionHash := definitions.ContentHash([]byte(selection + "\x00" + name))
	if setup.InstanceID != "" || setup.SubmissionHash != "" {
		if setup.SubmissionHash != submissionHash {
			return setup, "", model.Fail("revision_conflict", "This setup was already saved with different choices.")
		}
		return setup, "", nil
	}
	choice, err := validateSelectedChoice(setup, selection, expected)
	if err != nil {
		return setup, "", err
	}
	compiled, err := s.compileSetup(setup)
	if err != nil {
		return setup, "", err
	}
	if err = ensureUniqueSelection(tx, setup, compiled.Definition, selection); err != nil {
		return setup, "", err
	}
	setup.Inputs = maps.Clone(setup.Inputs)
	setup.Inputs[compiled.Definition.Setup.Select.Input] = selection
	setup.State, err = searchselect.SelectionState(compiled.Definition.Setup, choice, setup.State)
	if err != nil {
		return setup, "", err
	}
	instance, err := s.selectionInstance(tx, &setup, compiled, name)
	if err != nil {
		return setup, "", err
	}
	operation, err := s.admitSelection(tx, compiled, &setup, &instance, name, submissionHash)
	return setup, operation, err
}

// validateSelectedChoice checks the setup revision and confirms the choice belongs to current results.
func validateSelectedChoice(setup model.Setup, selection, expected string) (model.Choice, error) {
	if err := checkRevision(expected, revisionETag("setup", setup.ID, setup.Revision)); err != nil {
		return model.Choice{}, err
	}
	if setup.Status != "choose" {
		return model.Choice{}, model.Fail("invalid_input", "Search and choose a result before saving.")
	}
	for _, choice := range setup.Choices {
		if choice.ID == selection {
			return choice, nil
		}
	}
	return model.Choice{}, model.Fail("invalid_choice", "Choose a result from the current results.")
}

// compileSetup compiles the definition revision pinned to this setup.
func (s *Service) compileSetup(setup model.Setup) (*definitions.Compiled, error) {
	compiled, err := s.compiler.CompileYAML(setup.DefinitionYAML)
	if err != nil {
		return nil, err
	}
	compiled.Revision = setup.DefinitionRevision
	return compiled, nil
}

// ensureUniqueSelection prevents duplicate instances for the same provider selection.
func ensureUniqueSelection(tx *repository.IntegrationTx, setup model.Setup, definition model.Definition, selection string) error {
	return tx.Each("instances", func(id string, data []byte) error {
		if id == setup.TargetInstanceID {
			return nil
		}
		var instance model.Instance
		if err := repository.DecodeIntegrationRecord(data, &instance); err != nil {
			return err
		}
		if instance.DefinitionKey != setup.DefinitionKey {
			return nil
		}
		if instance.Inputs[definition.Setup.Select.Input] != selection {
			return nil
		}
		return model.Fail("already_configured", "This selection already has an integration. Open it from Integrations.")
	})
}

// selectionInstance reuses the reviewed target or creates a new integration identity.
func (s *Service) selectionInstance(tx *repository.IntegrationTx, setup *model.Setup, compiled *definitions.Compiled, name string) (model.Instance, error) {
	if setup.TargetInstanceID != "" {
		var instance model.Instance
		found, err := tx.Get("instances", setup.TargetInstanceID, &instance)
		if err != nil {
			return instance, err
		}
		if !found || instance.Revision != setup.TargetRevision {
			return instance, model.Fail("revision_conflict", "The integration changed while you were editing. Start its setup again.")
		}
		setup.Status = "saving"
		return instance, nil
	}
	if err := ensureInstanceCapacity(tx); err != nil {
		return model.Instance{}, err
	}
	id, err := allocateRecordID(tx, "instances", uuid.NewRandom)
	if err != nil {
		return model.Instance{}, err
	}
	now := s.clock.Now()
	instance := model.Instance{
		ID:              id,
		Name:            name,
		DefinitionKey:   setup.DefinitionKey,
		Revision:        1,
		Generation:      1,
		BindingID:       uuid.NewString(),
		BindingRevision: 1,
		Enabled:         true,
		SearchInput:     compiled.Definition.Setup.Search.Input,
		Inputs:          setup.Inputs,
		State:           setup.State,
		Credentials:     setup.Credentials,
		Origin:          compiled.Origin,
		Identities:      map[string]string{},
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	setup.InstanceID = id
	setup.Status = "completed"
	return instance, nil
}

// ensureInstanceCapacity enforces the configured limit on integration instances.
func ensureInstanceCapacity(tx *repository.IntegrationTx) error {
	count := 0
	if err := tx.Each("instances", func(string, []byte) error { count++; return nil }); err != nil {
		return err
	}
	if count >= 50 {
		return model.Fail("busy", "This installation supports up to 50 integrations.")
	}
	return nil
}

// admitSelection persists the selection, operation and new instance atomically.
func (s *Service) admitSelection(tx *repository.IntegrationTx, compiled *definitions.Compiled, setup *model.Setup, instance *model.Instance, name, submissionHash string) (string, error) {
	operation, err := s.makeOperation(tx, compiled, setup.Inputs, setup.State, setup.Credentials, compiled.Origin, compiled.Definition.Setup.Select.Request, "setup", instance.Generation)
	if err != nil {
		return "", err
	}
	operation.InstanceID = instance.ID
	operation.SetupID = setup.ID
	if setup.TargetInstanceID != "" {
		operation.Generation = setup.Generation
	}
	setup.OperationID = operation.ID
	setup.SubmissionHash = submissionHash
	setup.ConnectionName = name
	s.touch(setup)
	if setup.TargetInstanceID == "" {
		instance.OperationID = operation.ID
		now := s.clock.Now()
		instance.LastRenewal = &now
		if err = tx.Put("instances", instance.ID, *instance); err != nil {
			return "", err
		}
		retainSetupReceipt(setup, now)
	}
	if err = tx.Put("operations", operation.ID, operation); err != nil {
		return "", err
	}
	if err = tx.Put("setups", setup.ID, *setup); err != nil {
		return "", err
	}
	return operation.ID, nil
}

// A new connection owns private setup data after admission. The completion receipt
// retains replay identity without keeping a second copy of credentials or inputs.
func retainSetupReceipt(setup *model.Setup, now time.Time) {
	setup.Choices = []model.Choice{}
	setup.Inputs = nil
	setup.State = nil
	setup.Credentials = nil
	setup.DefinitionYAML = ""
	setup.ConnectionName = ""
	setup.ExpiresAt = now.Add(24 * time.Hour)
}
