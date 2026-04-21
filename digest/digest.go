// Package digest provides an HTTP middleware that computes and attaches
// a content digest header to every response, allowing clients to verify
// response integrity.
package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

const defaultHeader = "X-Content-Digest"

// Digester wraps an http.Handler and appends a SHA-256 digest of the
// response body to a configurable response header.
type Digester struct {
	next   http.Handler
	header string
}

// New returns a Digester that writes the digest into header.
// If header is empty, X-Content-Digest is used.
func New(next http.Handler, header string) *Digester {
	if next == nil {
		panic("digest: next handler must not be nil")
	}
	if header == "" {
		header = defaultHeader
	}
	return &Digester{next: next, header: header}
}

// Header returns the response header name used for the digest.
func (d *Digester) Header() string { return d.header }

// ServeHTTP buffers the downstream response, computes its SHA-256 digest,
// sets the digest header, then writes the buffered response to w.
func (d *Digester) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rec := &recorder{header: make(http.Header), code: http.StatusOK}
	d.next.ServeHTTP(rec, r)

	sum := sha256.Sum256(rec.body)
	hex := "sha256=" + hex.EncodeToString(sum[:])

	for k, vals := range rec.header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set(d.header, hex)
	w.WriteHeader(rec.code)
	_, _ = w.Write(rec.body)
}

type recorder struct {
	header http.Header
	body   []byte
	code   int
}

func (r *recorder) Header() http.Header { return r.header }

func (r *recorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return len(b), nil
}

func (r *recorder) WriteHeader(code int) { r.code = code }
