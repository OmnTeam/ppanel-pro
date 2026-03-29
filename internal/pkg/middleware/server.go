package middleware

import (
	"net/http"

	"github.com/go-kratos/kratos/v2/log"
)

// ServerMiddleware 服务器节点认证中间件
// 用于验证节点服务器的请求，通过secret_key验证
func ServerMiddleware(nodeSecret string, logger log.Logger) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 从查询参数中获取secret_key
			key := r.URL.Query().Get("secret_key")

			// 验证secret_key是否匹配
			if key == nodeSecret {
				handler.ServeHTTP(w, r)
				return
			}

			// 验证失败，返回403
			log.Warnf("[ServerMiddleware] Invalid secret_key from %s", r.RemoteAddr)
			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
}
