package middleware

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/OmnTeam/ppanel-pro/internal/conf"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

const (
	legacySubscribeConfigPath = "/api/subscribe"
	legacySubscribeTokenBase  = "/api/subscribe"
)

const modernSubscribeBase = "/v1/subscribe"

// SubscribePathCompatFilter enforces the old project's subscribe_path behavior:
// once a custom path is configured, only that path is allowed externally.
func SubscribePathCompatFilter(appConf *conf.Application) khttp.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if shouldBlockSubscribePath(appConf, r.URL.Path) {
				http.NotFound(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func shouldBlockSubscribePath(appConf *conf.Application, requestPath string) bool {
	customPath := currentSubscribePath(appConf)
	if customPath == "" || requestPath == "" {
		return false
	}

	for _, candidate := range subscribePathCandidates(customPath) {
		if requestPath == candidate {
			return false
		}
		prefix := strings.TrimRight(candidate, "/") + "/"
		if strings.HasPrefix(requestPath, prefix) {
			return false
		}
	}

	if customPath != legacySubscribeConfigPath && (requestPath == legacySubscribeConfigPath || strings.HasPrefix(requestPath, legacySubscribeTokenBase+"/")) {
		return true
	}
	if requestPath == modernSubscribeBase+"/config" || strings.HasPrefix(requestPath, modernSubscribeBase+"/") {
		return true
	}
	return false
}

func currentSubscribePath(appConf *conf.Application) string {
	if appConf == nil || appConf.Subscribe == nil {
		return legacySubscribeConfigPath
	}

	customPath := strings.TrimSpace(appConf.Subscribe.SubscribePath)
	if customPath == "" {
		customPath = legacySubscribeConfigPath
	}
	if !strings.HasPrefix(customPath, "/") {
		customPath = "/" + customPath
	}
	customPath = strings.TrimRight(customPath, "/")
	if customPath == "" {
		customPath = legacySubscribeConfigPath
	}
	return customPath
}

func subscribePathCandidates(customPath string) []string {
	candidates := []string{customPath}
	if subscribeGatewayModeEnabled() {
		candidates = append(candidates, "/sub"+customPath)
	}
	return candidates
}

func subscribeGatewayModeEnabled() bool {
	value, exists := os.LookupEnv("GATEWAY_MODE")
	if !exists || strings.TrimSpace(value) != "true" {
		return false
	}

	port, exists := os.LookupEnv("GATEWAY_PORT")
	if !exists || strings.TrimSpace(port) == "" {
		return false
	}

	if _, err := strconv.Atoi(strings.TrimSpace(port)); err != nil {
		return false
	}

	return true
}
