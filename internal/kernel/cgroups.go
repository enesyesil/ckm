package kernel

import (
	"fmt"
	"sync"
)

// CGroup represents a control group (like Linux cgroups) for resource limits
type CGroup struct {
	Name       string
	CPUShares  int64 // CPU weight (1024 = 1 CPU)
	MemoryMB   int64 // Memory limit in MB
	MemoryUsed int64 // Current memory usage
	mu         sync.RWMutex
}

// CGroupManager manages cgroups and resource limits
type CGroupManager struct {
	cgroups  map[string]*CGroup
	totalMB  int64 // Total system memory
	usedMB   int64 // Total used memory across all cgroups
	mu       sync.RWMutex
}

// NewCGroupManager creates a new cgroup manager with total memory capacity
func NewCGroupManager(totalMB int64) *CGroupManager {
	return &CGroupManager{
		cgroups: make(map[string]*CGroup),
		totalMB: totalMB,
		usedMB:  0,
	}
}

// CreateCGroup creates a new cgroup with resource limits
func (cgm *CGroupManager) CreateCGroup(name string, cpuShares int64, memoryMB int64) *CGroup {
	cgm.mu.Lock()
	defer cgm.mu.Unlock()

	cg := &CGroup{
		Name:       name,
		CPUShares:  cpuShares,
		MemoryMB:   memoryMB,
		MemoryUsed: 0,
	}
	cgm.cgroups[name] = cg
	return cg
}

// AllocateMemory allocates memory in a cgroup (returns false on OOM)
func (cgm *CGroupManager) AllocateMemory(cgroupName string, mb int64) bool {
	cgm.mu.RLock()
	cg, ok := cgm.cgroups[cgroupName]
	cgm.mu.RUnlock()

	if !ok {
		return false
	}

	cg.mu.Lock()
	defer cg.mu.Unlock()

	// Check OOM condition
	if cg.MemoryUsed+mb > cg.MemoryMB {
		return false
	}

	cg.MemoryUsed += mb
	return true
}

// FreeMemory frees memory in a cgroup
func (cgm *CGroupManager) FreeMemory(cgroupName string, mb int64) {
	cgm.mu.RLock()
	cg, ok := cgm.cgroups[cgroupName]
	cgm.mu.RUnlock()

	if !ok {
		return
	}

	cg.mu.Lock()
	defer cg.mu.Unlock()
	cg.MemoryUsed -= mb
	if cg.MemoryUsed < 0 {
		cg.MemoryUsed = 0
	}
}

// GetCGroup returns a cgroup by name
func (cgm *CGroupManager) GetCGroup(name string) (*CGroup, bool) {
	cgm.mu.RLock()
	defer cgm.mu.RUnlock()
	cg, ok := cgm.cgroups[name]
	return cg, ok
}

// Allocate allocates memory from the global pool (simple interface for workloads)
func (cgm *CGroupManager) Allocate(id string, mb int) bool {
	cgm.mu.Lock()
	defer cgm.mu.Unlock()

	if cgm.usedMB+int64(mb) > cgm.totalMB {
		fmt.Printf("[MEM] Not enough memory for %s (%dMB needed, %dMB available)\n", id, mb, cgm.totalMB-cgm.usedMB)
		return false
	}

	cgm.usedMB += int64(mb)
	fmt.Printf("[MEM] Allocated %dMB to %s (used: %dMB / %dMB)\n", mb, id, cgm.usedMB, cgm.totalMB)
	return true
}

// Free frees memory back to the global pool
func (cgm *CGroupManager) Free(id string, mb int) {
	cgm.mu.Lock()
	defer cgm.mu.Unlock()

	cgm.usedMB -= int64(mb)
	if cgm.usedMB < 0 {
		cgm.usedMB = 0
	}
	fmt.Printf("[MEM] Freed %dMB from %s (used: %dMB / %dMB)\n", mb, id, cgm.usedMB, cgm.totalMB)
}

// GetUsedMemory returns current total memory usage
func (cgm *CGroupManager) GetUsedMemory() int {
	cgm.mu.RLock()
	defer cgm.mu.RUnlock()
	return int(cgm.usedMB)
}

// GetTotalMemory returns total memory capacity
func (cgm *CGroupManager) GetTotalMemory() int {
	cgm.mu.RLock()
	defer cgm.mu.RUnlock()
	return int(cgm.totalMB)
}
