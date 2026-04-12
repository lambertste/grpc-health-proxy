// Package checkpoint provides a middleware that persists a lightweight
// request fingerprint (method + path + status) to an append-only in-memory
// log so that operators can inspect which requests completed successfully
// across a rolling window.
package checkpoint

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Entry records a single completed request.
type Entry struct {
	Timestamp time.Time
	Method    string
	Path      string
	Status    int
}

// String returns a human-readable representation of the entry.
func (e Entry) String() string {
	return fmt.Sprintf("%s %s %s %d", e.Timestamp.Format(time.RFC3339), e.Method, e.Path, e.Status)
}

// Log is a bounded, thread-safe append-only checkpoint log.
type Log struct {
	mu      sync.Mutex
	entries []Entry
	cap     int
}

// New creates a Log that retains at most capacity entries.
// If capacity is <= 0 it defaults to 1000.
func New(capacity int) *Log {
	if capacity <= 0 {
		capacity = 1000
	}
	return &Log{cap: capacity, entries: make([]Entry, 0, capacity)}
}

// record appends an entry, evicting the oldest when the log is full.
func (l *Log) record(e Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.entries) >= l.cap {
		l.entries = l.entries[1:]
	}
	l.entries = append(l.entries, e)
}

// Entries returns a snapshot of all current log entries.
func (l *Log) Entries() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Entry, len(l.entries))
	copy(out, l.entries)
	return out
}

// Len returns the current number of entries.
func (l *Log) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries)
}

// captureWriter wraps http.ResponseWriter to capture the status code.
type captureWriter struct {
	http.ResponseWriter
	status int
}

func (c *captureWriter) WriteHeader(code int) {
	c.status = code
	c.ResponseWriter.WriteHeader(code)
}

// Middleware returns an http.Handler that records every completed request
// into the provided Log before passing control to next.
func (l *Log) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cw := &captureWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(cw, r)
		l.record(Entry{
			Timestamp: time.Now(),
			Method:    r.Method,
			Path:      r.URL.Path,
			Status:    cw.status,
		})
	})
}
