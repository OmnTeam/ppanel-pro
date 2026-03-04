package responsecode

// Proxy Service 专用响应码表
// 基于 SaaS Response Code Design v4.0 规范
//
// 响应码格式：分类码(1位) + 模块码(3位) + 序号码(3位) = 7位数字
// 代理服务模块码：060
// 示例：2006003 = 成功状态(2) + 代理模块(060) + 序号(003)

const (
	// ==== 成功码 (2006003 - 2060999) ====

	// 用户管理相关
	UserCreated           = 2006003 // 用户创建成功
	UserUpdated           = 2006004 // 用户更新成功
	UserDeleted           = 2006005 // 用户删除成功
	UserListRetrieved     = 2006006 // 用户列表获取成功
	UserDetailRetrieved   = 2006007 // 用户详情获取成功
	PasswordUpdated       = 2006008 // 密码更新成功
	NotifySettingsUpdated = 2006009 // 通知设置更新成功

	// 订阅管理相关 (2006003100-2006003199)

	SubscribeCreated         = 2006010 // 订阅创建成功
	SubscribeUpdated         = 2006011 // 订阅更新成功
	SubscribeDeleted         = 2006012 // 订阅删除成功
	SubscribeListRetrieved   = 2006013 // 订阅列表获取成功
	SubscribeDetailRetrieved = 2006014 // 订阅详情获取成功
	SubscribeSorted          = 2006015 // 订阅排序成功

	// 订单管理相关 (2006003200-2006003299)
	OrderCreated         = 2006016 // 订单创建成功
	OrderUpdated         = 2006017 // 订单更新成功
	OrderDeleted         = 2006018 // 订单删除成功
	OrderListRetrieved   = 2006019 // 订单列表获取成功
	OrderDetailRetrieved = 2006020 // 订单详情获取成功
	OrderCancelled       = 2006021 // 订单取消成功
	OrderCompleted       = 2006022 // 订单完成成功

	// 支付管理相关 (2006003300-2006003399)
	PaymentCreated         = 2006023 // 支付方式创建成功
	PaymentUpdated         = 2006024 // 支付方式更新成功
	PaymentDeleted         = 2006025 // 支付方式删除成功
	PaymentListRetrieved   = 2006026 // 支付方式列表获取成功
	PaymentDetailRetrieved = 2006027 // 支付方式详情获取成功
	PaymentToggled         = 2006028 // 支付方式状态切换成功

	// 服务器管理相关 (2006003400-2006003499)
	ServerCreated         = 2006029 // 服务器创建成功
	ServerUpdated         = 2006030 // 服务器更新成功
	ServerDeleted         = 2006031 // 服务器删除成功
	ServerListRetrieved   = 2006032 // 服务器列表获取成功
	ServerDetailRetrieved = 2006033 // 服务器详情获取成功
	ServerToggled         = 2006034 // 服务器状态切换成功
	ServerSortUpdated     = 2006035 // 服务器排序更新成功

	// 节点管理相关 (2006003500-2006003599)
	NodeCreated         = 2006036 // 节点创建成功
	NodeUpdated         = 2006037 // 节点更新成功
	NodeDeleted         = 2006038 // 节点删除成功
	NodeListRetrieved   = 2006039 // 节点列表获取成功
	NodeDetailRetrieved = 2006040 // 节点详情获取成功
	NodeToggled         = 2006041 // 节点状态切换成功
	NodeSortUpdated     = 2006042 // 节点排序更新成功

	// 优惠券管理相关 (2006003600-2006003699)
	CouponCreated         = 2006043 // 优惠券创建成功
	CouponUpdated         = 2006044 // 优惠券更新成功
	CouponDeleted         = 2006045 // 优惠券删除成功
	CouponListRetrieved   = 2006046 // 优惠券列表获取成功
	CouponDetailRetrieved = 2006047 // 优惠券详情获取成功
	CouponValidated       = 2006048 // 优惠券验证成功
	CouponUsed            = 2006049 // 优惠券使用成功
	DiscountCalculated    = 2006050 // 折扣计算成功

	// 工单管理相关 (2006003700-2006003799)
	TicketCreated         = 2006051 // 工单创建成功
	TicketUpdated         = 2006052 // 工单更新成功
	TicketDeleted         = 2006053 // 工单删除成功
	TicketListRetrieved   = 2006054 // 工单列表获取成功
	TicketDetailRetrieved = 2006055 // 工单详情获取成功
	TicketStatusUpdated   = 2006056 // 工单状态更新成功

	// 系统日志管理相关 (2006003800-2006003899)
	SystemLogCreated         = 2006060 // 系统日志创建成功
	SystemLogDeleted         = 2006061 // 系统日志删除成功
	SystemLogListRetrieved   = 2006062 // 系统日志列表获取成功
	SystemLogDetailRetrieved = 2006063 // 系统日志详情获取成功

	// 认证管理相关 (2006070-2006079)
	UserCheckSuccess     = 2006070 // 用户检查成功
	UserLoginSuccess     = 2006071 // 用户登录成功
	UserRegisterSuccess  = 2006072 // 用户注册成功
	PasswordResetSuccess = 2006073 // 密码重置成功

	// Public User 相关 (2006080-2006101)
	UserInfoQuerySuccess       = 2006080 // 查询用户信息成功
	LoginLogQuerySuccess       = 2006081 // 查询登录日志成功
	BalanceLogQuerySuccess     = 2006082 // 查询余额日志成功
	CommissionLogQuerySuccess  = 2006083 // 查询佣金日志成功
	AffiliateQuerySuccess      = 2006084 // 查询推荐成功
	AffiliateListQuerySuccess  = 2006085 // 查询推荐列表成功
	OAuthMethodsQuerySuccess   = 2006086 // 查询OAuth方法成功
	UserSubscribeQuerySuccess  = 2006087 // 查询用户订阅成功
	SubscribeLogQuerySuccess   = 2006088 // 查询订阅日志成功
	SubscribeTokenResetSuccess = 2006089 // 重置订阅令牌成功
	PreUnsubscribeSuccess      = 2006090 // 预退订成功
	UnsubscribeSuccess         = 2006091 // 退订成功
	NotifyUpdateSuccess        = 2006092 // 更新通知设置成功
	PasswordUpdateSuccess      = 2006093 // 更新密码成功
	TelegramBindSuccess        = 2006094 // 绑定Telegram成功
	TelegramUnbindSuccess      = 2006095 // 解绑Telegram成功
	OAuthBindSuccess           = 2006096 // 绑定OAuth成功
	OAuthCallbackSuccess       = 2006097 // OAuth回调成功
	OAuthUnbindSuccess         = 2006098 // 解绑OAuth成功
	EmailVerifySuccess         = 2006099 // 验证邮箱成功
	MobileBindSuccess          = 2006100 // 绑定手机成功
	EmailBindSuccess           = 2006101 // 绑定邮箱成功

	// 设备管理相关 (2006268-2006270)
	UserDeviceListQuerySuccess       = 2006268 // 查询设备列表成功
	UserDeviceUnbindSuccess          = 2006269 // 解绑设备成功
	UserDeviceStatisticsQuerySuccess = 2006270 // 获取设备在线统计成功

	// Public Order 相关 (2006102-2006109)
	OrderCloseSuccess       = 2006102 // 关闭订单成功
	OrderDetailQuerySuccess = 2006103 // 查询订单详情成功
	OrderListQuerySuccess   = 2006104 // 查询订单列表成功
	OrderPreCreateSuccess   = 2006105 // 预创建订单成功
	PurchaseSuccess         = 2006106 // 购买成功
	RechargeSuccess         = 2006107 // 充值成功
	RenewalSuccess          = 2006108 // 续费成功
	TrafficResetSuccess     = 2006109 // 重置流量成功

	// Auth OAuth 相关 (2006110-2006112)
	OAuthLoginSuccess    = 2006110 // OAuth登录成功
	OAuthTokenGetSuccess = 2006111 // 获取OAuth令牌成功
	AppleCallbackSuccess = 2006112 // Apple回调成功

	// Public Ticket 相关 (2006113-2006117)
	UserTicketCreateSuccess       = 2006113 // 创建工单成功
	UserTicketListQuerySuccess    = 2006114 // 查询工单列表成功
	UserTicketDetailQuerySuccess  = 2006115 // 查询工单详情成功
	UserTicketStatusUpdateSuccess = 2006116 // 更新工单状态成功
	UserTicketFollowCreateSuccess = 2006117 // 创建工单跟进成功

	// Public Common 相关 (2006118-2006126)
	GetAdsSuccess                = 2006118 // 获取广告列表成功
	GetClientSuccess             = 2006119 // 获取客户端列表成功
	GetPrivacyPolicySuccess      = 2006120 // 获取隐私政策成功
	GetTosSuccess                = 2006121 // 获取服务条款成功
	GetGlobalConfigSuccess       = 2006122 // 获取全局配置成功
	GetStatSuccess               = 2006123 // 获取统计数据成功
	SendEmailCodeSuccess         = 2006124 // 发送邮箱验证码成功
	SendSmsCodeSuccess           = 2006125 // 发送短信验证码成功
	CheckVerificationCodeSuccess = 2006126 // 验证码校验成功

	// Admin Log 相关 (2006127-2006141)
	FilterBalanceLogSuccess              = 2006127 // 查询余额日志成功
	FilterCommissionLogSuccess           = 2006128 // 查询佣金日志成功
	FilterEmailLogSuccess                = 2006129 // 查询邮件日志成功
	FilterGiftLogSuccess                 = 2006130 // 查询礼品日志成功
	FilterLoginLogSuccess                = 2006131 // 查询登录日志成功
	GetMessageLogListSuccess             = 2006132 // 查询消息日志列表成功
	FilterMobileLogSuccess               = 2006133 // 查询短信日志成功
	FilterRegisterLogSuccess             = 2006134 // 查询注册日志成功
	FilterServerTrafficLogSuccess        = 2006135 // 查询服务器流量日志成功
	FilterSubscribeLogSuccess            = 2006136 // 查询订阅日志成功
	FilterResetSubscribeLogSuccess       = 2006137 // 查询重置订阅日志成功
	FilterUserSubscribeTrafficLogSuccess = 2006138 // 查询用户订阅流量日志成功
	FilterTrafficLogDetailsSuccess       = 2006139 // 查询流量日志详情成功
	GetLogSettingSuccess                 = 2006140 // 获取日志设置成功
	UpdateLogSettingSuccess              = 2006141 // 更新日志设置成功

	// Admin Ticket 相关 (2006142-2006145)
	AdminUpdateTicketStatusSuccess = 2006142 // 更新工单状态成功
	AdminGetTicketSuccess          = 2006143 // 获取工单详情成功
	AdminCreateTicketFollowSuccess = 2006144 // 创建工单跟进成功
	AdminGetTicketListSuccess      = 2006145 // 获取工单列表成功

	// Admin Order 相关 (2006146-2006148)
	AdminCreateOrderSuccess       = 2006146 // 创建订单成功
	AdminGetOrderListSuccess      = 2006147 // 获取订单列表成功
	AdminUpdateOrderStatusSuccess = 2006148 // 更新订单状态成功

	// Admin Console 相关 (2006149-2006152)
	QueryRevenueStatisticsSuccess = 2006149 // 查询营收统计成功
	QueryUserStatisticsSuccess    = 2006150 // 查询用户统计成功
	QueryTicketWaitReplySuccess   = 2006151 // 查询待回复工单成功
	QueryServerTotalDataSuccess   = 2006152 // 查询服务器总数据成功

	// Public Portal 相关 (2006153-2006158)
	GetSubscriptionSuccess            = 2006153 // 获取订阅列表成功
	PrePurchaseOrderSuccess           = 2006154 // 预购买订单成功
	PortalPurchaseSuccess             = 2006155 // 门户购买成功
	GetAvailablePaymentMethodsSuccess = 2006156 // 获取可用支付方式成功
	PurchaseCheckoutSuccess           = 2006157 // 购买结账成功
	QueryPurchaseOrderSuccess         = 2006158 // 查询购买订单成功

	// Public Subscribe 相关 (2006159)
	SubscribeQuerySuccess = 2006159 // 查询订阅列表成功

	// Public Announcement 相关 (2006160)
	AnnouncementQuerySuccess = 2006160 // 查询公告列表成功

	// Public Document 相关 (2006161)
	DocumentQuerySuccess = 2006161 // 查询文档成功

	// Admin Ads 相关 (2006162-2006166)
	AdminGetAdsListSuccess = 2006162 // 获取广告列表成功
	AdminGetAdsSuccess     = 2006163 // 获取广告详情成功
	AdminCreateAdsSuccess  = 2006164 // 创建广告成功
	AdminUpdateAdsSuccess  = 2006165 // 更新广告成功
	AdminDeleteAdsSuccess  = 2006166 // 删除广告成功

	// Admin Announcement 相关 (2006167-2006171)
	AdminCreateAnnouncementSuccess = 2006167 // 创建公告成功
	AdminUpdateAnnouncementSuccess = 2006168 // 更新公告成功
	AdminGetAnnouncementSuccess    = 2006169 // 获取公告详情成功
	AdminListAnnouncementsSuccess  = 2006170 // 获取公告列表成功
	AdminDeleteAnnouncementSuccess = 2006171 // 删除公告成功

	// Admin Application 相关 (2006172-2006176)
	AdminCreateSubscribeApplicationSuccess  = 2006172 // 创建订阅应用配置成功
	AdminPreviewSubscribeTemplateSuccess    = 2006173 // 预览订阅模板成功
	AdminUpdateSubscribeApplicationSuccess  = 2006174 // 更新订阅应用配置成功
	AdminDeleteSubscribeApplicationSuccess  = 2006175 // 删除订阅应用配置成功
	AdminGetSubscribeApplicationListSuccess = 2006176 // 获取订阅应用配置列表成功

	// Admin Coupon 相关 (2006177-2006181)
	AdminCreateCouponSuccess      = 2006177 // 创建优惠券成功
	AdminUpdateCouponSuccess      = 2006178 // 更新优惠券成功
	AdminDeleteCouponSuccess      = 2006179 // 删除优惠券成功
	AdminBatchDeleteCouponSuccess = 2006180 // 批量删除优惠券成功
	AdminGetCouponListSuccess     = 2006181 // 获取优惠券列表成功

	// Admin Payment 相关 (2006182-2006186)
	AdminCreatePaymentMethodSuccess  = 2006182 // 创建支付方式成功
	AdminUpdatePaymentMethodSuccess  = 2006183 // 更新支付方式成功
	AdminDeletePaymentMethodSuccess  = 2006184 // 删除支付方式成功
	AdminGetPaymentMethodListSuccess = 2006185 // 获取支付方式列表成功
	AdminGetPaymentPlatformSuccess   = 2006186 // 获取支付平台成功

	// Admin Document 相关 (2006187-2006192)
	AdminCreateDocumentSuccess      = 2006187 // 创建文档成功
	AdminUpdateDocumentSuccess      = 2006188 // 更新文档成功
	AdminDeleteDocumentSuccess      = 2006189 // 删除文档成功
	AdminBatchDeleteDocumentSuccess = 2006190 // 批量删除文档成功
	AdminGetDocumentListSuccess     = 2006191 // 获取文档列表成功
	AdminGetDocumentDetailSuccess   = 2006192 // 获取文档详情成功

	// Admin AuthMethod 相关 (2006193-2006199)
	AdminGetAuthMethodConfigSuccess    = 2006193 // 获取认证方法配置成功
	AdminUpdateAuthMethodConfigSuccess = 2006194 // 更新认证方法配置成功
	AdminGetEmailPlatformSuccess       = 2006195 // 获取邮件平台列表成功
	AdminGetSmsPlatformSuccess         = 2006196 // 获取短信平台列表成功
	AdminGetAuthMethodListSuccess      = 2006197 // 获取认证方法列表成功
	AdminTestEmailSendSuccess          = 2006198 // 测试邮件发送成功
	AdminTestSmsSendSuccess            = 2006199 // 测试短信发送成功

	// Admin Marketing 相关 (2006200-2006207)
	AdminCreateBatchSendEmailTaskSuccess    = 2006200 // 创建批量发送邮件任务成功
	AdminGetBatchSendEmailTaskListSuccess   = 2006201 // 获取批量发送邮件任务列表成功
	AdminStopBatchSendEmailTaskSuccess      = 2006202 // 停止批量发送邮件任务成功
	AdminGetPreSendEmailCountSuccess        = 2006203 // 获取预发送邮件数量成功
	AdminGetBatchSendEmailTaskStatusSuccess = 2006204 // 获取批量发送邮件任务状态成功
	AdminCreateQuotaTaskSuccess             = 2006205 // 创建配额任务成功
	AdminQueryQuotaTaskPreCountSuccess      = 2006206 // 查询配额任务预计数量成功
	AdminQueryQuotaTaskListSuccess          = 2006207 // 查询配额任务列表成功

	// Admin Subscribe 相关 (2006208-2006219)
	AdminCreateSubscribeSuccess           = 2006208 // 创建订阅套餐成功
	AdminUpdateSubscribeSuccess           = 2006209 // 更新订阅套餐成功
	AdminDeleteSubscribeSuccess           = 2006210 // 删除订阅套餐成功
	AdminBatchDeleteSubscribeSuccess      = 2006211 // 批量删除订阅套餐成功
	AdminGetSubscribeDetailsSuccess       = 2006212 // 获取订阅套餐详情成功
	AdminGetSubscribeListSuccess          = 2006213 // 获取订阅套餐列表成功
	AdminSubscribeSortSuccess             = 2006214 // 订阅套餐排序成功
	AdminCreateSubscribeGroupSuccess      = 2006215 // 创建订阅组成功
	AdminUpdateSubscribeGroupSuccess      = 2006216 // 更新订阅组成功
	AdminDeleteSubscribeGroupSuccess      = 2006217 // 删除订阅组成功
	AdminBatchDeleteSubscribeGroupSuccess = 2006218 // 批量删除订阅组成功
	AdminGetSubscribeGroupListSuccess     = 2006219 // 获取订阅组列表成功

	// Admin Server 相关 (2006220-2006234)
	AdminCreateServerSuccess         = 2006220 // 创建服务器成功
	AdminUpdateServerSuccess         = 2006221 // 更新服务器成功
	AdminDeleteServerSuccess         = 2006222 // 删除服务器成功
	AdminFilterServerListSuccess     = 2006223 // 获取服务器列表成功
	AdminGetServerProtocolsSuccess   = 2006224 // 获取服务器协议成功
	AdminCreateNodeSuccess           = 2006225 // 创建节点成功
	AdminUpdateNodeSuccess           = 2006226 // 更新节点成功
	AdminDeleteNodeSuccess           = 2006227 // 删除节点成功
	AdminFilterNodeListSuccess       = 2006228 // 获取节点列表成功
	AdminToggleNodeStatusSuccess     = 2006229 // 切换节点状态成功
	AdminQueryNodeTagSuccess         = 2006230 // 查询节点标签成功
	AdminHasMigrateServerNodeSuccess = 2006231 // 检查服务器节点迁移成功
	AdminMigrateServerNodeSuccess    = 2006232 // 迁移服务器节点成功
	AdminResetSortWithServerSuccess  = 2006233 // 重置服务器排序成功
	AdminResetSortWithNodeSuccess    = 2006234 // 重置节点排序成功

	// Admin User 相关 (2006235-2006243)
	AdminCreateUserSuccess               = 2006235 // 创建用户成功
	AdminDeleteUserSuccess               = 2006236 // 删除用户成功
	AdminBatchDeleteUserSuccess          = 2006237 // 批量删除用户成功
	AdminCurrentUserSuccess              = 2006238 // 获取当前用户成功
	AdminGetUserDetailSuccess            = 2006239 // 获取用户详情成功
	AdminGetUserListSuccess              = 2006240 // 获取用户列表成功
	AdminUpdateUserBasicInfoSuccess      = 2006241 // 更新用户基本信息成功
	AdminUpdateUserNotifySettingsSuccess = 2006242 // 更新用户通知设置成功
	AdminGetUserLoginLogsSuccess         = 2006243 // 获取用户登录日志成功

	// Admin System 相关 (2006244-2006267)
	AdminGetCurrencyConfigSuccess         = 2006244 // 获取货币配置成功
	AdminUpdateCurrencyConfigSuccess      = 2006245 // 更新货币配置成功
	AdminGetInviteConfigSuccess           = 2006246 // 获取邀请配置成功
	AdminUpdateInviteConfigSuccess        = 2006247 // 更新邀请配置成功
	AdminGetNodeConfigSuccess             = 2006248 // 获取节点配置成功
	AdminUpdateNodeConfigSuccess          = 2006249 // 更新节点配置成功
	AdminGetPrivacyPolicyConfigSuccess    = 2006250 // 获取隐私政策配置成功
	AdminUpdatePrivacyPolicyConfigSuccess = 2006251 // 更新隐私政策配置成功
	AdminGetRegisterConfigSuccess         = 2006252 // 获取注册配置成功
	AdminUpdateRegisterConfigSuccess      = 2006253 // 更新注册配置成功
	AdminGetSiteConfigSuccess             = 2006254 // 获取站点配置成功
	AdminUpdateSiteConfigSuccess          = 2006255 // 更新站点配置成功
	AdminGetSubscribeConfigSuccess        = 2006256 // 获取订阅配置成功
	AdminUpdateSubscribeConfigSuccess     = 2006257 // 更新订阅配置成功
	AdminGetTosConfigSuccess              = 2006258 // 获取服务条款配置成功
	AdminUpdateTosConfigSuccess           = 2006259 // 更新服务条款配置成功
	AdminGetVerifyCodeConfigSuccess       = 2006260 // 获取验证码配置成功
	AdminUpdateVerifyCodeConfigSuccess    = 2006261 // 更新验证码配置成功
	AdminGetVerifyConfigSuccess           = 2006262 // 获取验证配置成功
	AdminUpdateVerifyConfigSuccess        = 2006263 // 更新验证配置成功
	AdminGetNodeMultiplierSuccess         = 2006264 // 获取节点倍率成功
	AdminPreViewNodeMultiplierSuccess     = 2006265 // 预览节点倍率成功
	AdminSetNodeMultiplierSuccess         = 2006266 // 设置节点倍率成功
	AdminSettingTelegramBotSuccess        = 2006267 // 设置Telegram机器人成功

	// Admin Redemption 相关 (2006280-2006286)
	AdminCreateRedemptionCodeSuccess       = 2006280 // 创建兑换码成功
	AdminUpdateRedemptionCodeSuccess       = 2006281 // 更新兑换码成功
	AdminToggleRedemptionCodeStatusSuccess = 2006282 // 切换兑换码状态成功
	AdminDeleteRedemptionCodeSuccess       = 2006283 // 删除兑换码成功
	AdminBatchDeleteRedemptionCodeSuccess  = 2006284 // 批量删除兑换码成功
	AdminGetRedemptionCodeListSuccess      = 2006285 // 获取兑换码列表成功
	AdminGetRedemptionRecordListSuccess    = 2006286 // 获取兑换记录列表成功

	// Admin Tool 相关 (2006287-2006290)
	AdminGetSystemLogSuccess    = 2006287 // 获取系统日志成功
	AdminRestartSystemSuccess   = 2006288 // 重启系统成功
	AdminGetVersionSuccess      = 2006289 // 获取版本信息成功
	AdminQueryIPLocationSuccess = 2006290 // 查询IP地理位置成功

	// Admin Group 相关 (2006291-2006309)
	AdminGetUserGroupListSuccess       = 2006291 // 获取用户组列表成功
	AdminCreateUserGroupSuccess        = 2006292 // 创建用户组成功
	AdminUpdateUserGroupSuccess        = 2006293 // 更新用户组成功
	AdminDeleteUserGroupSuccess        = 2006294 // 删除用户组成功
	AdminUpdateUserUserGroupSuccess    = 2006295 // 更新用户的用户组成功
	AdminGetNodeGroupListSuccess       = 2006296 // 获取节点组列表成功
	AdminCreateNodeGroupSuccess        = 2006297 // 创建节点组成功
	AdminUpdateNodeGroupSuccess        = 2006298 // 更新节点组成功
	AdminDeleteNodeGroupSuccess        = 2006299 // 删除节点组成功
	AdminGetGroupConfigSuccess         = 2006300 // 获取分组配置成功
	AdminUpdateGroupConfigSuccess      = 2006301 // 更新分组配置成功
	AdminRecalculateGroupSuccess       = 2006302 // 重新计算分组成功
	AdminGetRecalculationStatusSuccess = 2006303 // 获取重新计算状态成功
	AdminGetGroupHistorySuccess        = 2006304 // 获取分组历史成功
	AdminGetGroupHistoryDetailSuccess  = 2006305 // 获取分组历史详情成功
	AdminExportGroupResultSuccess      = 2006306 // 导出分组结果成功
	AdminMigrateUsersSuccess           = 2006307 // 迁移用户成功
	AdminPreviewUserNodesSuccess       = 2006308 // 预览用户节点成功
	AdminResetGroupsSuccess            = 2006309 // 重置所有分组成功

	// ==== 业务错误码 (3006003 - 3060999) ====

	// 参数验证错误
	ErrInvalidUserID        = 3006003 // 无效的用户ID
	ErrInvalidTenantID      = 3006004 // 无效的租户ID
	ErrInvalidOrderID       = 3006005 // 无效的订单ID
	ErrInvalidSubscribeID   = 3006006 // 无效的订阅ID
	ErrInvalidPaymentID     = 3006007 // 无效的支付ID
	ErrInvalidServerID      = 3006008 // 无效的服务器ID
	ErrInvalidNodeID        = 3006009 // 无效的节点ID
	ErrInvalidCouponCode    = 3006010 // 无效的优惠券码
	ErrMissingRequiredParam = 3006011 // 缺少必需参数
	ErrInvalidParamFormat   = 3006012 // 参数格式错误

	// 数据不存在错误
	ErrUserNotFound                 = 3006013 // 用户不存在
	ErrOrderNotFound                = 3006014 // 订单不存在
	ErrSubscribeNotFound            = 3006015 // 订阅不存在
	ErrPaymentNotFound              = 3006016 // 支付方式不存在
	ErrServerNotFound               = 3006017 // 服务器不存在
	ErrNodeNotFound                 = 3006018 // 节点不存在
	ErrCouponNotFound               = 3006019 // 优惠券不存在
	ErrDeviceNotFound               = 3006020 // 设备不存在
	ErrAuthMethodNotFound           = 3006021 // 认证方法不存在
	ErrAnnouncementNotFound         = 3006022 // 公告不存在
	ErrDocumentNotFound             = 3006023 // 文档不存在
	ErrAdsNotFound                  = 3006024 // 广告不存在
	ErrSystemNotFound               = 3006025 // 系统配置不存在
	ErrSubscribeApplicationNotFound = 3006026 // 订阅申请不存在
	ErrServerGroupNotFound          = 3006056 // 服务器组不存在
	ErrTicketNotFound               = 3006057 // 工单不存在
	ErrSystemLogNotFound            = 3006060 // 系统日志不存在
	ErrTaskNotFound                 = 3006064 // 任务不存在
	ErrRedemptionCodeNotFound       = 3006072 // 兑换码不存在
	ErrUserGroupNotFound            = 3006073 // 用户组不存在

	// 业务逻辑错误
	ErrInvalidTaskType     = 3006065 // 无效的任务类型
	ErrInvalidTaskStatus   = 3006066 // 无效的任务状态
	ErrTaskCannotBeStopped = 3006067 // 任务无法停止

	// 数据冲突错误
	ErrUserAlreadyExists                 = 3006027 // 用户已存在
	ErrOrderAlreadyExists                = 3006028 // 订单已存在
	ErrSubscribeAlreadyExists            = 3006029 // 订阅已存在
	ErrPaymentAlreadyExists              = 3006030 // 支付方式已存在
	ErrServerAlreadyExists               = 3006031 // 服务器已存在
	ErrNodeAlreadyExists                 = 3006032 // 节点已存在
	ErrCouponAlreadyExists               = 3006033 // 优惠券已存在
	ErrDuplicateEmail                    = 3006034 // 邮箱已存在
	ErrDuplicateUsername                 = 3006035 // 用户名已存在
	ErrAnnouncementAlreadyExists         = 3006036 // 公告已存在
	ErrDocumentAlreadyExists             = 3006037 // 文档已存在
	ErrSystemAlreadyExists               = 3006038 // 系统配置已存在
	ErrAuthMethodAlreadyExists           = 3006039 // 认证方法已存在
	ErrSubscribeApplicationAlreadyExists = 3006040 // 订阅申请已存在

	// 业务逻辑错误
	ErrOrderCannotCancel       = 3006041 // 订单不能取消
	ErrOrderCannotComplete     = 3006042 // 订单不能完成
	ErrOrderCannotClose        = 3006071 // 订单不能关闭
	ErrCouponExpired           = 3006043 // 优惠券已过期
	ErrCouponNotAvailable      = 3006044 // 优惠券不可用
	ErrCouponUsedUp            = 3006045 // 优惠券已用完
	ErrCouponUserLimitExceeded = 3006046 // 优惠券用户使用次数超限
	ErrInsufficientBalance     = 3006047 // 余额不足
	ErrUserCommissionNotEnough = 3006074 // 佣金不足
	ErrOrderPaymentFailed      = 3006061 // 订单支付失败
	ErrDeviceLimitExceeded     = 3006048 // 设备数量超限
	ErrSubscribeExpired        = 3006049 // 订阅已过期
	ErrTrafficExceeded         = 3006050 // 流量超限
	ErrSubscribeInUse          = 3006069 // 订阅套餐正在使用中
	ErrInvalidOrderStatus      = 3006051 // 无效的订单状态
	ErrInvalidParameter        = 3006052 // 无效的参数
	ErrTitleRequired           = 3006053 // 标题不能为空
	ErrTypeRequired            = 3006054 // 类型不能为空
	ErrInvalidTimeRange        = 3006055 // 时间范围无效
	ErrInvalidTicketStatus     = 3006058 // 无效的工单状态
	ErrInvalidTicketPriority   = 3006059 // 无效的工单优先级
	ErrUnsupportedPlatform     = 3006068 // 不支持的支付平台

	// ==== 权限错误码 (4006003 - 4060999) ====

	// 认证错误
	ErrMissingAuthToken     = 4006003 // 缺少认证令牌
	ErrInvalidAuthToken     = 4006004 // 无效的认证令牌
	ErrAuthTokenExpired     = 4006005 // 认证令牌已过期
	ErrInvalidCredentials   = 4006006 // 无效的凭证
	ErrUserNotAuthenticated = 4006007 // 用户未认证
	ErrPasswordIncorrect    = 4006008 // 密码错误
	ErrAccountLocked        = 4006009 // 账户已锁定
	ErrAccountDisabled      = 4006010 // 账户已禁用

	// 授权错误
	ErrPermissionDenied       = 4006011 // 权限被拒绝
	ErrInsufficientPermission = 4006012 // 权限不足
	ErrResourceAccessDenied   = 4006013 // 资源访问被拒绝
	ErrOperationNotAllowed    = 4006014 // 操作不被允许
	ErrTenantAccessDenied     = 4006015 // 租户访问被拒绝
	ErrCrossTenantOperation   = 4006016 // 跨租户操作
	ErrNotResourceOwner       = 4006017 // 非资源所有者
	ErrInvalidAccess          = 4006018 // 无效访问

	// ==== 系统错误码 (5006003 - 5060999) ====

	// 数据库错误
	ErrDatabaseConnection  = 5006003 // 数据库连接失败
	ErrDatabaseQuery       = 5006004 // 数据库查询失败
	ErrDatabaseUpdate      = 5006005 // 数据库更新失败
	ErrDatabaseInsert      = 5006006 // 数据库插入失败
	ErrDatabaseDelete      = 5006007 // 数据库删除失败
	ErrDatabaseTransaction = 5006008 // 数据库事务失败
	ErrDatabaseConstraint  = 5006009 // 数据库约束错误
	ErrDatabaseTimeout     = 5006010 // 数据库超时

	// 缓存错误
	ErrCacheConnection  = 5006011 // 缓存连接失败
	ErrCacheGet         = 5006012 // 缓存获取失败
	ErrCacheSet         = 5006013 // 缓存设置失败
	ErrCacheDelete      = 5006014 // 缓存删除失败
	ErrCacheExpired     = 5006015 // 缓存已过期
	ErrCacheSerialize   = 5006016 // 缓存序列化失败
	ErrCacheDeserialize = 5006017 // 缓存反序列化失败

	// 内部服务错误
	ErrInternalError        = 5006018 // 内部错误
	ErrServiceUnavailable   = 5006019 // 服务不可用
	ErrServiceTimeout       = 5006020 // 服务超时
	ErrConfigurationError   = 5006021 // 配置错误
	ErrInitializationFailed = 5006022 // 初始化失败
	ErrResourceExhausted    = 5006023 // 资源耗尽

	// 外部服务错误
	ErrIPGeolocationFailed = 5006024 // IP地理位置查询失败
	ErrPaymentGatewayError = 5006025 // 支付网关错误
	ErrEmailSendFailed     = 5006026 // 邮件发送失败
	ErrSMSSendFailed       = 5006027 // 短信发送失败
	ErrThirdPartyAPIError  = 5006028 // 第三方API错误
)

// CodeMessages 响应码消息映射表（供scanner扫描）
var CodeMessages = map[int]string{
	// 成功码消息 - 用户管理
	UserCreated:           "用户创建成功",
	UserUpdated:           "用户更新成功",
	UserDeleted:           "用户删除成功",
	UserListRetrieved:     "用户列表获取成功",
	UserDetailRetrieved:   "用户详情获取成功",
	PasswordUpdated:       "密码更新成功",
	NotifySettingsUpdated: "通知设置更新成功",

	// 成功码消息 - 订阅管理
	SubscribeCreated:         "订阅创建成功",
	SubscribeUpdated:         "订阅更新成功",
	SubscribeDeleted:         "订阅删除成功",
	SubscribeListRetrieved:   "订阅列表获取成功",
	SubscribeDetailRetrieved: "订阅详情获取成功",
	SubscribeSorted:          "订阅排序成功",

	// 成功码消息 - 订单管理
	OrderCreated:         "订单创建成功",
	OrderUpdated:         "订单更新成功",
	OrderDeleted:         "订单删除成功",
	OrderListRetrieved:   "订单列表获取成功",
	OrderDetailRetrieved: "订单详情获取成功",
	OrderCancelled:       "订单取消成功",
	OrderCompleted:       "订单完成成功",

	// 成功码消息 - 支付管理
	PaymentCreated:         "支付方式创建成功",
	PaymentUpdated:         "支付方式更新成功",
	PaymentDeleted:         "支付方式删除成功",
	PaymentListRetrieved:   "支付方式列表获取成功",
	PaymentDetailRetrieved: "支付方式详情获取成功",
	PaymentToggled:         "支付方式状态切换成功",

	// 成功码消息 - 服务器管理
	ServerCreated:         "服务器创建成功",
	ServerUpdated:         "服务器更新成功",
	ServerDeleted:         "服务器删除成功",
	ServerListRetrieved:   "服务器列表获取成功",
	ServerDetailRetrieved: "服务器详情获取成功",
	ServerToggled:         "服务器状态切换成功",
	ServerSortUpdated:     "服务器排序更新成功",

	// 成功码消息 - 节点管理
	NodeCreated:         "节点创建成功",
	NodeUpdated:         "节点更新成功",
	NodeDeleted:         "节点删除成功",
	NodeListRetrieved:   "节点列表获取成功",
	NodeDetailRetrieved: "节点详情获取成功",
	NodeToggled:         "节点状态切换成功",
	NodeSortUpdated:     "节点排序更新成功",

	// 成功码消息 - 优惠券管理
	CouponCreated:         "优惠券创建成功",
	CouponUpdated:         "优惠券更新成功",
	CouponDeleted:         "优惠券删除成功",
	CouponListRetrieved:   "优惠券列表获取成功",
	CouponDetailRetrieved: "优惠券详情获取成功",
	CouponValidated:       "优惠券验证成功",
	CouponUsed:            "优惠券使用成功",
	DiscountCalculated:    "折扣计算成功",

	// 成功码消息 - 工单管理
	TicketCreated:         "工单创建成功",
	TicketUpdated:         "工单更新成功",
	TicketDeleted:         "工单删除成功",
	TicketListRetrieved:   "工单列表获取成功",
	TicketDetailRetrieved: "工单详情获取成功",
	TicketStatusUpdated:   "工单状态更新成功",

	// 成功码消息 - 系统日志管理
	SystemLogCreated:         "系统日志创建成功",
	SystemLogDeleted:         "系统日志删除成功",
	SystemLogListRetrieved:   "系统日志列表获取成功",
	SystemLogDetailRetrieved: "系统日志详情获取成功",

	// 成功码消息 - 认证管理
	UserCheckSuccess:     "用户检查成功",
	UserLoginSuccess:     "用户登录成功",
	UserRegisterSuccess:  "用户注册成功",
	PasswordResetSuccess: "密码重置成功",

	// 成功码消息 - Public User
	UserInfoQuerySuccess:       "查询用户信息成功",
	LoginLogQuerySuccess:       "查询登录日志成功",
	BalanceLogQuerySuccess:     "查询余额日志成功",
	CommissionLogQuerySuccess:  "查询佣金日志成功",
	AffiliateQuerySuccess:      "查询推荐成功",
	AffiliateListQuerySuccess:  "查询推荐列表成功",
	OAuthMethodsQuerySuccess:   "查询OAuth方法成功",
	UserSubscribeQuerySuccess:  "查询用户订阅成功",
	SubscribeLogQuerySuccess:   "查询订阅日志成功",
	SubscribeTokenResetSuccess: "重置订阅令牌成功",
	PreUnsubscribeSuccess:      "预退订成功",
	UnsubscribeSuccess:         "退订成功",
	NotifyUpdateSuccess:        "更新通知设置成功",
	PasswordUpdateSuccess:      "更新密码成功",
	TelegramBindSuccess:        "绑定Telegram成功",
	TelegramUnbindSuccess:      "解绑Telegram成功",
	OAuthBindSuccess:           "绑定OAuth成功",
	OAuthCallbackSuccess:       "OAuth回调成功",
	OAuthUnbindSuccess:         "解绑OAuth成功",
	EmailVerifySuccess:         "验证邮箱成功",
	MobileBindSuccess:          "绑定手机成功",
	EmailBindSuccess:           "绑定邮箱成功",

	// 成功码消息 - 设备管理
	UserDeviceListQuerySuccess:       "查询设备列表成功",
	UserDeviceUnbindSuccess:          "解绑设备成功",
	UserDeviceStatisticsQuerySuccess: "获取设备在线统计成功",

	// 成功码消息 - Public Order
	OrderCloseSuccess:       "关闭订单成功",
	OrderDetailQuerySuccess: "查询订单详情成功",
	OrderListQuerySuccess:   "查询订单列表成功",
	OrderPreCreateSuccess:   "预创建订单成功",
	PurchaseSuccess:         "购买成功",
	RechargeSuccess:         "充值成功",
	RenewalSuccess:          "续费成功",
	TrafficResetSuccess:     "重置流量成功",

	// 成功码消息 - Auth OAuth
	OAuthLoginSuccess:    "OAuth登录成功",
	OAuthTokenGetSuccess: "获取OAuth令牌成功",
	AppleCallbackSuccess: "Apple回调成功",

	// 成功码消息 - Public Ticket
	UserTicketCreateSuccess:       "创建工单成功",
	UserTicketListQuerySuccess:    "查询工单列表成功",
	UserTicketDetailQuerySuccess:  "查询工单详情成功",
	UserTicketStatusUpdateSuccess: "更新工单状态成功",
	UserTicketFollowCreateSuccess: "创建工单跟进成功",

	// 成功码消息 - Public Common
	GetAdsSuccess:                "获取广告列表成功",
	GetClientSuccess:             "获取客户端列表成功",
	GetPrivacyPolicySuccess:      "获取隐私政策成功",
	GetTosSuccess:                "获取服务条款成功",
	GetGlobalConfigSuccess:       "获取全局配置成功",
	GetStatSuccess:               "获取统计数据成功",
	SendEmailCodeSuccess:         "发送邮箱验证码成功",
	SendSmsCodeSuccess:           "发送短信验证码成功",
	CheckVerificationCodeSuccess: "验证码校验成功",

	// 成功码消息 - Admin Log
	FilterBalanceLogSuccess:              "查询余额日志成功",
	FilterCommissionLogSuccess:           "查询佣金日志成功",
	FilterEmailLogSuccess:                "查询邮件日志成功",
	FilterGiftLogSuccess:                 "查询礼品日志成功",
	FilterLoginLogSuccess:                "查询登录日志成功",
	GetMessageLogListSuccess:             "查询消息日志列表成功",
	FilterMobileLogSuccess:               "查询短信日志成功",
	FilterRegisterLogSuccess:             "查询注册日志成功",
	FilterServerTrafficLogSuccess:        "查询服务器流量日志成功",
	FilterSubscribeLogSuccess:            "查询订阅日志成功",
	FilterResetSubscribeLogSuccess:       "查询重置订阅日志成功",
	FilterUserSubscribeTrafficLogSuccess: "查询用户订阅流量日志成功",
	FilterTrafficLogDetailsSuccess:       "查询流量日志详情成功",
	GetLogSettingSuccess:                 "获取日志设置成功",
	UpdateLogSettingSuccess:              "更新日志设置成功",

	// 成功码消息 - Admin Ticket
	AdminUpdateTicketStatusSuccess: "更新工单状态成功",
	AdminGetTicketSuccess:          "获取工单详情成功",
	AdminCreateTicketFollowSuccess: "创建工单跟进成功",
	AdminGetTicketListSuccess:      "获取工单列表成功",

	// 成功码消息 - Admin Order
	AdminCreateOrderSuccess:       "创建订单成功",
	AdminGetOrderListSuccess:      "获取订单列表成功",
	AdminUpdateOrderStatusSuccess: "更新订单状态成功",

	// 成功码消息 - Admin Console
	QueryRevenueStatisticsSuccess: "查询营收统计成功",
	QueryUserStatisticsSuccess:    "查询用户统计成功",
	QueryTicketWaitReplySuccess:   "查询待回复工单成功",
	QueryServerTotalDataSuccess:   "查询服务器总数据成功",

	// 成功码消息 - Public Portal
	GetSubscriptionSuccess:            "获取订阅列表成功",
	PrePurchaseOrderSuccess:           "预购买订单成功",
	PortalPurchaseSuccess:             "门户购买成功",
	GetAvailablePaymentMethodsSuccess: "获取可用支付方式成功",
	PurchaseCheckoutSuccess:           "购买结账成功",
	QueryPurchaseOrderSuccess:         "查询购买订单成功",

	// 成功码消息 - Public Subscribe
	SubscribeQuerySuccess: "查询订阅列表成功",

	// 成功码消息 - Public Announcement
	AnnouncementQuerySuccess: "查询公告列表成功",

	// 成功码消息 - Public Document
	DocumentQuerySuccess: "查询文档成功",

	// 成功码消息 - Admin Ads
	AdminGetAdsListSuccess: "获取广告列表成功",
	AdminGetAdsSuccess:     "获取广告详情成功",
	AdminCreateAdsSuccess:  "创建广告成功",
	AdminUpdateAdsSuccess:  "更新广告成功",
	AdminDeleteAdsSuccess:  "删除广告成功",

	// 成功码消息 - Admin Announcement
	AdminCreateAnnouncementSuccess: "创建公告成功",
	AdminUpdateAnnouncementSuccess: "更新公告成功",
	AdminGetAnnouncementSuccess:    "获取公告详情成功",
	AdminListAnnouncementsSuccess:  "获取公告列表成功",
	AdminDeleteAnnouncementSuccess: "删除公告成功",

	// 成功码消息 - Admin Application
	AdminCreateSubscribeApplicationSuccess:  "创建订阅应用配置成功",
	AdminPreviewSubscribeTemplateSuccess:    "预览订阅模板成功",
	AdminUpdateSubscribeApplicationSuccess:  "更新订阅应用配置成功",
	AdminDeleteSubscribeApplicationSuccess:  "删除订阅应用配置成功",
	AdminGetSubscribeApplicationListSuccess: "获取订阅应用配置列表成功",

	// 成功码消息 - Admin Coupon
	AdminCreateCouponSuccess:      "创建优惠券成功",
	AdminUpdateCouponSuccess:      "更新优惠券成功",
	AdminDeleteCouponSuccess:      "删除优惠券成功",
	AdminBatchDeleteCouponSuccess: "批量删除优惠券成功",
	AdminGetCouponListSuccess:     "获取优惠券列表成功",

	// 成功码消息 - Admin Payment
	AdminCreatePaymentMethodSuccess:  "创建支付方式成功",
	AdminUpdatePaymentMethodSuccess:  "更新支付方式成功",
	AdminDeletePaymentMethodSuccess:  "删除支付方式成功",
	AdminGetPaymentMethodListSuccess: "获取支付方式列表成功",
	AdminGetPaymentPlatformSuccess:   "获取支付平台成功",

	// 成功码消息 - Admin Document
	AdminCreateDocumentSuccess:      "创建文档成功",
	AdminUpdateDocumentSuccess:      "更新文档成功",
	AdminDeleteDocumentSuccess:      "删除文档成功",
	AdminBatchDeleteDocumentSuccess: "批量删除文档成功",
	AdminGetDocumentListSuccess:     "获取文档列表成功",
	AdminGetDocumentDetailSuccess:   "获取文档详情成功",

	// 成功码消息 - Admin AuthMethod
	AdminGetAuthMethodConfigSuccess:    "获取认证方法配置成功",
	AdminUpdateAuthMethodConfigSuccess: "更新认证方法配置成功",
	AdminGetEmailPlatformSuccess:       "获取邮件平台列表成功",
	AdminGetSmsPlatformSuccess:         "获取短信平台列表成功",
	AdminGetAuthMethodListSuccess:      "获取认证方法列表成功",
	AdminTestEmailSendSuccess:          "测试邮件发送成功",
	AdminTestSmsSendSuccess:            "测试短信发送成功",

	// 成功码消息 - Admin Marketing
	AdminCreateBatchSendEmailTaskSuccess:    "创建批量发送邮件任务成功",
	AdminGetBatchSendEmailTaskListSuccess:   "获取批量发送邮件任务列表成功",
	AdminStopBatchSendEmailTaskSuccess:      "停止批量发送邮件任务成功",
	AdminGetPreSendEmailCountSuccess:        "获取预发送邮件数量成功",
	AdminGetBatchSendEmailTaskStatusSuccess: "获取批量发送邮件任务状态成功",
	AdminCreateQuotaTaskSuccess:             "创建配额任务成功",
	AdminQueryQuotaTaskPreCountSuccess:      "查询配额任务预计数量成功",
	AdminQueryQuotaTaskListSuccess:          "查询配额任务列表成功",

	// 成功码消息 - Admin Subscribe
	AdminCreateSubscribeSuccess:           "创建订阅套餐成功",
	AdminUpdateSubscribeSuccess:           "更新订阅套餐成功",
	AdminDeleteSubscribeSuccess:           "删除订阅套餐成功",
	AdminBatchDeleteSubscribeSuccess:      "批量删除订阅套餐成功",
	AdminGetSubscribeDetailsSuccess:       "获取订阅套餐详情成功",
	AdminGetSubscribeListSuccess:          "获取订阅套餐列表成功",
	AdminSubscribeSortSuccess:             "订阅套餐排序成功",
	AdminCreateSubscribeGroupSuccess:      "创建订阅组成功",
	AdminUpdateSubscribeGroupSuccess:      "更新订阅组成功",
	AdminDeleteSubscribeGroupSuccess:      "删除订阅组成功",
	AdminBatchDeleteSubscribeGroupSuccess: "批量删除订阅组成功",
	AdminGetSubscribeGroupListSuccess:     "获取订阅组列表成功",

	// 成功码消息 - Admin Server
	AdminCreateServerSuccess:         "创建服务器成功",
	AdminUpdateServerSuccess:         "更新服务器成功",
	AdminDeleteServerSuccess:         "删除服务器成功",
	AdminFilterServerListSuccess:     "获取服务器列表成功",
	AdminGetServerProtocolsSuccess:   "获取服务器协议成功",
	AdminCreateNodeSuccess:           "创建节点成功",
	AdminUpdateNodeSuccess:           "更新节点成功",
	AdminDeleteNodeSuccess:           "删除节点成功",
	AdminFilterNodeListSuccess:       "获取节点列表成功",
	AdminToggleNodeStatusSuccess:     "切换节点状态成功",
	AdminQueryNodeTagSuccess:         "查询节点标签成功",
	AdminHasMigrateServerNodeSuccess: "检查服务器节点迁移成功",
	AdminMigrateServerNodeSuccess:    "迁移服务器节点成功",
	AdminResetSortWithServerSuccess:  "重置服务器排序成功",
	AdminResetSortWithNodeSuccess:    "重置节点排序成功",

	// 成功码消息 - Admin User
	AdminCreateUserSuccess:               "创建用户成功",
	AdminDeleteUserSuccess:               "删除用户成功",
	AdminBatchDeleteUserSuccess:          "批量删除用户成功",
	AdminCurrentUserSuccess:              "获取当前用户成功",
	AdminGetUserDetailSuccess:            "获取用户详情成功",
	AdminGetUserListSuccess:              "获取用户列表成功",
	AdminUpdateUserBasicInfoSuccess:      "更新用户基本信息成功",
	AdminUpdateUserNotifySettingsSuccess: "更新用户通知设置成功",
	AdminGetUserLoginLogsSuccess:         "获取用户登录日志成功",

	// 成功码消息 - Admin System
	AdminGetCurrencyConfigSuccess:          "获取货币配置成功",
	AdminUpdateCurrencyConfigSuccess:       "更新货币配置成功",
	AdminGetInviteConfigSuccess:            "获取邀请配置成功",
	AdminUpdateInviteConfigSuccess:         "更新邀请配置成功",
	AdminGetNodeConfigSuccess:              "获取节点配置成功",
	AdminUpdateNodeConfigSuccess:           "更新节点配置成功",
	AdminGetPrivacyPolicyConfigSuccess:     "获取隐私政策配置成功",
	AdminUpdatePrivacyPolicyConfigSuccess:  "更新隐私政策配置成功",
	AdminGetRegisterConfigSuccess:          "获取注册配置成功",
	AdminUpdateRegisterConfigSuccess:       "更新注册配置成功",
	AdminGetSiteConfigSuccess:              "获取站点配置成功",
	AdminUpdateSiteConfigSuccess:           "更新站点配置成功",
	AdminGetSubscribeConfigSuccess:         "获取订阅配置成功",
	AdminUpdateSubscribeConfigSuccess:      "更新订阅配置成功",
	AdminGetTosConfigSuccess:               "获取服务条款配置成功",
	AdminUpdateTosConfigSuccess:            "更新服务条款配置成功",
	AdminGetVerifyCodeConfigSuccess:        "获取验证码配置成功",
	AdminUpdateVerifyCodeConfigSuccess:     "更新验证码配置成功",
	AdminGetVerifyConfigSuccess:            "获取验证配置成功",
	AdminUpdateVerifyConfigSuccess:         "更新验证配置成功",
	AdminGetNodeMultiplierSuccess:          "获取节点倍率成功",
	AdminPreViewNodeMultiplierSuccess:      "预览节点倍率成功",
	AdminSetNodeMultiplierSuccess:          "设置节点倍率成功",
	AdminSettingTelegramBotSuccess:         "设置Telegram机器人成功",
	AdminCreateRedemptionCodeSuccess:       "创建兑换码成功",
	AdminUpdateRedemptionCodeSuccess:       "更新兑换码成功",
	AdminToggleRedemptionCodeStatusSuccess: "切换兑换码状态成功",
	AdminDeleteRedemptionCodeSuccess:       "删除兑换码成功",
	AdminBatchDeleteRedemptionCodeSuccess:  "批量删除兑换码成功",
	AdminGetRedemptionCodeListSuccess:      "获取兑换码列表成功",
	AdminGetRedemptionRecordListSuccess:    "获取兑换记录列表成功",
	AdminGetSystemLogSuccess:               "获取系统日志成功",
	AdminRestartSystemSuccess:              "重启系统成功",
	AdminGetVersionSuccess:                 "获取版本信息成功",
	AdminQueryIPLocationSuccess:            "查询IP地理位置成功",
	AdminGetUserGroupListSuccess:           "获取用户组列表成功",
	AdminCreateUserGroupSuccess:            "创建用户组成功",
	AdminUpdateUserGroupSuccess:            "更新用户组成功",
	AdminDeleteUserGroupSuccess:            "删除用户组成功",
	AdminUpdateUserUserGroupSuccess:        "更新用户的用户组成功",
	AdminGetNodeGroupListSuccess:           "获取节点组列表成功",
	AdminCreateNodeGroupSuccess:            "创建节点组成功",
	AdminUpdateNodeGroupSuccess:            "更新节点组成功",
	AdminDeleteNodeGroupSuccess:            "删除节点组成功",
	AdminGetGroupConfigSuccess:             "获取分组配置成功",
	AdminUpdateGroupConfigSuccess:          "更新分组配置成功",
	AdminRecalculateGroupSuccess:           "重新计算分组成功",
	AdminGetRecalculationStatusSuccess:     "获取重新计算状态成功",
	AdminGetGroupHistorySuccess:            "获取分组历史成功",
	AdminGetGroupHistoryDetailSuccess:      "获取分组历史详情成功",
	AdminExportGroupResultSuccess:          "导出分组结果成功",
	AdminMigrateUsersSuccess:               "迁移用户成功",
	AdminPreviewUserNodesSuccess:           "预览用户节点成功",
	AdminResetGroupsSuccess:                "重置所有分组成功",

	// 业务错误码消息 - 参数验证
	ErrInvalidUserID:        "无效的用户ID",
	ErrInvalidTenantID:      "无效的租户ID",
	ErrInvalidOrderID:       "无效的订单ID",
	ErrInvalidSubscribeID:   "无效的订阅ID",
	ErrInvalidPaymentID:     "无效的支付ID",
	ErrInvalidServerID:      "无效的服务器ID",
	ErrInvalidNodeID:        "无效的节点ID",
	ErrInvalidCouponCode:    "无效的优惠券码",
	ErrMissingRequiredParam: "缺少必需参数",
	ErrInvalidParamFormat:   "参数格式错误",

	// 业务错误码消息 - 数据不存在
	ErrUserNotFound:                 "用户不存在",
	ErrOrderNotFound:                "订单不存在",
	ErrSubscribeNotFound:            "订阅不存在",
	ErrPaymentNotFound:              "支付方式不存在",
	ErrServerNotFound:               "服务器不存在",
	ErrNodeNotFound:                 "节点不存在",
	ErrCouponNotFound:               "优惠券不存在",
	ErrDeviceNotFound:               "设备不存在",
	ErrAuthMethodNotFound:           "认证方法不存在",
	ErrAnnouncementNotFound:         "公告不存在",
	ErrDocumentNotFound:             "文档不存在",
	ErrAdsNotFound:                  "广告不存在",
	ErrSystemNotFound:               "系统配置不存在",
	ErrSubscribeApplicationNotFound: "订阅申请不存在",
	ErrServerGroupNotFound:          "服务器组不存在",
	ErrTicketNotFound:               "工单不存在",
	ErrSystemLogNotFound:            "系统日志不存在",
	ErrTaskNotFound:                 "任务不存在",
	ErrRedemptionCodeNotFound:       "兑换码不存在",
	ErrUserGroupNotFound:            "用户组不存在",
	ErrInvalidTaskType:              "无效的任务类型",
	ErrInvalidTaskStatus:            "无效的任务状态",
	ErrTaskCannotBeStopped:          "任务无法停止（只有进行中的任务可以停止）",

	// 业务错误码消息 - 数据冲突
	ErrUserAlreadyExists:                 "用户已存在",
	ErrOrderAlreadyExists:                "订单已存在",
	ErrSubscribeAlreadyExists:            "订阅已存在",
	ErrPaymentAlreadyExists:              "支付方式已存在",
	ErrServerAlreadyExists:               "服务器已存在",
	ErrNodeAlreadyExists:                 "节点已存在",
	ErrCouponAlreadyExists:               "优惠券已存在",
	ErrDuplicateEmail:                    "邮箱已存在",
	ErrDuplicateUsername:                 "用户名已存在",
	ErrAnnouncementAlreadyExists:         "公告已存在",
	ErrDocumentAlreadyExists:             "文档已存在",
	ErrSystemAlreadyExists:               "系统配置已存在",
	ErrAuthMethodAlreadyExists:           "认证方法已存在",
	ErrSubscribeApplicationAlreadyExists: "订阅申请已存在",

	// 业务错误码消息 - 业务逻辑
	ErrOrderCannotCancel:       "订单不能取消",
	ErrOrderCannotComplete:     "订单不能完成",
	ErrOrderCannotClose:        "订单不能关闭",
	ErrCouponExpired:           "优惠券已过期",
	ErrCouponNotAvailable:      "优惠券不可用",
	ErrCouponUsedUp:            "优惠券已用完",
	ErrCouponUserLimitExceeded: "优惠券用户使用次数超限",
	ErrInsufficientBalance:     "余额不足",
	ErrUserCommissionNotEnough: "佣金不足",
	ErrOrderPaymentFailed:      "订单支付失败",
	ErrDeviceLimitExceeded:     "设备数量超限",
	ErrSubscribeExpired:        "订阅已过期",
	ErrTrafficExceeded:         "流量超限",
	ErrSubscribeInUse:          "订阅套餐正在使用中",
	ErrInvalidOrderStatus:      "无效的订单状态",
	ErrInvalidParameter:        "无效的参数",
	ErrTitleRequired:           "标题不能为空",
	ErrTypeRequired:            "类型不能为空",
	ErrInvalidTimeRange:        "时间范围无效",
	ErrInvalidTicketStatus:     "无效的工单状态",
	ErrInvalidTicketPriority:   "无效的工单优先级",
	ErrUnsupportedPlatform:     "不支持的支付平台",

	// 权限错误码消息 - 认证
	ErrMissingAuthToken:     "缺少认证令牌",
	ErrInvalidAuthToken:     "无效的认证令牌",
	ErrAuthTokenExpired:     "认证令牌已过期",
	ErrInvalidCredentials:   "无效的凭证",
	ErrUserNotAuthenticated: "用户未认证",
	ErrPasswordIncorrect:    "密码错误",
	ErrAccountLocked:        "账户已锁定",
	ErrAccountDisabled:      "账户已禁用",

	// 权限错误码消息 - 授权
	ErrPermissionDenied:       "权限被拒绝",
	ErrInsufficientPermission: "权限不足",
	ErrResourceAccessDenied:   "资源访问被拒绝",
	ErrOperationNotAllowed:    "操作不被允许",
	ErrTenantAccessDenied:     "租户访问被拒绝",
	ErrCrossTenantOperation:   "跨租户操作",
	ErrNotResourceOwner:       "非资源所有者",
	ErrInvalidAccess:          "无效访问",

	// 系统错误码消息 - 数据库
	ErrDatabaseConnection:  "数据库连接失败",
	ErrDatabaseQuery:       "数据库查询失败",
	ErrDatabaseUpdate:      "数据库更新失败",
	ErrDatabaseInsert:      "数据库插入失败",
	ErrDatabaseDelete:      "数据库删除失败",
	ErrDatabaseTransaction: "数据库事务失败",
	ErrDatabaseConstraint:  "数据库约束错误",
	ErrDatabaseTimeout:     "数据库超时",

	// 系统错误码消息 - 缓存
	ErrCacheConnection:  "缓存连接失败",
	ErrCacheGet:         "缓存获取失败",
	ErrCacheSet:         "缓存设置失败",
	ErrCacheDelete:      "缓存删除失败",
	ErrCacheExpired:     "缓存已过期",
	ErrCacheSerialize:   "缓存序列化失败",
	ErrCacheDeserialize: "缓存反序列化失败",

	// 系统错误码消息 - 内部服务
	ErrInternalError:        "内部错误",
	ErrServiceUnavailable:   "服务不可用",
	ErrServiceTimeout:       "服务超时",
	ErrConfigurationError:   "配置错误",
	ErrInitializationFailed: "初始化失败",
	ErrResourceExhausted:    "资源耗尽",

	// 系统错误码消息 - 外部服务
	ErrIPGeolocationFailed: "IP地理位置查询失败",
	ErrPaymentGatewayError: "支付网关错误",
	ErrEmailSendFailed:     "邮件发送失败",
	ErrSMSSendFailed:       "短信发送失败",
	ErrThirdPartyAPIError:  "第三方API错误",
}

// getCodeMessage 获取响应码对应的消息
func getCodeMessage(code int) string {
	if message, exists := CodeMessages[code]; exists {
		return message
	}
	return "未知响应码"
}
