package kernel

import (
	"fmt"
	"sort"
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
	fmt.Println("[+] Queued:", w.ID)
	s.queue = append(s.queue, FairWorkload{Workload: w, RunTime: 0})
}

// Run executes workloads fairly by prioritizing least runtime (legacy mode)
func (s *FairScheduler) Run() {
	for len(s.queue) > 0 {
		// Sort by least total runtime (fairness)
		sort.SliceStable(s.queue, func(i, j int) bool {
			return s.queue[i].RunTime < s.queue[j].RunTime
		})

		w := &s.queue[0]
		runTime := s.quantum
		if w.CPUTime < runTime {
			runTime = w.CPUTime
		}

		fmt.Printf("[>] Running %s (fair slice %v)\n", w.ID, runTime)
		time.Sleep(runTime)

		w.CPUTime -= runTime
		w.RunTime += runTime

		if w.CPUTime <= 0 {
			fmt.Printf("[✔] Completed %s\n", w.ID)
			s.queue = s.queue[1:] // Remove completed workload
		}
	}
}
