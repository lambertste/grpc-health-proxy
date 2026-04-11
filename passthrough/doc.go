// Package passthrough provides a conditional bypass middleware for HTTP
// handler chains.
//
// # Overview
//
// Certain requests — such as CORS preflight (OPTIONS), health probes, or
// static asset fetches — should skip expensive middleware like authentication
// or rate-limiting. passthrough lets you declare those conditions once and
// attach them at the router level.
//
// # Usage
//
//	h := passthrough.New(
//		myExpensiveHandler,
//		passthrough.Any(
//			passthrough.MethodIn(http.MethodOptions),
//			passthrough.PathExact("/healthz"),
//		),
//		cheapBypassHandler,
//	)
//	http.Handle("/", h)
//
// # Predicates
//
// Built-in predicates:
//   - PathExact  — exact URL path match
//   - MethodIn   — HTTP method membership
//   - Any        — logical OR over multiple predicates
//
// Custom predicates can be provided by implementing the Predicate function
// type: func(r *http.Request) bool.
package passthrough
