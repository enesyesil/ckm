package kernel

import "fmt"

// MultilevelScheduler uses different schedulers for different workload types
type MultilevelScheduler struct {
	vmQueue   Scheduler
	taskQueue Scheduler
}

// NewMultilevelScheduler creates a new multilevel scheduler
func NewMultilevelScheduler(vmSched Scheduler, taskSched Scheduler) *MultilevelScheduler {
	return &MultilevelScheduler{
		vmQueue:   vmSched,
		taskQueue: taskSched,
	}
}

// Add routes workload to appropriate queue based on type
func (m *MultilevelScheduler) Add(w Workload) {
	if w.Type == "vm" {
		m.vmQueue.Add(w)
	} else {
		m.taskQueue.Add(w)
	}
}

// Run executes workloads from both queues (legacy mode)
func (m *MultilevelScheduler) Run() {
	fmt.Println("[>>] Starting Multilevel Scheduler...")
	// Run task queue first, then VM queue
	m.taskQueue.Run()
	m.vmQueue.Run()
	fmt.Println("[✓] Multilevel Scheduler complete")
}
