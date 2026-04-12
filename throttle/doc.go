// Package throttle implements a concurrency-throttling middleware for HTTP
// servers.
//
// Unlike a rate limiter, which caps the number of requests per unit of time,
// a throttle caps the number of requests that are actively being handled at
// the same moment. Excess requests are held in a bounded queue; if no slot
// becomes available within the configured wait duration, or if the request
// context is cancelled, the middleware immediately responds with HTTP 429
// (Too Many Requests).
//
// # Usage
//
//	th := throttle.New(
//		10,              // max concurrent requests
//		20,              // backlog queue depth
//		3*time.Second,   // max time to wait for a slot
//	)
//
//	http.ListenAndServe(":8080", th.Middleware(myHandler))
//
// # Metrics
//
// Call Active() at any time to observe the number of in-flight requests.
package throttle
