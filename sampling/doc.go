// Package sampling provides probabilistic request sampling for the
// grpc-health-proxy sidecar.
//
// # Overview
//
// A Sampler is created with a target rate in the range [0.0, 1.0] and a
// Collector callback. For each incoming HTTP request the sampler performs
// an independent Bernoulli trial: if the trial succeeds, a shallow clone of
// the request is forwarded to the Collector before the request continues
// through the normal handler chain.
//
// The primary response path is never blocked by the collector; the callback
// is invoked synchronously but the caller controls whether it fans out to a
// goroutine.
//
// # Usage
//
//	sampler := sampling.New(0.1, func(r *http.Request) {
//		log.Printf("sampled %s %s", r.Method, r.URL.Path)
//	})
//	http.Handle("/", sampler.Middleware(myHandler))
package sampling
