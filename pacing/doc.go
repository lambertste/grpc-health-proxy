// Package pacing smooths outgoing or incoming request rates by enforcing a
// configurable minimum interval between successive calls.
//
// Unlike a rate-limiter that rejects excess requests, a pacer queues them and
// releases each one only after the required interval has passed, producing a
// steady, predictable throughput.
//
// Basic usage:
//
//	pacer := pacing.New(20 * time.Millisecond) // max ~50 req/s
//	mux.Handle("/", pacer.Middleware(myHandler))
//
// The pacer is safe for concurrent use.
package pacing
