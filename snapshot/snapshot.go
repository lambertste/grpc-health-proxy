// Package snapshot provides a point-in-time capture of HTTP handler metrics
// and state, useful for diagnostics and debug endpoints.
package snapshot

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Entry holds a single captured snapshot.
type Entry struct {
	Timestamp time.Time         `json:"timestamp"`
	Status    int               `json:"status"`
	Path      string            `json:"path"`
	Method    string            `json:"method"`
	LatencyMs int64             `json:"latency_ms"`
	Headers   map[string]string `json:"headers,omitempty"`
}

// Snapshotter captures recent request snapshots up to a fixed capacity.
type Snapshotter struct {
	mu       sync.Mutex
	entries  []Entry
	capacity int
}

// New creates a Snapshotter with the given capacity. If capacity is zero or
// negative it defaults to 100.
func New(capacity int) *Snapshotter {
	if capacity <= 0 {
		capacity = 100
	}
	return &Snapshotter{capacity: capacity}
}

// record appends an entry, evicting the oldest when at capacity.
func (s *Snapshotter) record(e Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.entries) >= s.capacity {
		s.entries = s.entries[1:]
	}
	s.entries = append(s.entries, e)
}

// Entries returns a copy of all captured entries.
func (s *Snapshotter) Entries() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Entry, len(s.entries))
	copy(out, s.entries)
	return out
}

// Len returns the current number of stored entries.
func (s *Snapshotter) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}

type responseWriter struct {
	http.ResponseWriter
	code int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.code = code
	rw.ResponseWriter.WriteHeader(code)
}

// Middleware wraps next, recording a snapshot for every request.
func (s *Snapshotter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rw, r)
		hdrs := make(map[string]string)
		for k := range r.Header {
			hdrs[k] = r.Header.Get(k)
		}
		s.record(Entry{
			Timestamp: start,
			Status:    rw.code,
			Path:      r.URL.Path,
			Method:    r.Method,
			LatencyMs: time.Since(start).Milliseconds(),
			Headers:   hdrs,
		})
	})
}

// Handler returns an http.Handler that serves the current snapshots as JSON.
func (s *Snapshotter) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.Entries()) //nolint:errcheck
	})
}
