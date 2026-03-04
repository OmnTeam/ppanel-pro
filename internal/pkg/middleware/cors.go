package middleware

import (
	"net/http"
	"strings"

	"github.com/OmnTeam/ppanel-pro/internal/conf"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// CORSFilter 返回一个 CORS HTTP Filter，使用配置文件设置
func CORSFilter(corsConfig *conf.Server_CORS) khttp.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 如果CORS被禁用，直接通过
			if corsConfig != nil && !corsConfig.Enable {
				next.ServeHTTP(w, r)
				return
			}

			// 设置默认值
			if corsConfig == nil {
				corsConfig = getDefaultCORSConfig()
			}

			// 获取请求的 Origin
			origin := r.Header.Get("Origin")

			// 设置允许的源
			if len(corsConfig.AllowedOrigins) > 0 {
				if contains(corsConfig.AllowedOrigins, "*") {
					if origin != "" {
						w.Header().Set("Access-Control-Allow-Origin", origin)
					} else {
						w.Header().Set("Access-Control-Allow-Origin", "*")
					}
				} else if contains(corsConfig.AllowedOrigins, origin) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			// 设置是否允许凭证
			if corsConfig.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// 设置允许的请求方法
			if len(corsConfig.AllowedMethods) > 0 {
				methods := strings.Join(corsConfig.AllowedMethods, ", ")
				w.Header().Set("Access-Control-Allow-Methods", methods)
			} else {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD")
			}

			// 设置允许的请求头
			if len(corsConfig.AllowedHeaders) > 0 {
				if contains(corsConfig.AllowedHeaders, "*") {
					requestHeaders := r.Header.Get("Access-Control-Request-Headers")
					if requestHeaders != "" {
						w.Header().Set("Access-Control-Allow-Headers", requestHeaders)
					} else {
						w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
					}
				} else {
					headers := strings.Join(corsConfig.AllowedHeaders, ", ")
					w.Header().Set("Access-Control-Allow-Headers", headers)
				}
			} else {
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
			}

			// 设置预检请求的有效期
			if corsConfig.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", string(rune(corsConfig.MaxAge)))
			} else {
				w.Header().Set("Access-Control-Max-Age", "86400")
			}

			// 设置允许客户端访问的响应头
			if len(corsConfig.ExposedHeaders) > 0 {
				headers := strings.Join(corsConfig.ExposedHeaders, ", ")
				w.Header().Set("Access-Control-Expose-Headers", headers)
			} else {
				w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Authorization")
			}

			// 添加 Vary 头
			w.Header().Add("Vary", "Origin")
			w.Header().Add("Vary", "Access-Control-Request-Method")
			w.Header().Add("Vary", "Access-Control-Request-Headers")

			// 处理 OPTIONS 预检请求
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// 继续处理其他请求
			next.ServeHTTP(w, r)
		})
	}
}

// getDefaultCORSConfig 返回默认的CORS配置
func getDefaultCORSConfig() *conf.Server_CORS {
	return &conf.Server_CORS{
		Enable:           true,
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           86400,
	}
}

// contains 检查字符串切片是否包含指定字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
