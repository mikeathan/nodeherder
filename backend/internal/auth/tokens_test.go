package auth_test

import (
    "node-herder/internal/auth"
    "strings"
    "testing"
    "time"
)

func TestGenerateAndValidateJWT(t *testing.T) {
    cfg := auth.WithDefaultJWTConfig()
    jwtService := auth.NewJWTService(cfg)

    userID := "123"
    username := "test"

    // Generate token
    token, err := jwtService.GenerateJWT(userID, username)
    if err != nil {
        t.Fatalf("failed to generate JWT: %v", err)
    }

    if token == "" {
        t.Fatal("generated token is empty")
    }

    // Validate token
    claims, err := jwtService.ValidateJWT(token)
    if err != nil {
        t.Fatalf("failed to validate JWT: %v", err)
    }

    if claims.UserID != userID {
        t.Errorf("expected UserID %s, got %s", userID, claims.UserID)
    }

    if claims.Username != username {
        t.Errorf("expected Username %s, got %s", username, claims.Username)
    }
}

func TestValidateJWT_InvalidToken(t *testing.T) {
    cfg := auth.WithDefaultJWTConfig()
    jwtService := auth.NewJWTService(cfg)

    invalidToken := "not.a.valid.token"

    _, err := jwtService.ValidateJWT(invalidToken)
    if err == nil {
        t.Fatal("expected error for invalid token, got nil")
    }
}

func TestValidateJWT_TamperedToken(t *testing.T) {
    cfg := auth.WithDefaultJWTConfig()
    jwtService := auth.NewJWTService(cfg)

    userID := "123"
    username := "test"
    token, _ := jwtService.GenerateJWT(userID, username)

    parts := strings.Split(token, ".")
    if len(parts) != 3 {
        t.Fatal("token does not have 3 parts")
    }
    parts[1] = "tampered"
    tamperedToken := strings.Join(parts, ".")

    _, err := jwtService.ValidateJWT(tamperedToken)
    if err == nil {
        t.Fatal("expected error for tampered token, got nil")
    }
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
    cfg := auth.JWTConfig{
        Secret:     []byte("test_secret"),
        Expiration: -time.Hour,
    }
    jwtService := auth.NewJWTService(cfg)

    userID := "123"
    username := "test"
    token, err := jwtService.GenerateJWT(userID, username)
    if err != nil {
        t.Fatalf("failed to generate token: %v", err)
    }

    _, err = jwtService.ValidateJWT(token)
    if err == nil {
        t.Fatal("expected error for expired token, got nil")
    }
}
