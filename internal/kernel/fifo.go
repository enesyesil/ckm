package kernel

import (
	"fmt"
	"time"
)

type FIFOScheduler struct {
	queue []Workload
}

func NewFIFOScheduler() *FIFOScheduler {
	return &FIFOScheduler{
		queue: []Workload{},
	}
}

func (s *FIFOScheduler) Add(w Workload) {
	fmt.Printf("[+] Queued PID %d (%s)\n", w.PID, w.ID)
	w.Status = "waiting"
	s.queue = append(s.queue, w)
}

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
