package shedding_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/yourorg/grpc-health-proxy/shedding"
)

// ExampleNew demonstrates wiring a Shedder into an HTTP handler chain.
func ExampleNew() {
	// Reject requests when > 60 % of the last 50 requests returned 5xx.
	s := shedding.New(0.6, 50)

	// Simulate a healthy upstream.
	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := s.Middleware(upstream)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	fmt.Println(rec.Code)
	// Output: 200
}
