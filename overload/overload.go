// Package overload provides adaptive load shedding based on system CPU usage.
// When CPU utilisation exceeds a configurable threshold the middleware begins
// probabilistically rejecting incoming requests, returning 503 Service
// Unavailable so that upstream callers can back off gracefully.
package overload

import (
	"math"
	"net/http"
	"sync/atomic"
	"time"
)

// Sampler is a function that returns the current CPU utilisation in the range
// [0.0, 1.0]. The default implementation is a stub; callers should supply a
// real sampler via Option.
type Sampler func() float64

// Guard sheds load when CPU utilisation exceeds the configured threshold.
type Guard struct {
	threshold float64
	sampler   Sampler
	last      atomic.Value // float64
	ticker    *time.Ticker
	done      chan struct{}
}

// New creates a Guard that starts shedding requests once CPU utilisation
// exceeds threshold (0.0–1.0). Utilisation is sampled every interval via
// sampler. A zero interval defaults to 1 second.
func New(threshold float64, interval time.Duration, sampler Sampler) *Guard {
	if threshold <= 0 || threshold > 1 {
		threshold = 0.8
	}
	if interval <= 0 {
		interval = time.Second
	}
	if sampler == nil {
		sampler = func() float64 { return 0 }
	}
	g := &Guard{
		threshold: threshold,
		sampler:   sampler,
		ticker:    time.NewTicker(interval),
		done:      make(chan struct{}),
	}
	g.last.Store(0.0)
	go g.poll()
	return g
}

// poll updates the cached CPU sample on every tick.
func (g *Guard) poll() {
	for {
		select {
		case <-g.ticker.C:
			g.last.Store(g.sampler())
		case <-g.done:
			return
		}
	}
}

// Stop halts background sampling. The Guard must not be used after Stop.
func (g *Guard) Stop() {
	g.ticker.Stop()
	close(g.done)
}

// Allow returns true when the request should be admitted. The rejection
// probability rises linearly from 0 at threshold to 1 at full saturation.
func (g *Guard) Allow() bool {
	cpu := g.last.Load().(float64)
	if cpu <= g.threshold {
		return true
	}
	// linear probability: 0 at threshold, 1 at 1.0
	span := 1.0 - g.threshold
	if span <= 0 {
		return false
	}
	rejectProb := math.Min((cpu-g.threshold)/span, 1.0)
	// Use a cheap deterministic sample derived from current nanoseconds.
	nano := float64(time.Now().UnixNano()%1000) / 1000.0
	return nano >= rejectProb
}

// Middleware returns an http.Handler that sheds load based on CPU utilisation.
func (g *Guard) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !g.Allow() {
			http.Error(w, "service overloaded", http.StatusServiceUnavailable)
			return
		}
		next.ServeHTTP(w, r)
	})
}
