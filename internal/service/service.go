package service

import (
	"github.com/google/wire"

	ads "github.com/OmnTeam/ppanel-pro/internal/service/admin/ads"
	announcement "github.com/OmnTeam/ppanel-pro/internal/service/admin/announcement"
	application "github.com/OmnTeam/ppanel-pro/internal/service/admin/application"
	authmethod "github.com/OmnTeam/ppanel-pro/internal/service/admin/authmethod"
	adminconsole "github.com/OmnTeam/ppanel-pro/internal/service/admin/console"
	admincoupon "github.com/OmnTeam/ppanel-pro/internal/service/admin/coupon"
	admindocument "github.com/OmnTeam/ppanel-pro/internal/service/admin/document"
	maingroup "github.com/OmnTeam/ppanel-pro/internal/service/admin/group"
	adminlog "github.com/OmnTeam/ppanel-pro/internal/service/admin/log"
	adminmarketing "github.com/OmnTeam/ppanel-pro/internal/service/admin/marketing"
	adminorder "github.com/OmnTeam/ppanel-pro/internal/service/admin/order"
	adminpayment "github.com/OmnTeam/ppanel-pro/internal/service/admin/payment"
	adminredemption "github.com/OmnTeam/ppanel-pro/internal/service/admin/redemption"
	adminserver "github.com/OmnTeam/ppanel-pro/internal/service/admin/server"
	adminsubscribe "github.com/OmnTeam/ppanel-pro/internal/service/admin/subscribe"
	adminsystem "github.com/OmnTeam/ppanel-pro/internal/service/admin/system"
	adminticket "github.com/OmnTeam/ppanel-pro/internal/service/admin/ticket"
	admintool "github.com/OmnTeam/ppanel-pro/internal/service/admin/tool"
	adminuser "github.com/OmnTeam/ppanel-pro/internal/service/admin/user"
	"github.com/OmnTeam/ppanel-pro/internal/service/auth"
	authoauth "github.com/OmnTeam/ppanel-pro/internal/service/auth/oauth"
	"github.com/OmnTeam/ppanel-pro/internal/service/common"
	publicorder "github.com/OmnTeam/ppanel-pro/internal/service/public"
	publicannouncement "github.com/OmnTeam/ppanel-pro/internal/service/public/announcement"
	publicdocument "github.com/OmnTeam/ppanel-pro/internal/service/public/document"
	publicpayment "github.com/OmnTeam/ppanel-pro/internal/service/public/payment"
	publicportal "github.com/OmnTeam/ppanel-pro/internal/service/public/portal"
	publicredemption "github.com/OmnTeam/ppanel-pro/internal/service/public/redemption"
	publicsubscribe "github.com/OmnTeam/ppanel-pro/internal/service/public/subscribe"
	publicsubscription "github.com/OmnTeam/ppanel-pro/internal/service/public/subscription"
	publicticket "github.com/OmnTeam/ppanel-pro/internal/service/public/ticket"
	publicuser "github.com/OmnTeam/ppanel-pro/internal/service/public/user"
	// Server模块服务
	"github.com/OmnTeam/ppanel-pro/internal/service/server"
)

// ProviderSet is service providers
var ProviderSet = wire.NewSet(
	ads.NewAdsService,
	announcement.NewAnnouncementService,
	application.NewSubscribeApplicationService,
	authmethod.NewAuthMethodService,
	adminconsole.NewConsoleService,
	admincoupon.NewCouponService,
	admindocument.NewDocumentService,
	adminlog.NewLogService,
	adminmarketing.NewMarketingService,
	adminorder.NewOrderService,
	adminpayment.NewPaymentService,
	adminserver.NewServerService,
	adminsubscribe.NewSubscribeService,
	adminsystem.NewSystemService,
	adminticket.NewTicketService,
	adminredemption.NewRedemptionService,
	admintool.NewToolService,
	maingroup.NewGroupService,
	// Admin User模块服务
	adminuser.NewUserService,
	adminuser.NewUserAuthMethodService,
	adminuser.NewUserDeviceService,
	adminuser.NewUserSubscribeService,
	// Auth模块服务
	auth.NewAuthService,
	// Auth OAuth模块服务
	authoauth.NewOAuthService,
	// Common模块服务
	common.NewCommonService,
	// Public Order模块服务
	publicorder.NewPublicOrderService,
	// Public Announcement模块服务
	publicannouncement.NewAnnouncementService,
	// Public Document模块服务
	publicdocument.NewDocumentService,
	// Public Payment模块服务
	publicpayment.NewPaymentService,
	// Public Portal模块服务
	publicportal.NewPortalService,
	// Public Subscribe模块服务
	publicsubscribe.NewSubscribeService,
	// Public Subscription模块服务（订阅配置生成）
	publicsubscription.NewPublicSubscriptionService,
	// Public Redemption模块服务
	publicredemption.NewRedemptionService,
	// Public Ticket模块服务
	publicticket.NewTicketService,
	// Public User模块服务
	publicuser.NewUserService,
	// Server模块服务
	server.NewServerService,
)
