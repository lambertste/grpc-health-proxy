package probe_test

import (
	"context"
	"fmt"
	"time"

	"github.com/yourorg/grpc-health-proxy/probe"
)

func ExampleRunner_Healthy() {
	r := probe.New(2 * time.Second)

	// Register a probe that always succeeds.
	r.Register("ping", func(_ context.Context) probe.Result {
		return probe.Result{OK: true, Message: "pong"}
	})

	ok, err := r.Healthy(context.Background())
	fmt.Println(ok, err)
	// Output: true <nil>
}

func ExampleRunner_RunAll() {
	r := probe.New(2 * time.Second)

	r.Register("db", func(_ context.Context) probe.Result {
		return probe.Result{OK: true, Message: "connected"}
	})
	r.Register("cache", func(_ context.Context) probe.Result {
		return probe.Result{OK: false, Message: "timeout"}
	})

	results := r.RunAll(context.Background())
	for name, res := range results {
		fmt.Printf("%s: ok=%v msg=%s\n", name, res.OK, res.Message)
	}
}
