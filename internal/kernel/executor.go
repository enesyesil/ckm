package kernel

import (
	"context"
	"sync"
	"time"

	"ckm/internal/common"
	"ckm/internal/runtime"
	"go.uber.org/zap"
)

// Executor runs workloads using Docker runtime with worker pool
type Executor struct {
	runtime        *runtime.DockerRuntime
	store          *WorkloadStore
	logger         *zap.Logger
	workerPool     chan struct{} // Limits concurrent executions
	circuitBreaker *common.CircuitBreaker
	wg             sync.WaitGroup
}

// NewExecutor creates a new workload executor with worker pool
func NewExecutor(dockerRuntime *runtime.DockerRuntime, store *WorkloadStore, logger *zap.Logger, maxWorkers int) *Executor {
	return &Executor{
		runtime:        dockerRuntime,
		store:          store,
		logger:         logger,
		workerPool:     make(chan struct{}, maxWorkers),
		circuitBreaker: common.NewCircuitBreaker(5, 30*time.Second), // Open after 5 failures, reset after 30s
	}
}

// Execute runs a workload in a container with circuit breaker protection
func (e *Executor) Execute(ctx context.Context, w *Workload) error {
	// Acquire worker slot (limits concurrency)
	e.workerPool <- struct{}{}
	defer func() { <-e.workerPool }()

	e.wg.Add(1)
	defer e.wg.Done()

	// Update status to running
	e.store.Update(w.ID, "running")
	w.StartedAt = time.Now()
	common.WorkloadsRunning.Inc()

	// Track execution time for metrics
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		common.WorkloadDurationSeconds.WithLabelValues(w.Type).Observe(duration.Seconds())
	}()

	// Use circuit breaker for Docker operations
	var containerID string
	err := e.circuitBreaker.Call(func() error {
		var createErr error
		containerID, createErr = e.runtime.CreateContainer(ctx, w.Image, w.Command, w.MemoryMB, int64(w.Priority*512))
		return createErr
	})
	if err != nil {
		e.store.Update(w.ID, "failed")
		common.WorkloadFailuresTotal.WithLabelValues(w.Type, "create").Inc()
		if err == common.ErrCircuitOpen {
			e.logger.Warn("Circuit breaker open, Docker operations paused", zap.String("workload", w.ID))
		}
		common.WorkloadsRunning.Dec()
		return err
	}

	w.ContainerID = containerID
	e.store.Add(w)

	// Track container startup time with circuit breaker
	startupStart := time.Now()
	err = e.circuitBreaker.Call(func() error {
		return e.runtime.StartContainer(ctx, containerID)
	})
	if err != nil {
		e.store.Update(w.ID, "failed")
		common.WorkloadFailuresTotal.WithLabelValues(w.Type, "start").Inc()
		common.WorkloadsRunning.Dec()
		return err
	}
	common.ContainerStartupTimeSeconds.Observe(time.Since(startupStart).Seconds())

	// Wait for container completion
	var exitCode int64
	err = e.circuitBreaker.Call(func() error {
		var waitErr error
		exitCode, waitErr = e.runtime.WaitContainer(ctx, containerID)
		return waitErr
	})
	if err != nil {
		e.store.Update(w.ID, "failed")
		common.WorkloadFailuresTotal.WithLabelValues(w.Type, "wait").Inc()
		common.WorkloadsRunning.Dec()
		return err
	}

	// Update status based on exit code
	if exitCode == 0 {
		e.store.Update(w.ID, "done")
		common.WorkloadCompleted.WithLabelValues(w.Type).Inc()
	} else {
		e.store.Update(w.ID, "failed")
		common.WorkloadFailuresTotal.WithLabelValues(w.Type, "exit").Inc()
	}

	common.WorkloadsRunning.Dec()

	// Cleanup container
	_ = e.runtime.RemoveContainer(ctx, containerID)
	return nil
}

// ExecuteAsync runs workload in background goroutine
func (e *Executor) ExecuteAsync(ctx context.Context, w *Workload) {
	go func() {
		if err := e.Execute(ctx, w); err != nil {
			e.logger.Error("Workload execution failed", zap.String("id", w.ID), zap.Error(err))
		}
	}()
}

// Wait waits for all running workloads to complete
func (e *Executor) Wait() {
	e.wg.Wait()
}

// StopContainer stops a container (exposed for API server)
func (e *Executor) StopContainer(ctx context.Context, containerID string) error {
	return e.runtime.StopContainer(ctx, containerID, 10*time.Second)
}
