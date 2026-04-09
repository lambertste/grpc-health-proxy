// Package health provides gRPC health check functionality.
//
// This package implements a health checker that connects to gRPC services
// and performs health checks using the standard gRPC health checking protocol
// as defined in https://github.com/grpc/grpc/blob/master/doc/health-checking.md
//
// The Checker type provides methods to:
//   - Establish connections to gRPC services
//   - Perform health checks on specific services
//   - Manage connection lifecycle
//
// Example usage:
//
//	checker, err := health.NewChecker("localhost:50051", 5*time.Second)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if err := checker.Connect(); err != nil {
//	    log.Fatal(err)
//	}
//	defer checker.Close()
//
//	ctx := context.Background()
//	status, err := checker.Check(ctx, "my.service")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if status == health.StatusServing {
//	    fmt.Println("Service is healthy")
//	}
//
package health
