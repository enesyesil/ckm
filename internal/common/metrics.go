package common

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Workload metrics
	WorkloadsRunning = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ckm_workloads_running_total",
			Help: "Number of workloads currently running",
		})

	WorkloadCompleted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ckm_workloads_completed_total",
			Help: "Total completed workloads by type",
		},
		[]string{"type"},
	)

	WorkloadDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ckm_workload_duration_seconds",
			Help:    "Workload execution duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type"},
	)

	WorkloadFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ckm_workload_failures_total",
			Help: "Total workload failures by type and reason",
		},
		[]string{"type", "reason"},
	)

	// Memory metrics
	MemoryUsed = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ckm_memory_usage_megabytes",
			Help: "Memory usage in MB",
		})

	// Scheduler metrics
	SchedulerQueueLength = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ckm_scheduler_queue_length",
			Help: "Number of workloads in scheduler queue",
		},
		[]string{"scheduler"},
	)

	// Container metrics
	ContainerStartupTimeSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "ckm_container_startup_time_seconds",
			Help:    "Container startup time in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0},
		},
	)

	// Container discovery metrics (real-time stats for all containers)
	DiscoveredContainers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ckm_discovered_containers_total",
			Help: "Number of running containers discovered",
		},
	)

	ContainerCPUPercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ckm_container_cpu_percent",
			Help: "Container CPU usage percentage",
		},
		[]string{"container", "image"},
	)

	ContainerMemoryBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ckm_container_memory_bytes",
			Help: "Container memory usage in bytes",
		},
		[]string{"container", "image"},
	)

	ContainerMemoryPercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ckm_container_memory_percent",
			Help: "Container memory usage percentage",
		},
		[]string{"container", "image"},
	)

	ContainerNetworkRxBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ckm_container_network_rx_bytes",
			Help: "Container network received bytes",
		},
		[]string{"container", "image"},
	)

	ContainerNetworkTxBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ckm_container_network_tx_bytes",
			Help: "Container network transmitted bytes",
		},
		[]string{"container", "image"},
	)

	ContainerBlockReadBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ckm_container_block_read_bytes",
			Help: "Container block read bytes",
		},
		[]string{"container", "image"},
	)

	ContainerBlockWriteBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ckm_container_block_write_bytes",
			Help: "Container block write bytes",
		},
		[]string{"container", "image"},
	)

	ContainerPIDs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ckm_container_pids",
			Help: "Number of processes in container",
		},
		[]string{"container", "image"},
	)
)

// InitMetrics registers all Prometheus metrics and starts metrics server
func InitMetrics() {
	prometheus.MustRegister(WorkloadsRunning)
	prometheus.MustRegister(WorkloadCompleted)
	prometheus.MustRegister(WorkloadDurationSeconds)
	prometheus.MustRegister(WorkloadFailuresTotal)
	prometheus.MustRegister(MemoryUsed)
	prometheus.MustRegister(SchedulerQueueLength)
	prometheus.MustRegister(ContainerStartupTimeSeconds)

	// Container discovery metrics
	prometheus.MustRegister(DiscoveredContainers)
	prometheus.MustRegister(ContainerCPUPercent)
	prometheus.MustRegister(ContainerMemoryBytes)
	prometheus.MustRegister(ContainerMemoryPercent)
	prometheus.MustRegister(ContainerNetworkRxBytes)
	prometheus.MustRegister(ContainerNetworkTxBytes)
	prometheus.MustRegister(ContainerBlockReadBytes)
	prometheus.MustRegister(ContainerBlockWriteBytes)
	prometheus.MustRegister(ContainerPIDs)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		_ = http.ListenAndServe(":9090", nil)
	}()
}
