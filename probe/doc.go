// Package probe provides composable health-probe primitives for
// grpc-health-proxy.
//
// # Overview
//
// A Runner holds a named set of Func probes. Each probe is a plain
// function that accepts a context and returns a Result indicating
// success or failure together with a human-readable message.
//
// # Usage
//
//	r := probe.New(3 * time.Second)
//
//	r.Register("grpc", func(ctx context.Context) probe.Result {
//		if err := checkGRPC(ctx); err != nil {
//			return probe.Result{OK: false, Message: err.Error()}
//		}
//		return probe.Result{OK: true, Message: "ok"}
//	})
//
//	ok, err := r.Healthy(ctx)
//
// All registered probes run concurrently; the runner is considered
// healthy only when every probe returns OK: true.
package probe
