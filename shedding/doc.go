// Package shedding provides adaptive load shedding for HTTP services.
//
// A Shedder maintains a sliding window of recent request outcomes and begins
// rejecting new requests with HTTP 503 once the observed error rate climbs
// above a configurable threshold. This protects downstream services from
// cascading failures by shedding excess load early.
//
// Basic usage:
//
//	// Shed load when more than 50 % of the last 100 requests failed.
//	shedder := shedding.New(0.5, 100)
//
//	http.Handle("/api", shedder.Middleware(myHandler))
//
// The window size controls how quickly the shedder reacts to changing
// conditions. A smaller window reacts faster but is more sensitive to
// transient spikes; a larger window smooths out noise at the cost of
// slower adaptation.
package shedding
