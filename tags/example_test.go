package tags_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"grpc-health-proxy/tags"
)

func ExampleMiddleware() {
	// A middleware that stamps a region tag on every request.
	stampRegion := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if tg := tags.FromContext(r.Context()); tg != nil {
				tg.Set("region", "eu-west-1")
			}
			next.ServeHTTP(w, r)
		})
	}

	// The final handler reads the tag.
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tg := tags.FromContext(r.Context())
		region, _ := tg.Get("region")
		fmt.Fprintln(w, region)
	})

	// Chain: inject tags → stamp region → final handler.
	handler := tags.Middleware(stampRegion(final))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Print(rec.Body.String())
	// Output: eu-west-1
}
