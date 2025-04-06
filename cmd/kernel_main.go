package main

import (
	"time"

	"../internal/kernel"
)

func main() {
    scheduler := kernel.NewFIFOScheduler()

    // Add simulated workloads
    scheduler.Add(kernel.Workload{
        ID:       "task-001",
        Type:     "task",
        CPUTime:  2 * time.Second,
        MemoryMB: 128,
    })

    scheduler.Add(kernel.Workload{
        ID:       "vm-001",
        Type:     "vm",
        CPUTime:  3 * time.Second,
        MemoryMB: 512,
    })

    scheduler.Run()
}
