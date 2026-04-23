// Package canary provides a middleware that routes a configurable
// percentage of traffic to a canary backend while sending the remainder
// to the primary handler. The split is decided per-request using a
// fast, lock-free random draw so there is no shared mutable state on
// the hot path.
package canary

import (
	"math/rand"
	"net/http"
	"sync/atomic"
)

// Canary splits inbound requests between a primary and a canary handler.
type Canary struct {
	primary http.Handler
	canary  http.Handler
	// rate is the fraction of requests [0,100] sent to the canary.
	rate atomic.Int64
}

// New returns a Canary that routes rate% of requests to canary and the
// rest to primary. rate is clamped to [0, 100].
func New(primary, canary http.Handler, rate int) *Canary {
	if primary == nil {
		panic("canary: primary handler must not be nil")
	}
	if canary == nil {
		panic("canary: canary handler must not be nil")
	}
	c := &Canary{primary: primary, canary: canary}
	c.SetRate(rate)
	return c
}

// SetRate updates the canary traffic percentage at runtime. Safe for
// concurrent use.
func (c *Canary) SetRate(rate int) {
	if rate < 0 {
		rate = 0
	}
	if rate > 100 {
		rate = 100
	}
	c.rate.Store(int64(rate))
}

// Rate returns the current canary traffic percentage.
func (c *Canary) Rate() int {
	return int(c.rate.Load())
}

// ServeHTTP implements http.Handler.
func (c *Canary) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//nolint:gosec // non-cryptographic use is intentional
	if rand.Intn(100) < c.Rate() {
		c.canary.ServeHTTP(w, r)
		return
	}
	c.primary.ServeHTTP(w, r)
}
