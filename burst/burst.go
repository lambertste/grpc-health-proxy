// Package burst provides a token-bucket based burst limiter that allows
// short spikes of traffic above a sustained rate, draining tokens at a
// configurable fill rate with a configurable maximum bucket capacity.
package burst

import (
	"net/http"
	"sync"
	"time"
)

// Limiter is a token-bucket burst limiter.
type Limiter struct {
	mu       sync.Mutex
	tokens   float64
	cap      float64
	rate     float64 // tokens per second
	lastFill time.Time
	now      func() time.Time
}

// New creates a Limiter with the given capacity (burst size) and fill rate
// (tokens added per second). Panics if cap or rate are not positive.
func New(cap float64, ratePerSec float64) *Limiter {
	if cap <= 0 {
		panic("burst: cap must be positive")
	}
	if ratePerSec <= 0 {
		panic("burst: ratePerSec must be positive")
	}
	return &Limiter{
		tokens:   cap,
		cap:      cap,
		rate:     ratePerSec,
		lastFill: time.Now(),
		now:      time.Now,
	}
}

// Allow reports whether one token is available and consumes it if so.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refill()
	if l.tokens < 1 {
		return false
	}
	l.tokens--
	return true
}

// Available returns the current token count (truncated to int).
func (l *Limiter) Available() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refill()
	return int(l.tokens)
}

func (l *Limiter) refill() {
	now := l.now()
	elapsed := now.Sub(l.lastFill).Seconds()
	l.lastFill = now
	l.tokens += elapsed * l.rate
	if l.tokens > l.cap {
		l.tokens = l.cap
	}
}

// Middleware returns an HTTP middleware that rejects requests with 429 when
// the burst limiter has no available tokens.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow() {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
