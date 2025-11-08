package server

import (
	nethttp "net/http"

	adsv1 "github.com/OmnTeam/ppanel-pro/api/admin/ads/v1"
	announcementv1 "github.com/OmnTeam/ppanel-pro/api/admin/announcement/v1"
	applicationv1 "github.com/OmnTeam/ppanel-pro/api/admin/application/v1"
	authmethodv1 "github.com/OmnTeam/ppanel-pro/api/admin/authmethod/v1"
	adminv1 "github.com/OmnTeam/ppanel-pro/api/admin/console/v1"
	admincouponv1 "github.com/OmnTeam/ppanel-pro/api/admin/coupon/v1"
	admindocumentv1 "github.com/OmnTeam/ppanel-pro/api/admin/document/v1"
	adminlogv1 "github.com/OmnTeam/ppanel-pro/api/admin/log/v1"
	adminmarketingv1 "github.com/OmnTeam/ppanel-pro/api/admin/marketing/v1"
	adminorderv1 "github.com/OmnTeam/ppanel-pro/api/admin/order/v1"
	adminpaymentv1 "github.com/OmnTeam/ppanel-pro/api/admin/payment/v1"
	adminserverv1 "github.com/OmnTeam/ppanel-pro/api/admin/server/v1"
	adminsubscribev1 "github.com/OmnTeam/ppanel-pro/api/admin/subscribe/v1"
	adminsystemv1 "github.com/OmnTeam/ppanel-pro/api/admin/system/v1"
	adminticketv1 "github.com/OmnTeam/ppanel-pro/api/admin/ticket/v1"
	adminuserv1 "github.com/OmnTeam/ppanel-pro/api/admin/user/v1"
	authoauthv1 "github.com/OmnTeam/ppanel-pro/api/auth/oauth/v1"
	publicannouncementv1 "github.com/OmnTeam/ppanel-pro/api/public/announcement/v1"
	publicauthv1 "github.com/OmnTeam/ppanel-pro/api/public/auth/v1"
	publiccommonv1 "github.com/OmnTeam/ppanel-pro/api/public/common/v1"
	publicdocumentv1 "github.com/OmnTeam/ppanel-pro/api/public/document/v1"
	publicorderv1 "github.com/OmnTeam/ppanel-pro/api/public/order/v1"
	publicpaymentv1 "github.com/OmnTeam/ppanel-pro/api/public/payment/v1"
	publicportalv1 "github.com/OmnTeam/ppanel-pro/api/public/portal/v1"
	publicsubscribev1 "github.com/OmnTeam/ppanel-pro/api/public/subscribe/v1"
	publicticketv1 "github.com/OmnTeam/ppanel-pro/api/public/ticket/v1"
	publicuserv1 "github.com/OmnTeam/ppanel-pro/api/public/user/v1"
	"github.com/OmnTeam/ppanel-pro/internal/conf"
	"github.com/OmnTeam/ppanel-pro/internal/data"
	"github.com/OmnTeam/ppanel-pro/internal/pkg/middleware"
	adsservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/ads"
	announcementservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/announcement"
	applicationservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/application"
	authmethodservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/authmethod"
	adminconsoleservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/console"
	admincouponservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/coupon"
	admindocumentservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/document"
	adminlogservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/log"
	adminmarketingservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/marketing"
	adminorderservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/order"
	adminpaymentservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/payment"
	adminserverservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/server"
	adminsubscribeservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/subscribe"
	adminsystemservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/system"
	adminticketservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/ticket"
	adminuserservice "github.com/OmnTeam/ppanel-pro/internal/service/admin/user"
	authservice "github.com/OmnTeam/ppanel-pro/internal/service/auth"
	authoauthservice "github.com/OmnTeam/ppanel-pro/internal/service/auth/oauth"
	commonservice "github.com/OmnTeam/ppanel-pro/internal/service/common"
	publicorderservice "github.com/OmnTeam/ppanel-pro/internal/service/public"
	publicannouncementservice "github.com/OmnTeam/ppanel-pro/internal/service/public/announcement"
	publicdocumentservice "github.com/OmnTeam/ppanel-pro/internal/service/public/document"
	publicpaymentservice "github.com/OmnTeam/ppanel-pro/internal/service/public/payment"
	publicportalservice "github.com/OmnTeam/ppanel-pro/internal/service/public/portal"
	publicsubscribeservice "github.com/OmnTeam/ppanel-pro/internal/service/public/subscribe"
	publicticketservice "github.com/OmnTeam/ppanel-pro/internal/service/public/ticket"
	publicuserservice "github.com/OmnTeam/ppanel-pro/internal/service/public/user"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server
func NewHTTPServer(c *conf.Server, ads *adsservice.AdsService, announcement *announcementservice.AnnouncementService, application *applicationservice.SubscribeApplicationService, authmethod *authmethodservice.AuthMethodService, adminConsole *adminconsoleservice.ConsoleService, adminCoupon *admincouponservice.CouponService, adminDocument *admindocumentservice.DocumentService, adminLog *adminlogservice.LogService, adminMarketing *adminmarketingservice.MarketingService, adminOrder *adminorderservice.OrderService, adminPayment *adminpaymentservice.PaymentService, adminServer *adminserverservice.ServerService, adminSubscribe *adminsubscribeservice.SubscribeService, adminSystem *adminsystemservice.SystemService, adminTicket *adminticketservice.TicketService, adminUser *adminuserservice.UserService, adminUserAuthMethod *adminuserservice.UserAuthMethodService, adminUserDevice *adminuserservice.UserDeviceService, adminUserSubscribe *adminuserservice.UserSubscribeService, auth *authservice.AuthService, oauthSvc *authoauthservice.OAuthService, commonSvc *commonservice.CommonService, publicOrder *publicorderservice.PublicOrderService, publicAnnouncement *publicannouncementservice.AnnouncementService, publicDocument *publicdocumentservice.DocumentService, publicPayment *publicpaymentservice.PaymentService, publicPortal *publicportalservice.PortalService, publicSubscribe *publicsubscribeservice.SubscribeService, publicTicket *publicticketservice.TicketService, publicUser *publicuserservice.UserService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Filter(middleware.CORSFilter()), // CORS Filter 必须在最前面
		http.Middleware(
			middleware.JWTAuth(), // JWT authentication middleware
		),
		http.ErrorEncoder(CustomErrorEncoder),       // 使用自定义错误编码器，所有错误返回HTTP 200
		http.RequestDecoder(CustomRequestDecoder),   // 使用自定义请求解码器，处理前端空对象问题
		http.ResponseEncoder(CustomResponseEncoder), // 使用自定义响应编码器，解决 int64 序列化问题
		http.StrictSlash(false),                     // 禁用尾部斜杠自动重定向，通过手动注册两个路由来支持
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
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
	// Public Subscribe模块服务注册
	publicsubscribev1.RegisterSubscribeHTTPServer(srv, publicSubscribe)
	// Public Ticket模块服务注册
	publicticketv1.RegisterTicketHTTPServer(srv, publicTicket)
	// Public User模块服务注册
	publicuserv1.RegisterUserHTTPServer(srv, publicUser)

	// 注册WebSocket端点
	srv.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// 从查询参数获取设备ID和令牌
		deviceID := r.URL.Query().Get("device_id")
		token := r.URL.Query().Get("token")

		if deviceID == "" || token == "" {
			nethttp.Error(w, "Missing device_id or token", nethttp.StatusBadRequest)
			return
		}

		// 获取全局设备管理器
		deviceManager := data.GetGlobalDeviceManager()
		if deviceManager == nil {
			nethttp.Error(w, "Device manager not available", nethttp.StatusInternalServerError)
			return
		}

		// 验证令牌并获取用户信息（这里简化处理，实际应该验证token）
		// 暂时使用硬编码方式，应该从token验证获取
		maxDevices := 5    // 默认最大设备数
		userID := int64(1) // 暂时硬编码，应该从token验证获取

		// 使用设备管理器处理WebSocket连接
		err := deviceManager.AddDevice(w, r, token, userID, deviceID, maxDevices)
		if err != nil {
			nethttp.Error(w, "WebSocket upgrade failed", nethttp.StatusInternalServerError)
			return
		}
	})

	return srv
}
