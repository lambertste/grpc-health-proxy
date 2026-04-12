// Package reqlog provides a ring-buffer backed request log that retains the
// most recent N HTTP requests for inspection and debugging.
package reqlog

import (
	"net/http"
	"sync"
	"time"
)

// Entry holds a snapshot of a single HTTP request.
type Entry struct {
	Timestamp  time.Time
	Method     string
	Path       string
	RemoteAddr string
	Status     int
	Duration   time.Duration
}

// Logger is a fixed-capacity ring buffer that records recent HTTP requests.
type Logger struct {
	mu      sync.Mutex
	buf     []Entry
	cap     int
	head    int
	count   int
}

const defaultCapacity = 100

// New returns a Logger that retains at most capacity entries.
// If capacity is zero the default of 100 is used.
func New(capacity int) *Logger {
	if capacity <= 0 {
		capacity = defaultCapacity
	}
	return &Logger{
		buf: make([]Entry, capacity),
		cap: capacity,
	}
}

// record appends an entry, overwriting the oldest when the buffer is full.
func (l *Logger) record(e Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buf[l.head] = e
	l.head = (l.head + 1) % l.cap
	if l.count < l.cap {
		l.count++
	}
}

// Entries returns a copy of all retained entries, oldest first.
func (l *Logger) Entries() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Entry, l.count)
	start := (l.head - l.count + l.cap) % l.cap
	for i := 0; i < l.count; i++ {
		out[i] = l.buf[(start+i)%l.cap]
	}
	return out
}

// Len returns the number of entries currently stored.
func (l *Logger) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.count
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Middleware wraps next and records each request into the Logger.
func (l *Logger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, r)
		l.record(Entry{
			Timestamp:  start,
			Method:     r.Method,
			Path:       r.URL.Path,
			RemoteAddr: r.RemoteAddr,
			Status:     rec.status,
			Duration:   time.Since(start),
		})
	})
}
