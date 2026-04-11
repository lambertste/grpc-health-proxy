// Package coalesce implements request coalescing (also known as request
// deduplication or "singleflight"). It is useful when many concurrent
// callers issue logically identical upstream requests — for example,
// multiple HTTP handlers all checking the health of the same backend at
// the same instant.
//
// # Usage
//
//	g := coalesce.New()
//
//	result, err := g.Do(ctx, "backend-health", func(ctx context.Context) (interface{}, error) {
//		return checker.Check(ctx)
//	})
//
// Only one call to checker.Check is in flight at any given time for the
// key "backend-health". All other callers block and share the result.
//
// # Behaviour
//
//   - Keys are scoped to the Group; different keys execute independently.
//   - Once fn returns, the key is removed so the next call starts fresh.
//   - Errors are shared identically to values.
package coalesce
