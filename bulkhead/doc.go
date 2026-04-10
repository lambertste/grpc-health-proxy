// Package bulkhead provides a concurrency-limiting HTTP middleware inspired
// by the bulkhead pattern from resilience engineering.
//
// A Bulkhead caps the number of requests that may be processed at the same
// time. When the cap is reached, new requests are rejected immediately with
// an HTTP 503 Service Unavailable response rather than being queued, which
// bounds memory growth and preserves latency for requests that are admitted.
//
// # Usage
//
//	b := bulkhead.New(50)          // allow up to 50 concurrent requests
//	http.Handle("/", b.Middleware(myHandler))
//
// The Bulkhead is safe for concurrent use by multiple goroutines and uses
// lock-free compare-and-swap operations internally.
package bulkhead
