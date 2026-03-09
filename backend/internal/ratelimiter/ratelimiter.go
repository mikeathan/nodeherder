package ratelimiter

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mutex     sync.Mutex
	lastWrite time.Time
	store     map[string]time.Time // In-memory store for device IDs and last write times
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		mutex:     sync.Mutex{},
		lastWrite: time.Time{},
		store:     map[string]time.Time{},
	}
}

func (rl *RateLimiter) AllowWrite(id string, rateLimit time.Duration) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// Check if device exists in the in-memory store
	if lastWrite, ok := rl.store[id]; ok {
		if now.Sub(lastWrite) < rateLimit {
			return false // Rate limit exceeded
		}
	}

	// Update lastWrite time and store in map
	rl.lastWrite = now
	rl.store[id] = now

	return true
}
