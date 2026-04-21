// Package stamp injects a fixed set of response headers (a "stamp") onto
// every outbound HTTP response. Typical uses include attaching a deployment
// region, a build version, or an environment name so that clients and
// load-balancers can identify which instance served a request.
package stamp

import "net/http"

// Stamp holds the ordered list of header key/value pairs to inject.
type Stamp struct {
	pairs []string // interleaved key, value, key, value …
}

// New creates a Stamp from an even-length list of key/value pairs.
// It panics if args is empty or has an odd length.
func New(kv ...string) *Stamp {
	if len(kv) == 0 || len(kv)%2 != 0 {
		panic("stamp: New requires a non-empty, even number of key/value arguments")
	}
	return &Stamp{pairs: kv}
}

// Headers returns a copy of the configured key/value pairs.
func (s *Stamp) Headers() []string {
	out := make([]string, len(s.pairs))
	copy(out, s.pairs)
	return out
}

// Middleware returns an http.Handler that writes all stamp headers before
// delegating to next.
func (s *Stamp) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		for i := 0; i < len(s.pairs)-1; i += 2 {
			h.Set(s.pairs[i], s.pairs[i+1])
		}
		next.ServeHTTP(w, r)
	})
}
