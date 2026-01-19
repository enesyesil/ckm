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
	fmt.Println("[+] Queued:", w.ID)
	w.Status = "waiting"
	s.queue = append(s.queue, w)
}

// Run executes workloads in round-robin order (legacy mode - workloads executed via Executor in API mode)
func (s *RoundRobinScheduler) Run() {
	for len(s.queue) > 0 {
		w := s.queue[0]
		s.queue = s.queue[1:]

		timeSlice := s.quantum
		if w.CPUTime < s.quantum {
			timeSlice = w.CPUTime
		}

		fmt.Printf("[>] Running %s (%s) for %v\n", w.ID, w.Type, timeSlice)
		w.Status = "running"
		time.Sleep(timeSlice)
		w.CPUTime -= timeSlice

		if w.CPUTime > 0 {
			fmt.Printf("[↻] %s not finished, re-queueing (%v left)\n", w.ID, w.CPUTime)
			s.queue = append(s.queue, w)
		} else {
			w.Status = "done"
			fmt.Printf("[✔] Completed %s\n", w.ID)
		}
	}
}
