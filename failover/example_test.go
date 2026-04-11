package failover_test

import (
	"context"
	"fmt"

	"github.com/yourorg/grpc-health-proxy/failover"
)

// ExamplePool_Do demonstrates using a Pool to perform a health check with
// automatic failover to a secondary target.
func ExamplePool_Do() {
	pool := failover.New([]string{"primary:443", "secondary:443"})

	// Simulate a primary that is down and a secondary that is healthy.
	healthy := map[string]bool{
		"primary:443":   false,
		"secondary:443": true,
	}

	target, err := pool.Do(context.Background(), func(_ context.Context, addr string) error {
		if !healthy[addr] {
			return fmt.Errorf("%s is unhealthy", addr)
		}
		return nil
	})
	if err != nil {
		fmt.Println("all targets unhealthy")
		return
	}
	fmt.Println("serving from:", target)
	// Output: serving from: secondary:443
}
