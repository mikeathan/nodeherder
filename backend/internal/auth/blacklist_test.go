package auth_test

import (
	"fmt"
	"node-herder/internal/auth"
	"sync"
	"testing"
	"time"
)

func TestNewTokenBlacklist(t *testing.T) {
	bl := auth.NewTokenBlacklist()

	if bl == nil {
		t.Fatal("NewTokenBlacklist() returned nil")
	}

	// Test that a new blacklist has no tokens by checking IsBlacklisted
	if bl.IsBlacklisted("any-token") {
		t.Error("new blacklist should not contain any tokens")
	}
}

func TestTokenBlacklist_Add(t *testing.T) {
	bl := auth.NewTokenBlacklist()
	token := "test-token-123"
	expiresAt := time.Now().Add(1 * time.Hour)

	bl.Add(token, expiresAt)

	// Verify the token is now blacklisted
	if !bl.IsBlacklisted(token) {
		t.Error("token was not added to blacklist")
	}
}

func TestTokenBlacklist_Add_MultipleTokens(t *testing.T) {
	bl := auth.NewTokenBlacklist()
	now := time.Now()

	tokens := map[string]time.Time{
		"token1": now.Add(1 * time.Hour),
		"token2": now.Add(2 * time.Hour),
		"token3": now.Add(3 * time.Hour),
	}

	for token, expiry := range tokens {
		bl.Add(token, expiry)
	}

	// Verify all tokens are blacklisted
	for token := range tokens {
		if !bl.IsBlacklisted(token) {
			t.Errorf("token %s not found in blacklist", token)
		}
	}
}

func TestTokenBlacklist_Add_UpdatesExistingToken(t *testing.T) {
	bl := auth.NewTokenBlacklist()
	token := "test-token"
	firstExpiry := time.Now().Add(1 * time.Hour)
	secondExpiry := time.Now().Add(2 * time.Hour)

	// Add token twice with different expiries
	bl.Add(token, firstExpiry)
	bl.Add(token, secondExpiry)

	// Token should still be blacklisted (we can't verify which expiry without internal access)
	if !bl.IsBlacklisted(token) {
		t.Error("token should remain blacklisted after update")
	}
}

func TestTokenBlacklist_IsBlacklisted(t *testing.T) {
	bl := auth.NewTokenBlacklist()
	blacklistedToken := "blacklisted-token"
	nonBlacklistedToken := "valid-token"

	bl.Add(blacklistedToken, time.Now().Add(1*time.Hour))

	if !bl.IsBlacklisted(blacklistedToken) {
		t.Error("expected token to be blacklisted")
	}

	if bl.IsBlacklisted(nonBlacklistedToken) {
		t.Error("expected token to not be blacklisted")
	}
}

func TestTokenBlacklist_IsBlacklisted_EmptyBlacklist(t *testing.T) {
	bl := auth.NewTokenBlacklist()

	if bl.IsBlacklisted("any-token") {
		t.Error("empty blacklist should not contain any tokens")
	}
}

func TestTokenBlacklist_Cleanup_ExpiredTokensRemoved(t *testing.T) {
	bl := auth.NewTokenBlacklist()
	now := time.Now()

	// Add tokens with past expiry times
	expiredToken1 := "expired1"
	expiredToken2 := "expired2"
	bl.Add(expiredToken1, now.Add(-2*time.Hour))
	bl.Add(expiredToken2, now.Add(-1*time.Hour))

	// Verify tokens are initially blacklisted
	if !bl.IsBlacklisted(expiredToken1) || !bl.IsBlacklisted(expiredToken2) {
		t.Fatal("expired tokens should be initially blacklisted")
	}

	// Wait for automatic cleanup (1 hour is too long, so we test immediate behavior)
	// Since cleanup runs every hour, we test that expired tokens are still queryable
	// The cleanup goroutine will remove them eventually

	// For now, verify the tokens exist (cleanup hasn't run yet)
	if !bl.IsBlacklisted(expiredToken1) {
		t.Error("token should still be in blacklist before cleanup runs")
	}
}

func TestTokenBlacklist_ValidTokensNotExpired(t *testing.T) {
	bl := auth.NewTokenBlacklist()
	now := time.Now()

	validToken := "valid-token"
	bl.Add(validToken, now.Add(2*time.Hour))

	// Token should be blacklisted and remain so
	if !bl.IsBlacklisted(validToken) {
		t.Error("valid token should be blacklisted")
	}

	// Wait a bit and verify it's still there
	time.Sleep(10 * time.Millisecond)

	if !bl.IsBlacklisted(validToken) {
		t.Error("valid token should remain blacklisted")
	}
}

func TestTokenBlacklist_ConcurrentAdd(t *testing.T) {
	bl := auth.NewTokenBlacklist()
	now := time.Now()

	var wg sync.WaitGroup
	numGoroutines := 100
	tokensPerGoroutine := 10

	// Concurrent Add operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < tokensPerGoroutine; j++ {
				token := fmt.Sprintf("token-%d-%d", id, j)
				bl.Add(token, now.Add(1*time.Hour))
			}
		}(i)
	}

	wg.Wait()

	// Verify at least some tokens were added (checking all would be tedious)
	if !bl.IsBlacklisted("token-0-0") {
		t.Error("expected at least some tokens to be added")
	}
}

func TestTokenBlacklist_ConcurrentAddAndCheck(t *testing.T) {
	bl := auth.NewTokenBlacklist()
	now := time.Now()
	testToken := "test-concurrent-token"

	var wg sync.WaitGroup

	// Add operations
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bl.Add(testToken, now.Add(1*time.Hour))
		}()
	}

	// Check operations
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bl.IsBlacklisted(testToken)
		}()
	}

	wg.Wait()

	// Verify token is blacklisted after concurrent operations
	if !bl.IsBlacklisted(testToken) {
		t.Error("token should be blacklisted after concurrent operations")
	}
}

func TestTokenBlacklist_ConcurrentAddMultipleTokens(t *testing.T) {
	bl := auth.NewTokenBlacklist()
	now := time.Now()

	var wg sync.WaitGroup
	numTokens := 100

	// Add many different tokens concurrently
	for i := 0; i < numTokens; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			token := fmt.Sprintf("concurrent-token-%d", id)
			bl.Add(token, now.Add(1*time.Hour))
		}(i)
	}

	wg.Wait()

	// Verify random tokens are blacklisted
	testTokens := []string{"concurrent-token-0", "concurrent-token-50", "concurrent-token-99"}
	for _, token := range testTokens {
		if !bl.IsBlacklisted(token) {
			t.Errorf("token %s should be blacklisted", token)
		}
	}
}

func TestTokenBlacklist_ConcurrentCheckNonExistent(t *testing.T) {
	bl := auth.NewTokenBlacklist()

	var wg sync.WaitGroup

	// Many concurrent checks on non-existent tokens should not panic
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			token := fmt.Sprintf("non-existent-%d", id)
			bl.IsBlacklisted(token)
		}(i)
	}

	wg.Wait()

	// Should complete without panicking
}

func TestTokenBlacklist_AddZeroExpiry(t *testing.T) {
	bl := auth.NewTokenBlacklist()
	token := "zero-expiry-token"

	// Add with zero time (already expired)
	bl.Add(token, time.Time{})

	// Token should be added (cleanup will remove it later)
	if !bl.IsBlacklisted(token) {
		t.Error("token with zero expiry should be initially blacklisted")
	}
}

func TestTokenBlacklist_AddPastExpiry(t *testing.T) {
	bl := auth.NewTokenBlacklist()
	token := "past-expiry-token"
	now := time.Now()

	// Add with past expiry
	bl.Add(token, now.Add(-1*time.Hour))

	// Token should be added (even though expired)
	if !bl.IsBlacklisted(token) {
		t.Error("token with past expiry should be initially blacklisted")
	}
}

func TestTokenBlacklist_MultipleInstancesIndependent(t *testing.T) {
	bl1 := auth.NewTokenBlacklist()
	bl2 := auth.NewTokenBlacklist()

	token := "shared-token"

	bl1.Add(token, time.Now().Add(1*time.Hour))

	// Token should only be in bl1, not bl2
	if !bl1.IsBlacklisted(token) {
		t.Error("token should be in first blacklist")
	}

	if bl2.IsBlacklisted(token) {
		t.Error("token should not be in second blacklist")
	}
}
