// Package priority provides a weighted priority queue middleware for HTTP
// servers. It classifies incoming requests into three tiers — Low, Normal,
// and High — and enforces independent concurrency limits per tier so that
// low-priority traffic is shed first when the system is under load.
//
// # Basic usage
//
//	q := priority.New(
//		priority.HeaderClassifier("X-Priority"),
//		10,  // max concurrent Low requests
//		50,  // max concurrent Normal requests
//		0,   // High is uncapped
//	)
//
//	http.ListenAndServe(":8080", q.Middleware(myHandler))
//
// # Classifiers
//
// A Classifier is any func(*http.Request) Level. The package ships with
// HeaderClassifier, PathPrefixClassifier, and ChainClassifier helpers, but
// callers may supply any custom logic.
//
// # Limits
//
// A limit of 0 means the tier is uncapped. When a request exceeds its tier
// limit the middleware responds immediately with 429 Too Many Requests.
package priority
