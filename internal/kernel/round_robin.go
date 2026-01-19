package kernel

import (
	"fmt"
	"time"
)

// RoundRobinScheduler schedules workloads in round-robin fashion with time quantum
type RoundRobinScheduler struct {
	queue   []Workload
	quantum time.Duration
}

// NewRoundRobinScheduler creates a new round-robin scheduler
func NewRoundRobinScheduler(quantum time.Duration) *RoundRobinScheduler {
	return &RoundRobinScheduler{
		queue:   []Workload{},
		quantum: quantum,
	}
}

// Add adds a workload to the queue
func (s *RoundRobinScheduler) Add(w Workload) {
	fmt.Printf("[RR] Queued: %s\n", w.ID)
	w.Status = "waiting"
	s.queue = append(s.queue, w)
}

// Run is a no-op; actual execution happens via Executor
func (s *RoundRobinScheduler) Run() {
	// Workloads are executed asynchronously by the Executor
}
