// Package warmup provides a middleware that holds traffic during service
// initialisation and releases it once the service signals readiness.
//
// # Overview
//
// Many services require a brief warm-up period after startup — loading caches,
// establishing connection pools, or pre-computing data — before they can serve
// requests reliably. The warmup package exposes a simple Warmup type that
// tracks this state and a Middleware that returns 503 Service Unavailable
// until the service is considered ready.
//
// # Usage
//
//	w := warmup.New(10 * time.Second) // auto-ready after 10 s
//	// or call w.MarkReady() as soon as your init is done
//	http.Handle("/", w.Middleware(myHandler))
package warmup
