// Package replay provides request recording and replay for HTTP handlers.
//
// A Recorder wraps any http.Handler and captures a bounded ring-buffer of
// incoming requests. Captured entries can later be replayed verbatim against
// any handler — useful for canary validation, regression testing, and
// production traffic mirroring in development environments.
//
// # Usage
//
//	rec := replay.New(500)          // keep last 500 requests
//	http.Handle("/", rec.Middleware(myHandler))
//
//	// later — replay captured traffic against a new handler version
//	results := rec.Replay(candidateHandler)
//
// # Thread Safety
//
// All methods on Recorder are safe for concurrent use.
package replay
