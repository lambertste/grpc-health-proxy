package replay_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/your-org/grpc-health-proxy/replay"
)

func ExampleNew() {
	// Create a recorder with capacity for 100 requests.
	rec := replay.New(100)

	// Wrap your handler to start recording.
	production := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := rec.Middleware(production)

	// Simulate some incoming traffic.
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	fmt.Println("recorded:", rec.Len())

	// Replay the captured traffic against a candidate handler.
	candidate := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	results := rec.Replay(candidate)
	fmt.Println("replayed:", len(results))

	// Output:
	// recorded: 3
	// replayed: 3
}
