// Package tls provides helpers for constructing *crypto/tls.Config values
// from file-based PEM credentials.
//
// # Usage
//
// Populate a Config with optional paths to a CA certificate and/or a
// client certificate/key pair, then call Build to obtain a *tls.Config
// ready for use with google.golang.org/grpc/credentials.NewTLS.
//
//	config := tls.Config{
//		CACert:     "/etc/certs/ca.pem",
//		ClientCert: "/etc/certs/client.pem",
//		ClientKey:  "/etc/certs/client-key.pem",
//	}
//	tlsCfg, err := config.Build()
//	if err != nil {
//		log.Fatal(err)
//	}
//	creds := credentials.NewTLS(tlsCfg)
package tls
