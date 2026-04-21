// Package digest implements an HTTP middleware that computes a SHA-256
// digest of every response body and attaches it as a response header.
//
// # Overview
//
// Clients can use the digest to verify that the response they received
// was not corrupted or tampered with in transit. The header value follows
// the format:
//
//	sha256=<hex-encoded-hash>
//
// # Usage
//
//	d := digest.New(next, "X-Content-Digest")
//	http.ListenAndServe(":8080", d)
//
// If the header argument is empty, X-Content-Digest is used.
package digest
