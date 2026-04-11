package sampling_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/yourorg/grpc-health-proxy/sampling"
)

// ExampleNew demonstrates wiring a Sampler into an HTTP handler chain.
func ExampleNew() {
	var logged int
	collector := func(r *http.Request) {
		logged++
	}

	// Sample every request (rate = 1.0) for demonstration purposes.
	s := sampling.New(1.0, collector)

	handler := s.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(logged)
	// Output:
	// 200
	// 1
}
