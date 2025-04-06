package kernel

import (
	"fmt"
	"sync"
)

type MemoryManager struct {
    totalMB   int
    usedMB    int
    mu        sync.Mutex
}

func NewMemoryManager(totalMB int) *MemoryManager {
    return &MemoryManager{
        totalMB: totalMB,
        usedMB:  0,
    }
}

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

func (m *MemoryManager) Free(id string, mb int) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.usedMB -= mb
    if m.usedMB < 0 {
        m.usedMB = 0
    }
    fmt.Printf("[MEM] 🔁 Freed %dMB from %s (used: %dMB / %dMB)\n", mb, id, m.usedMB, m.totalMB)
}

func (m *MemoryManager) GetUsedMemory() int {
    m.mu.Lock()
    defer m.mu.Unlock()
    return m.usedMB
}