package kernel

import (
	"fmt"
)

// FIFOScheduler schedules workloads in first-in-first-out order
type FIFOScheduler struct {
	queue []Workload
}

// NewFIFOScheduler creates a new FIFO scheduler
func NewFIFOScheduler() *FIFOScheduler {
	return &FIFOScheduler{
		queue: []Workload{},
	}
}

// Add adds a workload to the queue
func (s *FIFOScheduler) Add(w Workload) {
	fmt.Printf("[FIFO] Queued PID %d (%s)\n", w.PID, w.ID)
	w.Status = "waiting"
	s.queue = append(s.queue, w)
}

// Run is a no-op; actual execution happens via Executor
func (s *FIFOScheduler) Run() {
	// Workloads are executed asynchronously by the Executor
}
