package kernel

import (
	"sync"
	"time"
)

// WorkloadStore manages workload state in memory (thread-safe)
type WorkloadStore struct {
	workloads map[string]*Workload
	mu        sync.RWMutex
}

// NewWorkloadStore creates a new workload store
func NewWorkloadStore() *WorkloadStore {
	return &WorkloadStore{
		workloads: make(map[string]*Workload),
	}
}

// Add stores a workload
func (s *WorkloadStore) Add(w *Workload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w.CreatedAt = time.Now()
	s.workloads[w.ID] = w
}

// Get retrieves a workload by ID
func (s *WorkloadStore) Get(id string) (*Workload, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, ok := s.workloads[id]
	return w, ok
}

// GetAll returns all workloads
func (s *WorkloadStore) GetAll() []*Workload {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Workload, 0, len(s.workloads))
	for _, w := range s.workloads {
		result = append(result, w)
	}
	return result
}

// Update updates workload status
func (s *WorkloadStore) Update(id string, status string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	w, ok := s.workloads[id]
	if !ok {
		return false
	}
	w.Status = status
	if status == "done" || status == "failed" {
		w.CompletedAt = time.Now()
	}
	return true
}

// Delete removes a workload
func (s *WorkloadStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.workloads, id)
}
