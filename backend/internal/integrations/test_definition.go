package integrations

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/google/uuid"

	"node-herder/internal/integrations/capabilities"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

// TestDefinition previews only a saved-schedule GET using an existing confirmed binding. It never
// changes that binding, the active definition, cache validators, snapshots or consumer resources.
// Offline mode validates a supplied sample without network I/O; live mode is explicit bounded work.
func (s *Service) TestDefinition(key, source, instanceID, mode string, sample json.RawMessage, review *DestinationReview) (SetupDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out SetupDTO
	if e := s.ready(); e != nil {
		return out, e
	}
	if mode != "offline" && mode != "live" {
		return out, model.Fail("invalid_input", "Choose an offline sample or a live saved-schedule test.")
	}
	if len(sample) > 256<<10 {
		return out, model.Fail("invalid_input", "Use a sample of at most 256 KiB.")
	}
	validation, e := s.validateDefinition(key, source, nil)
	if e != nil {
		return out, e
	}
	if mode == "live" {
		if e = reviewDestination(validation, review); e != nil {
			return out, e
		}
	}
	c, e := CompileYAML(source)
	if e != nil {
		return out, e
	}
	if e = s.hasCapacity(); e != nil {
		return out, e
	}
	var admitted string
	e = s.repo.Update(func(tx *repository.IntegrationTx) error {
		var i model.Instance
		ok, e := tx.Get("instances", instanceID, &i)
		if e != nil {
			return e
		}
		if !ok || i.DefinitionKey != key {
			return model.Fail("invalid_input", "Choose an existing connection using this provider.")
		}
		if mode == "live" && !i.Enabled {
			return model.Fail("disabled", "Enable the connection before testing a live request.")
		}
		credentials := map[string]model.Credential{}
		for slot, cred := range i.Credentials {
			credentials[slot] = cred
		}
		activeCredential := credential(c.Definition, i.Credentials)
		if c.Origin != i.Origin || activeCredential != nil && activeCredential.Origin != c.Origin {
			approved := false
			if review != nil && review.Origin == c.Origin {
				for _, given := range review.Reassociate {
					cred := credential(c.Definition, i.Credentials)
					revision := uint64(0)
					if cred != nil {
						revision = cred.Revision
					}
					if given.InstanceID == i.ID && given.Revision == i.Revision && given.CredentialRevision == revision && given.CredentialSlot == c.Definition.Auth.SecretSlot {
						approved = true
					}
				}
			}
			if mode == "live" && !approved {
				return model.Fail("destination_review_required", "Approve sending this connection's saved inputs and token to the test destination.")
			}
			if cred := credential(c.Definition, credentials); cred != nil {
				cred.Origin = c.Origin
				credentials[c.Definition.Auth.SecretSlot] = *cred
			}
		}
		count := 0
		if e = tx.Each("setups", func(_ string, b []byte) error {
			var p model.Setup
			if e := repository.DecodeIntegrationRecord(b, &p); e != nil {
				return e
			}
			if p.TestOnly || p.Status != "completed" {
				count++
			}
			return nil
		}); e != nil {
			return e
		}
		if count >= 32 {
			return model.Fail("busy", "Cancel or wait for an existing setup or test to expire.")
		}
		now := s.clock.Now()
		setupID, e := allocateRecordID(tx, "setups", uuid.NewRandom)
		if e != nil {
			return e
		}
		p := model.Setup{
			ID:                 setupID,
			DefinitionKey:      key,
			DefinitionYAML:     source,
			DefinitionRevision: 1,
			Revision:           1,
			Generation:         1,
			TargetInstanceID:   i.ID,
			TargetRevision:     i.Revision,
			TestOnly:           true,
			Status:             "testing",
			Inputs:             i.Inputs,
			State:              i.State,
			Choices:            []model.Choice{},
			Credentials:        credentials,
			CreatedAt:          now,
			ExpiresAt:          now.Add(30 * time.Minute),
			AbsoluteExpiry:     now.Add(30 * time.Minute),
		}
		if mode == "offline" {
			decoder := json.NewDecoder(bytes.NewReader(sample))
			decoder.UseNumber()
			var body any
			if e = decoder.Decode(&body); e != nil {
				return model.Fail("invalid_input", "The sample is not valid JSON.")
			}
			var extra any
			if decoder.Decode(&extra) != io.EOF {
				return model.Fail("invalid_input", "Use one JSON sample.")
			}
			result, e := capabilities.NormalizeWaste(body, c.Definition.Mapping, capabilities.WasteContext{
				InstanceID:       i.ID,
				ProviderID:       c.Definition.ID,
				ProviderVersion:  "sha256:" + c.Hash,
				ProfileRevision:  1,
				BindingRevision:  i.BindingRevision,
				SelectedProperty: i.Inputs[c.Definition.Setup.Select.Input],
				Now:              now,
			})
			if e != nil {
				return e
			}
			p.Status = "completed"
			p.Preview = &result.Snapshot
			p.Inputs = nil
			p.State = nil
			p.Credentials = nil
		} else {
			o, e := s.makeOperation(tx, c, p.Inputs, p.State, p.Credentials, c.Origin, c.Definition.Refresh.Request, "test", p.Generation)
			if e != nil {
				return e
			}
			o.SetupID = p.ID
			o.InstanceID = i.ID
			p.OperationID = o.ID
			if e = tx.Put("operations", o.ID, o); e != nil {
				return e
			}
			admitted = o.ID
		}
		if e = tx.Put("setups", p.ID, p); e != nil {
			return e
		}
		out = setupDTO(p)
		return nil
	})
	if e == nil && admitted != "" {
		s.enqueue(admitted)
	}
	return out, e
}
