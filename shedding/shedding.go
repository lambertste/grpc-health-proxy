// Package shedding implements adaptive load shedding based on a sliding
// error-rate window. Requests are rejected with HTTP 503 once the observed
// error rate exceeds a configurable threshold.
package shedding

import (
	"net/http"
	"sync"
)

// Shedder tracks request outcomes and sheds load when the error rate is too
// high.
type Shedder struct {
	mu        sync.Mutex
	total     int
	errors    int
	threshold float64 // 0–1, e.g. 0.5 means 50 % errors triggers shedding
	window    int     // number of most-recent requests to consider
	history   []bool  // circular buffer; true == error
	pos       int
}

// New returns a Shedder that sheds load once the error rate within the last
// window requests exceeds threshold. threshold is clamped to [0, 1] and
// window must be positive (defaults to 100).
func New(threshold float64, window int) *Shedder {
	if threshold < 0 {
		threshold = 0
	}
	if threshold > 1 {
		threshold = 1
	}
	if window <= 0 {
		window = 100
	}
	return &Shedder{
		threshold: threshold,
		window:    window,
		history:   make([]bool, window),
	}
}

// Record records a request outcome. isError should be true when the upstream
// returned an error or a 5xx status.
func (s *Shedder) Record(isError bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.total < s.window {
		s.total++
	} else {
		// evict oldest sample
		if s.history[s.pos] {
			s.errors--
		}
	}
	s.history[s.pos] = isError
	if isError {
		s.errors++
	}
	s.pos = (s.pos + 1) % s.window
}

// Allow returns false when the current error rate exceeds the threshold and
// there are enough samples to make a reliable decision (at least window/2).
func (s *Shedder) Allow() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.total < s.window/2 {
		return true
	}
	return float64(s.errors)/float64(s.total) < s.threshold
}

// Middleware wraps next, recording outcomes and shedding load as needed.
func (s *Shedder) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.Allow() {
			http.Error(w, "service unavailable – load shedding active", http.StatusServiceUnavailable)
			return
		}
		rec := &statusRecorder{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rec, r)
		s.Record(rec.code >= 500)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	code int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.code = code
	r.ResponseWriter.WriteHeader(code)
}
