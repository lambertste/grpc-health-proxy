// Package reflect provides an HTTP middleware that mirrors inbound request
// metadata back to the caller as response headers.
//
// This is useful during development and debugging: clients can inspect the
// reflected headers to verify which method, path, remote address, and protocol
// the proxy or load balancer forwarded to the upstream.
//
// # Usage
//
//	rfl := reflect.New(
//		next,
//		reflect.WithMethodHeader("X-Debug-Method"),
//		reflect.WithPathHeader("X-Debug-Path"),
//	)
//	http.ListenAndServe(":8080", rfl)
//
// By default the middleware injects four headers:
//
//	X-Reflect-Method  – the HTTP verb (GET, POST, …)
//	X-Reflect-Path    – the URL path
//	X-Reflect-Remote  – the client remote address
//	X-Reflect-Proto   – the HTTP protocol version
//
// All header names are configurable via the With* option functions.
package reflect
