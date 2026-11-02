package operations

// Acquires and caches OAuth2 client-credentials tokens without persisting them.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"node-herder/internal/integrations/destination"
	model "node-herder/models/integrations"
)

const (
	oauthMaxResponseBytes = 64 << 10
	oauthDefaultLifetime  = 5 * time.Minute
	oauthSafetyMargin     = 60 * time.Second
	oauthCacheLimit       = 32
)

type cachedToken struct {
	Value     string
	ExpiresAt time.Time
}

var tokenCache = struct {
	sync.Mutex
	entries map[string]cachedToken
}{entries: map[string]cachedToken{}}

// prepareAttempt obtains a bearer token for OAuth definitions; other modes pass through.
func prepareAttempt(ctx context.Context, client Doer, attempt Attempt, now time.Time) (Attempt, error) {
	if attempt.Definition.Auth.Mode != "oauth2_client_credentials" {
		return attempt, nil
	}
	token, err := oauthToken(ctx, client, attempt, now)
	if err != nil {
		return attempt, err
	}
	attempt.Token = token
	return attempt, nil
}

// executeWithRetry performs one request and refreshes an OAuth token at most once on 401/403.
func executeWithRetry(ctx context.Context, client Doer, attempt Attempt, now time.Time) (Response, error) {
	response, err := executeSingle(ctx, client, attempt, now)
	if err != nil || attempt.Definition.Auth.Mode != "oauth2_client_credentials" ||
		(response.Status != http.StatusUnauthorized && response.Status != http.StatusForbidden) {
		return response, err
	}
	invalidateOAuthToken(attempt)
	refreshed, tokenErr := prepareAttempt(ctx, client, attempt, now)
	if tokenErr != nil {
		return response, tokenErr
	}
	response, err = executeSingle(ctx, client, refreshed, now)
	if err == nil && (response.Status == http.StatusUnauthorized || response.Status == http.StatusForbidden) {
		return response, model.Fail("reconnect_required", "The provider rejected the OAuth client credentials.")
	}
	return response, err
}

// oauthToken returns a cached token or posts one declared client-credentials request.
func oauthToken(ctx context.Context, client Doer, attempt Attempt, now time.Time) (string, error) {
	auth := attempt.Definition.Auth
	clientID, ok := attempt.Credentials[auth.ClientIDSlot]
	if !ok {
		return "", model.Fail("credential_required", "Save the OAuth client ID in Setup.")
	}
	clientSecret, ok := attempt.Credentials[auth.ClientSecretSlot]
	if !ok {
		return "", model.Fail("credential_required", "Save the OAuth client secret in Setup.")
	}
	tokenURL, err := destination.Parse(auth.TokenURL)
	if err != nil {
		return "", model.Fail("invalid_definition", "The OAuth token endpoint is invalid.")
	}
	tokenOrigin := destination.Origin(tokenURL)
	if clientID.Origin != tokenOrigin || clientSecret.Origin != tokenOrigin {
		return "", model.Fail("reconnect_required", "Review the OAuth token destination before sending saved credentials.")
	}
	if !validCredentialValue(clientID.Value) || !validCredentialValue(clientSecret.Value) {
		return "", model.Fail("invalid_credential", "Enter valid OAuth client credentials.")
	}
	key := tokenCacheKey(auth.TokenURL, auth.Scope, clientID.Value, clientSecret.Value)
	if token, ok := cachedOAuthToken(key, now); ok {
		return token, nil
	}
	style := auth.TokenStyle
	if style == "" {
		style = "basic_form"
	}
	styles := []string{style}
	if style != "form_body" {
		styles = append(styles, "form_body")
	}
	var body []byte
	var status int
	var retryAt *time.Time
	for _, candidate := range styles {
		body, status, retryAt, err = sendOAuthToken(ctx, client, tokenURL, auth, clientID.Value, clientSecret.Value, candidate, now)
		if err != nil || (status != http.StatusBadRequest && status != http.StatusUnsupportedMediaType) {
			break
		}
	}
	if err != nil {
		return "", err
	}
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return "", model.Fail("reconnect_required", "The provider rejected the OAuth client credentials.")
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		failure := model.Fail("provider_unavailable", "The token endpoint is unavailable. Try again later.")
		failure.RetryAt = retryAt
		return "", failure
	}
	var payload map[string]any
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.UseNumber()
	if decoder.Decode(&payload) != nil {
		return "", model.Fail("invalid_response", "The token endpoint returned invalid JSON.")
	}
	if auth.TokenResponseRoot != "" {
		if selected, ok := Pointer(payload, auth.TokenResponseRoot); ok {
			if object, ok := selected.(map[string]any); ok {
				payload = object
			}
		}
	}
	if _, ok := payload["access_token"]; !ok {
		if data, ok := payload["data"].(map[string]any); ok {
			payload = data
		}
	}
	token, ok := payload["access_token"].(string)
	if !ok || !validCredentialValue(token) {
		return "", model.Fail("invalid_response", "The token endpoint returned an invalid access token.")
	}
	if tokenType, _ := payload["token_type"].(string); tokenType != "" && !strings.EqualFold(tokenType, "bearer") {
		return "", model.Fail("invalid_response", "The token endpoint returned an unsupported token type.")
	}
	lifetime := oauthDefaultLifetime
	if number, ok := payload["expires_in"].(json.Number); ok {
		if seconds, err := number.Int64(); err == nil && seconds > 0 && seconds <= 86400 {
			lifetime = time.Duration(seconds) * time.Second
		}
	}
	cacheOAuthToken(key, token, now, lifetime)
	return token, nil
}

// sendOAuthToken performs one bounded token request in the requested wire style.
func sendOAuthToken(
	ctx context.Context,
	client Doer,
	tokenURL *url.URL,
	auth model.AuthConfig,
	clientID, clientSecret, style string,
	now time.Time,
) ([]byte, int, *time.Time, error) {
	var body []byte
	contentType := "application/x-www-form-urlencoded"
	switch style {
	case "json_body":
		payload := map[string]string{
			"grant_type":    "client_credentials",
			"client_id":     clientID,
			"client_secret": clientSecret,
		}
		if auth.Scope != "" {
			payload["scope"] = auth.Scope
		}
		body, _ = json.Marshal(payload)
		contentType = "application/json"
	case "form_body":
		form := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {clientID},
			"client_secret": {clientSecret},
		}
		if auth.Scope != "" {
			form.Set("scope", auth.Scope)
		}
		body = []byte(form.Encode())
	default:
		form := url.Values{"grant_type": {"client_credentials"}}
		if auth.Scope != "" {
			form.Set("scope", auth.Scope)
		}
		body = []byte(form.Encode())
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL.String(), strings.NewReader(string(body)))
	if err != nil {
		return nil, 0, nil, model.Fail("invalid_definition", "Could not construct the OAuth token request.")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", contentType)
	if style == "basic_form" {
		request.SetBasicAuth(clientID, clientSecret)
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, 0, nil, model.Fail("provider_unavailable", "The token endpoint could not be reached. Try again later.")
	}
	if response == nil || response.Body == nil {
		return nil, 0, nil, model.Fail("invalid_response", "The token endpoint returned no response.")
	}
	defer response.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(response.Body, oauthMaxResponseBytes+1))
	if readErr != nil || len(body) > oauthMaxResponseBytes {
		return nil, 0, nil, model.Fail("invalid_response", "The token endpoint response exceeded its size limit.")
	}
	return body, response.StatusCode, RetryAfter(response.Header.Get("Retry-After"), now), nil
}

// tokenCacheKey avoids retaining the client secret in the in-memory cache key.
func tokenCacheKey(tokenURL, scope, clientID, clientSecret string) string {
	secretHash := sha256.Sum256([]byte(clientSecret))
	return tokenURL + "\x00" + scope + "\x00" + clientID + "\x00" + hex.EncodeToString(secretHash[:])
}

// cachedOAuthToken returns a token that remains safe to use beyond the request deadline.
func cachedOAuthToken(key string, now time.Time) (string, bool) {
	tokenCache.Lock()
	defer tokenCache.Unlock()
	entry, ok := tokenCache.entries[key]
	if !ok || !entry.ExpiresAt.After(now.Add(10*time.Second)) {
		if ok {
			delete(tokenCache.entries, key)
		}
		return "", false
	}
	return entry.Value, true
}

// cacheOAuthToken stores a bounded in-memory token with an expiry safety margin.
func cacheOAuthToken(key, token string, now time.Time, lifetime time.Duration) {
	margin := oauthSafetyMargin
	if lifetime <= 2*oauthSafetyMargin {
		margin = lifetime / 2
	}
	tokenCache.Lock()
	defer tokenCache.Unlock()
	if len(tokenCache.entries) >= oauthCacheLimit {
		clear(tokenCache.entries)
	}
	tokenCache.entries[key] = cachedToken{Value: token, ExpiresAt: now.Add(lifetime - margin)}
}

// invalidateOAuthToken removes a rejected token so the next attempt fetches a fresh one.
func invalidateOAuthToken(attempt Attempt) {
	auth := attempt.Definition.Auth
	clientID, idOK := attempt.Credentials[auth.ClientIDSlot]
	clientSecret, secretOK := attempt.Credentials[auth.ClientSecretSlot]
	if !idOK || !secretOK {
		return
	}
	key := tokenCacheKey(auth.TokenURL, auth.Scope, clientID.Value, clientSecret.Value)
	tokenCache.Lock()
	delete(tokenCache.entries, key)
	tokenCache.Unlock()
}

// ClearOAuthTokenCache drops all in-memory client-credentials tokens at service shutdown.
func ClearOAuthTokenCache() {
	tokenCache.Lock()
	clear(tokenCache.entries)
	tokenCache.Unlock()
}
