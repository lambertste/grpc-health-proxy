package tls

import "os"

// ConfigFromEnv builds a Config by reading well-known environment variables.
//
//	GRPC_TLS_CA_CERT            – path to PEM-encoded CA certificate
//	GRPC_TLS_CLIENT_CERT        – path to PEM-encoded client certificate
//	GRPC_TLS_CLIENT_KEY         – path to PEM-encoded client private key
//	GRPC_TLS_INSECURE_SKIP_VERIFY – set to "true" to skip server verification
func ConfigFromEnv() Config {
	return Config{
		CACert:             os.Getenv("GRPC_TLS_CA_CERT"),
		ClientCert:         os.Getenv("GRPC_TLS_CLIENT_CERT"),
		ClientKey:          os.Getenv("GRPC_TLS_CLIENT_KEY"),
		InsecureSkipVerify: os.Getenv("GRPC_TLS_INSECURE_SKIP_VERIFY") == "true",
	}
}
