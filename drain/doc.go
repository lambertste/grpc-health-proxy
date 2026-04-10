// Package drain implements graceful shutdown for the grpc-health-proxy HTTP
// server. When the process receives a termination signal the server should
// stop accepting new connections while allowing in-flight health-check
// requests to finish naturally.
//
// Usage:
//
//	d := drain.New(10 * time.Second)
//
//	// Wrap your router so every request is tracked.
//	http.ListenAndServe(addr, d.Middleware(router))
//
//	// On SIGTERM / SIGINT:
//	if err := d.Shutdown(ctx); err != nil {
//		log.Printf("drain: shutdown deadline exceeded: %v", err)
//	}
//
// The Middleware method returns 503 Service Unavailable for any request that
// arrives after Shutdown has been called, ensuring that upstream load
// balancers quickly remove the instance from rotation.
package drain
