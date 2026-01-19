package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ckm/internal/kernel"
	"go.uber.org/zap"
)

// setupTestServer creates a test server without Docker runtime
func setupTestServer() *Server {
	logger := zap.NewNop()
	store := kernel.NewWorkloadStore()
	cgroups := kernel.NewCGroupManager(1024)
	scheduler := kernel.NewRoundRobinScheduler(time.Second)

	// Create server without executor (for API testing only)
	s := &Server{
		store:     store,
		scheduler: scheduler,
		cgroups:   cgroups,
		logger:    logger,
	}
	return s
}

// TestHealthCheck tests the health check endpoint
func TestHealthCheck(t *testing.T) {
	s := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	s.healthCheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["status"] != "healthy" {
		t.Errorf("Expected status healthy, got %s", response["status"])
	}
}

// TestListWorkloadsEmpty tests listing workloads when empty
func TestListWorkloadsEmpty(t *testing.T) {
	s := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/workloads", nil)
	w := httptest.NewRecorder()

	s.listWorkloads(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var workloads []*kernel.Workload
	json.Unmarshal(w.Body.Bytes(), &workloads)

	if len(workloads) != 0 {
		t.Errorf("Expected 0 workloads, got %d", len(workloads))
	}
}

// TestListWorkloads tests listing workloads
func TestListWorkloads(t *testing.T) {
	s := setupTestServer()

	// Add workloads to store
	s.store.Add(&kernel.Workload{ID: "test-1", Type: "task"})
	s.store.Add(&kernel.Workload{ID: "test-2", Type: "container"})

	req := httptest.NewRequest("GET", "/api/v1/workloads", nil)
	w := httptest.NewRecorder()

	s.listWorkloads(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var workloads []*kernel.Workload
	json.Unmarshal(w.Body.Bytes(), &workloads)

	if len(workloads) != 2 {
		t.Errorf("Expected 2 workloads, got %d", len(workloads))
	}
}

// TestRespondJSON tests JSON response helper
func TestRespondJSON(t *testing.T) {
	s := setupTestServer()

	w := httptest.NewRecorder()
	s.respondJSON(w, http.StatusOK, map[string]string{"key": "value"})

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type application/json")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestRespondError tests error response helper
func TestRespondError(t *testing.T) {
	s := setupTestServer()

	w := httptest.NewRecorder()
	s.respondError(w, http.StatusBadRequest, "test error")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["error"] != "test error" {
		t.Errorf("Expected error 'test error', got '%s'", response["error"])
	}
}

// TestCreateWorkloadRequest tests request parsing
func TestCreateWorkloadRequest(t *testing.T) {
	reqBody := CreateWorkloadRequest{
		ID:       "test-1",
		Type:     "container",
		MemoryMB: 256,
		Image:    "alpine:latest",
		Command:  []string{"echo", "hello"},
		Priority: 1,
	}

	body, _ := json.Marshal(reqBody)
	reader := bytes.NewReader(body)

	var parsed CreateWorkloadRequest
	json.NewDecoder(reader).Decode(&parsed)

	if parsed.ID != "test-1" {
		t.Errorf("Expected ID test-1, got %s", parsed.ID)
	}
	if parsed.MemoryMB != 256 {
		t.Errorf("Expected MemoryMB 256, got %d", parsed.MemoryMB)
	}
}
