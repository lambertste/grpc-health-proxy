// Package burst implements a token-bucket burst limiter for HTTP services.
//
// A Limiter is created with a maximum capacity (burst size) and a fill rate
// (tokens per second). Tokens are added continuously up to the capacity. Each
// request consumes one token; requests that arrive when the bucket is empty are
// rejected with HTTP 429.
//
// # Usage
//
//	limiter := burst.New(20, 5) // burst of 20, refill 5/sec
//	http.Handle("/", limiter.Middleware(myHandler))
//
// The limiter is safe for concurrent use.
package burst
