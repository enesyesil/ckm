// +build integration

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"ckm/internal/common"
	"ckm/internal/kernel"
	"ckm/internal/runtime"
	"go.uber.org/zap"
)

// Integration tests require Docker to be running
// Run with: go test -tags=integration ./...

// TestIntegrationWorkloadLifecycle tests full workload lifecycle with real Docker
func TestIntegrationWorkloadLifecycle(t *testing.T) {
	// Initialize components
	logger := zap.NewNop()
	common.InitLogger()
	common.InitMetrics()

	store := kernel.NewWorkloadStore()
	cgroups := kernel.NewCGroupManager(1024)
	scheduler := kernel.NewRoundRobinScheduler(time.Second)

	dockerClient, err := runtime.NewDockerClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	dockerRuntime := runtime.NewDockerRuntime(dockerClient, logger)

	executor := kernel.NewExecutor(dockerRuntime, store, logger, 5)
	server := NewServer(store, executor, scheduler, cgroups, logger)

	// Start server in background
	go server.Start(":18080")
	time.Sleep(100 * time.Millisecond)

	// Create workload
	reqBody := CreateWorkloadRequest{
		ID:       "integration-test-1",
		Type:     "container",
		MemoryMB: 64,
		Image:    "alpine:latest",
		Command:  []string{"echo", "hello from integration test"},
		Priority: 1,
	}

	body, _ := json.Marshal(reqBody)
	resp, err := http.Post("http://localhost:18080/api/v1/workloads", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create workload: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	// Wait for completion
	time.Sleep(5 * time.Second)

	// Check workload status
	resp, err = http.Get("http://localhost:18080/api/v1/workloads/integration-test-1")
	if err != nil {
		t.Fatalf("Failed to get workload: %v", err)
	}
	defer resp.Body.Close()

	var workload kernel.Workload
	json.NewDecoder(resp.Body).Decode(&workload)

	if workload.Status != "done" {
		t.Errorf("Expected status done, got %s", workload.Status)
	}
}

// TestIntegrationHealthCheck tests health check with running server
func TestIntegrationHealthCheck(t *testing.T) {
	logger := zap.NewNop()
	store := kernel.NewWorkloadStore()
	cgroups := kernel.NewCGroupManager(1024)
	scheduler := kernel.NewRoundRobinScheduler(time.Second)

	dockerClient, err := runtime.NewDockerClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	dockerRuntime := runtime.NewDockerRuntime(dockerClient, logger)

	executor := kernel.NewExecutor(dockerRuntime, store, logger, 5)
	server := NewServer(store, executor, scheduler, cgroups, logger)

	// Start server in background
	go server.Start(":18081")
	time.Sleep(100 * time.Millisecond)

	// Check health
	resp, err := http.Get("http://localhost:18081/api/v1/health")
	if err != nil {
		t.Fatalf("Failed to check health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)

	if result["status"] != "healthy" {
		t.Errorf("Expected healthy, got %s", result["status"])
	}
}
