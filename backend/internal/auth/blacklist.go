package auth

import (
	"sync"
	"time"
)

// TokenBlacklist tracks revoked tokens until they naturally expire
type TokenBlacklist struct {
	mu     sync.RWMutex
	tokens map[string]time.Time // token -> expiry time
}

func NewTokenBlacklist() *TokenBlacklist {
	bl := &TokenBlacklist{
		tokens: make(map[string]time.Time),
	}

	// Cleanup expired tokens every hour
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			bl.cleanup()
		}
	}()

	return bl
}

func (bl *TokenBlacklist) Add(token string, expiresAt time.Time) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.tokens[token] = expiresAt
}

func (bl *TokenBlacklist) IsBlacklisted(token string) bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	expiresAt, exists := bl.tokens[token]
	if !exists {
		return false
	}

	// Token is blacklisted if it hasn't expired yet
	return time.Now().Before(expiresAt)
}

func (bl *TokenBlacklist) cleanup() {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	now := time.Now()
	for token, expiresAt := range bl.tokens {
		if now.After(expiresAt) {
			delete(bl.tokens, token)
		}
	}
}
