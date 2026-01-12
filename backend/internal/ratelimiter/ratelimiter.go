package ratelimiter

import (
	"sync"
	"time"
)

type DeviceRateLimiter struct {
	mutex     sync.Mutex
	lastWrite time.Time
	store     map[string]time.Time // In-memory store for device IDs and last write times
}

func NewDeviceRateLimiter() *DeviceRateLimiter {
	return &DeviceRateLimiter{
		mutex:     sync.Mutex{},
		lastWrite: time.Time{},
		store:     map[string]time.Time{},
	}
}

func (rl *DeviceRateLimiter) AllowWrite(id string, rateLimit time.Duration) bool {
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

// fixed window rate limiter
type WindowRateLimiter struct {
	mu          sync.Mutex
	capacity    int           // max requests per window
	remaining   int           // requests left in current window
	window      time.Duration // length of the window
	windowStart time.Time     // start of the current window
}

func NewWindowRateLimiter(capacity int, window time.Duration) *WindowRateLimiter {
	now := time.Now()
	return &WindowRateLimiter{
		capacity:    capacity,
		remaining:   capacity,
		window:      window,
		windowStart: now,
	}
}

func (rl *WindowRateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Reset window if expired
	if now.Sub(rl.windowStart) >= rl.window {
		rl.remaining = rl.capacity
		rl.windowStart = now
	}

	if rl.remaining <= 0 {
		return false
	}

	rl.remaining--
	return true
}
