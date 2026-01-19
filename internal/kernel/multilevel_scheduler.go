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
		fmt.Printf("[Multilevel] Routing %s to VM queue\n", w.ID)
		m.vmQueue.Add(w)
	} else {
		fmt.Printf("[Multilevel] Routing %s to task queue\n", w.ID)
		m.taskQueue.Add(w)
	}
}

// Run is a no-op; actual execution happens via Executor
func (m *MultilevelScheduler) Run() {
	// Workloads are executed asynchronously by the Executor
}
