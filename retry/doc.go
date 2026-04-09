// Package retry implements a simple, context-aware retry mechanism used by the
// health checker when transient gRPC errors are encountered.
//
// # Usage
//
//	p := retry.DefaultPolicy()          // 3 attempts, 200 ms delay
//	err := retry.Do(ctx, p, func(ctx context.Context) error {
//	    return checker.Check(ctx, serviceName)
//	})
//
// # Policy
//
// A [Policy] controls two knobs:
//   - MaxAttempts – total number of invocations (first call + retries).
//   - Delay       – fixed pause between consecutive attempts.
//
// The context passed to [Do] is forwarded to every invocation of fn and is
// also checked between attempts; if it is cancelled the function returns
// immediately with the context error.
package retry
