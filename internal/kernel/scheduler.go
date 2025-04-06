package kernel

import (
	"fmt"
	"time"
)

// Workload represents both tasks and VMs.
type Workload struct {
    ID        string
    Type      string // "task" or "vm"
    CPUTime   time.Duration
    MemoryMB  int
    Status    string // "waiting", "running", "done"
}

// An interface for pluggable scheduling algorithms.
type Scheduler interface {
    Add(Workload)
    Run()
}

// FIFOScheduler is a basic First-In-First-Out scheduler.
type FIFOScheduler struct {
    queue []Workload
}

// NewFIFOScheduler creates a new FIFO scheduler.
func NewFIFOScheduler() *FIFOScheduler {
    return &FIFOScheduler{
        queue: []Workload{},
    }
}

// Add appends a workload to the FIFO queue.
func (s *FIFOScheduler) Add(w Workload) {
    fmt.Println("[+] Queued:", w.ID)
    w.Status = "waiting"
    s.queue = append(s.queue, w)
}

// Run processes each workload in arrival order.
func (s *FIFOScheduler) Run() {
    for len(s.queue) > 0 {
        w := s.queue[0]
        s.queue = s.queue[1:]

        fmt.Printf("[>] Running %s (%s) for %v\n", w.ID, w.Type, w.CPUTime)
        w.Status = "running"
        time.Sleep(w.CPUTime) // Simulate CPU work
        w.Status = "done"
        fmt.Printf("[✔] Completed %s\n", w.ID)
    }
}
