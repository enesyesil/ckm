package kernel

import (
	"testing"
)

// TestCGroupManagerCreateCGroup tests cgroup creation
func TestCGroupManagerCreateCGroup(t *testing.T) {
	cgm := NewCGroupManager(2048)

	cg := cgm.CreateCGroup("test", 1024, 512)

	if cg.Name != "test" {
		t.Errorf("Expected name test, got %s", cg.Name)
	}
	if cg.CPUShares != 1024 {
		t.Errorf("Expected CPUShares 1024, got %d", cg.CPUShares)
	}
	if cg.MemoryMB != 512 {
		t.Errorf("Expected MemoryMB 512, got %d", cg.MemoryMB)
	}
}

// TestCGroupManagerAllocateMemory tests memory allocation in cgroup
func TestCGroupManagerAllocateMemory(t *testing.T) {
	cgm := NewCGroupManager(2048)
	cgm.CreateCGroup("test", 1024, 512)

	// Should succeed
	if !cgm.AllocateMemory("test", 256) {
		t.Error("Expected allocation to succeed")
	}

	cg, _ := cgm.GetCGroup("test")
	if cg.MemoryUsed != 256 {
		t.Errorf("Expected MemoryUsed 256, got %d", cg.MemoryUsed)
	}
}

// TestCGroupManagerAllocateMemoryOOM tests OOM condition
func TestCGroupManagerAllocateMemoryOOM(t *testing.T) {
	cgm := NewCGroupManager(2048)
	cgm.CreateCGroup("test", 1024, 256)

	// First allocation should succeed
	if !cgm.AllocateMemory("test", 256) {
		t.Error("Expected first allocation to succeed")
	}

	// Second allocation should fail (OOM)
	if cgm.AllocateMemory("test", 1) {
		t.Error("Expected second allocation to fail (OOM)")
	}
}

// TestCGroupManagerFreeMemory tests memory deallocation
func TestCGroupManagerFreeMemory(t *testing.T) {
	cgm := NewCGroupManager(2048)
	cgm.CreateCGroup("test", 1024, 512)

	cgm.AllocateMemory("test", 256)
	cgm.FreeMemory("test", 256)

	cg, _ := cgm.GetCGroup("test")
	if cg.MemoryUsed != 0 {
		t.Errorf("Expected MemoryUsed 0, got %d", cg.MemoryUsed)
	}
}

// TestCGroupManagerGetCGroup tests retrieving cgroup
func TestCGroupManagerGetCGroup(t *testing.T) {
	cgm := NewCGroupManager(2048)
	cgm.CreateCGroup("test", 1024, 512)

	cg, ok := cgm.GetCGroup("test")
	if !ok {
		t.Error("Expected cgroup to be found")
	}
	if cg.Name != "test" {
		t.Errorf("Expected name test, got %s", cg.Name)
	}

	// Non-existent cgroup
	_, ok = cgm.GetCGroup("non-existent")
	if ok {
		t.Error("Expected cgroup to not be found")
	}
}

// TestCGroupManagerGlobalAllocate tests global memory allocation
func TestCGroupManagerGlobalAllocate(t *testing.T) {
	cgm := NewCGroupManager(1024)

	// Should succeed
	if !cgm.Allocate("test-1", 256) {
		t.Error("Expected allocation to succeed")
	}

	if cgm.GetUsedMemory() != 256 {
		t.Errorf("Expected 256 MB used, got %d", cgm.GetUsedMemory())
	}
}

// TestCGroupManagerGlobalAllocateFull tests allocation when memory is full
func TestCGroupManagerGlobalAllocateFull(t *testing.T) {
	cgm := NewCGroupManager(512)

	// First allocation should succeed
	if !cgm.Allocate("test-1", 512) {
		t.Error("Expected first allocation to succeed")
	}

	// Second allocation should fail
	if cgm.Allocate("test-2", 1) {
		t.Error("Expected second allocation to fail")
	}
}

// TestCGroupManagerGlobalFree tests global memory deallocation
func TestCGroupManagerGlobalFree(t *testing.T) {
	cgm := NewCGroupManager(1024)

	cgm.Allocate("test-1", 512)
	cgm.Free("test-1", 512)

	if cgm.GetUsedMemory() != 0 {
		t.Errorf("Expected 0 MB used after free, got %d", cgm.GetUsedMemory())
	}
}
