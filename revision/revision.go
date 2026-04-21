// Package revision stamps every HTTP response with a build revision header.
// The header value is set once at construction time and injected into every
// outbound response, making it easy to correlate traffic with a specific
// deployed artifact.
package revision

import (
	"net/http"
	"strings"
)

const defaultHeader = "X-Revision"

// Stamper wraps an http.Handler and injects a build revision header.
type Stamper struct {
	next     http.Handler
	header   string
	revision string
}

// New returns a Stamper that writes revision into header on every response.
// If header is empty, X-Revision is used. If revision is empty or all
// whitespace the middleware is a no-op pass-through.
func New(next http.Handler, header, revision string) *Stamper {
	if next == nil {
		panic("revision: next handler must not be nil")
	}
	h := strings.TrimSpace(header)
	if h == "" {
		h = defaultHeader
	}
	return &Stamper{
		next:     next,
		header:   h,
		revision: strings.TrimSpace(revision),
	}
}

// Header returns the response header name used by the stamper.
func (s *Stamper) Header() string { return s.header }

// Revision returns the revision string injected into responses.
func (s *Stamper) Revision() string { return s.revision }

// ServeHTTP implements http.Handler.
func (s *Stamper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.revision != "" {
		w.Header().Set(s.header, s.revision)
	}
	s.next.ServeHTTP(w, r)
}
