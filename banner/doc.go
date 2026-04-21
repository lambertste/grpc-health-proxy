// Package banner provides HTTP middleware that attaches a fixed set of
// response headers to every reply.
//
// # Use-cases
//
//   - Advertising the serving component (X-Served-By)
//   - Exposing the deployed API version (X-API-Version)
//   - Broadcasting deprecation notices (X-Deprecation-Date)
//   - Injecting CORS or security headers in a single place
//
// # Quick start
//
//	b := banner.New(
//	    "X-Served-By",    "grpc-health-proxy",
//	    "X-API-Version",  "v2",
//	)
//	http.Handle("/", b.Middleware(myHandler))
//
// # Thread safety
//
// A Banner must not be mutated (via Set) after its middleware has been
// handed to an HTTP server. Concurrent reads from multiple goroutines
// are safe once the Banner is fully initialised.
package banner
