package operations

// Applies the configured authentication only to approved provider requests.

import (
	"net/http"
	"strings"

	model "node-herder/models/integrations"
)

// authenticateRequest forwards a valid credential only when the selected destination permits it.
func authenticateRequest(request *http.Request, attempt Attempt, forwardCredentials bool) error {
	if !forwardCredentials || attempt.Definition.Auth.Mode == "none" {
		return nil
	}
	credential := attempt.Credential
	if credential == nil {
		if attempt.Definition.Auth.Mode != "optional_bearer" {
			return model.Fail("credential_required", "Add the provider credential in Setup.")
		}
		return nil
	}
	// A saved credential is bound to its reviewed origin, even if the slot name matches.
	if credential.Origin != attempt.ApprovedOrigin {
		return model.Fail("reconnect_required", "The credential belongs to another provider destination.")
	}
	if !validCredentialValue(credential.Value) {
		return model.Fail("invalid_credential", "Enter a valid provider token.")
	}
	switch attempt.Definition.Auth.Mode {
	case "bearer", "optional_bearer":
		request.Header.Set("Authorization", "Bearer "+credential.Value)
	case "apikey_header":
		request.Header.Set("apikey", credential.Value)
	case "apikey_query":
		query := request.URL.Query()
		query.Set("key", credential.Value)
		request.URL.RawQuery = query.Encode()
	default:
		return model.Fail("invalid_definition", "Unsupported provider authentication.")
	}
	return nil
}

// validCredentialValue rejects empty, oversized and line-breaking credential values.
func validCredentialValue(value string) bool {
	return value != "" && len(value) <= 8192 && !strings.ContainsAny(value, "\r\n")
}
