package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"grpc-health-proxy/config"
	"grpc-health-proxy/health"
)

// Server represents the HTTP server that exposes health check endpoints
type Server struct {
	config  *config.Config
	checker *health.Checker
	server  *http.Server
}

// New creates a new HTTP server instance
func New(cfg *config.Config, checker *health.Checker) *Server {
	mux := http.NewServeMux()
	s := &Server{
		config:  cfg,
		checker: checker,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/ready", s.handleReady)
	mux.HandleFunc("/readyz", s.handleReady)

	return s
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
	log.Printf("Starting HTTP server on port %d", s.config.HTTPPort)
	return s.server.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.config.Timeout)
	defer cancel()

	status, err := s.checker.Check(ctx, "")
	if err != nil {
		s.writeErrorResponse(w, http.StatusServiceUnavailable, err)
		return
	}

	s.writeHealthResponse(w, status)
}

// handleReady handles readiness check requests
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.config.Timeout)
	defer cancel()

	status, err := s.checker.Check(ctx, "")
	if err != nil {
		s.writeErrorResponse(w, http.StatusServiceUnavailable, err)
		return
	}

	s.writeHealthResponse(w, status)
}

// writeHealthResponse writes a health check response
func (s *Server) writeHealthResponse(w http.ResponseWriter, status string) {
	w.Header().Set("Content-Type", "application/json")
	if status == "SERVING" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}

// writeErrorResponse writes an error response
func (s *Server) writeErrorResponse(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
