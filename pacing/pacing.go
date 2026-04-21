// Package pacing provides a token-bucket-style request pacer that smooths
// traffic by enforcing a steady inter-request interval rather than allowing
// burst-then-silence patterns.
package pacing

import (
	"net/http"
	"sync"
	"time"
)

// Pacer enforces a minimum interval between successive requests.
type Pacer struct {
	mu       sync.Mutex
	last     time.Time
	interval time.Duration
	clock    func() time.Time
	sleep    func(time.Duration)
}

// New creates a Pacer that enforces at least interval between requests.
// If interval is zero or negative it defaults to 10 ms.
func New(interval time.Duration) *Pacer {
	if interval <= 0 {
		interval = 10 * time.Millisecond
	}
	return &Pacer{
		interval: interval,
		clock:    time.Now,
		sleep:    time.Sleep,
	}
}

// Interval returns the configured inter-request interval.
func (p *Pacer) Interval() time.Duration { return p.interval }

// Wait blocks until the pacer allows the next request to proceed.
func (p *Pacer) Wait() {
	p.mu.Lock()
	now := p.clock()
	var delay time.Duration
	if !p.last.IsZero() {
		next := p.last.Add(p.interval)
		if next.After(now) {
			delay = next.Sub(now)
		}
	}
	p.last = now.Add(delay)
	p.mu.Unlock()

	if delay > 0 {
		p.sleep(delay)
	}
}

// Middleware returns an http.Handler that paces inbound requests before
// forwarding them to next.
func (p *Pacer) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.Wait()
		next.ServeHTTP(w, r)
	})
}
