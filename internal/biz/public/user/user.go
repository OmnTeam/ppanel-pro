package user

import (
	"context"
)

// UserRepo Public User数据仓库接口
type UserRepo interface {
	// QueryUserInfo 查询用户信息
	// 从context获取当前用户
	// 返回用户信息，AuthIdentifier需要脱敏
	// AuthMethods需要按照email、mobile优先级排序
	QueryUserInfo(ctx context.Context, userID int64) (*UserInfo, error)

	// GetLoginLog 获取登录日志
	// 从proxy_system_log表查询type=6(Login)的日志
	GetLoginLog(ctx context.Context, userID int64, page, size int64) ([]*LoginLog, int64, error)

	// QueryUserBalanceLog 查询用户余额日志
	// 从proxy_system_log表查询type=8(Balance)的日志
	// 固定查询全部，page=1, size=99999
	QueryUserBalanceLog(ctx context.Context, userID int64) ([]*BalanceLog, int64, error)

	// QueryUserCommissionLog 查询用户佣金日志
	// 从proxy_system_log表查询type=9(Commission)的日志
	QueryUserCommissionLog(ctx context.Context, userID int64, page, size int64) ([]*CommissionLog, int64, error)

	// QueryUserAffiliate 查询用户推荐数量
	// 查询referer_id=当前用户的用户数量
	// 查询commission日志求和
	QueryUserAffiliate(ctx context.Context, userID int64) (registers int64, totalCommission int64, err error)

	// QueryUserAffiliateList 查询用户推荐列表
	// 查询referer_id=当前用户的用户列表
	// 需要查询每个用户的AuthMethod并脱敏
	QueryUserAffiliateList(ctx context.Context, userID int64, page, size int64) ([]*UserAffiliate, int64, error)

	// GetOAuthMethods 获取OAuth方法
	// 查询当前用户的所有AuthMethods
	GetOAuthMethods(ctx context.Context, userID int64) ([]*AuthMethod, error)

	// QueryUserSubscribe 查询用户订阅
	// 查询status in (0,1,2,3)的订阅
	// 需要解析Discount字段并计算ResetTime
	QueryUserSubscribe(ctx context.Context, userID int64) ([]*UserSubscribe, int64, error)

	// GetSubscribeLog 获取订阅日志
	// 从proxy_system_log查询type=20(Subscribe)的日志
	GetSubscribeLog(ctx context.Context, userID int64, page, size int64) ([]*UserSubscribeLog, int64, error)

	// ResetUserSubscribeToken 重置订阅令牌
	// 验证UserSubscribeId属于当前用户
	// 生成新Token和UUID
	// 清除缓存
	ResetUserSubscribeToken(ctx context.Context, userID, userSubscribeID int64) error

	// PreUnsubscribe 预退订
	// 调用CalculateRemainingAmount计算退款金额
	PreUnsubscribe(ctx context.Context, userID, id int64) (deductionAmount int64, err error)

	// Unsubscribe 退订
	// 验证订阅状态 status in (0,1,2)
	// 使用事务处理退款逻辑
	// 清除缓存
	Unsubscribe(ctx context.Context, userID, id int64) error

	// UpdateUserNotify 更新通知设置
	// 更新用户的4个通知开关
	UpdateUserNotify(ctx context.Context, userID int64, enableLoginNotify, enableBalanceNotify, enableSubscribeNotify, enableTradeNotify bool) error

	// UpdateUserPassword 更新密码
	// 使用加密函数加密密码
	UpdateUserPassword(ctx context.Context, userID int64, password string) error

	// BindTelegram 绑定Telegram
	// 从JWT session获取，生成Telegram Bot URL
	// 过期时间5分钟
	BindTelegram(ctx context.Context, session string, botName string) (url string, expiredAt int64, err error)

	// UnbindTelegram 解绑Telegram
	// 查询telegram AuthMethod并删除
	// 发送Telegram通知（需要telegram bot）
	UnbindTelegram(ctx context.Context, userID int64) error

	// BindOAuth 绑定OAuth
	// 根据method（google/apple/github/facebook）生成OAuth URL
	// 生成state code并存入Redis（5分钟过期）
	BindOAuth(ctx context.Context, method, redirect string) (string, error)

	// BindOAuthCallback OAuth回调
	// 验证state code
	// 交换token获取用户信息
	// 创建AuthMethod记录
	BindOAuthCallback(ctx context.Context, userID int64, method string, callback string) error

	// UnbindOAuth 解绑OAuth
	// 验证method不能为email/mobile
	// 删除对应AuthMethod
	UnbindOAuth(ctx context.Context, userID int64, method string) error

	// VerifyEmail 验证邮箱
	// 从Redis验证code
	// 设置AuthMethod.Verified=true
	VerifyEmail(ctx context.Context, userID int64, email, code string) error

	// UpdateBindMobile 更新绑定手机
	// 验证手机验证码
	// 检查手机号是否已被其他用户绑定
	// 创建或更新mobile AuthMethod
	UpdateBindMobile(ctx context.Context, userID int64, areaCode, mobile, code string) error

	// UpdateBindEmail 更新绑定邮箱
	// 检查邮箱是否已被其他用户绑定
	// 创建或更新email AuthMethod（Verified=false）
	UpdateBindEmail(ctx context.Context, userID int64, email string) error

	// DeviceWSConnect 设备WebSocket连接
	// 升级WebSocket连接并添加设备到管理器
	DeviceWSConnect(ctx context.Context) error

	// GetDeviceList 获取设备列表
	// 查询用户的所有设备
	GetDeviceList(ctx context.Context, userID int64) ([]*UserDevice, int64, error)

	// UnbindDevice 解绑设备
	// 验证设备属于当前用户并删除设备
	UnbindDevice(ctx context.Context, userID, deviceID int64) error

	// GetDeviceOnlineStatistics 获取设备在线统计
	// 获取设备连接统计信息
	GetDeviceOnlineStatistics(ctx context.Context, userID int64) (*DeviceOnlineStatistics, error)
}

// UserInfo 用户信息
type UserInfo struct {
	ID                    int64
	Avatar                string
	Balance               int64
	Commission            int64
	ReferralPercentage    int32
	OnlyFirstPurchase     bool
	GiftAmount            int64
	Telegram              int64
	ReferCode             string
	RefererID             int64
	Enable                bool
	IsAdmin               bool
	EnableBalanceNotify   bool
	EnableLoginNotify     bool
	EnableSubscribeNotify bool
	EnableTradeNotify     bool
	AuthMethods           []*AuthMethod
	CreatedAt             int64
	UpdatedAt             int64
}

// AuthMethod 认证方法
type AuthMethod struct {
	AuthType       string
	AuthIdentifier string
	Verified       bool
}

// LoginLog 登录日志
type LoginLog struct {
	ID        int64
	UserID    int64
	LoginIP   string
	UserAgent string
	Success   bool
	Timestamp int64
}

// BalanceLog 余额日志
type BalanceLog struct {
	Type      int32
	UserID    int64
	Amount    int64
	OrderNo   string
	Balance   int64
	Timestamp int64
}

// CommissionLog 佣金日志
type CommissionLog struct {
	Type      int32
	UserID    int64
	Amount    int64
	OrderNo   string
	Timestamp int64
}

// UserAffiliate 用户推荐信息
type UserAffiliate struct {
	Avatar       string
	Identifier   string
	RegisteredAt int64
	Enable       bool
}

// UserSubscribe 用户订阅信息
type UserSubscribe struct {
	ID          int64
	UserID      int64
	OrderID     int64
	SubscribeID int64
	Subscribe   *Subscribe
	StartTime   int64
	ExpireTime  int64
	FinishedAt  int64
	ResetTime   int64
	Traffic     int64
	Download    int64
	Upload      int64
	Token       string
	Status      int32
	CreatedAt   int64
	UpdatedAt   int64
}

// Subscribe 订阅套餐信息
type Subscribe struct {
	ID             int64
	Name           string
	Description    string
	Price          int64
	Traffic        int64
	DeviceLimit    int32
	SpeedLimit     int32
	UnitTime       int32
	UnitPrice      int64
	ResetCycle     int32
	DeductionRatio int32
	AllowDeduction bool
	Discount       []*SubscribeDiscount
	Enable         bool
	CreatedAt      int64
	UpdatedAt      int64
}

// SubscribeDiscount 订阅折扣信息
type SubscribeDiscount struct {
	Quantity   int32
	Percentage int32
}

// UserSubscribeLog 用户订阅日志
type UserSubscribeLog struct {
	ID              int64
	UserID          int64
	UserSubscribeID int64
	Token           string
	IP              string
	UserAgent       string
	Timestamp       int64
}

// UserDevice 用户设备信息
type UserDevice struct {
	ID         int64
	IP         string
	Identifier string
	UserAgent  string
	Online     bool
	Enabled    bool
	CreatedAt  int64
	UpdatedAt  int64
}

// WeeklyStat 每周统计
type WeeklyStat struct {
	Day     int32
	DayName string
	Hours   float64
}

// ConnectionRecords 连接记录
type ConnectionRecords struct {
	CurrentContinuousDays   int64
	HistoryContinuousDays   int64
	LongestSingleConnection int64
}

// DeviceOnlineStatistics 设备在线统计
type DeviceOnlineStatistics struct {
	WeeklyStats      []*WeeklyStat
	ConnectionRecords *ConnectionRecords
}

// UserUseCase Public User用例
type UserUseCase struct {
	repo UserRepo
}

// NewUserUseCase 创建Public User用例
func NewUserUseCase(repo UserRepo) *UserUseCase {
	return &UserUseCase{repo: repo}
}

// QueryUserInfo 查询用户信息
func (uc *UserUseCase) QueryUserInfo(ctx context.Context, userID int64) (*UserInfo, error) {
	return uc.repo.QueryUserInfo(ctx, userID)
}

// GetLoginLog 获取登录日志
func (uc *UserUseCase) GetLoginLog(ctx context.Context, userID int64, page, size int64) ([]*LoginLog, int64, error) {
	return uc.repo.GetLoginLog(ctx, userID, page, size)
}

// QueryUserBalanceLog 查询用户余额日志
func (uc *UserUseCase) QueryUserBalanceLog(ctx context.Context, userID int64) ([]*BalanceLog, int64, error) {
	return uc.repo.QueryUserBalanceLog(ctx, userID)
}

// QueryUserCommissionLog 查询用户佣金日志
func (uc *UserUseCase) QueryUserCommissionLog(ctx context.Context, userID int64, page, size int64) ([]*CommissionLog, int64, error) {
	return uc.repo.QueryUserCommissionLog(ctx, userID, page, size)
}

// QueryUserAffiliate 查询用户推荐数量
func (uc *UserUseCase) QueryUserAffiliate(ctx context.Context, userID int64) (registers int64, totalCommission int64, err error) {
	return uc.repo.QueryUserAffiliate(ctx, userID)
}

// QueryUserAffiliateList 查询用户推荐列表
func (uc *UserUseCase) QueryUserAffiliateList(ctx context.Context, userID int64, page, size int64) ([]*UserAffiliate, int64, error) {
	return uc.repo.QueryUserAffiliateList(ctx, userID, page, size)
}

// GetOAuthMethods 获取OAuth方法
func (uc *UserUseCase) GetOAuthMethods(ctx context.Context, userID int64) ([]*AuthMethod, error) {
	return uc.repo.GetOAuthMethods(ctx, userID)
}

// QueryUserSubscribe 查询用户订阅
func (uc *UserUseCase) QueryUserSubscribe(ctx context.Context, userID int64) ([]*UserSubscribe, int64, error) {
	return uc.repo.QueryUserSubscribe(ctx, userID)
}

// GetSubscribeLog 获取订阅日志
func (uc *UserUseCase) GetSubscribeLog(ctx context.Context, userID int64, page, size int64) ([]*UserSubscribeLog, int64, error) {
	return uc.repo.GetSubscribeLog(ctx, userID, page, size)
}

// ResetUserSubscribeToken 重置订阅令牌
func (uc *UserUseCase) ResetUserSubscribeToken(ctx context.Context, userID, userSubscribeID int64) error {
	return uc.repo.ResetUserSubscribeToken(ctx, userID, userSubscribeID)
}

// PreUnsubscribe 预退订
func (uc *UserUseCase) PreUnsubscribe(ctx context.Context, userID, id int64) (int64, error) {
	return uc.repo.PreUnsubscribe(ctx, userID, id)
}

// Unsubscribe 退订
func (uc *UserUseCase) Unsubscribe(ctx context.Context, userID, id int64) error {
	return uc.repo.Unsubscribe(ctx, userID, id)
}

// UpdateUserNotify 更新通知设置
func (uc *UserUseCase) UpdateUserNotify(ctx context.Context, userID int64, enableLoginNotify, enableBalanceNotify, enableSubscribeNotify, enableTradeNotify bool) error {
	return uc.repo.UpdateUserNotify(ctx, userID, enableLoginNotify, enableBalanceNotify, enableSubscribeNotify, enableTradeNotify)
}

// UpdateUserPassword 更新密码
func (uc *UserUseCase) UpdateUserPassword(ctx context.Context, userID int64, password string) error {
	return uc.repo.UpdateUserPassword(ctx, userID, password)
}

// BindTelegram 绑定Telegram
func (uc *UserUseCase) BindTelegram(ctx context.Context, session string, botName string) (string, int64, error) {
	return uc.repo.BindTelegram(ctx, session, botName)
}

// UnbindTelegram 解绑Telegram
func (uc *UserUseCase) UnbindTelegram(ctx context.Context, userID int64) error {
	return uc.repo.UnbindTelegram(ctx, userID)
}

// BindOAuth 绑定OAuth
func (uc *UserUseCase) BindOAuth(ctx context.Context, method, redirect string) (string, error) {
	return uc.repo.BindOAuth(ctx, method, redirect)
}

// BindOAuthCallback OAuth回调
func (uc *UserUseCase) BindOAuthCallback(ctx context.Context, userID int64, method string, callback string) error {
	return uc.repo.BindOAuthCallback(ctx, userID, method, callback)
}

// UnbindOAuth 解绑OAuth
func (uc *UserUseCase) UnbindOAuth(ctx context.Context, userID int64, method string) error {
	return uc.repo.UnbindOAuth(ctx, userID, method)
}

// VerifyEmail 验证邮箱
func (uc *UserUseCase) VerifyEmail(ctx context.Context, userID int64, email, code string) error {
	return uc.repo.VerifyEmail(ctx, userID, email, code)
}

// UpdateBindMobile 更新绑定手机
func (uc *UserUseCase) UpdateBindMobile(ctx context.Context, userID int64, areaCode, mobile, code string) error {
	return uc.repo.UpdateBindMobile(ctx, userID, areaCode, mobile, code)
}

// UpdateBindEmail 更新绑定邮箱
func (uc *UserUseCase) UpdateBindEmail(ctx context.Context, userID int64, email string) error {
	return uc.repo.UpdateBindEmail(ctx, userID, email)
}

// DeviceWSConnect 设备WebSocket连接
func (uc *UserUseCase) DeviceWSConnect(ctx context.Context) error {
	return uc.repo.DeviceWSConnect(ctx)
}

// GetDeviceList 获取设备列表
func (uc *UserUseCase) GetDeviceList(ctx context.Context, userID int64) ([]*UserDevice, int64, error) {
	return uc.repo.GetDeviceList(ctx, userID)
}

// UnbindDevice 解绑设备
func (uc *UserUseCase) UnbindDevice(ctx context.Context, userID, deviceID int64) error {
	return uc.repo.UnbindDevice(ctx, userID, deviceID)
}

// GetDeviceOnlineStatistics 获取设备在线统计
func (uc *UserUseCase) GetDeviceOnlineStatistics(ctx context.Context, userID int64) (*DeviceOnlineStatistics, error) {
	return uc.repo.GetDeviceOnlineStatistics(ctx, userID)
}
