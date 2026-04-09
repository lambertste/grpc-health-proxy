// Package cache implements a lightweight, thread-safe, TTL-based in-memory
// cache for gRPC health check results.
//
// # Overview
//
// When the proxy receives an HTTP health request it can serve the answer from
// cache rather than performing a live gRPC call on every request. This reduces
// latency and protects upstream services from thundering-herd traffic during
// high-volume load-balancer polling.
//
// # Usage
//
//	c := cache.New(3 * time.Second)
//
//	// Store a result.
//	c.Set("myservice", true)
//
//	// Retrieve a result.
//	if healthy, ok := c.Get("myservice"); ok {
//		// serve from cache
//	}
//
//	// Remove a single entry.
//	c.Invalidate("myservice")
//
//	// Evict all stale entries (call periodically if desired).
//	c.Purge()
//
// A TTL of zero or less disables caching entirely; every Get call will return
// a miss and every Set call is a no-op.
package cache
