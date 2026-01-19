package kernel

import (
	"testing"
	"time"
)

// TestWorkloadStoreAdd tests adding workloads
func TestWorkloadStoreAdd(t *testing.T) {
	s := NewWorkloadStore()

	w := &Workload{ID: "test-1", Type: "task"}
	s.Add(w)

	got, ok := s.Get("test-1")
	if !ok {
		t.Error("Expected workload to be found")
	}
	if got.ID != "test-1" {
		t.Errorf("Expected ID test-1, got %s", got.ID)
	}
	if got.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}

// TestWorkloadStoreGet tests retrieving workloads
func TestWorkloadStoreGet(t *testing.T) {
	s := NewWorkloadStore()

	// Should not find non-existent workload
	_, ok := s.Get("non-existent")
	if ok {
		t.Error("Expected workload to not be found")
	}
}

// TestWorkloadStoreGetAll tests retrieving all workloads
func TestWorkloadStoreGetAll(t *testing.T) {
	s := NewWorkloadStore()

	s.Add(&Workload{ID: "test-1"})
	s.Add(&Workload{ID: "test-2"})
	s.Add(&Workload{ID: "test-3"})

	all := s.GetAll()
	if len(all) != 3 {
		t.Errorf("Expected 3 workloads, got %d", len(all))
	}
}

// TestWorkloadStoreUpdate tests updating workload status
func TestWorkloadStoreUpdate(t *testing.T) {
	s := NewWorkloadStore()

	w := &Workload{ID: "test-1", Status: "waiting"}
	s.Add(w)

	// Update status
	if !s.Update("test-1", "running") {
		t.Error("Expected update to succeed")
	}

	got, _ := s.Get("test-1")
	if got.Status != "running" {
		t.Errorf("Expected status running, got %s", got.Status)
	}
}

// TestWorkloadStoreUpdateCompleted tests updating to completed status
func TestWorkloadStoreUpdateCompleted(t *testing.T) {
	s := NewWorkloadStore()

	w := &Workload{ID: "test-1", Status: "running"}
	s.Add(w)

	// Update to done
	s.Update("test-1", "done")

	got, _ := s.Get("test-1")
	if got.CompletedAt.IsZero() {
		t.Error("Expected CompletedAt to be set")
	}
}

// TestWorkloadStoreDelete tests deleting workloads
func TestWorkloadStoreDelete(t *testing.T) {
	s := NewWorkloadStore()

	s.Add(&Workload{ID: "test-1"})
	s.Delete("test-1")

	_, ok := s.Get("test-1")
	if ok {
		t.Error("Expected workload to be deleted")
	}
}

// TestWorkloadStoreConcurrent tests concurrent access
func TestWorkloadStoreConcurrent(t *testing.T) {
	s := NewWorkloadStore()
	done := make(chan bool)

	// Add from multiple goroutines
	for i := 0; i < 100; i++ {
		go func(id int) {
			s.Add(&Workload{ID: "test"})
			time.Sleep(time.Millisecond)
			s.GetAll()
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}
