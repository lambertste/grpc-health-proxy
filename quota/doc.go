// Package quota provides a fixed-window per-key request quota enforcer.
//
// A Quota tracks how many times a given key (such as a client IP address or
// API token) has been seen within a rolling fixed-width time window. Once the
// configured limit is reached all further Allow calls for that key return
// false until the window resets.
//
// # Basic usage
//
//	// Allow each client IP up to 100 requests per minute.
//	q := quota.New(100, time.Minute)
//
//	http.Handle("/api", q.Middleware(
//	    func(r *http.Request) string { return r.RemoteAddr },
//	    myHandler,
//	))
//
// # Thread safety
//
// All methods on Quota are safe for concurrent use.
package quota
