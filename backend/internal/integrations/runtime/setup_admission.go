package runtime

// Creates setups and admits searches with capacity and revision checks.

import (
	"time"

	"github.com/google/uuid"

	"node-herder/internal/integrations/flows/searchselect"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

// createSetup allocates a setup and optionally binds it to an existing instance revision.
func (s *Service) createSetup(tx *repository.IntegrationTx, key, targetID, targetETag string) (model.Setup, error) {
	if err := s.ensureSetupCapacity(tx); err != nil {
		return model.Setup{}, err
	}
	record, compiled, err := s.definition(tx, key)
	if err != nil {
		return model.Setup{}, err
	}
	if record.DiskError != "" {
		return model.Setup{}, model.Fail("definition_unavailable", "Repair or reload this provider before setup.")
	}
	id, err := allocateRecordID(tx, "setups", uuid.NewRandom)
	if err != nil {
		return model.Setup{}, err
	}
	now := s.clock.Now()
	setup := model.Setup{
		ID:                 id,
		DefinitionKey:      key,
		DefinitionYAML:     compiled.YAML,
		DefinitionRevision: record.Revision,
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
		if err = s.bindSetupTarget(tx, &setup, targetID, targetETag); err != nil {
			return model.Setup{}, err
		}
	}
	err = tx.Put("setups", id, setup)
	return setup, err
}

// ensureSetupCapacity bounds the number of active setup records.
func (s *Service) ensureSetupCapacity(tx *repository.IntegrationTx) error {
	now := s.clock.Now()
	count := 0
	err := tx.Each("setups", func(_ string, data []byte) error {
		var setup model.Setup
		if err := repository.DecodeIntegrationRecord(data, &setup); err != nil {
			return err
		}
		if setup.Status == "completed" || setup.Status == "expired" {
			return nil
		}
		if now.Before(setup.ExpiresAt) {
			count++
		}
		return nil
	})
	if err != nil {
		return err
	}
	if count >= 32 {
		return model.Fail("busy", "Finish or cancel another setup before starting one.")
	}
	return nil
}

// bindSetupTarget checks provider and revision before copying an existing binding into setup.
func (s *Service) bindSetupTarget(tx *repository.IntegrationTx, setup *model.Setup, targetID, expected string) error {
	var instance model.Instance
	found, err := tx.Get("instances", targetID, &instance)
	if err != nil {
		return err
	}
	if !found {
		return model.Fail("not_found", "Integration not found.")
	}
	if instance.DefinitionKey != setup.DefinitionKey {
		return model.Fail("invalid_input", "Use the existing provider to change the selection.")
	}
	if err = checkRevision(expected, revisionETag("instance", instance.ID, instance.Revision)); err != nil {
		return err
	}
	setup.TargetInstanceID = instance.ID
	setup.TargetRevision = instance.Revision
	setup.Credentials = instance.Credentials
	return nil
}

// admitSearch validates search input and stores the setup and queued operation together.
func (s *Service) admitSearch(tx *repository.IntegrationTx, id, query, expected string) (model.Setup, string, error) {
	setup, err := s.getSetup(tx, id)
	if err != nil {
		return setup, "", err
	}
	if err = checkRevision(expected, revisionETag("setup", id, setup.Revision)); err != nil {
		return setup, "", err
	}
	if setup.InstanceID != "" {
		return setup, "", model.Fail("revision_conflict", "This setup was already saved.")
	}
	compiled, err := s.compileSetup(setup)
	if err != nil {
		return setup, "", err
	}
	if err = s.invalidateSetup(tx, &setup); err != nil {
		return setup, "", err
	}
	inputName := compiled.Definition.Setup.Search.Input
	normalized, err := searchselect.NormalizeInput(query, compiled.Definition.Inputs[inputName])
	if err != nil {
		return setup, "", err
	}
	requestID := compiled.Definition.Setup.Search.Request
	// A declared postcode service can replace the normal town search for UK postcodes.
	if alternative := compiled.Definition.Setup.Search.Postcode; alternative != nil && searchselect.IsUKPostcode(normalized) {
		requestID = alternative.Request
		normalized, err = searchselect.NormalizeInput(normalized, model.Input{Type: "postcode"})
		if err != nil {
			return setup, "", err
		}
	}
	resetSearch(&setup, inputName, normalized)
	s.touch(&setup)
	operation, err := s.makeOperation(tx, compiled, setup.Inputs, setup.State, setup.Credentials, compiled.Origin, requestID, "search", setup.Generation)
	if err != nil {
		return setup, "", err
	}
	operation.SetupID = setup.ID
	setup.OperationID = operation.ID
	if err = tx.Put("operations", operation.ID, operation); err != nil {
		return setup, "", err
	}
	err = tx.Put("setups", setup.ID, setup)
	return setup, operation.ID, err
}

// resetSearch clears selections and state that belong to a previous search.
func resetSearch(setup *model.Setup, inputName, value string) {
	setup.Inputs = map[string]string{inputName: value}
	setup.State = map[string]any{}
	setup.Choices = []model.Choice{}
	setup.Status = "searching"
	setup.SubmissionHash = ""
	setup.ConnectionName = ""
	setup.LastError = nil
}

// ensureOperationCapacity bounds the number of unfinished provider operations.
func (s *Service) ensureOperationCapacity(tx *repository.IntegrationTx) error {
	now := s.clock.Now()
	count := 0
	err := tx.Each("operations", func(_ string, data []byte) error {
		var operation model.Operation
		if err := repository.DecodeIntegrationRecord(data, &operation); err != nil {
			return err
		}
		if operation.Status == "done" || operation.Status == "failed" {
			return nil
		}
		if now.Before(operation.ExpiresAt) {
			count++
		}
		return nil
	})
	if err != nil {
		return err
	}
	if count >= 64 {
		return model.Fail("busy", "Too many pending operations. Complete or cancel an existing lookup first.")
	}
	return nil
}
