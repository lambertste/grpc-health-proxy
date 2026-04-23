package canary_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/your-org/grpc-health-proxy/canary"
)

func ExampleNew() {
	primary := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "primary")
	})
	v2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "canary-v2")
	})

	// Route 10 % of traffic to the v2 canary.
	handler := canary.New(primary, v2, 10)

	// Demonstrate that the rate can be adjusted at runtime.
	handler.SetRate(25)
	fmt.Println("rate:", handler.Rate())

	// A single request will always go to the primary at rate=0.
	handler.SetRate(0)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	fmt.Print(rec.Body.String())

	// Output:
	// rate: 25
	// primary
}
