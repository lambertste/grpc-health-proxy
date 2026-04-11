// Package sampling provides request sampling middleware that forwards
// a configurable fraction of requests to a collector for analysis or
// debugging without affecting the primary response path.
package sampling

import (
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// Collector receives sampled requests for further processing.
type Collector func(r *http.Request)

// Sampler decides which requests to forward to the collector.
type Sampler struct {
	mu        sync.Mutex
	rate      float64 // 0.0 – 1.0
	collector Collector
	rng       *rand.Rand
}

// New creates a Sampler that forwards approximately rate*100 % of requests
// to collector. rate is clamped to [0.0, 1.0].
func New(rate float64, collector Collector) *Sampler {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	if collector == nil {
		collector = func(*http.Request) {}
	}
	return &Sampler{
		rate:      rate,
		collector: collector,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Allow reports whether the current request should be sampled.
func (s *Sampler) Allow() bool {
	if s.rate == 0 {
		return false
	}
	if s.rate == 1 {
		return true
	}
	s.mu.Lock()
	v := s.rng.Float64()
	s.mu.Unlock()
	return v < s.rate
}

// Middleware returns an http.Handler that samples incoming requests and
// forwards clones to the collector before passing through to next.
func (s *Sampler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.Allow() {
			s.collector(r.Clone(r.Context()))
		}
		next.ServeHTTP(w, r)
	})
}
