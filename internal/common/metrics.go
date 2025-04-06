package common

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    WorkloadsRunning = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "ckm_workloads_running_total",
            Help: "Number of workloads currently running",
        })

    MemoryUsed = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "ckm_memory_usage_megabytes",
            Help: "Simulated memory usage in MB",
        })

    PageFaults = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "ckm_page_faults_total",
            Help: "Number of simulated page faults",
        })

    WorkloadCompleted = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ckm_workloads_completed_total",
            Help: "Total completed workloads by type",
        },
        []string{"type"},
    )
)

func InitMetrics() {
    prometheus.MustRegister(WorkloadsRunning)
    prometheus.MustRegister(MemoryUsed)
    prometheus.MustRegister(PageFaults)
    prometheus.MustRegister(WorkloadCompleted)

    go func() {
        http.Handle("/metrics", promhttp.Handler())
        _ = http.ListenAndServe(":9090", nil)
    }()
}
