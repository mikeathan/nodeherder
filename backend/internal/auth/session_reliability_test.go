package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"node-herder/internal/auth"
	"testing"
	"time"
)

func TestSessionCookieMatchesTokenExpiry(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("FRONTEND_BASE_URL", "http://localhost:4100")
	for _, duration := range []time.Duration{24 * time.Hour, 2 * time.Hour} {
		t.Run(duration.String(), func(t *testing.T) {
			jwt := auth.NewJWTService(auth.JWTConfig{Secret: []byte("test-secret"), Expiration: duration})
			offline := auth.NewOfflineOAuth(jwt, auth.NewTokenBlacklist(), "")
			w := httptest.NewRecorder()
			offline.(auth.OfflineAuth).HandleOfflineStart(w, httptest.NewRequest("GET", "/api/auth/offline/start", nil))
			cookie := assertCookie(t, w.Result().Cookies(), auth.CookieNameSession, true)
			if cookie == nil {
				t.Fatal("missing session cookie")
			}
			claims, err := jwt.ValidateJWT(cookie.Value)
			if err != nil {
				t.Fatal(err)
			}
			if cookie.MaxAge < int(duration.Seconds())-1 || cookie.MaxAge > int(duration.Seconds()) {
				t.Fatalf("cookie MaxAge=%d, want approximately %v", cookie.MaxAge, duration)
			}
			if !cookie.Expires.Equal(claims.ExpiresAt.Time) || !cookie.HttpOnly || cookie.Path != "/" || cookie.SameSite != http.SameSiteLaxMode {
				t.Fatalf("cookie persistence or flags do not match the session")
			}
		})
	}
}

func TestAuthModeAndLocalStartUseSamePolicy(t *testing.T) {
	for _, tc := range []struct {
		name, enabled, address, mode string
		startStatus                  int
	}{
		{"local enabled", "true", "192.168.1.2:1234", "local", 302},
		{"external", "true", "203.0.113.1:1234", "google", 403},
		{"disabled", "false", "192.168.1.2:1234", "google", 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET_KEY", "test-secret")
			t.Setenv("OFFLINE_STRICT_LOCAL", tc.enabled)
			t.Setenv("APP_ENV", "development")
			t.Setenv("FRONTEND_BASE_URL", "http://localhost:4100")
			p := auth.NewProvider(auth.WithDefaultJWTConfig(), "")
			modeProvider, ok := p.OAuth().(interface {
				HandleMode(http.ResponseWriter, *http.Request)
			})
			if !ok {
				t.Fatal("missing mode handler")
			}
			request := httptest.NewRequest("GET", "/api/auth/mode", nil)
			request.RemoteAddr = tc.address
			w := httptest.NewRecorder()
			modeProvider.HandleMode(w, request)
			var body struct {
				Mode auth.AuthMode `json:"mode"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if w.Code != 200 || string(body.Mode) != tc.mode || w.Header().Get("Cache-Control") != "no-store" || len(w.Result().Cookies()) != 0 {
				t.Fatalf("unexpected mode response: %d %s", w.Code, w.Body.String())
			}
			w = httptest.NewRecorder()
			p.OAuth().(auth.OfflineAuth).HandleOfflineStart(w, request)
			if w.Code != tc.startStatus {
				t.Fatalf("local start=%d, want %d", w.Code, tc.startStatus)
			}
			if tc.startStatus == 403 && len(w.Result().Cookies()) != 0 {
				t.Fatal("denied request issued cookie")
			}
		})
	}
}
