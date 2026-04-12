// Package watchdog provides a periodic health-check loop that calls a
// user-supplied probe function on a fixed interval and broadcasts the
// current health state to any number of listeners.
package watchdog

import (
	"context"
	"sync"
	"time"
)

const defaultInterval = 5 * time.Second

// Probe is a function that returns nil when the target is healthy.
type Probe func(ctx context.Context) error

// Watchdog runs a Probe on a fixed interval and exposes the latest result.
type Watchdog struct {
	probe    Probe
	interval time.Duration

	mu      sync.RWMutex
	lastErr error
	healthy bool
}

// New creates a Watchdog that calls probe every interval.
// If interval is zero the default of 5 s is used.
func New(probe Probe, interval time.Duration) *Watchdog {
	if interval <= 0 {
		interval = defaultInterval
	}
	return &Watchdog{
		probe:    probe,
		interval: interval,
		healthy:  true, // optimistic until first probe fires
	}
}

// Run starts the probe loop. It blocks until ctx is cancelled.
func (w *Watchdog) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Run an immediate probe before waiting for the first tick.
	w.runProbe(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runProbe(ctx)
		}
	}
}

func (w *Watchdog) runProbe(ctx context.Context) {
	err := w.probe(ctx)
	w.mu.Lock()
	w.lastErr = err
	w.healthy = err == nil
	w.mu.Unlock()
}

// Healthy returns true when the most recent probe succeeded.
func (w *Watchdog) Healthy() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.healthy
}

// LastErr returns the error from the most recent probe, or nil.
func (w *Watchdog) LastErr() error {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.lastErr
}
