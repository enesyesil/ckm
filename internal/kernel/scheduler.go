package kernel

import (
	"sync"
	"time"
)

// Workload is the core unit handled by all schedulers
type Workload struct {
	ID          string        // Unique identifier
	PID         int           // Process ID
	Type        string        // "container", "task", "vm"
	CPUTime     time.Duration // Expected execution time
	MemoryMB    int           // Memory limit in MB
	Status      string        // "waiting", "running", "done", "failed"
	Priority    int           // Scheduling priority (lower = higher)
	FilePath    string        // Source file path
	Image       string        // Docker image name
	Command     []string      // Container command
	CreatedAt   time.Time     // Creation timestamp
	StartedAt   time.Time     // Start timestamp
	CompletedAt time.Time     // Completion timestamp
	ContainerID string        // Docker container ID
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