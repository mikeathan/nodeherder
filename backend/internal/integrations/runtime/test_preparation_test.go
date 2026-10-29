package runtime

// Checks credential review before testing a changed provider destination.

import (
	"testing"

	"node-herder/internal/integrations/definitions"
	model "node-herder/models/integrations"
)

func TestPrepareTestCredentialsRequiresCurrentDestinationReview(t *testing.T) {
	instance := model.Instance{ID: "instance", Revision: 4, Origin: "https://old.example", Credentials: map[string]model.Credential{
		"key": {Value: "secret", Origin: "https://old.example", Revision: 2},
	}}
	compiled := &definitions.Compiled{Origin: "https://new.example", Definition: model.Definition{Auth: model.AuthConfig{Mode: "bearer", SecretSlot: "key"}}}
	grant := CredentialReview{InstanceID: "instance", Revision: 4, CredentialRevision: 2, CredentialSlot: "key"}
	for _, tc := range []struct {
		name   string
		review *DestinationReview
		allow  bool
	}{
		{name: "missing review"},
		{name: "wrong origin", review: &DestinationReview{Origin: instance.Origin, Reassociate: []CredentialReview{grant}}},
		{name: "stale credential", review: &DestinationReview{Origin: compiled.Origin, Reassociate: []CredentialReview{{InstanceID: instance.ID, Revision: 4, CredentialRevision: 1, CredentialSlot: "key"}}}},
		{name: "approved", review: &DestinationReview{Origin: compiled.Origin, Reassociate: []CredentialReview{grant}}, allow: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			credentials, err := prepareTestCredentials(instance, compiled, "live", tc.review)
			if (err == nil) != tc.allow {
				t.Fatalf("allowed = %v, error = %v", err == nil, err)
			}
			if tc.allow && (credentials["key"].Origin != compiled.Origin || instance.Credentials["key"].Origin != instance.Origin) {
				t.Fatal("test credentials must be copied and reassociated without changing the saved credential")
			}
		})
	}
}
