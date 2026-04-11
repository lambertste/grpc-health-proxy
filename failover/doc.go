// Package failover implements a round-robin failover pool for upstream gRPC
// health targets.
//
// # Overview
//
// A Pool is constructed with an ordered list of target addresses. On each
// call to Do the pool selects a starting index using a monotonically
// incrementing counter (round-robin) and iterates through the targets,
// invoking the caller-supplied CheckFn until one succeeds.
//
// # Usage
//
//	pool := failover.New([]string{
//		"primary:443",
//		"secondary:443",
//	})
//
//	target, err := pool.Do(ctx, func(ctx context.Context, addr string) error {
//		return checker.Check(ctx, addr)
//	})
//
// If all targets fail, Do returns ErrAllUnhealthy.
package failover
