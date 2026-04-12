// Package deadletter provides a dead-letter queue for capturing and
// inspecting requests that could not be successfully processed.
package deadletter

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// Entry holds a captured request and the upstream response status.
type Entry struct {
	Timestamp  time.Time
	Method     string
	Path       string
	StatusCode int
}

// Queue is a bounded, thread-safe dead-letter queue.
type Queue struct {
	mu      sync.Mutex
	entries []Entry
	cap     int
}

// New returns a Queue that retains at most capacity entries.
// If capacity is zero or negative it defaults to 100.
func New(capacity int) *Queue {
	if capacity <= 0 {
		capacity = 100
	}
	return &Queue{cap: capacity}
}

// record appends an entry, evicting the oldest when full.
func (q *Queue) record(e Entry) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.entries) >= q.cap {
		q.entries = q.entries[1:]
	}
	q.entries = append(q.entries, e)
}

// Entries returns a snapshot of all captured entries.
func (q *Queue) Entries() []Entry {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]Entry, len(q.entries))
	copy(out, q.entries)
	return out
}

// Len returns the current number of entries.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.entries)
}

// Middleware wraps next and records any response whose status code satisfies
// the provided predicate into the queue.
func (q *Queue) Middleware(predicate func(statusCode int) bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)

		if predicate(rec.Code) {
			q.record(Entry{
				Timestamp:  time.Now(),
				Method:     r.Method,
				Path:       r.URL.Path,
				StatusCode: rec.Code,
			})
		}

		for k, vs := range rec.Header() {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(rec.Code)
		_, _ = w.Write(rec.Body.Bytes())
	})
}

// IsError is a convenience predicate that captures 5xx responses.
func IsError(code int) bool { return code >= 500 }
