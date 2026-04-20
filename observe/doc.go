// Package observe provides a thin, backend-agnostic HTTP middleware for
// capturing per-request telemetry.
//
// # Overview
//
// Wrap any http.Handler with observe.New, supplying a Sink callback.  After
// each request completes the middleware calls the Sink with an Event that
// contains the HTTP method, URL path, response status code, and total
// handler latency.
//
// # Usage
//
//	var requestsTotal int64
//
//	sink := func(e observe.Event) {
//		atomic.AddInt64(&requestsTotal, 1)
//		log.Printf("method=%s path=%s status=%d latency=%s",
//			e.Method, e.Path, e.StatusCode, e.Latency)
//	}
//
//	h := observe.New(myHandler, sink)
//	http.ListenAndServe(":8080", h)
//
// The Sink is called synchronously in the request goroutine.  If you need
// non-blocking behaviour, wrap the Sink body in a goroutine or channel send.
package observe
