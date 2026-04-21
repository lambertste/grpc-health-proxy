// Package overload implements adaptive CPU-based load shedding for HTTP
// services.
//
// A Guard periodically samples CPU utilisation through a pluggable Sampler
// function. When utilisation exceeds the configured threshold the Guard begins
// probabilistically rejecting requests: the rejection probability increases
// linearly from 0 % at the threshold to 100 % at full saturation (1.0).
//
// Usage:
//
//	g := overload.New(0.75, time.Second, myCPUSampler)
//	defer g.Stop()
//
//	http.Handle("/", g.Middleware(myHandler))
//
// The Sampler is intentionally decoupled from the package so that callers can
// integrate with any system-metrics library (e.g. gopsutil, procfs) without
// adding a hard dependency to this module.
package overload
