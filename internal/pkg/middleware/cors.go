package middleware

import (
	"net/http"

	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

// CORSFilter 返回一个 CORS HTTP Filter，放行所有跨域请求
func CORSFilter() khttp.FilterFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 获取请求的 Origin
			origin := r.Header.Get("Origin")

			// 动态设置允许的源
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				// 允许携带凭证
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			// 设置允许的请求方法（放行所有常用方法）
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD")

			// 动态设置允许的请求头
			requestHeaders := r.Header.Get("Access-Control-Request-Headers")
			if requestHeaders != "" {
				w.Header().Set("Access-Control-Allow-Headers", requestHeaders)
			} else {
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, X-Token")
			}

			// 设置预检请求的有效期（24小时）
			w.Header().Set("Access-Control-Max-Age", "86400")

			// 设置允许客户端访问的响应头
			w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Authorization, X-Token")

			// 添加 Vary 头，告诉浏览器根据 Origin 缓存
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
