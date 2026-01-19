package common

import (
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota // Normal operation
	StateOpen                       // Failing, reject requests
	StateHalfOpen                   // Testing if service recovered
)

// CircuitBreaker prevents cascading failures by stopping requests when service is failing
type CircuitBreaker struct {
	maxFailures int           // Failures before opening circuit
	timeout     time.Duration // Time before attempting half-open
	failures    int           // Current failure count
	state       CircuitState
	lastFailure time.Time
	mu          sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		timeout:     timeout,
		state:       StateClosed,
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	
	// Check if circuit is open
	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.state = StateHalfOpen // Try again
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
	}
	cb.mu.Unlock()

	// Execute function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()
		
		// Open circuit if threshold reached
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}
		return err
	}

	// Success - reset if half-open
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.failures = 0
	} else {
		cb.failures = 0 // Reset on success
	}

	return nil
}

// GetState returns current circuit state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

var ErrCircuitOpen = &CircuitError{Message: "Circuit breaker is open"}

// CircuitError represents a circuit breaker error
type CircuitError struct {
	Message string
}

func (e *CircuitError) Error() string {
	return e.Message
}
