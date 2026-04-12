// Package pendingreq implements a pending-request limiter for HTTP servers.
//
// Unlike a pure concurrency limiter that rejects all requests beyond the
// active limit, pendingreq counts every request that has been accepted but
// not yet completed. This makes it suitable for protecting upstreams from
// queue pile-ups when combined with asynchronous or buffered transports.
//
// # Basic usage
//
//	l := pendingreq.New(100)
//	http.ListenAndServe(":8080", l.Middleware(myHandler))
//
// When the number of in-flight requests reaches the configured maximum,
// the middleware returns HTTP 429 Too Many Requests to new arrivals until
// a slot is freed.
package pendingreq
