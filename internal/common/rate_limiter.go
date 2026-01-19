package common

import (
	"sync"
	"time"
)

// RateLimiter implements token bucket algorithm for rate limiting
type RateLimiter struct {
	tokens     float64    // Current tokens
	capacity   float64    // Max tokens
	rate       float64    // Tokens per second
	lastRefill time.Time  // Last refill time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate float64, capacity float64) *RateLimiter {
	return &RateLimiter{
		tokens:     capacity,
		capacity:  capacity,
		rate:       rate,
		lastRefill: time.Now(),
	}
}

// Allow checks if request is allowed (consumes one token)
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens = min(rl.capacity, rl.tokens+elapsed*rl.rate)
	rl.lastRefill = now

	// Check if token available
	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}

	return false
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait() {
	for !rl.Allow() {
		time.Sleep(10 * time.Millisecond)
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
