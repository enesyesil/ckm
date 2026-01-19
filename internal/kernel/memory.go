package kernel

import (
	"fmt"
	"sync"
)

// MemoryManager manages memory allocation (thread-safe)
type MemoryManager struct {
	totalMB int
	usedMB  int
	mu      sync.Mutex
}

// NewMemoryManager creates a new memory manager with total capacity
func NewMemoryManager(totalMB int) *MemoryManager {
	return &MemoryManager{
		totalMB: totalMB,
		usedMB:  0,
	}
}

// Allocate allocates memory for a workload (returns false if insufficient)
func (m *MemoryManager) Allocate(id string, mb int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.usedMB+mb > m.totalMB {
		fmt.Printf("[MEM] Not enough memory for %s (%dMB needed, %dMB available)\n", id, mb, m.totalMB-m.usedMB)
		return false
	}

	m.usedMB += mb
	fmt.Printf("[MEM]  Allocated %dMB to %s (used: %dMB / %dMB)\n", mb, id, m.usedMB, m.totalMB)
	return true
}

// Free frees memory allocated to a workload
func (m *MemoryManager) Free(id string, mb int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.usedMB -= mb
	if m.usedMB < 0 {
		m.usedMB = 0
	}
	fmt.Printf("[MEM] 🔁 Freed %dMB from %s (used: %dMB / %dMB)\n", mb, id, m.usedMB, m.totalMB)
}

// GetUsedMemory returns current memory usage
func (m *MemoryManager) GetUsedMemory() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.usedMB
}