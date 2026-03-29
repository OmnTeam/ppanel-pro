package middleware

import (
	"net/http"
	"strings"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// LegacyPathCompatFilter keeps old Gin-style trailing-slash API paths working.
func LegacyPathCompatFilter() khttp.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(r.URL.Path) <= 1 || !strings.HasSuffix(r.URL.Path, "/") {
				next.ServeHTTP(w, r)
				return
			}

			trimmed := strings.TrimSuffix(r.URL.Path, "/")
			if trimmed == "" {
				trimmed = "/"
			}

			clone := r.Clone(r.Context())
			clone.URL.Path = trimmed
			if clone.URL.RawPath != "" {
				clone.URL.RawPath = strings.TrimSuffix(clone.URL.RawPath, "/")
			}
			clone.RequestURI = clone.URL.RequestURI()

			next.ServeHTTP(w, clone)
		})
	}
}
