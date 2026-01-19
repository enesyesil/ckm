package common

import (
	"errors"
	"testing"
	"time"
)

// TestCircuitBreakerClosed tests normal operation
func TestCircuitBreakerClosed(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)

	// Should succeed
	err := cb.Call(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if cb.GetState() != StateClosed {
		t.Errorf("Expected state Closed, got %d", cb.GetState())
	}
}

// TestCircuitBreakerOpen tests circuit opening after failures
func TestCircuitBreakerOpen(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)

	// Fail 3 times to open circuit
	for i := 0; i < 3; i++ {
		cb.Call(func() error {
			return errors.New("failure")
		})
	}

	if cb.GetState() != StateOpen {
		t.Errorf("Expected state Open, got %d", cb.GetState())
	}

	// Next call should fail with circuit open error
	err := cb.Call(func() error {
		return nil
	})

	if err != ErrCircuitOpen {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}
}

// TestCircuitBreakerHalfOpen tests half-open state after timeout
func TestCircuitBreakerHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(1, 10*time.Millisecond)

	// Fail once to open circuit
	cb.Call(func() error {
		return errors.New("failure")
	})

	// Wait for timeout
	time.Sleep(20 * time.Millisecond)

	// Next call should succeed and close circuit
	err := cb.Call(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if cb.GetState() != StateClosed {
		t.Errorf("Expected state Closed, got %d", cb.GetState())
	}
}

// TestCircuitBreakerReset tests reset on success
func TestCircuitBreakerReset(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)

	// Fail twice
	for i := 0; i < 2; i++ {
		cb.Call(func() error {
			return errors.New("failure")
		})
	}

	// Succeed once - should reset failure count
	cb.Call(func() error {
		return nil
	})

	// Fail once more - should not open circuit
	cb.Call(func() error {
		return errors.New("failure")
	})

	if cb.GetState() != StateClosed {
		t.Errorf("Expected state Closed after reset, got %d", cb.GetState())
	}
}
