// Package fence provides a token-bucket based request fencing mechanism
// that temporarily blocks traffic to a backend when its error rate exceeds
// a configurable threshold, allowing it time to recover.
package fence

import (
	"net/http"
	"sync"
	"time"
)

// Fence tracks error rates and blocks requests when the backend is unhealthy.
type Fence struct {
	mu          sync.Mutex
	threshold   float64
	window      time.Duration
	cooldown    time.Duration
	total       int
	errors      int
	windowStart time.Time
	blockedUntil time.Time
}

// New creates a Fence that opens (blocks traffic) when the error rate within
// window exceeds threshold (0.0–1.0), and remains open for cooldown duration.
// Panics if threshold is not in (0, 1] or window/cooldown are zero.
func New(threshold float64, window, cooldown time.Duration) *Fence {
	if threshold <= 0 || threshold > 1 {
		panic("fence: threshold must be in (0, 1]")
	}
	if window <= 0 {
		panic("fence: window must be positive")
	}
	if cooldown <= 0 {
		panic("fence: cooldown must be positive")
	}
	return &Fence{
		threshold:   threshold,
		window:      window,
		cooldown:    cooldown,
		windowStart: time.Now(),
	}
}

// Allow returns true if the fence is closed (traffic permitted).
func (f *Fence) Allow() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := time.Now()
	if now.Before(f.blockedUntil) {
		return false
	}
	if now.Sub(f.windowStart) >= f.window {
		f.total = 0
		f.errors = 0
		f.windowStart = now
	}
	return true
}

// Record records the outcome of a request. success=false counts as an error.
// When the error rate exceeds the threshold, the fence opens for cooldown.
func (f *Fence) Record(success bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := time.Now()
	if now.Sub(f.windowStart) >= f.window {
		f.total = 0
		f.errors = 0
		f.windowStart = now
	}
	f.total++
	if !success {
		f.errors++
	}
	if f.total > 0 && float64(f.errors)/float64(f.total) >= f.threshold {
		f.blockedUntil = now.Add(f.cooldown)
		f.total = 0
		f.errors = 0
		f.windowStart = now.Add(f.cooldown)
	}
}

// Middleware wraps next, returning 503 when the fence is open and recording
// outcomes based on the upstream response status code.
func (f *Fence) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !f.Allow() {
			http.Error(w, "service temporarily fenced", http.StatusServiceUnavailable)
			return
		}
		rec := &statusRecorder{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rec, r)
		f.Record(rec.code < 500)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	code int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.code = code
	s.ResponseWriter.WriteHeader(code)
}
