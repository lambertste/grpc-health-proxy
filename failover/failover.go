// Package failover provides a round-robin failover strategy for upstream
// gRPC health targets. When the primary target fails, requests are forwarded
// to the next available target in the pool.
package failover

import (
	"context"
	"errors"
	"sync/atomic"
)

// ErrNoTargets is returned when the target pool is empty.
var ErrNoTargets = errors.New("failover: no targets configured")

// ErrAllUnhealthy is returned when every target in the pool fails.
var ErrAllUnhealthy = errors.New("failover: all targets unhealthy")

// CheckFn is a function that performs a health check against a named target.
// It returns nil on success or a non-nil error on failure.
type CheckFn func(ctx context.Context, target string) error

// Pool holds an ordered list of targets and a monotonic counter used to
// implement round-robin selection of the initial target.
type Pool struct {
	targets []string
	counter atomic.Uint64
}

// New creates a Pool from the provided target addresses.
// It panics when targets is empty.
func New(targets []string) *Pool {
	if len(targets) == 0 {
		panic(ErrNoTargets)
	}
	p := &Pool{targets: make([]string, len(targets))}
	copy(p.targets, targets)
	return p
}

// Targets returns a copy of the configured target list.
func (p *Pool) Targets() []string {
	out := make([]string, len(p.targets))
	copy(out, p.targets)
	return out
}

// Do calls check against each target in round-robin order until one succeeds
// or all have been tried. The winning target address is returned together with
// a nil error. If every target fails, ErrAllUnhealthy is returned.
func (p *Pool) Do(ctx context.Context, check CheckFn) (string, error) {
	n := uint64(len(p.targets))
	start := p.counter.Add(1) - 1
	for i := uint64(0); i < n; i++ {
		target := p.targets[(start+i)%n]
		if err := check(ctx, target); err == nil {
			return target, nil
		}
	}
	return "", ErrAllUnhealthy
}
