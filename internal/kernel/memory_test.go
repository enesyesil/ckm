package kernel

import (
	"testing"
)

// TestMemoryManagerAllocate tests memory allocation
func TestMemoryManagerAllocate(t *testing.T) {
	m := NewMemoryManager(1024)

	// Should succeed
	if !m.Allocate("test-1", 256) {
		t.Error("Expected allocation to succeed")
	}

	if m.GetUsedMemory() != 256 {
		t.Errorf("Expected 256 MB used, got %d", m.GetUsedMemory())
	}
}

// TestMemoryManagerAllocateFull tests allocation when memory is full
func TestMemoryManagerAllocateFull(t *testing.T) {
	m := NewMemoryManager(512)

	// First allocation should succeed
	if !m.Allocate("test-1", 512) {
		t.Error("Expected first allocation to succeed")
	}

	// Second allocation should fail
	if m.Allocate("test-2", 1) {
		t.Error("Expected second allocation to fail")
	}
}

// TestMemoryManagerFree tests memory deallocation
func TestMemoryManagerFree(t *testing.T) {
	m := NewMemoryManager(1024)

	m.Allocate("test-1", 512)
	m.Free("test-1", 512)

	if m.GetUsedMemory() != 0 {
		t.Errorf("Expected 0 MB used after free, got %d", m.GetUsedMemory())
	}
}

// TestMemoryManagerConcurrent tests concurrent access
func TestMemoryManagerConcurrent(t *testing.T) {
	m := NewMemoryManager(10000)

	done := make(chan bool)

	// Allocate from multiple goroutines
	for i := 0; i < 100; i++ {
		go func(id int) {
			m.Allocate("test", 10)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	if m.GetUsedMemory() != 1000 {
		t.Errorf("Expected 1000 MB used, got %d", m.GetUsedMemory())
	}
}
