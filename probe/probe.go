// Package probe provides configurable readiness and liveness probe
// strategies that can be composed to gate traffic or signal health.
package probe

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Result holds the outcome of a single probe execution.
type Result struct {
	OK      bool
	Message string
	Latency time.Duration
}

// Func is a function that performs a single probe check.
type Func func(ctx context.Context) Result

// Runner executes one or more probe functions and aggregates results.
type Runner struct {
	mu      sync.RWMutex
	probes  map[string]Func
	timeout time.Duration
}

const defaultTimeout = 5 * time.Second

// New creates a Runner with the given per-probe timeout.
// If timeout is zero the default of 5 s is used.
func New(timeout time.Duration) *Runner {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return &Runner{
		probes:  make(map[string]Func),
		timeout: timeout,
	}
}

// Register adds a named probe function to the runner.
// Registering the same name twice overwrites the previous entry.
func (r *Runner) Register(name string, fn Func) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.probes[name] = fn
}

// RunAll executes every registered probe concurrently and returns a
// map of name → Result. The runner's timeout is applied to each probe.
func (r *Runner) RunAll(ctx context.Context) map[string]Result {
	r.mu.RLock()
	names := make([]string, 0, len(r.probes))
	fns := make([]Func, 0, len(r.probes))
	for n, f := range r.probes {
		names = append(names, n)
		fns = append(fns, f)
	}
	r.mu.RUnlock()

	type entry struct {
		name string
		res  Result
	}
	ch := make(chan entry, len(names))

	for i, fn := range fns {
		go func(name string, fn Func) {
			pCtx, cancel := context.WithTimeout(ctx, r.timeout)
			defer cancel()
			start := time.Now()
			res := fn(pCtx)
			res.Latency = time.Since(start)
			ch <- entry{name, res}
		}(names[i], fn)
	}

	out := make(map[string]Result, len(names))
	for range names {
		e := <-ch
		out[e.name] = e.res
	}
	return out
}

// Healthy returns true only when every registered probe reports OK.
func (r *Runner) Healthy(ctx context.Context) (bool, error) {
	results := r.RunAll(ctx)
	for name, res := range results {
		if !res.OK {
			return false, fmt.Errorf("probe %q failed: %s", name, res.Message)
		}
	}
	return true, nil
}
