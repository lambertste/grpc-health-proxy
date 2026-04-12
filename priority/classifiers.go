package priority

import (
	"net/http"
	"strings"
)

// HeaderClassifier returns a Classifier that reads a priority level from the
// named HTTP header. Expected values are "low", "high"; anything else (or
// absent) maps to Normal.
func HeaderClassifier(header string) Classifier {
	return func(r *http.Request) Level {
		switch strings.ToLower(r.Header.Get(header)) {
		case "low":
			return Low
		case "high":
			return High
		default:
			return Normal
		}
	}
}

// PathPrefixClassifier returns a Classifier that assigns High to requests
// whose path starts with highPrefix, Low to those starting with lowPrefix,
// and Normal to all others.
func PathPrefixClassifier(highPrefix, lowPrefix string) Classifier {
	return func(r *http.Request) Level {
		switch {
		case highPrefix != "" && strings.HasPrefix(r.URL.Path, highPrefix):
			return High
		case lowPrefix != "" && strings.HasPrefix(r.URL.Path, lowPrefix):
			return Low
		default:
			return Normal
		}
	}
}

// ChainClassifier returns a Classifier that tries each classifier in order
// and returns the first non-Normal result, falling back to Normal.
func ChainClassifier(classifiers ...Classifier) Classifier {
	return func(r *http.Request) Level {
		for _, c := range classifiers {
			if lvl := c(r); lvl != Normal {
				return lvl
			}
		}
		return Normal
	}
}
