package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
)

// PaymentParams 支付参数
type PaymentParams struct {
	Platform string `uri:"platform"`
	Token    string `uri:"token"`
}

// NotifyMiddleware 支付通知中间件
// 用于处理支付平台的回调通知
func NotifyMiddleware(getPaymentConfigFunc func(ctx context.Context, token string) (*PaymentConfig, error), logger log.Logger) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// 从URI中获取platform和token参数
			// URI格式: /api/public/payment/{platform}/{token}/notify
			pathParts := strings.Split(r.URL.Path, "/")
			if len(pathParts) < 6 {
				http.Error(w, "Invalid URI format", http.StatusBadRequest)
				return
			}

			// pathParts格式: ["", "api", "public", "payment", platform, token, "notify"]
			_ = pathParts[4] // platform (暂未使用)
			token := pathParts[5]

			// 根据token获取支付配置
			config, err := getPaymentConfigFunc(ctx, token)
			if err != nil {
				log.Errorf("[NotifyMiddleware] Failed to get payment config: %v", err)
				http.Error(w, "Invalid payment token", http.StatusBadRequest)
				return
			}

			// 将平台和配置信息存入context
			ctx = context.WithValue(ctx, "platform", config.Platform)
			ctx = context.WithValue(ctx, "payment", config)
			ctx = context.WithValue(ctx, "paymentToken", token)

			// 使用更新后的context继续处理请求
			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// PaymentConfig 支付配置
type PaymentConfig struct {
	Platform     string
	PaymentToken string
	// 其他支付相关配置...
}
