package runtime

// Projects saved resources and prepares normalized operation results for publication.

import (
	"time"

	"node-herder/internal/integrations/capabilities"
	"node-herder/internal/integrations/definitions"
	"node-herder/internal/integrations/flows/searchselect"
	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

// projectResource is the shared read boundary for saved resources and setup previews. A removed
// codec or corrupt payload becomes visible feedback without deleting or rewriting saved evidence.
func (s *Service) projectResource(resource model.Resource, now time.Time) (model.Resource, error) {
	handler, err := s.capabilities.Resolve(resource.ContractVersion)
	if err != nil {
		return model.Resource{}, err
	}
	return handler.Project(resource, now)
}

// finishUnchanged advances check time only; provider verification age and content identity
// remain untouched by HTTP 304. It runs under the same transaction as operation completion.
func (s *Service) finishUnchanged(tx *repository.IntegrationTx, op *model.Operation, owner operationOwner) error {
	if owner.instance.Snapshot == nil {
		return s.failOperation(tx, op, owner, model.Fail("invalid_response", "No saved integration data exists for this response."))
	}
	i := owner.instance
	handler, err := s.capabilities.Resolve(i.Snapshot.ContractVersion)
	if err != nil {
		return s.failOperation(tx, op, owner, failure(err))
	}
	checked, err := handler.Checked(*i.Snapshot, s.clock.Now())
	if err != nil {
		return s.failOperation(tx, op, owner, failure(err))
	}
	i.Snapshot = &checked
	s.recordRun(&i, *op, "done", "The saved data is unchanged; source verification time was preserved.")
	i.OperationID = ""
	i.LastError = nil
	s.finishInstanceSuccess(&i)
	op.Status = "done"
	clearAttempt(op)
	if e := tx.Put("instances", i.ID, i); e != nil {
		return e
	}
	return tx.Put("operations", op.ID, *op)
}

// normalizeOperationResult preserves IDs for a compatible refresh. A candidate for a different
// binding starts fresh identities; previews borrow neither an existing resource ID nor its history.
func (s *Service) normalizeOperationResult(tx *repository.IntegrationTx, op model.Operation, owner operationOwner, c *definitions.Compiled, response operations.Response) (capabilities.Result, error) {
	selected := op.Inputs[c.Definition.Setup.Select.Input]
	prior := owner.instance.Snapshot
	ids := owner.instance.Identities
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
	return s.capabilities.Normalize(c.Definition, response.Root, capabilities.Context{
		NewIdentity:     snapshotIdentityAllocator(tx),
		InstanceID:      op.InstanceID,
		ResourceID:      resource,
		ProviderID:      c.Definition.ID,
		ProviderVersion: "sha256:" + op.DefinitionHash,
		ProfileRevision: op.ProfileRevision,
		BindingRevision: bindingRevision,
		SelectedValue:   selected,
		SelectedLabel:   searchselect.SelectionLabel(op.State),
		Now:             s.clock.Now(),
		Previous:        prior,
		Identities:      ids,
	})
}
