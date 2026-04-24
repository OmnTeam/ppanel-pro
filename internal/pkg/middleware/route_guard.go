package middleware

import (
	"net/http"
	"strings"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// LegacyRouteGuardFilter blocks new-only HTTP routes that do not exist in the old project.
func LegacyRouteGuardFilter() khttp.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isBlockedLegacyExtraPath(r.URL.Path) {
				http.NotFound(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isBlockedLegacyExtraPath(path string) bool {
	switch path {
	case "/v1/admin/server/migrate/has",
		"/v1/admin/server/migrate/run",
		"/v1/admin/group/migrate",
		"/v1/auth/check-telephone":
		return true
	}

	if isBlockedGeneratedPaymentNotifyPath(path) {
		return true
	}

	return false
}

func isBlockedGeneratedPaymentNotifyPath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 5 {
		return false
	}

	if parts[0] != "v1" || parts[1] != "payment" || parts[4] != "notify" || strings.TrimSpace(parts[2]) == "" {
		return false
	}

	switch parts[3] {
	case "alipay", "epay", "stripe":
		return true
	default:
		return false
	}
}
