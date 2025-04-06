package kernel

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"ckm/internal/common"
)

// Workload is the core unit handled by all schedulers.
type Workload struct {
	ID       string        // Unique workload name (e.g., "task-001")
	PID      int           // Unique process identifier
	Type     string        // "task", "vm", or future "container"
	CPUTime  time.Duration // Time the task is expected to run
	MemoryMB int           // Memory requested
	Status   string        // "waiting", "running", "done"
	Priority int           // For priority schedulers
	FilePath string        // Optional: source script (e.g., .ipynb, .py)
}

// Scheduler is the interface implemented by all strategies (FIFO, RR, etc.)
type Scheduler interface {
	Add(Workload)
	Run()
}

// --- PID Generation (Thread-Safe) ---

var (
	pidCounter = 1000
	pidMutex   sync.Mutex
)

// NextPID returns a globally unique process ID (like a real OS)
func NextPID() int {
	pidMutex.Lock()
	defer pidMutex.Unlock()
	pidCounter++
	return pidCounter

}

// ClassifyWorkload assigns a type and priority based on file extension
func ClassifyWorkload(file string) (string, int) {
	ext := filepath.Ext(file)

	switch ext {
	case ".ipynb":
		return "notebook", 0 // highest priority, needs responsiveness
	case ".py", ".sh":
		return "task", 1 // medium priority
	case ".iso", ".qcow2":
		return "vm", 2 // lowest priority
	default:
		return "task", 2 // fallback
	}
}

	// ChooseScheduler selects the best scheduler based on workload type distribution
func ChooseScheduler(raws []common.RawWorkload) Scheduler {
	typeCounts := map[string]int{}

	for _, raw := range raws {
		typ, _ := ClassifyWorkload(raw.FilePath)
		typeCounts[typ]++
	}

	fmt.Println("[INFO] Workload types detected:", typeCounts)

	switch {
	case typeCounts["notebook"] > typeCounts["vm"] && typeCounts["notebook"] > 0:
		fmt.Println("[CKM] Using Fair Scheduler for notebook-heavy workload")
		return NewFairScheduler(1 * time.Second)

	case typeCounts["vm"] > typeCounts["task"] && typeCounts["vm"] > 0:
		fmt.Println("[CKM] Using Round Robin Scheduler for VM-heavy workload")
		return NewRoundRobinScheduler(1 * time.Second)

	default:
		fmt.Println("[CKM] Using Multilevel Scheduler for mixed workload")
		vmSched := NewRoundRobinScheduler(1 * time.Second)
		taskSched := NewPriorityScheduler()
		return NewMultilevelScheduler(vmSched, taskSched)
	}
}