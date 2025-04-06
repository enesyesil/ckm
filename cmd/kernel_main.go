package main

import (
	"ckm/internal/common"
	"ckm/internal/kernel"
	"fmt"
	"time"
)

func main() {

	common.InitMetrics() // init the prometheus metrics

    // Create a Round Robin Scheduler with 1-second quantum
    scheduler := kernel.NewRoundRobinScheduler(1 * time.Second)
    memory := kernel.NewMemoryManager(1024)

    rawWorkloads, err := common.LoadWorkloads("configs/workloads.yaml")
    if err != nil {
        fmt.Println("Failed to load config:", err)
        return
    }

    var accepted []kernel.Workload

    for _, raw := range rawWorkloads {
        wl := kernel.Workload{
            ID:       raw.ID,
            Type:     raw.Type,
            CPUTime:  common.ParseCPUTime(raw.CPUTime),
            MemoryMB: raw.MemoryMB,
        }

        ok := memory.Allocate(wl.ID, wl.MemoryMB)
        if ok {
             // Update memory metric
            common.MemoryUsed.Set(float64(memory.GetUsedMemory()))

            // Track running count
            common.WorkloadsRunning.Inc()


            scheduler.Add(wl)
            accepted = append(accepted, wl)
        }
    }

    scheduler.Run()

    for _, wl := range accepted {

        memory.Free(wl.ID, wl.MemoryMB)

	   // Update metrics
	    common.WorkloadsRunning.Dec()
		common.WorkloadCompleted.WithLabelValues(wl.Type).Inc()
		common.MemoryUsed.Set(float64(memory.GetUsedMemory()))
    }

	fmt.Println("All workloads complete. Metrics available at http://localhost:9090/metrics")
}
