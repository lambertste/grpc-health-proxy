// Package tls provides helpers for loading and configuring TLS credentials
// used when dialling the upstream gRPC health endpoint.
package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// Config holds the paths required to build a TLS configuration.
type Config struct {
	// CACert is the path to the PEM-encoded CA certificate used to verify
	// the upstream server's certificate. Leave empty to use the system pool.
	CACert string

	// ClientCert and ClientKey are paths to a PEM-encoded certificate/key pair
	// used for mutual TLS. Both must be set together or both left empty.
	ClientCert string
	ClientKey  string

	// InsecureSkipVerify disables server certificate verification. Should only
	// be used in development environments.
	InsecureSkipVerify bool
}

// Build constructs a *tls.Config from the receiver. It returns an error if
// any of the referenced files cannot be read or parsed.
func (c Config) Build() (*tls.Config, error) {
	tlsCfg := &tls.Config{
		InsecureSkipVerify: c.InsecureSkipVerify, //nolint:gosec // intentional opt-in
	}

	if c.CACert != "" {
		pem, err := os.ReadFile(c.CACert)
		if err != nil {
			return nil, fmt.Errorf("tls: read CA cert %q: %w", c.CACert, err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("tls: no valid certificates found in %q", c.CACert)
		}
		tlsCfg.RootCAs = pool
	}

	if c.ClientCert != "" || c.ClientKey != "" {
		if c.ClientCert == "" || c.ClientKey == "" {
			return nil, fmt.Errorf("tls: client cert and key must both be provided")
		}
		cert, err := tls.LoadX509KeyPair(c.ClientCert, c.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("tls: load client key pair: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	return tlsCfg, nil
}
