package tls_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	tlspkg "github.com/yourorg/grpc-health-proxy/tls"
)

func writeTempPEM(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, data, 0o600); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

func generateSelfSigned(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyBytes, _ := x509.MarshalECPrivateKey(priv)
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	return
}

func TestBuild_Defaults(t *testing.T) {
	cfg := tlspkg.Config{}
	tlsCfg, err := cfg.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tlsCfg == nil {
		t.Fatal("expected non-nil tls.Config")
	}
}

func TestBuild_WithCACert(t *testing.T) {
	certPEM, _ := generateSelfSigned(t)
	dir := t.TempDir()
	caPath := writeTempPEM(t, dir, "ca.pem", certPEM)

	cfg := tlspkg.Config{CACert: caPath}
	tlsCfg, err := cfg.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tlsCfg.RootCAs == nil {
		t.Fatal("expected RootCAs to be set")
	}
}

func TestBuild_WithClientCert(t *testing.T) {
	certPEM, keyPEM := generateSelfSigned(t)
	dir := t.TempDir()
	certPath := writeTempPEM(t, dir, "cert.pem", certPEM)
	keyPath := writeTempPEM(t, dir, "key.pem", keyPEM)

	cfg := tlspkg.Config{ClientCert: certPath, ClientKey: keyPath}
	tlsCfg, err := cfg.Build()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tlsCfg.Certificates) != 1 {
		t.Fatalf("expected 1 certificate, got %d", len(tlsCfg.Certificates))
	}
}

func TestBuild_MissingKeyReturnsError(t *testing.T) {
	certPEM, _ := generateSelfSigned(t)
	dir := t.TempDir()
	certPath := writeTempPEM(t, dir, "cert.pem", certPEM)

	cfg := tlspkg.Config{ClientCert: certPath}
	_, err := cfg.Build()
	if err == nil {
		t.Fatal("expected error when key is missing")
	}
}

func TestBuild_BadCACertReturnsError(t *testing.T) {
	dir := t.TempDir()
	caPath := writeTempPEM(t, dir, "ca.pem", []byte("not a cert"))

	cfg := tlspkg.Config{CACert: caPath}
	_, err := cfg.Build()
	if err == nil {
		t.Fatal("expected error for invalid CA cert")
	}
}
