package auth_test

import (
	"net/http"
	"net/http/httptest"
	"node-herder/internal/auth"
	"testing"
)

func TestAuthMiddleware_ValidToken(t *testing.T) {
	jwtService := auth.NewJWTService(auth.WithDefaultJWTConfig())
	blacklist := auth.NewTokenBlacklist()
	token, err := jwtService.GenerateJWT("123", "test@example.com")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := auth.Auth(jwtService, blacklist)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetUserFromContext(r)
		if !ok {
			t.Fatal("user not found in context")
		}
		if user.UserID != "123" {
			t.Errorf("expected user 123, got %s", user.UserID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Result().StatusCode)
	}
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	jwtService := auth.NewJWTService(auth.WithDefaultJWTConfig())
	blacklist := auth.NewTokenBlacklist()

	handler := auth.Auth(jwtService, blacklist)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Result().StatusCode)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	jwtService := auth.NewJWTService(auth.WithDefaultJWTConfig())
	blacklist := auth.NewTokenBlacklist()

	handler := auth.Auth(jwtService, blacklist)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "invalid.token.here")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Result().StatusCode)
	}
}
