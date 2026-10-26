package runtime

import (
	"node-herder/internal/integrations/definitions"
	"time"

	"github.com/google/uuid"

	"node-herder/internal/integrations/capabilities"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

// finishSnapshot publishes the resource, identity reservations and operation completion in one
// transaction. A reconfiguration swaps its binding only after the candidate result is validated.
func (s *Service) finishSnapshot(tx *repository.IntegrationTx, o *model.Operation, owner operationOwner, c *definitions.Compiled, result capabilities.Result, state map[string]any, etag string) error {
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
	if e := reserveResourceIdentities(tx, result.Resource, result.Reservations, i.ID); e != nil {
		return e
	}
	s.recordRun(&i, *o, "done", "Integration data saved.")
	i.ResourceID = result.Resource.ResourceID
	i.Snapshot = &result.Resource
	i.Identities = result.Identities
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
