package auth

import (
	"sync"
	"time"
)

type TokenBlacklist struct {
	mu           sync.RWMutex
	tokens       map[string]time.Time // token -> expiry time
	listeners    map[string]map[uint64]func()
	nextListener uint64
	nextCleanup  time.Time
}

func NewTokenBlacklist() *TokenBlacklist {
	bl := &TokenBlacklist{
		tokens:      make(map[string]time.Time),
		listeners:   make(map[string]map[uint64]func()),
		nextCleanup: time.Now().Add(time.Hour),
	}

	return bl
}

func (bl *TokenBlacklist) Add(token string, expiresAt time.Time) {
	bl.mu.Lock()
	// Sweep on writes rather than retain an unowned goroutine per provider.
	if now := time.Now(); !now.Before(bl.nextCleanup) {
		bl.cleanupLocked(now)
		bl.nextCleanup = now.Add(time.Hour)
	}
	bl.tokens[token] = expiresAt
	callbacks := bl.listeners[token]
	delete(bl.listeners, token)
	bl.mu.Unlock()
	for _, callback := range callbacks {
		callback()
	}
}

func (bl *TokenBlacklist) IsBlacklisted(token string) bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	_, exists := bl.tokens[token]
	return exists
}

func (bl *TokenBlacklist) cleanupLocked(now time.Time) {
	for token, expiresAt := range bl.tokens {
		if now.After(expiresAt) {
			delete(bl.tokens, token)
		}
	}
}

// OnRevoke observes a token until revoked or canceled. Callbacks run outside the
// blacklist lock and must return promptly; the owner must call the returned cleanup.
func (bl *TokenBlacklist) OnRevoke(token string, callback func()) func() {
	bl.mu.Lock()
	if _, revoked := bl.tokens[token]; revoked {
		bl.mu.Unlock()
		callback()
		return func() {}
	}
	bl.nextListener++
	id := bl.nextListener
	if bl.listeners[token] == nil {
		bl.listeners[token] = make(map[uint64]func())
	}
	bl.listeners[token][id] = callback
	bl.mu.Unlock()
	return func() {
		bl.mu.Lock()
		defer bl.mu.Unlock()
		delete(bl.listeners[token], id)
		if len(bl.listeners[token]) == 0 {
			delete(bl.listeners, token)
		}
	}
}
