// Package keepalive provides HTTP middleware that advertises connection
// keep-alive parameters to downstream clients via standard response headers.
//
// # Overview
//
// Modern HTTP/1.1 connections are persistent by default, but many clients and
// load balancers rely on explicit Connection and Keep-Alive headers to decide
// how long to reuse a connection and how many requests to pipeline over it.
//
// # Usage
//
//	h := keepalive.New(myHandler, keepalive.Policy{
//		Timeout:     45 * time.Second,
//		MaxRequests: 200,
//	})
//	http.ListenAndServe(":8080", h)
//
// Zero values are replaced with defaults (30 s timeout, 100 max requests).
package keepalive
