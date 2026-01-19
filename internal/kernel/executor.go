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
	runtime    *runtime.DockerRuntime
	store      *WorkloadStore
	logger     *zap.Logger
	workerPool chan struct{} // Limits concurrent executions
	wg         sync.WaitGroup
}

// NewExecutor creates a new workload executor with worker pool
func NewExecutor(dockerRuntime *runtime.DockerRuntime, store *WorkloadStore, logger *zap.Logger, maxWorkers int) *Executor {
	return &Executor{
		runtime:    dockerRuntime,
		store:      store,
		logger:     logger,
		workerPool: make(chan struct{}, maxWorkers),
	}
}

// Execute runs a workload in a container
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

	// Create container with resource limits
	containerID, err := e.runtime.CreateContainer(ctx, w.Image, w.Command, w.MemoryMB, int64(w.Priority*512))
	if err != nil {
		e.store.Update(w.ID, "failed")
		common.WorkloadFailuresTotal.WithLabelValues(w.Type, "create").Inc()
		return err
	}

	w.ContainerID = containerID
	e.store.Add(w)

	// Track container startup time
	startupStart := time.Now()
	if err := e.runtime.StartContainer(ctx, containerID); err != nil {
		e.store.Update(w.ID, "failed")
		common.WorkloadFailuresTotal.WithLabelValues(w.Type, "start").Inc()
		return err
	}
	common.ContainerStartupTimeSeconds.Observe(time.Since(startupStart).Seconds())

	// Wait for container completion
	exitCode, err := e.runtime.WaitContainer(ctx, containerID)
	if err != nil {
		e.store.Update(w.ID, "failed")
		common.WorkloadFailuresTotal.WithLabelValues(w.Type, "wait").Inc()
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
