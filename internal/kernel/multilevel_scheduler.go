package kernel

import "fmt"

type MultilevelScheduler struct {
	vmQueue   Scheduler
	taskQueue Scheduler
}

func NewMultilevelScheduler(vmSched Scheduler, taskSched Scheduler) *MultilevelScheduler {
	return &MultilevelScheduler{
		vmQueue:   vmSched,
		taskQueue: taskSched,
	}
}

func (m *MultilevelScheduler) Add(w Workload) {
	if w.Type == "vm" {
		m.vmQueue.Add(w)
	} else {
		m.taskQueue.Add(w)
	}
}

func (m *MultilevelScheduler) Run() {
	fmt.Println("[>>] Starting Multilevel Scheduler...")
	// Alternate or run sequentially (simple strategy for now)
	m.taskQueue.Run()
	m.vmQueue.Run()
	fmt.Println("[✓] Multilevel Scheduler complete")
}
