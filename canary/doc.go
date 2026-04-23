// Package canary implements a traffic-splitting middleware for gradual
// rollouts.
//
// # Overview
//
// A Canary wraps two http.Handler values — a stable primary and a
// canary release — and routes each incoming request to one of them
// based on a configurable percentage.  The split decision is made
// independently for every request using a uniform random draw, so
// there is no per-connection affinity and no shared lock on the hot
// path.
//
// # Usage
//
//	primary := myapp.StableHandler()
//	v2     := myapp.CanaryHandler()
//
//	// Send 5 % of requests to the canary.
//	h := canary.New(primary, v2, 5)
//	http.Handle("/", h)
//
//	// Increase the canary share at runtime (e.g. after metrics look good).
//	h.SetRate(20)
//
// # Thread Safety
//
// SetRate and Rate are safe for concurrent use.  The rate is stored in
// an atomic integer so reads and writes never block each other.
package canary
