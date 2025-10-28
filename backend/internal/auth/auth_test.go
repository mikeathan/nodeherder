package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"node-herder/internal/auth"
	"os"
	"strings"
	"testing"
)

// Test helpers
func setupTestProvider(t *testing.T, forceOfflineMode bool) *auth.Provider {
	t.Helper()

	// Set environment for testing
	// When OFFLINE_STRICT_LOCAL=false: always offline (dev mode)
	// When OFFLINE_STRICT_LOCAL=true: offline for non-local, online for local IPs
	if forceOfflineMode {
		os.Setenv("OFFLINE_STRICT_LOCAL", "false")
	} else {
		os.Setenv("OFFLINE_STRICT_LOCAL", "true")
	}

	cfg := auth.WithDefaultJWTConfig()
	callbackURL := "http://localhost:8080/api/auth/callback"
	return auth.NewProvider(cfg, callbackURL)
}

func assertCookie(t *testing.T, cookies []*http.Cookie, name string, shouldExist bool) *http.Cookie {
	t.Helper()
	for _, c := range cookies {
		if c.Name == name {
			if !shouldExist {
				t.Errorf("cookie %s should not exist", name)
			}
			return c
		}
	}
	if shouldExist {
		t.Errorf("expected cookie %s to exist", name)
	}
	return nil
}

func assertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) map[string]interface{} {
	t.Helper()
	if w.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, w.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return body
}

func TestGoogleOAuth_HandleLogin(t *testing.T) {
	// OFFLINE_STRICT_LOCAL=true + local IP (127.0.0.1) = online (Google OAuth)
	// because: decide() returns !IsLocalRequest() = !true = false (use online)
	os.Setenv("OFFLINE_STRICT_LOCAL", "true")

	provider := setupTestProvider(t, false)
	oauth := provider.OAuth()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/login", nil)
	// Use local IP to trigger online mode when OFFLINE_STRICT_LOCAL=true
	req.RemoteAddr = "127.0.0.1:12345"

	oauth.HandleLogin(w, req)

	body := assertJSONResponse(t, w, http.StatusOK)

	// Should return OAuth URL
	url, ok := body["url"].(string)
	if !ok || url == "" {
		t.Fatalf("expected non-empty oauth url in response")
	}

	// URL should contain Google OAuth parameters
	if !strings.Contains(url, "accounts.google.com") {
		t.Errorf("expected Google OAuth URL, got: %s", url)
	}
	if !strings.Contains(url, "code_challenge") {
		t.Errorf("expected PKCE code_challenge in URL")
	}

	// Should set state cookie
	stateCookie := assertCookie(t, w.Result().Cookies(), auth.CookieNameOAuthState, true)
	if stateCookie == nil {
		return
	}
	if !stateCookie.HttpOnly {
		t.Errorf("state cookie should be HttpOnly")
	}
	if stateCookie.Value == "" {
		t.Errorf("state cookie should have a value")
	}

	// Should set PKCE cookie
	pkceCookie := assertCookie(t, w.Result().Cookies(), auth.CookieNameOAuthPKCE, true)
	if pkceCookie == nil {
		return
	}
	if !pkceCookie.HttpOnly {
		t.Errorf("PKCE cookie should be HttpOnly")
	}

	// PKCE cookie format: state.verifier
	parts := strings.Split(pkceCookie.Value, ".")
	if len(parts) < 2 {
		t.Errorf("expected PKCE cookie format 'state.verifier', got: %s", pkceCookie.Value)
	} else if parts[0] != stateCookie.Value {
		t.Errorf("PKCE state prefix should match state cookie, expected %s, got %s", stateCookie.Value, parts[0])
	}
}

func TestGoogleOAuth_HandleCallback_MissingParams(t *testing.T) {
	os.Setenv("OFFLINE_STRICT_LOCAL", "true")
	provider := setupTestProvider(t, false)
	oauth := provider.OAuth()

	tests := []struct {
		name  string
		url   string
		error string
	}{
		{"no params", "/api/auth/callback", "Missing code or state"},
		{"only code", "/api/auth/callback?code=abc", "Missing code or state"},
		{"only state", "/api/auth/callback?state=xyz", "Missing code or state"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.url, nil)
			req.RemoteAddr = "127.0.0.1:12345" // Local IP to force online when OFFLINE_STRICT_LOCAL=true

			oauth.HandleCallback(w, req)

			body := assertJSONResponse(t, w, http.StatusBadRequest)
			if errMsg, ok := body["error"].(string); !ok || !strings.Contains(errMsg, tt.error) {
				t.Errorf("expected error containing %q, got: %v", tt.error, body["error"])
			}
		})
	}
}

func TestGoogleOAuth_HandleCallback_InvalidState(t *testing.T) {
	os.Setenv("OFFLINE_STRICT_LOCAL", "true")
	provider := setupTestProvider(t, false)
	oauth := provider.OAuth()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/auth/callback?code=abc&state=invalid", nil)
	req.RemoteAddr = "127.0.0.1:12345" // Local IP to force online

	oauth.HandleCallback(w, req)

	body := assertJSONResponse(t, w, http.StatusBadRequest)
	if errMsg, ok := body["error"].(string); !ok || !strings.Contains(errMsg, "Invalid state") {
		t.Errorf("expected 'Invalid state' error, got: %v", body["error"])
	}
}

func TestGoogleOAuth_HandleLogout_NoToken(t *testing.T) {
	provider := setupTestProvider(t, false)
	oauth := provider.OAuth()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/logout", nil)

	oauth.HandleLogout(w, req)

	body := assertJSONResponse(t, w, http.StatusInternalServerError)
	if errMsg, ok := body["error"].(string); !ok || !strings.Contains(errMsg, "no auth token found") {
		t.Errorf("expected 'no auth token found' error, got: %v", body["error"])
	}
}

func TestGoogleOAuth_HandleLogout_WithValidToken(t *testing.T) {
	provider := setupTestProvider(t, false)
	oauth := provider.OAuth()

	// Create a valid JWT
	cfg := auth.WithDefaultJWTConfig()
	jwtService := auth.NewJWTService(cfg)
	token, err := jwtService.GenerateJWT("test-user-id", "Test User")
	if err != nil {
		t.Fatalf("failed to generate JWT: %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieNameSession, Value: token})

	oauth.HandleLogout(w, req)

	body := assertJSONResponse(t, w, http.StatusOK)
	if status, ok := body["status"].(string); !ok || status != "success" {
		t.Errorf("expected status 'success', got: %v", body["status"])
	}

	// Session cookie should be cleared
	sessionCookie := assertCookie(t, w.Result().Cookies(), auth.CookieNameSession, true)
	if sessionCookie != nil && sessionCookie.MaxAge != -1 {
		t.Errorf("expected session cookie to be cleared (MaxAge=-1)")
	}
}

func TestGoogleOAuth_HandleMe_Unauthorized(t *testing.T) {
	provider := setupTestProvider(t, false)
	oauth := provider.OAuth()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/auth/me", nil)

	oauth.HandleMe(w, req)

	body := assertJSONResponse(t, w, http.StatusUnauthorized)
	if errMsg, ok := body["error"].(string); !ok || errMsg != "unauthorized" {
		t.Errorf("expected 'unauthorized' error, got: %v", body["error"])
	}
}

func TestOfflineOAuth_HandleLogin(t *testing.T) {
	// OFFLINE_STRICT_LOCAL=true + external IP = offline
	// because: decide() returns !IsLocalRequest() = !false = true (use offline)
	os.Setenv("OFFLINE_STRICT_LOCAL", "true")
	provider := setupTestProvider(t, false)
	oauth := provider.OAuth()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/login", nil)
	req.RemoteAddr = "203.0.113.1:12345" // External IP to trigger offline

	oauth.HandleLogin(w, req)

	body := assertJSONResponse(t, w, http.StatusOK)

	// Should return offline start URL
	url, ok := body["url"].(string)
	if !ok || url == "" {
		t.Fatalf("expected non-empty url in response")
	}

	if !strings.Contains(url, "/offline/start") {
		t.Errorf("expected offline start URL, got: %s", url)
	}

	// Should NOT set OAuth state/PKCE cookies (offline flow doesn't need them)
	assertCookie(t, w.Result().Cookies(), auth.CookieNameOAuthState, false)
	assertCookie(t, w.Result().Cookies(), auth.CookieNameOAuthPKCE, false)
}

func TestOfflineOAuth_HandleCallback_NotSupported(t *testing.T) {
	// OFFLINE_STRICT_LOCAL=true + external IP = offline (callback not supported)
	os.Setenv("OFFLINE_STRICT_LOCAL", "true")
	provider := setupTestProvider(t, false)
	oauth := provider.OAuth()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/auth/callback?code=abc&state=xyz", nil)
	req.RemoteAddr = "203.0.113.1:12345" // External IP for offline

	oauth.HandleCallback(w, req)

	body := assertJSONResponse(t, w, http.StatusNotFound)
	if errMsg, ok := body["error"].(string); !ok || !strings.Contains(errMsg, "not supported") {
		t.Errorf("expected 'not supported' error, got: %v", body["error"])
	}
}

func TestOfflineOAuth_HandleOfflineStart(t *testing.T) {
	os.Setenv("OFFLINE_STRICT_LOCAL", "true")
	provider := setupTestProvider(t, false)

	// Access the delegator and call HandleOfflineStart directly
	type offlineStarter interface {
		HandleOfflineStart(http.ResponseWriter, *http.Request)
	}

	delegator, ok := provider.OAuth().(offlineStarter)
	if !ok {
		t.Fatal("provider.OAuth() should support HandleOfflineStart")
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/auth/offline/start", nil)

	delegator.HandleOfflineStart(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Should set session cookie with JWT
	sessionCookie := assertCookie(t, w.Result().Cookies(), auth.CookieNameSession, true)
	if sessionCookie == nil {
		return
	}
	if sessionCookie.Value == "" {
		t.Errorf("session cookie should have a JWT token")
	}

	// Response should be HTML for popup
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("expected HTML content type, got: %s", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "<html") && !strings.Contains(body, "window.opener") {
		t.Errorf("expected HTML popup response")
	}
}

func TestOfflineOAuth_HandleLogout(t *testing.T) {
	os.Setenv("OFFLINE_STRICT_LOCAL", "true")
	provider := setupTestProvider(t, false)
	oauth := provider.OAuth()

	// Create a valid JWT for offline user
	cfg := auth.WithDefaultJWTConfig()
	jwtService := auth.NewJWTService(cfg)
	token, err := jwtService.GenerateJWT("local", "Offline User")
	if err != nil {
		t.Fatalf("failed to generate JWT: %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	req.RemoteAddr = "203.0.113.1:12345" // External IP for offline
	req.AddCookie(&http.Cookie{Name: auth.CookieNameSession, Value: token})

	oauth.HandleLogout(w, req)

	body := assertJSONResponse(t, w, http.StatusOK)
	if status, ok := body["status"].(string); !ok || status != "success" {
		t.Errorf("expected status 'success', got: %v", body["status"])
	}
}

func TestAuthDelegator_PicksCorrectImplementation(t *testing.T) {
	os.Setenv("OFFLINE_STRICT_LOCAL", "true")

	cfg := auth.WithDefaultJWTConfig()
	callbackURL := "http://localhost:8080/api/auth/callback"
	provider := auth.NewProvider(cfg, callbackURL)
	oauth := provider.OAuth()

	tests := []struct {
		name          string
		remoteAddr    string
		expectOffline bool
	}{
		{"localhost IP should use online", "127.0.0.1:12345", false},
		{"loopback IPv6 should use online", "[::1]:12345", false},
		{"external host should use offline", "203.0.113.1:12345", true},
		{"external IPv6 should use offline", "[2001:db8::1]:12345", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/auth/login", nil)
			req.RemoteAddr = tt.remoteAddr

			oauth.HandleLogin(w, req)

			body := assertJSONResponse(t, w, http.StatusOK)
			url, _ := body["url"].(string)

			isOffline := strings.Contains(url, "/offline/start")
			if isOffline != tt.expectOffline {
				t.Errorf("expected offline=%v, got offline=%v (url: %s)", tt.expectOffline, isOffline, url)
			}
		})
	}
}

func TestAuthDelegator_RegisterRoutes(t *testing.T) {
	cfg := auth.WithDefaultJWTConfig()
	callbackURL := "http://localhost:8080/api/auth/callback"
	provider := auth.NewProvider(cfg, callbackURL)

	// Mock registrar to track registered routes
	registered := []routeInfo{}

	mockRegistrar := &mockRouteRegistrar{
		routes: &registered,
	}

	provider.OAuth().RegisterRoutes(mockRegistrar)

	expectedRoutes := []routeInfo{
		{"POST", "/api/auth/login", true},
		{"POST", "/api/auth/logout", true},
		{"GET", "/api/auth/callback", true},
		{"GET", "/api/auth/offline/start", true},
		{"GET", "/api/auth/me", false},
	}

	if len(registered) != len(expectedRoutes) {
		t.Fatalf("expected %d routes, got %d", len(expectedRoutes), len(registered))
	}

	for i, expected := range expectedRoutes {
		if registered[i].method != expected.method ||
			registered[i].path != expected.path ||
			registered[i].public != expected.public {
			t.Errorf("route %d: expected %+v, got %+v", i, expected, registered[i])
		}
	}
}

// Mock route registrar for testing
type routeInfo struct {
	method string
	path   string
	public bool
}

type mockRouteRegistrar struct {
	routes *[]routeInfo
}

func (m *mockRouteRegistrar) GET(path string, handler http.Handler) {
	*m.routes = append(*m.routes, routeInfo{"GET", path, false})
}

func (m *mockRouteRegistrar) POST(path string, handler http.Handler) {
	*m.routes = append(*m.routes, routeInfo{"POST", path, false})
}

func (m *mockRouteRegistrar) PublicGET(path string, handler http.Handler) {
	*m.routes = append(*m.routes, routeInfo{"GET", path, true})
}

func (m *mockRouteRegistrar) PublicPOST(path string, handler http.Handler) {
	*m.routes = append(*m.routes, routeInfo{"POST", path, true})
}
