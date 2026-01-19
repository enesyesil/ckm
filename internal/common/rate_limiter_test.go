package common

import (
	"testing"
	"time"
)

// TestRateLimiterAllow tests allowing requests
func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(10, 10) // 10 tokens/sec, 10 capacity

	// Should allow 10 requests immediately
	for i := 0; i < 10; i++ {
		if !rl.Allow() {
			t.Errorf("Expected request %d to be allowed", i)
		}
	}

	// 11th request should be denied
	if rl.Allow() {
		t.Error("Expected 11th request to be denied")
	}
}

// TestRateLimiterRefill tests token refill
func TestRateLimiterRefill(t *testing.T) {
	rl := NewRateLimiter(100, 1) // 100 tokens/sec, 1 capacity

	// Use the token
	rl.Allow()

	// Should be denied immediately
	if rl.Allow() {
		t.Error("Expected request to be denied")
	}

	// Wait for refill
	time.Sleep(20 * time.Millisecond)

	// Should be allowed after refill
	if !rl.Allow() {
		t.Error("Expected request to be allowed after refill")
	}
}

// TestRateLimiterWait tests blocking wait
func TestRateLimiterWait(t *testing.T) {
	rl := NewRateLimiter(100, 1) // 100 tokens/sec, 1 capacity

	// Use the token
	rl.Allow()

	// Wait should block and then succeed
	start := time.Now()
	rl.Wait()
	elapsed := time.Since(start)

	if elapsed < 5*time.Millisecond {
		t.Errorf("Expected Wait to block, elapsed: %v", elapsed)
	}
}
