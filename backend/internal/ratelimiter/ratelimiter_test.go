package ratelimiter

import (
	"testing"
	"time"
)

func TestWindowRateLimiter_Capacity(t *testing.T) {
	limiter := NewWindowRateLimiter(2, time.Second)

	if !limiter.Allow() {
		t.Fatal("expected first allow to pass")
	}
	if !limiter.Allow() {
		t.Fatal("expected second allow to pass")
	}
	if limiter.Allow() {
		t.Fatal("expected third allow to be rate limited")
	}
}

func TestWindowRateLimiter_ResetsAfterWindow(t *testing.T) {
	window := 20 * time.Millisecond
	limiter := NewWindowRateLimiter(1, window)

	if !limiter.Allow() {
		t.Fatal("expected first allow to pass")
	}
	if limiter.Allow() {
		t.Fatal("expected second allow to be rate limited in same window")
	}

	time.Sleep(window + 10*time.Millisecond)

	if !limiter.Allow() {
		t.Fatal("expected allow to pass after window reset")
	}
}
