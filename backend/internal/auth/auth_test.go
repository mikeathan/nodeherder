package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"node-herder/internal/auth"
	"testing"
)

func TestGoogleOAuth_HandleLogin(t *testing.T) {
	cfg := auth.WithDefaultJWTConfig()
	callbackURL := "http://localhost:8080/auth/callback"
	provider := auth.NewProvider(cfg, callbackURL)
	oauth := provider.OAuth()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/login", nil)

	oauth.HandleLogin(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.StatusCode)
	}

	// Expect JSON body with an 'url' field
	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["url"] == "" {
		t.Fatalf("expected non-empty oauth url in response")
	}

	// Expect both state and PKCE cookies to be set
	var stateCookie, pkceCookie *http.Cookie
	for _, c := range res.Cookies() {
		switch c.Name {
		case auth.CookieNameOAuthState:
			stateCookie = c
		case auth.CookieNameOAuthPKCE:
			pkceCookie = c
		}
	}
	if stateCookie == nil {
		t.Fatalf("expected %s cookie to be set", auth.CookieNameOAuthState)
	}
	if pkceCookie == nil {
		t.Fatalf("expected %s cookie to be set", auth.CookieNameOAuthPKCE)
	}
	if !stateCookie.HttpOnly || !pkceCookie.HttpOnly {
		t.Errorf("expected cookies to be HttpOnly")
	}
	if stateCookie.Value == "" || pkceCookie.Value == "" {
		t.Errorf("expected non-empty cookie values")
	}
	// pkce cookie must have format state.verifier and state must match the state cookie
	if dot := indexByteTest(pkceCookie.Value, '.'); dot <= 0 {
		t.Errorf("expected pkce cookie value to contain a dot separator")
	} else {
		if pkceCookie.Value[:dot] != stateCookie.Value {
			t.Errorf("expected pkce cookie state prefix to match state cookie")
		}
	}
}

// indexByteTest is a tiny helper for the test to avoid importing strings.
func indexByteTest(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
