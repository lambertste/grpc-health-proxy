package admission_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/grpc-health-proxy/admission"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestAllow_NoPredicatesAlwaysAdmits(t *testing.T) {
	c := admission.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if !c.Allow(req) {
		t.Fatal("expected request to be admitted with no predicates")
	}
	if c.Admitted() != 1 {
		t.Fatalf("admitted counter: got %d, want 1", c.Admitted())
	}
}

func TestAllow_PredicateRejects(t *testing.T) {
	rejectAll := func(_ *http.Request) bool { return false }
	c := admission.New(rejectAll)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if c.Allow(req) {
		t.Fatal("expected request to be rejected")
	}
	if c.Rejected() != 1 {
		t.Fatalf("rejected counter: got %d, want 1", c.Rejected())
	}
}

func TestAllow_AllPredicatesMustPass(t *testing.T) {
	allowAll := func(_ *http.Request) bool { return true }
	rejectAll := func(_ *http.Request) bool { return false }
	c := admission.New(allowAll, rejectAll)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if c.Allow(req) {
		t.Fatal("expected rejection when any predicate fails")
	}
}

func TestMiddleware_Returns503WhenRejected(t *testing.T) {
	rejectAll := func(_ *http.Request) bool { return false }
	c := admission.New(rejectAll)
	h := c.Middleware(http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status: got %d, want 503", rec.Code)
	}
}

func TestMiddleware_PassesThroughWhenAdmitted(t *testing.T) {
	allowAll := func(_ *http.Request) bool { return true }
	c := admission.New(allowAll)
	h := c.Middleware(http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
}

func TestCounters_TrackBothOutcomes(t *testing.T) {
	calls := 0
	pred := func(_ *http.Request) bool {
		calls++
		return calls%2 == 0 // admit even-numbered calls
	}
	c := admission.New(pred)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for i := 0; i < 4; i++ {
		c.Allow(req)
	}
	if c.Admitted() != 2 {
		t.Fatalf("admitted: got %d, want 2", c.Admitted())
	}
	if c.Rejected() != 2 {
		t.Fatalf("rejected: got %d, want 2", c.Rejected())
	}
}
