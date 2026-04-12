// Package tags provides lightweight request-scoped tagging for HTTP handlers.
//
// Tags are key-value string pairs attached to each request via context.
// They are useful for propagating metadata (e.g. tenant ID, feature flags,
// environment labels) from early middleware to downstream handlers and
// observability tooling without modifying function signatures.
//
// # Usage
//
//	// Wrap your handler with the tagging middleware:
//	//   mux := tags.Middleware(yourHandler)
//
//	// Inside any downstream handler:
//	//   tg := tags.FromContext(r.Context())
//	//   if tg != nil {
//	//       tg.Set("region", "us-east-1")
//	//       region, _ := tg.Get("region")
//	//   }
//
// Tags instances are safe for concurrent use.
package tags
