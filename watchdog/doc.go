// Package watchdog provides a lightweight periodic health-check loop.
//
// A Watchdog wraps a user-supplied Probe function and calls it on a
// configurable interval. The latest health state is available at any
// time via Healthy and LastErr, making it easy to integrate with
// liveness/readiness endpoints or circuit breakers.
//
// # Basic usage
//
//	probe := func(ctx context.Context) error {
//		return grpcConn.Invoke(ctx, "/grpc.health.v1.Health/Check", req, resp)
//	}
//
//	wd := watchdog.New(probe, 5*time.Second)
//	go wd.Run(ctx)
//
//	// Later, check the cached result without blocking:
//	if !wd.Healthy() {
//		log.Println("upstream gRPC service is unhealthy:", wd.LastErr())
//	}
package watchdog
