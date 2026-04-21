// Package epoch provides a middleware that injects the current Unix
// timestamp (in seconds) into every HTTP response as a configurable header.
// It is useful for debugging cache freshness, clock skew, and response
// ordering in distributed systems.
package epoch

import (
	"fmt"
	"net/http"
	"time"
)

const defaultHeader = "X-Epoch"

// Stamper injects a Unix-epoch timestamp header into every response.
type Stamper struct {
	header string
	now    func() time.Time
}

// New creates a Stamper that writes the current Unix timestamp to header.
// If header is empty the default "X-Epoch" is used.
func New(header string) *Stamper {
	if header == "" {
		header = defaultHeader
	}
	return &Stamper{
		header: header,
		now:    time.Now,
	}
}

// Header returns the response header name used by this Stamper.
func (s *Stamper) Header() string { return s.header }

// Middleware returns an http.Handler that stamps every response with the
// current Unix epoch second before delegating to next.
func (s *Stamper) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(s.header, fmt.Sprintf("%d", s.now().Unix()))
		next.ServeHTTP(w, r)
	})
}
