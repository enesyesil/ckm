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

	PageFaults = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ckm_page_faults_total",
			Help: "Number of simulated page faults",
		})

	// Scheduler metrics
	SchedulerQueueLength = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ckm_scheduler_queue_length",
			Help: "Number of workloads in scheduler queue",
		},
		[]string{"scheduler"},
	)

	SchedulerDecisionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ckm_scheduler_decisions_total",
			Help: "Total scheduler selection decisions",
		},
		[]string{"scheduler"},
	)

	// Resource utilization
	ResourceUtilization = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ckm_resource_utilization_percent",
			Help: "Resource utilization percentage",
		},
		[]string{"resource"},
	)

	// Container metrics
	ContainerStartupTimeSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "ckm_container_startup_time_seconds",
			Help:    "Container startup time in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0},
		},
	)
)

// InitMetrics registers all Prometheus metrics and starts metrics server
func InitMetrics() {
	prometheus.MustRegister(WorkloadsRunning)
	prometheus.MustRegister(WorkloadCompleted)
	prometheus.MustRegister(WorkloadDurationSeconds)
	prometheus.MustRegister(WorkloadFailuresTotal)
	prometheus.MustRegister(MemoryUsed)
	prometheus.MustRegister(PageFaults)
	prometheus.MustRegister(SchedulerQueueLength)
	prometheus.MustRegister(SchedulerDecisionsTotal)
	prometheus.MustRegister(ResourceUtilization)
	prometheus.MustRegister(ContainerStartupTimeSeconds)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		_ = http.ListenAndServe(":9090", nil)
	}()
}
