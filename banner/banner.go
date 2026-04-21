// Package banner provides a middleware that injects a configurable
// response header (or set of headers) into every HTTP response.
// It is useful for advertising server identity, API version, or
// deprecation notices without modifying individual handlers.
package banner

import (
	"net/http"
)

// Banner holds a set of key/value pairs that will be added to every
// response written through its middleware.
type Banner struct {
	headers map[string]string
}

// New creates a Banner that will attach the supplied key/value pairs
// as response headers. Duplicate keys are silently overwritten; the
// last value wins.
func New(pairs ...string) *Banner {
	if len(pairs)%2 != 0 {
		panic("banner: New requires an even number of arguments (key, value, ...)")
	}
	h := make(map[string]string, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		h[pairs[i]] = pairs[i+1]
	}
	return &Banner{headers: h}
}

// Headers returns a copy of the configured key/value pairs.
func (b *Banner) Headers() map[string]string {
	copy := make(map[string]string, len(b.headers))
	for k, v := range b.headers {
		copy[k] = v
	}
	return copy
}

// Set adds or replaces a single header entry on the banner.
func (b *Banner) Set(key, value string) {
	b.headers[key] = value
}

// Middleware returns an http.Handler that writes all banner headers
// before delegating to next.
func (b *Banner) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range b.headers {
			w.Header().Set(k, v)
		}
		next.ServeHTTP(w, r)
	})
}
