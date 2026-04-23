// Package reflect provides a middleware that injects request metadata
// (method, path, remote address, and protocol) as HTTP response headers
// for debugging and observability purposes.
package reflect

import (
	"net/http"
)

const (
	defaultMethodHeader  = "X-Reflect-Method"
	defaultPathHeader    = "X-Reflect-Path"
	defaultRemoteHeader  = "X-Reflect-Remote"
	defaultProtoHeader   = "X-Reflect-Proto"
)

// Reflector mirrors request metadata into response headers.
type Reflector struct {
	next         http.Handler
	methodHeader string
	pathHeader   string
	remoteHeader string
	protoHeader  string
}

// Option configures a Reflector.
type Option func(*Reflector)

// WithMethodHeader overrides the response header used for the HTTP method.
func WithMethodHeader(h string) Option {
	return func(r *Reflector) {
		if h != "" {
			r.methodHeader = h
		}
	}
}

// WithPathHeader overrides the response header used for the request path.
func WithPathHeader(h string) Option {
	return func(r *Reflector) {
		if h != "" {
			r.pathHeader = h
		}
	}
}

// WithRemoteHeader overrides the response header used for the remote address.
func WithRemoteHeader(h string) Option {
	return func(r *Reflector) {
		if h != "" {
			r.remoteHeader = h
		}
	}
}

// WithProtoHeader overrides the response header used for the protocol.
func WithProtoHeader(h string) Option {
	return func(r *Reflector) {
		if h != "" {
			r.protoHeader = h
		}
	}
}

// New creates a new Reflector that wraps next.
// It panics if next is nil.
func New(next http.Handler, opts ...Option) *Reflector {
	if next == nil {
		panic("reflect: next handler must not be nil")
	}
	r := &Reflector{
		next:         next,
		methodHeader: defaultMethodHeader,
		pathHeader:   defaultPathHeader,
		remoteHeader: defaultRemoteHeader,
		protoHeader:  defaultProtoHeader,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// ServeHTTP injects request metadata into the response headers before
// delegating to the wrapped handler.
func (r *Reflector) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set(r.methodHeader, req.Method)
	w.Header().Set(r.pathHeader, req.URL.Path)
	w.Header().Set(r.remoteHeader, req.RemoteAddr)
	w.Header().Set(r.protoHeader, req.Proto)
	r.next.ServeHTTP(w, req)
}
