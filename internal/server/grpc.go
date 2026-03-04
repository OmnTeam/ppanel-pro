package server

import (
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
	publicauthv1 "github.com/OmnTeam/ppanel-pro/api/public/auth/v1"
	publiccommonv1 "github.com/OmnTeam/ppanel-pro/api/public/common/v1"
	publicorderv1 "github.com/OmnTeam/ppanel-pro/api/public/order/v1"
	publicportalv1 "github.com/OmnTeam/ppanel-pro/api/public/portal/v1"
	publicticketv1 "github.com/OmnTeam/ppanel-pro/api/public/ticket/v1"
	publicuserv1 "github.com/OmnTeam/ppanel-pro/api/public/user/v1"
	"github.com/OmnTeam/ppanel-pro/internal/conf"
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
	publicportalservice "github.com/OmnTeam/ppanel-pro/internal/service/public/portal"
	publicticketservice "github.com/OmnTeam/ppanel-pro/internal/service/public/ticket"
	publicuserservice "github.com/OmnTeam/ppanel-pro/internal/service/public/user"

	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server
func NewGRPCServer(c *conf.Server, ads *adsservice.AdsService, announcement *announcementservice.AnnouncementService, application *applicationservice.SubscribeApplicationService, authmethod *authmethodservice.AuthMethodService, adminConsole *adminconsoleservice.ConsoleService, adminCoupon *admincouponservice.CouponService, adminDocument *admindocumentservice.DocumentService, adminLog *adminlogservice.LogService, adminMarketing *adminmarketingservice.MarketingService, adminOrder *adminorderservice.OrderService, adminPayment *adminpaymentservice.PaymentService, adminServer *adminserverservice.ServerService, adminSubscribe *adminsubscribeservice.SubscribeService, adminSystem *adminsystemservice.SystemService, adminTicket *adminticketservice.TicketService, adminRedemption *adminredemptionservice.RedemptionService, adminTool *admintoolservice.ToolService, adminGroup *maingroupservice.GroupService, adminUser *adminuserservice.UserService, adminUserAuthMethod *adminuserservice.UserAuthMethodService, adminUserDevice *adminuserservice.UserDeviceService, adminUserSubscribe *adminuserservice.UserSubscribeService, auth *authservice.AuthService, oauthSvc *authoauthservice.OAuthService, commonSvc *commonservice.CommonService, publicOrder *publicorderservice.PublicOrderService, publicPortal *publicportalservice.PortalService, publicTicket *publicticketservice.TicketService, publicUser *publicuserservice.UserService) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			middleware.JWTAuth(c.Auth), // JWT authentication middleware，使用配置
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	adsv1.RegisterAdsServiceServer(srv, ads)
	announcementv1.RegisterAnnouncementServiceServer(srv, announcement)
	applicationv1.RegisterSubscribeApplicationServiceServer(srv, application)
	authmethodv1.RegisterAuthMethodServiceServer(srv, authmethod)
	adminv1.RegisterAdminConsoleServer(srv, adminConsole)
	admincouponv1.RegisterCouponServiceServer(srv, adminCoupon)
	admindocumentv1.RegisterDocumentServiceServer(srv, adminDocument)
	adminlogv1.RegisterLogServiceServer(srv, adminLog)
	adminmarketingv1.RegisterMarketingServiceServer(srv, adminMarketing)
	adminorderv1.RegisterOrderServiceServer(srv, adminOrder)
	adminpaymentv1.RegisterPaymentServiceServer(srv, adminPayment)
	adminserverv1.RegisterServerServiceServer(srv, adminServer)
	adminsubscribev1.RegisterSubscribeServer(srv, adminSubscribe)
	adminsystemv1.RegisterSystemServiceServer(srv, adminSystem)
	adminticketv1.RegisterTicketServer(srv, adminTicket)
	adminredemptionv1.RegisterRedemptionServer(srv, adminRedemption)
	admintoolv1.RegisterToolServer(srv, adminTool)
	maingroupv1.RegisterGroupServer(srv, adminGroup)
	// Admin User模块服务注册
	adminuserv1.RegisterUserServiceServer(srv, adminUser)
	adminuserv1.RegisterUserAuthMethodServiceServer(srv, adminUserAuthMethod)
	adminuserv1.RegisterUserDeviceServiceServer(srv, adminUserDevice)
	adminuserv1.RegisterUserSubscribeServiceServer(srv, adminUserSubscribe)
	// Auth模块服务注册
	publicauthv1.RegisterAuthServer(srv, auth)
	// Auth OAuth模块服务注册
	authoauthv1.RegisterOAuthServer(srv, oauthSvc)
	// Common模块服务注册
	publiccommonv1.RegisterCommonServer(srv, commonSvc)
	// Public Order模块服务注册
	publicorderv1.RegisterPublicOrderServer(srv, publicOrder)
	// Public Portal模块服务注册
	publicportalv1.RegisterPortalServer(srv, publicPortal)
	// Public Ticket模块服务注册
	publicticketv1.RegisterTicketServer(srv, publicTicket)
	// Public User模块服务注册
	publicuserv1.RegisterUserServer(srv, publicUser)
	return srv
}
