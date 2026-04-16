package server

import (
	"errors"
	"fmt"
	nethttp "net/http"
	"os"
	"strconv"
	"strings"

	adsv1 "github.com/OmnTeam/ppanel-pro/api/admin/ads/v1"
	announcementv1 "github.com/OmnTeam/ppanel-pro/api/admin/announcement/v1"
	applicationv1 "github.com/OmnTeam/ppanel-pro/api/admin/application/v1"
	authmethodv1 "github.com/OmnTeam/ppanel-pro/api/admin/authmethod/v1"
	adminv1 "github.com/OmnTeam/ppanel-pro/api/admin/console/v1"
	admincouponv1 "github.com/OmnTeam/ppanel-pro/api/admin/coupon/v1"
	admindocumentv1 "github.com/OmnTeam/ppanel-pro/api/admin/document/v1"
	maingroupv1 "github.com/OmnTeam/ppanel-pro/api/admin/group/v1"
	adminlogv1 "github.com/OmnTeam/ppanel-pro/api/admin/log/v1"
	adminmarketingv1 "github.com/OmnTeam/ppanel-pro/api/admin/marketing/v1"
	adminorderv1 "github.com/OmnTeam/ppanel-pro/api/admin/order/v1"
	adminpaymentv1 "github.com/OmnTeam/ppanel-pro/api/admin/payment/v1"
	adminredemptionv1 "github.com/OmnTeam/ppanel-pro/api/admin/redemption/v1"
	adminserverv1 "github.com/OmnTeam/ppanel-pro/api/admin/server/v1"
	adminsubscribev1 "github.com/OmnTeam/ppanel-pro/api/admin/subscribe/v1"
	adminsystemv1 "github.com/OmnTeam/ppanel-pro/api/admin/system/v1"
	adminticketv1 "github.com/OmnTeam/ppanel-pro/api/admin/ticket/v1"
	admintoolv1 "github.com/OmnTeam/ppanel-pro/api/admin/tool/v1"
	adminuserv1 "github.com/OmnTeam/ppanel-pro/api/admin/user/v1"
	authoauthv1 "github.com/OmnTeam/ppanel-pro/api/auth/oauth/v1"
	publicannouncementv1 "github.com/OmnTeam/ppanel-pro/api/public/announcement/v1"
	publicauthv1 "github.com/OmnTeam/ppanel-pro/api/public/auth/v1"
	publiccommonv1 "github.com/OmnTeam/ppanel-pro/api/public/common/v1"
	publicdocumentv1 "github.com/OmnTeam/ppanel-pro/api/public/document/v1"
	publicorderv1 "github.com/OmnTeam/ppanel-pro/api/public/order/v1"
	publicpaymentv1 "github.com/OmnTeam/ppanel-pro/api/public/payment/v1"
	publicportalv1 "github.com/OmnTeam/ppanel-pro/api/public/portal/v1"
	publicredemptionv1 "github.com/OmnTeam/ppanel-pro/api/public/redemption/v1"
	publicsubscribev1 "github.com/OmnTeam/ppanel-pro/api/public/subscribe/v1"
	subscriptionv1 "github.com/OmnTeam/ppanel-pro/api/public/subscription/v1"
	publicticketv1 "github.com/OmnTeam/ppanel-pro/api/public/ticket/v1"
	publicuserv1 "github.com/OmnTeam/ppanel-pro/api/public/user/v1"
	subscriptionbiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/subscription"
	"github.com/OmnTeam/ppanel-pro/internal/conf"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	appmiddleware "github.com/OmnTeam/ppanel-pro/internal/middleware"
	"github.com/OmnTeam/ppanel-pro/internal/pkg/middleware"
	adsservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/ads"
	announcementservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/announcement"
	applicationservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/application"
	authmethodservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/authmethod"
	adminconsoleservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/console"
	admincouponservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/coupon"
	admindocumentservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/document"
	maingroupservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/group"
	adminlogservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/log"
	adminmarketingservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/marketing"
	adminorderservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/order"
	adminpaymentservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/payment"
	adminredemptionservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/redemption"
	adminserverservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/server"
	adminsubscribeservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/subscribe"
	adminsystemservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/system"
	adminticketservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/ticket"
	admintoolservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/tool"
	adminuserservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/user"
	authservice "github.com/OmnTeam/ppanel-pro/internal/service/auth"
	authoauthservice "github.com/OmnTeam/ppanel-pro/internal/service/auth/oauth"
	commonservice "github.com/OmnTeam/ppanel-pro/internal/service/common"
	publicorderservice "github.com/OmnTeam/ppanel-pro/internal/service/public"
	publicannouncementservice "github.com/OmnTeam/ppanel-pro/internal/service/public/announcement"
	publicdocumentservice "github.com/OmnTeam/ppanel-pro/internal/service/public/document"
	publicpaymentservice "github.com/OmnTeam/ppanel-pro/internal/service/public/payment"
	publicportalservice "github.com/OmnTeam/ppanel-pro/internal/service/public/portal"
	publicredemptionservice "github.com/OmnTeam/ppanel-pro/internal/service/public/redemption"
	publicsubscribeservice "github.com/OmnTeam/ppanel-pro/internal/service/public/subscribe"
	publicsubscription "github.com/OmnTeam/ppanel-pro/internal/service/public/subscription"
	publicticketservice "github.com/OmnTeam/ppanel-pro/internal/service/public/ticket"
	publicuserservice "github.com/OmnTeam/ppanel-pro/internal/service/public/user"
	serverservice "github.com/OmnTeam/ppanel-pro/internal/service/server"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server
func NewHTTPServer(c *conf.Server, appConf *conf.Application, authMiddleware *appmiddleware.ServiceContext, authCompat *data.AuthCompat, d *data.Data, ads *adsservice.AdsService, announcement *announcementservice.AnnouncementService, application *applicationservice.SubscribeApplicationService, authmethod *authmethodservice.AuthMethodService, adminConsole *adminconsoleservice.ConsoleService, adminCoupon *admincouponservice.CouponService, adminDocument *admindocumentservice.DocumentService, adminLog *adminlogservice.LogService, adminMarketing *adminmarketingservice.MarketingService, adminOrder *adminorderservice.OrderService, adminPayment *adminpaymentservice.PaymentService, adminServer *adminserverservice.ServerService, adminSubscribe *adminsubscribeservice.SubscribeService, adminSystem *adminsystemservice.SystemService, adminTicket *adminticketservice.TicketService, adminRedemption *adminredemptionservice.RedemptionService, adminTool *admintoolservice.ToolService, adminGroup *maingroupservice.GroupService, adminUser *adminuserservice.UserService, adminUserAuthMethod *adminuserservice.UserAuthMethodService, adminUserDevice *adminuserservice.UserDeviceService, adminUserSubscribe *adminuserservice.UserSubscribeService, auth *authservice.AuthService, oauthSvc *authoauthservice.OAuthService, commonSvc *commonservice.CommonService, publicOrder *publicorderservice.PublicOrderService, publicAnnouncement *publicannouncementservice.AnnouncementService, publicDocument *publicdocumentservice.DocumentService, publicPayment *publicpaymentservice.PaymentService, publicPortal *publicportalservice.PortalService, publicRedemption *publicredemptionservice.RedemptionService, publicSubscribe *publicsubscribeservice.SubscribeService, publicSubscription *publicsubscription.PublicSubscriptionService, publicTicket *publicticketservice.TicketService, publicUser *publicuserservice.UserService, server *serverservice.ServerService, logger log.Logger) *http.Server {
	if c == nil {
		c = &conf.Server{}
	}
	httpConf := c.GetHttp()

	var opts = []http.ServerOption{
		http.Filter(
			middleware.TraceMiddleware(logger),
			middleware.CORSFilter(c.Cors),                                 // CORS Filter 必须最外层
			middleware.LegacyPathCompatFilter(),                           // 兼容旧项目尾斜杠路由
			middleware.LegacyRouteGuardFilter(),                           // 屏蔽新项目多出的非兼容路由
			middleware.DeviceMiddleware(getDeviceConfig(appConf), logger), // 设备加解密需保留在 HTTP Filter 链中
		),
		http.Middleware(
			recovery.Recovery(),
			middleware.Logging(logger), // Logging middleware，记录请求日志
			authMiddleware.Auth(),      // 对齐旧项目认证语义
		),
		http.ErrorEncoder(CustomErrorEncoder),       // 使用自定义错误编码器，所有错误返回HTTP 200
		http.RequestDecoder(CustomRequestDecoder),   // 使用自定义请求解码器，处理前端空对象问题
		http.ResponseEncoder(CustomResponseEncoder), // 使用自定义响应编码器，解决 int64 序列化问题
		http.StrictSlash(false),                     // 禁用尾部斜杠自动重定向，通过手动注册两个路由来支持
		http.NotFoundHandler(newCORSAwareFallbackHandler(c.Cors, nethttp.StatusNotFound)),
		http.MethodNotAllowedHandler(newCORSAwareFallbackHandler(c.Cors, nethttp.StatusMethodNotAllowed)),
	}
	if httpConf != nil && httpConf.Network != "" {
		opts = append(opts, http.Network(httpConf.Network))
	}
	if httpConf != nil && httpConf.Addr != "" {
		opts = append(opts, http.Address(httpConf.Addr))
	}
	if httpConf != nil && httpConf.Timeout != nil {
		opts = append(opts, http.Timeout(httpConf.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	registerLegacyCompatRoutes(srv, authCompat, auth, oauthSvc, d, appConf, adminGroup, adminPayment, adminSystem, adminTicket, publicOrder, publicPayment, publicPortal, publicTicket, publicUser, logger)
	adsv1.RegisterAdsServiceHTTPServer(srv, ads)
	announcementv1.RegisterAnnouncementServiceHTTPServer(srv, announcement)
	applicationv1.RegisterSubscribeApplicationServiceHTTPServer(srv, application)
	authmethodv1.RegisterAuthMethodServiceHTTPServer(srv, authmethod)
	adminv1.RegisterAdminConsoleHTTPServer(srv, adminConsole)
	admincouponv1.RegisterCouponServiceHTTPServer(srv, adminCoupon)
	admindocumentv1.RegisterDocumentServiceHTTPServer(srv, adminDocument)
	adminlogv1.RegisterLogServiceHTTPServer(srv, adminLog)
	adminmarketingv1.RegisterMarketingServiceHTTPServer(srv, adminMarketing)
	adminorderv1.RegisterOrderServiceHTTPServer(srv, adminOrder)
	adminpaymentv1.RegisterPaymentServiceHTTPServer(srv, adminPayment)
	adminserverv1.RegisterServerServiceHTTPServer(srv, adminServer)
	adminsubscribev1.RegisterSubscribeHTTPServer(srv, adminSubscribe)
	adminsystemv1.RegisterSystemServiceHTTPServer(srv, adminSystem)
	adminticketv1.RegisterTicketHTTPServer(srv, adminTicket)
	adminredemptionv1.RegisterRedemptionHTTPServer(srv, adminRedemption)
	admintoolv1.RegisterToolHTTPServer(srv, adminTool)
	maingroupv1.RegisterGroupHTTPServer(srv, adminGroup)
	// Admin User模块服务注册
	adminuserv1.RegisterUserServiceHTTPServer(srv, adminUser)
	adminuserv1.RegisterUserAuthMethodServiceHTTPServer(srv, adminUserAuthMethod)
	adminuserv1.RegisterUserDeviceServiceHTTPServer(srv, adminUserDevice)
	adminuserv1.RegisterUserSubscribeServiceHTTPServer(srv, adminUserSubscribe)
	// Auth模块服务注册
	publicauthv1.RegisterAuthHTTPServer(srv, auth)
	// Auth OAuth模块服务注册
	authoauthv1.RegisterOAuthHTTPServer(srv, oauthSvc)
	// Common模块服务注册
	publiccommonv1.RegisterCommonHTTPServer(srv, commonSvc)
	// Public Order模块服务注册
	publicorderv1.RegisterPublicOrderHTTPServer(srv, publicOrder)
	// Public Announcement模块服务注册
	publicannouncementv1.RegisterAnnouncementHTTPServer(srv, publicAnnouncement)
	// Public Document模块服务注册
	publicdocumentv1.RegisterDocumentHTTPServer(srv, publicDocument)
	// Public Payment模块服务注册
	publicpaymentv1.RegisterPaymentHTTPServer(srv, publicPayment)
	// Public Portal模块服务注册
	publicportalv1.RegisterPortalHTTPServer(srv, publicPortal)
	// Public Redemption模块服务注册
	publicredemptionv1.RegisterRedemptionServiceHTTPServer(srv, publicRedemption)
	// Public Subscribe模块服务注册
	publicsubscribev1.RegisterSubscribeHTTPServer(srv, publicSubscribe)
	// Public Subscription模块服务注册（订阅配置生成）
	subscriptionv1.RegisterSubscriptionHTTPServer(srv, publicSubscription)
	// Public Ticket模块服务注册
	publicticketv1.RegisterTicketHTTPServer(srv, publicTicket)
	// Public User模块服务注册
	publicuserv1.RegisterUserHTTPServer(srv, publicUser)
	// 注册兼容的订阅配置端点（与原项目格式完全一致）
	for _, subscribePath := range subscribeCompatPaths(appConf) {
		srv.HandleFunc(subscribePath, handleSubscribeConfig(publicSubscription))
	}

	return srv
}

func getDeviceConfig(appConf *conf.Application) *conf.Device {
	if appConf == nil {
		return nil
	}
	return appConf.Device
}

// handleSubscribeConfig 处理订阅配置请求（与原项目完全兼容）
func handleSubscribeConfig(subscriptionSvc *publicsubscription.PublicSubscriptionService) nethttp.HandlerFunc {
	return func(w nethttp.ResponseWriter, r *nethttp.Request) {
		token := r.Header.Get("token")
		if token == "" {
			token = r.URL.Query().Get("token")
		}

		userAgent := r.UserAgent()
		clientIP := getClientIP(r)

		requestURI := r.RequestURI
		requestHost := r.Host
		gatewayMode := legacyGatewayModeEnabled()

		req := &subscriptionv1.GetSubscribeConfigRequest{
			Token: token,
		}

		if target := r.URL.Query().Get("target"); target != "" {
			req.Target = target
		}
		if r.URL.Query().Get("list") == "true" {
			req.List = true
		}
		if r.URL.Query().Get("emoji") == "true" {
			req.Emoji = true
		}
		if r.URL.Query().Get("udp") == "true" {
			req.Udp = true
		}
		if r.URL.Query().Get("tfo") == "true" {
			req.Tfo = true
		}
		if r.URL.Query().Get("scv") == "true" {
			req.Scv = true
		}
		if filter := r.URL.Query().Get("filter"); filter != "" {
			req.Filter = filter
		}
		if sortValue := strings.TrimSpace(r.URL.Query().Get("sort")); sortValue != "" {
			if parsed, err := strconv.ParseInt(sortValue, 10, 32); err == nil {
				req.Sort = int32(parsed)
			}
		}
		if r.URL.Query().Get("expand") == "true" {
			req.Expand = true
		}

		ctx := r.Context()
		queryParams := getQueryMap(r)
		ctx = middleware.WithUserAgent(ctx, userAgent)
		ctx = middleware.WithClientIP(ctx, clientIP)
		ctx = middleware.WithRequestURI(ctx, requestURI)
		ctx = middleware.WithRequestHost(ctx, requestHost)
		ctx = middleware.WithGatewayMode(ctx, gatewayMode)
		ctx = middleware.WithQueryParams(ctx, queryParams)

		if err := subscriptionSvc.ValidateLegacyRequest(ctx, token, requestHost, userAgent); err != nil {
			if errors.Is(err, subscriptionbiz.ErrLegacyAccessDenied) {
				nethttp.Error(w, "Access denied", nethttp.StatusForbidden)
				return
			}
			nethttp.Error(w, "Internal Server Error", nethttp.StatusInternalServerError)
			return
		}

		resp, err := subscriptionSvc.GetSubscribeConfig(ctx, req)
		if err != nil || resp == nil || resp.Code != 0 || resp.Data == nil {
			nethttp.Error(w, "Internal Server Error", nethttp.StatusInternalServerError)
			return
		}

		if resp.Data.Header != "" {
			w.Header().Set("subscription-userinfo", resp.Data.Header)
		}
		if resp.Data.ContentType != "" {
			w.Header().Set("Content-Type", resp.Data.ContentType)
		}
		if resp.Data.Filename != "" {
			w.Header().Set("content-disposition", fmt.Sprintf("attachment;filename*=UTF-8''%s", resp.Data.Filename))
		}

		w.WriteHeader(nethttp.StatusOK)
		w.Write([]byte(resp.Data.Config))
	}
}

// getClientIP 获取客户端真实IP（支持代理）
func getClientIP(r *nethttp.Request) string {
	// 尝试从X-Forwarded-For获取
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从X-Real-IP获取
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用RemoteAddr
	return r.RemoteAddr
}

func getQueryMap(r *nethttp.Request) map[string]string {
	result := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			result[key] = values[0]
		}
	}
	return result
}

func subscribeCompatPaths(appConf *conf.Application) []string {
	paths := []string{
		"/v1/subscribe/config",
	}

	if appConf != nil && appConf.Subscribe != nil {
		if customPath := strings.TrimSpace(appConf.Subscribe.SubscribePath); customPath != "" {
			if !strings.HasPrefix(customPath, "/") {
				customPath = "/" + customPath
			}
			paths = append(paths, customPath)
		}
	}

	seen := make(map[string]struct{}, len(paths))
	result := make([]string, 0, len(paths))
	for _, item := range paths {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func legacyGatewayModeEnabled() bool {
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
