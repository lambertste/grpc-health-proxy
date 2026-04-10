// Package shadow provides a request shadowing middleware that duplicates
// incoming HTTP requests to a secondary backend for dark-launch testing
// without affecting the primary response.
package shadow

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

// Doer is a minimal interface for sending HTTP requests.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Shadow duplicates each request to a shadow backend asynchronously.
type Shadow struct {
	target  string
	client  Doer
	timeout time.Duration
}

// New returns a Shadow that mirrors requests to target.
// timeout controls how long the shadow request may run; zero uses 2 s.
func New(target string, client Doer, timeout time.Duration) *Shadow {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	return &Shadow{target: target, client: client, timeout: timeout}
}

// Middleware returns an http.Handler that forwards the request to next and
// concurrently mirrors it to the shadow backend, discarding the shadow response.
func (s *Shadow) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body []byte
		if r.Body != nil {
			body, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(body))
		}

		go s.mirror(r, body)

		next.ServeHTTP(w, r)
	})
}

func (s *Shadow) mirror(orig *http.Request, body []byte) {
	url := s.target + orig.RequestURI
	req, err := http.NewRequest(orig.Method, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	for k, vs := range orig.Header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}
