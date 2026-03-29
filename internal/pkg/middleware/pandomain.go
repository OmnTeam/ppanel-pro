package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-kratos/kratos/v2/log"

	subscriptionbiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/subscription"
	"github.com/OmnTeam/ppanel-pro/internal/conf"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
)

// PanDomainMiddleware 泛域名订阅中间件
// 用于处理泛域名的订阅请求，例如: token.example.com
func PanDomainMiddleware(subscribeConfig *conf.Subscribe, uc *subscriptionbiz.SubscriptionUseCase, logger log.Logger) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 如果泛域名功能未启用，直接跳过
			if subscribeConfig == nil || !subscribeConfig.PanDomain {
				handler.ServeHTTP(w, r)
				return
			}

			// 只处理根路径请求
			if r.URL.Path != "/" {
				handler.ServeHTTP(w, r)
				return
			}

			// 拦截浏览器请求
			ua := r.UserAgent()

			// User-Agent限制检查
			if subscribeConfig.UserAgentLimit {
				if ua == "" {
					http.Error(w, "Access denied", http.StatusForbidden)
					return
				}

				browserKeywords := tool.RemoveDuplicateElements(strings.Split(subscribeConfig.UserAgentList, "\n")...)
				var allow = false

				// TODO: 查询客户端列表
				// clients, err := svc.ClientModel.List(r.Context())
				// if err != nil {
				// 	logger.Errorf("[PanDomainMiddleware] Query client list failed: %v", err)
				// }
				// for _, item := range clients {
				// 	u := strings.ToLower(item.UserAgent)
				// 	u = strings.Trim(u, " ")
				// 	browserKeywords = append(browserKeywords, u)
				// }

				for _, keyword := range browserKeywords {
					keyword = strings.ToLower(strings.Trim(keyword, " "))
					if keyword == "" {
						continue
					}
					if strings.Contains(strings.ToLower(ua), keyword) {
						allow = true
					}
				}

				if !allow {
					http.Error(w, "Access denied", http.StatusForbidden)
					return
				}
			}

			// 解析域名获取token
			domain := r.Host
			domainArr := strings.Split(domain, ".")
			if len(domainArr) < 2 {
				http.Error(w, "Invalid domain", http.StatusBadRequest)
				return
			}

			domainFirst := domainArr[0]
			// domainFlag := ""
			// if len(domainArr) > 2 {
			// 	domainFlag = domainArr[1]
			// }

			// 构建订阅请求
			ctx := context.Background()
			// 设置请求上下文信息
			ctx = context.WithValue(ctx, "userAgent", ua)
			ctx = context.WithValue(ctx, "clientIP", GetClientIP(r.Context()))
			ctx = context.WithValue(ctx, "requestURI", r.RequestURI)
			ctx = context.WithValue(ctx, "requestHost", r.Host)
			ctx = context.WithValue(ctx, "gatewayMode", false)

			// 调用订阅服务获取配置
			configBytes, header, err := GetSubscribeConfigByToken(ctx, uc, domainFirst, ua, r)
			if err != nil {
				log.Errorf("[PanDomainMiddleware] Failed to get subscribe config: %v", err)
				http.Error(w, "Failed to get subscribe config", http.StatusInternalServerError)
				return
			}

			// 设置响应头
			w.Header().Set("subscription-userinfo", header)
			w.Header().Set("Content-Type", "application/octet-stream; charset=UTF-8")

			// 返回订阅配置
			w.WriteHeader(http.StatusOK)
			w.Write(configBytes)
		})
	}
}

// GetSubscribeConfigByToken 根据token获取订阅配置
// 这个函数需要调用subscription service来获取配置
func GetSubscribeConfigByToken(ctx context.Context, uc *subscriptionbiz.SubscriptionUseCase, token, userAgent string, r *http.Request) ([]byte, string, error) {
	// TODO: 调用订阅服务
	// 这里需要调用 subscription service 的 GetSubscribeConfig 方法
	// 由于原项目中这个逻辑直接调用subscribeLogic.Handler
	// 在新架构中，我们需要调用对应的service

	// 临时返回空实现
	return nil, "", nil
}
