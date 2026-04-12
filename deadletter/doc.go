// Package deadletter implements a bounded dead-letter queue middleware for
// capturing HTTP requests that result in error responses.
//
// # Overview
//
// When a downstream handler returns a response whose status code satisfies a
// caller-supplied predicate, the request metadata is appended to an in-memory
// queue for later inspection or alerting.
//
// # Usage
//
//	q := deadletter.New(200)          // retain at most 200 entries
//
//	wrapped := q.Middleware(
//	    deadletter.IsError,           // capture 5xx responses
//	    myHandler,
//	)
//
//	// Inspect captured entries at any time:
//	for _, e := range q.Entries() {
//	    log.Printf("%s %s => %d at %s", e.Method, e.Path, e.StatusCode, e.Timestamp)
//	}
//
// # Capacity
//
// The queue is bounded. When the capacity is reached the oldest entry is
// evicted before the new one is appended, keeping memory usage constant.
package deadletter
