// Package sticky provides session-affinity (sticky routing) middleware for
// HTTP services.
//
// # Overview
//
// Sticky extracts a routing key from every inbound request and uses a
// consistent hash to select one of the registered backends. Requests that
// share the same key are always forwarded to the same backend for the
// lifetime of the process, which is useful when upstream services maintain
// in-memory session state.
//
// # Extractors
//
// Two built-in extractors are provided:
//
//   - CookieExtractor – uses a named cookie value as the routing key.
//   - HeaderExtractor – uses a named HTTP header value as the routing key.
//
// Any func(r *http.Request) string can serve as a custom Extractor.
//
// # Usage
//
//	h := sticky.New(
//		sticky.CookieExtractor("session_id"),
//		backendA,
//		backendB,
//		backendC,
//	)
//	http.Handle("/", h)
package sticky
