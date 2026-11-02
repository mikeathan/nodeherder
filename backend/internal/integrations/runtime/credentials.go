package runtime

// Resolves declared credential slots without exposing secret values.

import (
	"node-herder/internal/integrations/destination"
	model "node-herder/models/integrations"
)

// credentialSlot identifies the primary slot used for operation revision fencing.
func credentialSlot(definition model.Definition) string {
	if definition.Auth.Mode == "oauth2_client_credentials" {
		return definition.Auth.ClientSecretSlot
	}
	return definition.Auth.SecretSlot
}

// credential selects the primary slot used for operation revision fencing.
func credential(definition model.Definition, credentials map[string]model.Credential) *model.Credential {
	slot := credentialSlot(definition)
	if slot == "" {
		return nil
	}
	value, ok := credentials[slot]
	if !ok {
		return nil
	}
	if definition.Auth.Mode == "oauth2_client_credentials" {
		if clientID, exists := credentials[definition.Auth.ClientIDSlot]; exists {
			// Both OAuth slots are credential material: changing either must fence old work.
			value.Revision = clientID.Revision*1_000_000 + value.Revision
		}
	}
	return &value
}

// credentialOrigin binds client-credentials slots to the token endpoint, not the data origin.
func credentialOrigin(definition model.Definition, providerOrigin, slot string) string {
	if definition.Auth.Mode == "oauth2_client_credentials" &&
		(slot == definition.Auth.ClientIDSlot || slot == definition.Auth.ClientSecretSlot) {
		if target, err := destination.Parse(definition.Auth.TokenURL); err == nil {
			return destination.Origin(target)
		}
	}
	return providerOrigin
}

// credentialReviewOrigin exposes a separate approval origin only for OAuth token requests.
// Other credential modes are covered by the provider data-origin review.
func credentialReviewOrigin(definition model.Definition) string {
	if definition.Auth.Mode != "oauth2_client_credentials" {
		return ""
	}
	return credentialOrigin(definition, "", credentialSlot(definition))
}

// credentialsConfigured reports whether every declared slot is present.
func credentialsConfigured(definition model.Definition, credentials map[string]model.Credential) bool {
	slots := definition.Auth.Slots()
	if len(slots) == 0 {
		return false
	}
	for _, slot := range slots {
		if _, ok := credentials[slot]; !ok {
			return false
		}
	}
	return true
}
