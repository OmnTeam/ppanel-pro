package biz

import (
	"github.com/google/wire"

	ads "github.com/OmnTeam/ppanel-pro/internal/biz/admin/ads"
	announcement "github.com/OmnTeam/ppanel-pro/internal/biz/admin/announcement"
	application "github.com/OmnTeam/ppanel-pro/internal/biz/admin/application"
	authmethod "github.com/OmnTeam/ppanel-pro/internal/biz/admin/authmethod"
	adminconsole "github.com/OmnTeam/ppanel-pro/internal/biz/admin/console"
	admincoupon "github.com/OmnTeam/ppanel-pro/internal/biz/admin/coupon"
	admindocument "github.com/OmnTeam/ppanel-pro/internal/biz/admin/document"
	maingroup "github.com/OmnTeam/ppanel-pro/internal/biz/admin/group"
	adminlog "github.com/OmnTeam/ppanel-pro/internal/biz/admin/log"
	adminmarketing "github.com/OmnTeam/ppanel-pro/internal/biz/admin/marketing"
	adminorder "github.com/OmnTeam/ppanel-pro/internal/biz/admin/order"
	adminpayment "github.com/OmnTeam/ppanel-pro/internal/biz/admin/payment"
	adminredemption "github.com/OmnTeam/ppanel-pro/internal/biz/admin/redemption"
	adminserver "github.com/OmnTeam/ppanel-pro/internal/biz/admin/server"
	adminsubscribe "github.com/OmnTeam/ppanel-pro/internal/biz/admin/subscribe"
	adminsystem "github.com/OmnTeam/ppanel-pro/internal/biz/admin/system"
	adminticket "github.com/OmnTeam/ppanel-pro/internal/biz/admin/ticket"
	admintool "github.com/OmnTeam/ppanel-pro/internal/biz/admin/tool"
	adminuser "github.com/OmnTeam/ppanel-pro/internal/biz/admin/user"
	"github.com/OmnTeam/ppanel-pro/internal/biz/auth"
	authoauth "github.com/OmnTeam/ppanel-pro/internal/biz/auth/oauth"
	"github.com/OmnTeam/ppanel-pro/internal/biz/common"
	publicorder "github.com/OmnTeam/ppanel-pro/internal/biz/public"
	publicannouncement "github.com/OmnTeam/ppanel-pro/internal/biz/public/announcement"
	publicdocument "github.com/OmnTeam/ppanel-pro/internal/biz/public/document"
	publicpayment "github.com/OmnTeam/ppanel-pro/internal/biz/public/payment"
	publicportal "github.com/OmnTeam/ppanel-pro/internal/biz/public/portal"
	publicsubscribe "github.com/OmnTeam/ppanel-pro/internal/biz/public/subscribe"
	publicticket "github.com/OmnTeam/ppanel-pro/internal/biz/public/ticket"
	publicuser "github.com/OmnTeam/ppanel-pro/internal/biz/public/user"
	publicwithdrawal "github.com/OmnTeam/ppanel-pro/internal/biz/public/withdrawal"
	// Server模块用例
	server "github.com/OmnTeam/ppanel-pro/internal/biz/server"
)

// ProviderSet is biz providers
var ProviderSet = wire.NewSet(
	ads.NewAdsUsecase,
	announcement.NewAnnouncementUsecase,
	application.NewSubscribeApplicationUsecase,
	authmethod.NewAuthMethodUsecase,
	adminconsole.NewConsoleUsecase,
	admincoupon.NewCouponUseCase,
	admindocument.NewDocumentUsecase,
	adminlog.NewSystemLogUsecase,
	adminlog.NewTrafficLogUsecase,
	adminlog.NewLogSettingUsecase,
	adminmarketing.NewMarketingUsecase,
	adminorder.NewOrderUseCase,
	adminpayment.NewPaymentUsecase,
	adminserver.NewServerUsecase,
	adminserver.NewNodeUsecase,
	adminserver.NewMigrationUsecase,
	adminsubscribe.NewSubscribeUseCase,
	adminsystem.NewSystemUsecase,
	adminticket.NewTicketUseCase,
	adminredemption.NewRedemptionUseCase,
	admintool.NewToolUseCase,
	maingroup.NewGroupUseCase,
	// Admin User模块用例
	adminuser.NewUserUsecase,
	adminuser.NewAuthMethodUsecase,
	adminuser.NewDeviceUsecase,
	adminuser.NewSubscribeUsecase,
	// Auth模块用例
	auth.NewAuthUsecase,
	// Auth OAuth模块用例
	authoauth.NewOAuthUseCase,
	// Common模块用例
	common.NewCommonUsecase,
	// Public Order模块用例
	publicorder.NewOrderUsecase,
	// Public Announcement模块用例
	publicannouncement.NewAnnouncementUseCase,
	// Public Document模块用例
	publicdocument.NewDocumentUseCase,
	// Public Payment模块用例
	publicpayment.NewPaymentUseCase,
	// Public Portal模块用例
	publicportal.NewPortalUseCase,
	// Public Subscribe模块用例
	publicsubscribe.NewSubscribeUseCase,
	// Public Ticket模块用例
	publicticket.NewTicketUseCase,
	// Public User模块用例
	publicuser.NewUserUseCase,
	// Public Withdrawal模块用例
	publicwithdrawal.NewWithdrawalUsecase,
	// Server模块用例
	server.NewServerNodeUsecase,
)
