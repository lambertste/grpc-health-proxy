package admission_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/grpc-health-proxy/admission"
)

func TestMethodAllowlist_AdmitsAllowedMethod(t *testing.T) {
	p := admission.MethodAllowlist("GET", "POST")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if !p(req) {
		t.Fatal("expected GET to be admitted")
	}
}

func TestMethodAllowlist_RejectsForbiddenMethod(t *testing.T) {
	p := admission.MethodAllowlist("GET")
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	if p(req) {
		t.Fatal("expected DELETE to be rejected")
	}
}

func TestPathPrefix_AdmitsMatchingPath(t *testing.T) {
	p := admission.PathPrefix("/health", "/ready")
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	if !p(req) {
		t.Fatal("expected /health/live to be admitted")
	}
}

func TestPathPrefix_RejectsNonMatchingPath(t *testing.T) {
	p := admission.PathPrefix("/health")
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	if p(req) {
		t.Fatal("expected /metrics to be rejected")
	}
}

func TestHeaderRequired_AdmitsPresentHeader(t *testing.T) {
	p := admission.HeaderRequired("X-Api-Key", "secret")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Api-Key", "secret")
	if !p(req) {
		t.Fatal("expected request with correct header to be admitted")
	}
}

func TestHeaderRequired_RejectsMissingHeader(t *testing.T) {
	p := admission.HeaderRequired("X-Api-Key", "")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if p(req) {
		t.Fatal("expected request without header to be rejected")
	}
}

func TestHeaderRequired_RejectsWrongValue(t *testing.T) {
	p := admission.HeaderRequired("X-Api-Key", "correct")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Api-Key", "wrong")
	if p(req) {
		t.Fatal("expected request with wrong header value to be rejected")
	}
}
