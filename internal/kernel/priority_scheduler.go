package kernel

import (
	"fmt"
	"sort"
	"time"
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
	fmt.Println("[+] Queued:", w.ID)
	w.Status = "waiting"
	s.queue = append(s.queue, w)
}

// Run executes workloads sorted by priority (legacy mode)
func (s *PriorityScheduler) Run() {
	// Sort by priority (lower = higher priority)
	sort.SliceStable(s.queue, func(i, j int) bool {
		return s.queue[i].Priority < s.queue[j].Priority
	})

	for _, w := range s.queue {
		fmt.Printf("[>] Running %s (priority %d) for %v\n", w.ID, w.Priority, w.CPUTime)
		time.Sleep(w.CPUTime)
		fmt.Printf("[✔] Completed %s\n", w.ID)
	}
}
