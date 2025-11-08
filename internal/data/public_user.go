package data

import (
	"context"
	"fmt"

	userbiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/user"
	"github.com/go-kratos/kratos/v2/log"
	kratoslog "github.com/go-kratos/kratos/v2/log"
)

type userRepo struct {
	data *Data
	log  *log.Helper
}

// NewPublicUserRepo 创建Public User数据仓储实例
func NewPublicUserRepo(d *Data, logger kratoslog.Logger) userbiz.UserRepo {
	return &userRepo{
		data: d,
		log:  log.NewHelper(logger),
	}
}

// QueryUserInfo 查询用户信息
func (r *userRepo) QueryUserInfo(ctx context.Context, userID int64) (*userbiz.UserInfo, error) {
	return &userbiz.UserInfo{}, nil
}

// GetLoginLog 获取登录日志
func (r *userRepo) GetLoginLog(ctx context.Context, userID int64, page, size int64) ([]*userbiz.LoginLog, int64, error) {
	return []*userbiz.LoginLog{}, 0, nil
}

// QueryUserBalanceLog 查询用户余额日志
func (r *userRepo) QueryUserBalanceLog(ctx context.Context, userID int64) ([]*userbiz.BalanceLog, int64, error) {
	return []*userbiz.BalanceLog{}, 0, nil
}

// QueryUserCommissionLog 查询用户佣金日志
func (r *userRepo) QueryUserCommissionLog(ctx context.Context, userID int64, page, size int64) ([]*userbiz.CommissionLog, int64, error) {
	return []*userbiz.CommissionLog{}, 0, nil
}

// QueryUserAffiliate 查询用户推荐数量
func (r *userRepo) QueryUserAffiliate(ctx context.Context, userID int64) (int64, int64, error) {
	return 0, 0, nil
}

// QueryUserAffiliateList 查询用户推荐列表
func (r *userRepo) QueryUserAffiliateList(ctx context.Context, userID int64, page, size int64) ([]*userbiz.UserAffiliate, int64, error) {
	return []*userbiz.UserAffiliate{}, 0, nil
}

// GetOAuthMethods 获取OAuth方法
func (r *userRepo) GetOAuthMethods(ctx context.Context, userID int64) ([]*userbiz.AuthMethod, error) {
	return []*userbiz.AuthMethod{}, nil
}

// QueryUserSubscribe 查询用户订阅
func (r *userRepo) QueryUserSubscribe(ctx context.Context, userID int64) ([]*userbiz.UserSubscribe, int64, error) {
	return []*userbiz.UserSubscribe{}, 0, nil
}

// GetSubscribeLog 获取订阅日志
func (r *userRepo) GetSubscribeLog(ctx context.Context, userID int64, page, size int64) ([]*userbiz.UserSubscribeLog, int64, error) {
	return []*userbiz.UserSubscribeLog{}, 0, nil
}

// ResetUserSubscribeToken 重置订阅令牌
func (r *userRepo) ResetUserSubscribeToken(ctx context.Context, userID, userSubscribeID int64) error {
	return nil
}

// PreUnsubscribe 预退订
func (r *userRepo) PreUnsubscribe(ctx context.Context, userID, id int64) (int64, error) {
	return 0, nil
}

// Unsubscribe 退订
func (r *userRepo) Unsubscribe(ctx context.Context, userID, id int64) error {
	return nil
}

// UpdateUserNotify 更新用户通知设置
func (r *userRepo) UpdateUserNotify(ctx context.Context, userID int64, enableLoginNotify, enableBalanceNotify, enableSubscribeNotify, enableTradeNotify bool) error {
	return nil
}

// UpdateUserPassword 更新用户密码
func (r *userRepo) UpdateUserPassword(ctx context.Context, userID int64, password string) error {
	return nil
}

// BindTelegram 绑定Telegram
func (r *userRepo) BindTelegram(ctx context.Context, session string, botName string) (string, int64, error) {
	return "", 0, nil
}

// UnbindTelegram 解绑Telegram
func (r *userRepo) UnbindTelegram(ctx context.Context, userID int64) error {
	return nil
}

// BindOAuth 绑定OAuth
func (r *userRepo) BindOAuth(ctx context.Context, method, redirect string) (string, error) {
	return "", nil
}

// BindOAuthCallback OAuth回调
func (r *userRepo) BindOAuthCallback(ctx context.Context, userID int64, method string, callback string) error {
	return nil
}

// UnbindOAuth 解绑OAuth
func (r *userRepo) UnbindOAuth(ctx context.Context, userID int64, method string) error {
	return nil
}

// VerifyEmail 验证邮箱
func (r *userRepo) VerifyEmail(ctx context.Context, userID int64, email, code string) error {
	return nil
}

// UpdateBindMobile 更新绑定手机
func (r *userRepo) UpdateBindMobile(ctx context.Context, userID int64, areaCode, mobile, code string) error {
	return nil
}

// UpdateBindEmail 更新绑定邮箱
func (r *userRepo) UpdateBindEmail(ctx context.Context, userID int64, email string) error {
	return nil
}

// DeviceWSConnect 设备WebSocket连接
func (r *userRepo) DeviceWSConnect(ctx context.Context) error {
	// 这个方法实际上在数据层不需要做太多工作
	// 主要的WebSocket连接逻辑在服务层和设备管理器中处理
	// 这里只是确保设备管理器可用
	if r.data.DeviceManager() == nil {
		return fmt.Errorf("device manager is not available")
	}

	r.log.Infof("DeviceWSConnect called - device manager is available")
	return nil
}

// GetDeviceList 获取设备列表 - 完整实现
func (r *userRepo) GetDeviceList(ctx context.Context, userID int64) ([]*userbiz.UserDevice, int64, error) {
	if r.data.DeviceManager() == nil {
		r.log.Errorf("DeviceManager is nil")
		return []*userbiz.UserDevice{}, 0, nil
	}

	devices, err := r.data.DeviceManager().GetUserDevices(userID)
	if err != nil {
		r.log.Errorf("Failed to get devices for user %d: %v", userID, err)
		return nil, 0, err
	}

	// 转换为业务层对象
	userDevices := make([]*userbiz.UserDevice, 0, len(devices))
	for _, device := range devices {
		userDevices = append(userDevices, &userbiz.UserDevice{
			ID:         device.ID,
			IP:         device.IP,
			Identifier: device.Identifier,
			UserAgent:  device.UserAgent,
			Online:     device.Online,
			Enabled:    device.Enabled,
			CreatedAt:  device.CreatedAt,
			UpdatedAt:  device.UpdatedAt,
		})
	}

	return userDevices, int64(len(userDevices)), nil
}

// UnbindDevice 解绑设备 - 完整实现
func (r *userRepo) UnbindDevice(ctx context.Context, userID, deviceID int64) error {
	if r.data.DeviceManager() == nil {
		r.log.Errorf("DeviceManager is nil")
		return fmt.Errorf("device manager not available")
	}

	err := r.data.DeviceManager().RemoveDevice(userID, deviceID)
	if err != nil {
		r.log.Errorf("Failed to remove device %d for user %d: %v", deviceID, userID, err)
		return err
	}

	r.log.Infof("Successfully unbound device %d for user %d", deviceID, userID)
	return nil
}

// GetDeviceOnlineStatistics 获取设备在线统计 - 完整实现
func (r *userRepo) GetDeviceOnlineStatistics(ctx context.Context, userID int64) (*userbiz.DeviceOnlineStatistics, error) {
	if r.data.DeviceManager() == nil {
		r.log.Errorf("DeviceManager is nil")
		return &userbiz.DeviceOnlineStatistics{}, nil
	}

	stats, err := r.data.DeviceManager().GetUserDeviceStatistics(userID)
	if err != nil {
		r.log.Errorf("Failed to get device statistics for user %d: %v", userID, err)
		return nil, err
	}

	// 转换每周统计
	weeklyStats := make([]*userbiz.WeeklyStat, 0, len(stats.WeeklyStats))
	for _, stat := range stats.WeeklyStats {
		weeklyStats = append(weeklyStats, &userbiz.WeeklyStat{
			Day:     stat.Day,
			DayName: stat.DayName,
			Hours:   stat.Hours,
		})
	}

	// 转换连接记录
	connectionRecords := &userbiz.ConnectionRecords{
		CurrentContinuousDays:   stats.ConnectionRecords.CurrentContinuousDays,
		HistoryContinuousDays:   stats.ConnectionRecords.HistoryContinuousDays,
		LongestSingleConnection: stats.ConnectionRecords.LongestSingleConnection,
	}

	return &userbiz.DeviceOnlineStatistics{
		WeeklyStats:       weeklyStats,
		ConnectionRecords: connectionRecords,
	}, nil
}