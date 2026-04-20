// Package clamp provides an HTTP middleware that constrains the status code
// written by a wrapped handler to a configurable [min, max] range.
//
// # Motivation
//
// Some third-party or generated handlers may emit status codes that are
// outside the range a downstream load balancer or API gateway can interpret
// correctly. Wrapping such handlers with a Clamper normalises their output
// without modifying the handler itself.
//
// # Usage
//
//	import "github.com/your-org/grpc-health-proxy/clamp"
//
//	// Allow only 2xx–4xx to reach the load balancer.
//	h := clamp.New(myHandler, 200, 499)
//	http.ListenAndServe(":8080", h)
//
// Zero values for min and max fall back to 100 and 599 respectively.
// New panics if min > max.
package clamp
