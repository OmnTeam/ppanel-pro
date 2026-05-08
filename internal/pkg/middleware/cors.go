package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/OmnTeam/npanel-pro/internal/conf"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

var defaultCORSConfig = &conf.Server_CORS{
	Enable:           true,
	AllowedOrigins:   []string{"*"},
	AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"},
	AllowedHeaders:   []string{"*"},
	ExposedHeaders:   []string{"Content-Length", "Content-Type", "Authorization"},
	AllowCredentials: true,
	MaxAge:           86400,
}

// CORSFilter returns a Kratos HTTP filter backed by server.cors config.
func CORSFilter(corsConfig *conf.Server_CORS) khttp.FilterFunc {
	config := normalizeCORSConfig(corsConfig)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !config.Enable {
				next.ServeHTTP(w, r)
				return
			}

			ApplyCORSHeaders(w.Header(), r, config)
			if IsCORSPreflightRequest(r) {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ApplyCORSHeaders applies the configured CORS headers onto the response.
func ApplyCORSHeaders(header http.Header, r *http.Request, corsConfig *conf.Server_CORS) {
	config := normalizeCORSConfig(corsConfig)
	origin := strings.TrimSpace(r.Header.Get("Origin"))

	if allowedOrigin := resolveAllowedOrigin(config, origin); allowedOrigin != "" {
		header.Set("Access-Control-Allow-Origin", allowedOrigin)
	}
	if config.AllowCredentials {
		header.Set("Access-Control-Allow-Credentials", "true")
	}

	header.Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
	header.Set("Access-Control-Allow-Headers", resolveAllowedHeaders(r, config))
	header.Set("Access-Control-Max-Age", strconv.FormatInt(int64(config.MaxAge), 10))

	if len(config.ExposedHeaders) > 0 {
		header.Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
	}

	appendVary(header, "Origin", "Access-Control-Request-Method", "Access-Control-Request-Headers")
}

// IsCORSPreflightRequest reports whether the request is a browser preflight.
func IsCORSPreflightRequest(r *http.Request) bool {
	return r.Method == http.MethodOptions &&
		strings.TrimSpace(r.Header.Get("Origin")) != "" &&
		strings.TrimSpace(r.Header.Get("Access-Control-Request-Method")) != ""
}

func normalizeCORSConfig(corsConfig *conf.Server_CORS) *conf.Server_CORS {
	if corsConfig == nil {
		return cloneCORSConfig(defaultCORSConfig)
	}

	config := cloneCORSConfig(corsConfig)
	if len(config.AllowedOrigins) == 0 {
		config.AllowedOrigins = append([]string(nil), defaultCORSConfig.AllowedOrigins...)
	}
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = append([]string(nil), defaultCORSConfig.AllowedMethods...)
	}
	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = append([]string(nil), defaultCORSConfig.AllowedHeaders...)
	}
	if len(config.ExposedHeaders) == 0 {
		config.ExposedHeaders = append([]string(nil), defaultCORSConfig.ExposedHeaders...)
	}
	if config.MaxAge <= 0 {
		config.MaxAge = defaultCORSConfig.MaxAge
	}

	return config
}

func resolveAllowedOrigin(config *conf.Server_CORS, origin string) string {
	if len(config.AllowedOrigins) == 0 {
		return ""
	}

	if contains(config.AllowedOrigins, "*") {
		if config.AllowCredentials {
			return origin
		}
		return "*"
	}

	if origin != "" && contains(config.AllowedOrigins, origin) {
		return origin
	}

	return ""
}

func resolveAllowedHeaders(r *http.Request, config *conf.Server_CORS) string {
	if len(config.AllowedHeaders) == 0 {
		return ""
	}
	if contains(config.AllowedHeaders, "*") {
		if requestHeaders := strings.TrimSpace(r.Header.Get("Access-Control-Request-Headers")); requestHeaders != "" {
			return requestHeaders
		}
		return "Content-Type, Authorization, X-Requested-With, Accept, Origin"
	}
	return strings.Join(config.AllowedHeaders, ", ")
}

func appendVary(header http.Header, values ...string) {
	seen := make(map[string]struct{})
	merged := make([]string, 0, len(values))

	for _, existing := range header.Values("Vary") {
		for _, part := range strings.Split(existing, ",") {
			token := strings.TrimSpace(part)
			if token == "" {
				continue
			}
			if _, ok := seen[token]; ok {
				continue
			}
			seen[token] = struct{}{}
			merged = append(merged, token)
		}
	}

	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			token := strings.TrimSpace(part)
			if token == "" {
				continue
			}
			if _, ok := seen[token]; ok {
				continue
			}
			seen[token] = struct{}{}
			merged = append(merged, token)
		}
	}

	if len(merged) > 0 {
		header.Set("Vary", strings.Join(merged, ", "))
	}
}

func cloneCORSConfig(corsConfig *conf.Server_CORS) *conf.Server_CORS {
	if corsConfig == nil {
		return nil
	}
	return &conf.Server_CORS{
		Enable:           corsConfig.Enable,
		AllowedOrigins:   append([]string(nil), corsConfig.AllowedOrigins...),
		AllowedMethods:   append([]string(nil), corsConfig.AllowedMethods...),
		AllowedHeaders:   append([]string(nil), corsConfig.AllowedHeaders...),
		ExposedHeaders:   append([]string(nil), corsConfig.ExposedHeaders...),
		AllowCredentials: corsConfig.AllowCredentials,
		MaxAge:           corsConfig.MaxAge,
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.TrimSpace(s) == item {
			return true
		}
	}
	return false
}
