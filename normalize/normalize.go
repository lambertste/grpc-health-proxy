// Package normalize provides HTTP middleware that canonicalises request paths
// before they reach downstream handlers.
//
// It cleans dot segments, collapses duplicate slashes, and optionally strips a
// configurable path prefix so that upstream routing rules stay simple.
package normalize

import (
	"net/http"
	"path"
	"strings"
)

// Option configures a Normalizer.
type Option func(*Normalizer)

// WithPrefix returns an Option that strips the given prefix from every request
// path before forwarding. The prefix is stripped after path cleaning.
func WithPrefix(prefix string) Option {
	return func(n *Normalizer) {
		n.prefix = path.Clean("/" + strings.TrimPrefix(prefix, "/"))
	}
}

// WithTrailingSlash returns an Option that controls whether a trailing slash is
// appended to (true) or removed from (false) the cleaned path. The default
// behaviour leaves the trailing slash unchanged.
func WithTrailingSlash(add bool) Option {
	return func(n *Normalizer) {
		n.trailingSlash = &add
	}
}

// Normalizer is an HTTP middleware that cleans and optionally transforms
// request paths.
type Normalizer struct {
	next          http.Handler
	prefix        string
	trailingSlash *bool
}

// New creates a Normalizer that wraps next with path normalisation.
func New(next http.Handler, opts ...Option) *Normalizer {
	n := &Normalizer{next: next}
	for _, o := range opts {
		o(n)
	}
	return n
}

// ServeHTTP cleans the request path, strips any configured prefix, and
// forwards the request to the wrapped handler.
func (n *Normalizer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := path.Clean("/" + r.URL.Path)

	if n.prefix != "" && n.prefix != "/" {
		trimmed := strings.TrimPrefix(p, n.prefix)
		if trimmed == "" {
			trimmed = "/"
		}
		p = trimmed
	}

	if n.trailingSlash != nil {
		if *n.trailingSlash {
			if !strings.HasSuffix(p, "/") {
				p += "/"
			}
		} else {
			p = strings.TrimRight(p, "/")
			if p == "" {
				p = "/"
			}
		}
	}

	r2 := r.Clone(r.Context())
	r2.URL.Path = p
	n.next.ServeHTTP(w, r2)
}
