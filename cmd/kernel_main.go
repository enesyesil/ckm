package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ckm/internal/api"
	"ckm/internal/common"
	"ckm/internal/kernel"
	"ckm/internal/runtime"
	"go.uber.org/zap"
)

func main() {
	// Initialize structured logging
	common.InitLogger()
	logger := common.Logger
	defer logger.Sync()

	// Initialize Prometheus metrics
	common.InitMetrics()
	logger.Info("Metrics server started on :9090")

	// Create components
	memory := kernel.NewMemoryManager(1024)
	store := kernel.NewWorkloadStore()
	scheduler := kernel.NewRoundRobinScheduler(1 * time.Second)

	// Initialize Docker runtime
	dockerRuntime, err := runtime.NewDockerRuntime(logger)
	if err != nil {
		logger.Fatal("Failed to initialize Docker runtime", zap.Error(err))
	}

	// Create executor with worker pool (max 10 concurrent workloads)
	executor := kernel.NewExecutor(dockerRuntime, store, logger, 10)

	// Create API server
	server := api.NewServer(store, executor, scheduler, memory, logger)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigHandler := kernel.NewSignalHandler()
	sigHandler.RegisterHandler(syscall.SIGTERM, func() {
		logger.Info("Received SIGTERM, shutting down gracefully...")
		cancel()
	})
	sigHandler.RegisterHandler(syscall.SIGINT, func() {
		logger.Info("Received SIGINT, shutting down gracefully...")
		cancel()
	})
	sigHandler.Start(ctx)

	// Start API server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("Starting API server on :8080")
		if err := server.Start(":8080"); err != nil {
			serverErr <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case <-ctx.Done():
		logger.Info("Shutting down...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		
		// Wait for running workloads
		executor.Wait()
		
		// Shutdown API server
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Error shutting down server", zap.Error(err))
		}
		logger.Info("Shutdown complete")
	case err := <-serverErr:
		logger.Fatal("Server error", zap.Error(err))
	}
}
