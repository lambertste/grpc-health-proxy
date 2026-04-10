// Package healthz provides a simple liveness and readiness endpoint
// for the proxy itself, separate from the upstream gRPC health checks.
package healthz

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// Status represents the self-health status of the proxy.
type Status struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
}

// Handler handles liveness and readiness HTTP endpoints for the proxy itself.
type Handler struct {
	ready   atomic.Bool
	started time.Time
}

// New creates a new Handler. The proxy begins in a not-ready state.
func New() *Handler {
	return &Handler{started: time.Now()}
}

// SetReady marks the proxy as ready to serve traffic.
func (h *Handler) SetReady(ready bool) {
	h.ready.Store(ready)
}

// Live handles GET /healthz/live — always returns 200 while the process is up.
func (h *Handler) Live(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Status{ //nolint:errcheck
		Status:    "ok",
		Timestamp: time.Now().UTC(),
		Uptime:    time.Since(h.started).Round(time.Second).String(),
	})
}

// Ready handles GET /healthz/ready — returns 200 when ready, 503 otherwise.
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !h.ready.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(Status{ //nolint:errcheck
			Status:    "not_ready",
			Timestamp: time.Now().UTC(),
			Uptime:    time.Since(h.started).Round(time.Second).String(),
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Status{ //nolint:errcheck
		Status:    "ready",
		Timestamp: time.Now().UTC(),
		Uptime:    time.Since(h.started).Round(time.Second).String(),
	})
}

// Register mounts the liveness and readiness endpoints on mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/healthz/live", h.Live)
	mux.HandleFunc("/healthz/ready", h.Ready)
}
