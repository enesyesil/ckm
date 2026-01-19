package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"ckm/internal/common"
	"ckm/internal/kernel"
	"go.uber.org/zap"

	"github.com/gorilla/mux"
)

// Server provides REST API for workload management
type Server struct {
	router      *mux.Router
	store       *kernel.WorkloadStore
	executor    *kernel.Executor
	scheduler   kernel.Scheduler
	cgroups     *kernel.CGroupManager
	rateLimiter *common.RateLimiter
	logger      *zap.Logger
	httpServer  *http.Server
}

// NewServer creates a new API server
func NewServer(store *kernel.WorkloadStore, executor *kernel.Executor, scheduler kernel.Scheduler, cgroups *kernel.CGroupManager, logger *zap.Logger) *Server {
	s := &Server{
		router:      mux.NewRouter(),
		store:       store,
		executor:    executor,
		scheduler:   scheduler,
		cgroups:     cgroups,
		rateLimiter: common.NewRateLimiter(100, 50), // 100 req/sec, burst of 50
		logger:      logger,
	}
	s.setupRoutes()
	return s
}

// setupRoutes configures API endpoints
func (s *Server) setupRoutes() {
	api := s.router.PathPrefix("/api/v1").Subrouter()
	api.Use(s.rateLimitMiddleware)
	api.HandleFunc("/workloads", s.createWorkload).Methods("POST")
	api.HandleFunc("/workloads", s.listWorkloads).Methods("GET")
	api.HandleFunc("/workloads/{id}", s.getWorkload).Methods("GET")
	api.HandleFunc("/workloads/{id}", s.deleteWorkload).Methods("DELETE")
	api.HandleFunc("/health", s.healthCheck).Methods("GET")
}

// rateLimitMiddleware applies rate limiting to all API requests
func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.rateLimiter != nil && !s.rateLimiter.Allow() {
			s.respondError(w, http.StatusTooManyRequests, "Rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// createWorkload handles POST /api/v1/workloads
func (s *Server) createWorkload(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Create workload with PID
	wl := &kernel.Workload{
		ID:       req.ID,
		PID:      kernel.NextPID(),
		Type:     req.Type,
		MemoryMB: req.MemoryMB,
		Image:    req.Image,
		Command:  req.Command,
		Priority: req.Priority,
		Status:   "waiting",
	}

	// Allocate memory via cgroups
	if !s.cgroups.Allocate(wl.ID, wl.MemoryMB) {
		s.respondError(w, http.StatusInsufficientStorage, "Not enough memory")
		return
	}

	// Add to store and scheduler
	s.store.Add(wl)
	s.scheduler.Add(*wl)

	// Update metrics
	common.MemoryUsed.Set(float64(s.cgroups.GetUsedMemory()))
	common.SchedulerQueueLength.WithLabelValues("default").Inc()

	// Execute asynchronously
	ctx := context.Background()
	s.executor.ExecuteAsync(ctx, wl)

	s.respondJSON(w, http.StatusCreated, wl)
}

// listWorkloads handles GET /api/v1/workloads
func (s *Server) listWorkloads(w http.ResponseWriter, r *http.Request) {
	workloads := s.store.GetAll()
	s.respondJSON(w, http.StatusOK, workloads)
}

// getWorkload handles GET /api/v1/workloads/{id}
func (s *Server) getWorkload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	wl, ok := s.store.Get(vars["id"])
	if !ok {
		s.respondError(w, http.StatusNotFound, "Workload not found")
		return
	}
	s.respondJSON(w, http.StatusOK, wl)
}

// deleteWorkload handles DELETE /api/v1/workloads/{id}
func (s *Server) deleteWorkload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	wl, ok := s.store.Get(vars["id"])
	if !ok {
		s.respondError(w, http.StatusNotFound, "Workload not found")
		return
	}

	// Stop container if running
	if wl.ContainerID != "" && wl.Status == "running" {
		ctx := context.Background()
		_ = s.executor.StopContainer(ctx, wl.ContainerID)
	}

	// Free memory and delete
	s.cgroups.Free(wl.ID, wl.MemoryMB)
	s.store.Delete(wl.ID)
	common.MemoryUsed.Set(float64(s.cgroups.GetUsedMemory()))

	s.respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// healthCheck handles GET /api/v1/health
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// Start starts the HTTP server
func (s *Server) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	s.logger.Info("Starting API server", zap.String("addr", addr))
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Helper functions
func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) respondError(w http.ResponseWriter, status int, message string) {
	s.respondJSON(w, status, map[string]string{"error": message})
}

// CreateWorkloadRequest represents workload creation request
type CreateWorkloadRequest struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	MemoryMB int      `json:"memory_mb"`
	Image    string   `json:"image"`
	Command  []string `json:"command"`
	Priority int      `json:"priority"`
}
