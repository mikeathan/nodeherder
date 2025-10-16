package auth_test

import (
	"net/http"
	"net/http/httptest"
	"node-herder/internal/auth"
	"testing"
)

func TestGoogleOAuth_HandleLogin(t *testing.T) {
	cfg := auth.WithDefaultJWTConfig()
	provider := auth.NewProvider(cfg)
	oauth := provider.OAuth()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/login", nil)

	oauth.HandleLogin(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("expected redirect, got %d", res.StatusCode)
	}

	cookie := res.Cookies()[0]
	if cookie.Name != auth.AuthStateCookie {
		t.Errorf("expected oauthstate cookie, got %s", cookie.Name)
	}
}
