// Package sticky implements session-affinity (sticky routing) middleware.
// It extracts a routing key from each request — via a cookie, header, or
// custom extractor — and consistently forwards the request to the same
// backend bucket so that stateful upstream services receive related
// requests from the same client.
package sticky

import (
	"hash/fnv"
	"net/http"
)

// Extractor derives a routing key from an HTTP request.
type Extractor func(r *http.Request) string

// Selector maps a routing key to one of the supplied handlers.
type Selector func(key string, backends []http.Handler) http.Handler

// Sticky routes each request to a consistent backend determined by the
// routing key returned by ext.
type Sticky struct {
	extract   Extractor
	select_  Selector
	backends []http.Handler
}

// New creates a Sticky middleware. It panics if backends is empty or ext
// is nil.
func New(ext Extractor, backends ...http.Handler) *Sticky {
	if ext == nil {
		panic("sticky: extractor must not be nil")
	}
	if len(backends) == 0 {
		panic("sticky: at least one backend is required")
	}
	return &Sticky{
		extract:  ext,
		select_: hashSelector,
		backends: backends,
	}
}

// ServeHTTP satisfies http.Handler.
func (s *Sticky) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := s.extract(r)
	s.select_(key, s.backends).ServeHTTP(w, r)
}

// CookieExtractor returns an Extractor that uses the value of the named
// cookie as the routing key. If the cookie is absent the empty string is
// returned, which still produces a stable (zero) bucket.
func CookieExtractor(name string) Extractor {
	return func(r *http.Request) string {
		c, err := r.Cookie(name)
		if err != nil {
			return ""
		}
		return c.Value
	}
}

// HeaderExtractor returns an Extractor that uses the value of the named
// HTTP header as the routing key.
func HeaderExtractor(name string) Extractor {
	return func(r *http.Request) string {
		return r.Header.Get(name)
	}
}

// hashSelector consistently maps key to a backend index via FNV-32a.
func hashSelector(key string, backends []http.Handler) http.Handler {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	idx := int(h.Sum32()) % len(backends)
	return backends[idx]
}
