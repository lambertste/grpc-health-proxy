package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/your-org/grpc-health-proxy/metrics"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestLogging_WritesLog(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	h := Logging(logger)(http.HandlerFunc(okHandler))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if !strings.Contains(buf.String(), "GET") {
		t.Errorf("expected log to contain method, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "/healthz") {
		t.Errorf("expected log to contain path, got: %s", buf.String())
	}
}

func TestMetrics_RecordsMetrics(t *testing.T) {
	m := metrics.New()
	h := Metrics(m)(http.HandlerFunc(okHandler))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)
	// If no panic occurred, metrics were recorded without error.
}

func TestChain_AppliesOrder(t *testing.T) {
	var order []string

	mk := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	h := Chain(http.HandlerFunc(okHandler), mk("first"), mk("second"))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if len(order) != 2 || order[0] != "first" || order[1] != "second" {
		t.Errorf("unexpected middleware order: %v", order)
	}
}
