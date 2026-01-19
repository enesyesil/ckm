package main

import (
	"context"
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
	cgroups := kernel.NewCGroupManager(1024) // 1024 MB total memory
	store := kernel.NewWorkloadStore()
	scheduler := kernel.NewRoundRobinScheduler(1 * time.Second)

	// Initialize shared Docker client
	dockerClient, err := runtime.NewDockerClient()
	if err != nil {
		logger.Fatal("Failed to initialize Docker client", zap.Error(err))
	}

	// Initialize Docker runtime using shared client
	dockerRuntime := runtime.NewDockerRuntime(dockerClient, logger)

	// Create executor with worker pool (max 10 concurrent workloads)
	executor := kernel.NewExecutor(dockerRuntime, store, logger, 10)

	// Start container discovery service using shared client (monitors ALL running containers)
	discovery := runtime.NewContainerDiscovery(dockerClient, logger, 5*time.Second)

	// Create API server
	server := api.NewServer(store, executor, scheduler, cgroups, logger)

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

	// Start container discovery in background
	go discovery.Start(ctx)
	logger.Info("Container discovery started (monitoring all Docker containers)")

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
