package kernel

import (
	"fmt"
	"time"
)

// FairWorkload extends Workload with runtime tracking for fairness
type FairWorkload struct {
	Workload
	RunTime time.Duration
}

// FairScheduler schedules workloads fairly based on total runtime
type FairScheduler struct {
	queue   []FairWorkload
	quantum time.Duration
}

// NewFairScheduler creates a new fair scheduler
func NewFairScheduler(quantum time.Duration) *FairScheduler {
	return &FairScheduler{
		queue:   []FairWorkload{},
		quantum: quantum,
	}
}

// Add adds a workload to the queue
func (s *FairScheduler) Add(w Workload) {
	fmt.Printf("[Fair] Queued: %s\n", w.ID)
	s.queue = append(s.queue, FairWorkload{Workload: w, RunTime: 0})
}

// Run is a no-op; actual execution happens via Executor
func (s *FairScheduler) Run() {
	// Workloads are executed asynchronously by the Executor
}
