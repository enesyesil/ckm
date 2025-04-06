package kernel

import (
	"fmt"
	"sort"
	"time"
)

type FairWorkload struct {
	Workload
	RunTime time.Duration
}

type FairScheduler struct {
	queue   []FairWorkload
	quantum time.Duration
}

func NewFairScheduler(quantum time.Duration) *FairScheduler {
	return &FairScheduler{
		queue:   []FairWorkload{},
		quantum: quantum,
	}
}

func (s *FairScheduler) Add(w Workload) {
	fmt.Println("[+] Queued:", w.ID)
	s.queue = append(s.queue, FairWorkload{Workload: w, RunTime: 0})
}

func (s *FairScheduler) Run() {
	for len(s.queue) > 0 {
		// Sort by least total runtime (simulating fairness)
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
			s.queue = s.queue[1:] // remove from queue
		}
	}
}
