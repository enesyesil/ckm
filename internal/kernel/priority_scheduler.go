package kernel

import (
	"fmt"
)

// PriorityScheduler schedules workloads by priority (lower number = higher priority)
type PriorityScheduler struct {
	queue []Workload
}

// NewPriorityScheduler creates a new priority scheduler
func NewPriorityScheduler() *PriorityScheduler {
	return &PriorityScheduler{
		queue: []Workload{},
	}
}

// Add adds a workload to the queue
func (s *PriorityScheduler) Add(w Workload) {
	fmt.Printf("[Priority] Queued: %s (priority %d)\n", w.ID, w.Priority)
	w.Status = "waiting"
	s.queue = append(s.queue, w)
}

// Run is a no-op; actual execution happens via Executor
func (s *PriorityScheduler) Run() {
	// Workloads are executed asynchronously by the Executor
}
