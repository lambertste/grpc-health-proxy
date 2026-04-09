// Package ratelimit implements a token-bucket rate limiter for the
// grpc-health-proxy HTTP server.
//
// # Overview
//
// A Limiter is created with a sustained token refill rate (tokens per second)
// and a maximum burst capacity. Each incoming request consumes one token. When
// the bucket is empty the Limiter denies the request and the Middleware helper
// responds with HTTP 429 Too Many Requests.
//
// # Usage
//
//	limiter := ratelimit.New(100, 200) // 100 req/s, burst of 200
//	http.Handle("/healthz", limiter.Middleware(healthHandler))
//
// The Limiter is safe for concurrent use.
package ratelimit
