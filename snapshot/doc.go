// Package snapshot provides ring-buffer capture of recent HTTP request
// metadata for use in diagnostics and debug sidecars.
//
// # Overview
//
// A Snapshotter wraps an http.Handler via Middleware, recording an Entry for
// every request that passes through. Entries include the timestamp, HTTP
// method, path, response status code, approximate latency in milliseconds,
// and a flat copy of the request headers.
//
// The buffer is bounded: once capacity is reached the oldest entry is evicted
// to make room for the newest, so memory usage stays constant.
//
// # Usage
//
//	s := snapshot.New(200)          // keep last 200 requests
//	http.Handle("/api", s.Middleware(myHandler))
//	http.Handle("/_debug/snapshots", s.Handler()) // JSON dump
//
// # Thread Safety
//
// All methods are safe for concurrent use.
package snapshot
