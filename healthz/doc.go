// Package healthz exposes liveness and readiness HTTP endpoints for the
// grpc-health-proxy process itself.
//
// These endpoints are distinct from the upstream gRPC health checks translated
// by the health package. They allow orchestrators (Kubernetes, Nomad, etc.) to
// determine whether the proxy process is alive and ready to accept traffic.
//
// Endpoints
//
//	 GET /healthz/live  — liveness probe; always 200 while the process runs.
//	 GET /healthz/ready — readiness probe; 200 once SetReady(true) is called,
//	                      503 otherwise.
//
// Usage
//
//	h := healthz.New()
//	h.Register(mux)
//	// … after initialisation is complete:
//	h.SetReady(true)
package healthz
