package admission

import (
	"net/http"
	"strings"
)

// MethodAllowlist returns a Predicate that admits only requests whose
// HTTP method is in the provided list.
func MethodAllowlist(methods ...string) Predicate {
	allowed := make(map[string]struct{}, len(methods))
	for _, m := range methods {
		allowed[strings.ToUpper(m)] = struct{}{}
	}
	return func(r *http.Request) bool {
		_, ok := allowed[r.Method]
		return ok
	}
}

// PathPrefix returns a Predicate that admits only requests whose URL
// path starts with one of the given prefixes.
func PathPrefix(prefixes ...string) Predicate {
	return func(r *http.Request) bool {
		for _, p := range prefixes {
			if strings.HasPrefix(r.URL.Path, p) {
				return true
			}
		}
		return false
	}
}

// HeaderRequired returns a Predicate that admits only requests that
// carry the specified header with the given value. An empty value
// means the header must merely be present.
func HeaderRequired(name, value string) Predicate {
	return func(r *http.Request) bool {
		v := r.Header.Get(name)
		if v == "" {
			return false
		}
		if value == "" {
			return true
		}
		return v == value
	}
}
