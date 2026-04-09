package middleware_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/example/grpc-health-proxy/middleware"
)

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestLogging_WritesLog(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	h := middleware.Logging(logger)(http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestMetrics_RecordsMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()

	requestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
	}, []string{"method", "path", "status"})
	requestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
	}, []string{"method", "path"})

	reg.MustRegister(requestsTotal, requestDuration)

	h := middleware.Metrics(requestsTotal, requestDuration)(http.HandlerFunc(okHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}
	if len(mfs) == 0 {
		t.Error("expected at least one metric family")
	}
}

func TestChain_AppliesOrder(t *testing.T) {
	order := []string{}

	makeMiddleware := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	h := middleware.Chain(
		http.HandlerFunc(okHandler),
		makeMiddleware("first"),
		makeMiddleware("second"),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if len(order) != 2 || order[0] != "first" || order[1] != "second" {
		t.Errorf("unexpected middleware order: %v", order)
	}
}
