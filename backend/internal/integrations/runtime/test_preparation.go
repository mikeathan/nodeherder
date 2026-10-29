package runtime

// Prepares credentials and capacity for isolated provider-definition tests.

import (
	"node-herder/internal/integrations/definitions"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

// prepareTestCredentials copies credentials and validates a live cross-origin test's review.
func prepareTestCredentials(i model.Instance, c *definitions.Compiled, mode string, review *DestinationReview) (map[string]model.Credential, error) {
	credentials := map[string]model.Credential{}
	for slot, cred := range i.Credentials {
		credentials[slot] = cred
	}
	activeCredential := credential(c.Definition, i.Credentials)
	if c.Origin != i.Origin || activeCredential != nil && activeCredential.Origin != c.Origin {
		if mode == "live" && !testDestinationApproved(i, c, activeCredential, review) {
			return nil, model.Fail("destination_review_required", "Approve sending this integration's saved inputs and token to the test destination.")
		}
		if cred := credential(c.Definition, credentials); cred != nil {
			cred.Origin = c.Origin
			credentials[c.Definition.Auth.SecretSlot] = *cred
		}
	}
	return credentials, nil
}

// testDestinationApproved checks whether the test destination and exact credential revision were approved.
func testDestinationApproved(instance model.Instance, compiled *definitions.Compiled, active *model.Credential, review *DestinationReview) bool {
	if review == nil || review.Origin != compiled.Origin {
		return false
	}
	var credentialRevision uint64
	if active != nil {
		credentialRevision = active.Revision
	}
	return hasCredentialReview(review, instance.ID, instance.Revision, credentialRevision, compiled.Definition.Auth.SecretSlot)
}

// ensureTestCapacity bounds temporary provider tests and unfinished setups.
func ensureTestCapacity(tx *repository.IntegrationTx) error {
	count := 0
	if e := tx.Each("setups", func(_ string, b []byte) error {
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
	return nil
}
