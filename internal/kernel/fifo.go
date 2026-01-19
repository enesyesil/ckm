package kernel

import (
	"fmt"
	"time"
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
	fmt.Printf("[+] Queued PID %d (%s)\n", w.PID, w.ID)
	w.Status = "waiting"
	s.queue = append(s.queue, w)
}

// Run executes workloads in FIFO order (legacy mode)
func (s *FIFOScheduler) Run() {
	for len(s.queue) > 0 {
		w := s.queue[0]
		s.queue = s.queue[1:]

		fmt.Printf("[>] Running PID %d (%s) for %v\n", w.PID, w.ID, w.CPUTime)
		w.Status = "running"
		time.Sleep(w.CPUTime)
		w.Status = "done"
		fmt.Printf("[✔] Completed PID %d (%s)\n", w.PID, w.ID)
	}
}
