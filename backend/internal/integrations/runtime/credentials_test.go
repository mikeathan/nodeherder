package runtime

// Checks credential-slot and OAuth token-origin review semantics.

import (
	"testing"

	model "node-herder/models/integrations"
)

func TestOAuthCredentialReviewUsesTokenOrigin(t *testing.T) {
	definition := model.Definition{Auth: model.AuthConfig{
		Mode:             "oauth2_client_credentials",
		TokenURL:         "https://auth.example/token",
		ClientIDSlot:     "client_id",
		ClientSecretSlot: "client_secret",
	}}
	if got := credentialSlot(definition); got != "client_secret" {
		t.Fatalf("credential slot = %q", got)
	}
	validation := ValidationDTO{
		Origin:                   "https://api.example",
		PreviousOrigin:           "https://api.example",
		CredentialOrigin:         "https://auth.example",
		PreviousCredentialOrigin: "https://old-auth.example",
	}
	if err := reviewDestination(validation, nil); err == nil {
		t.Fatal("token-origin change was accepted without review")
	}
	if err := reviewDestination(validation, &DestinationReview{
		Origin:           "https://api.example",
		CredentialOrigin: "https://auth.example",
	}); err != nil {
		t.Fatal(err)
	}
}
