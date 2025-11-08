package log

import (
	"context"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/go-kratos/kratos/v2/log"
)

// SystemLogRepo 系统日志仓库接口
type SystemLogRepo interface {
	// FilterSystemLog 过滤系统日志（通用方法）
	FilterSystemLog(ctx context.Context, params *FilterParams) ([]*ent.ProxySystemLog, int64, error)
}

// FilterParams 过滤参数
type FilterParams struct {
	Page     int32
	Size     int32
	Type     int8   // 日志类型
	Date     string // 日期过滤
	ObjectID *int64 // 对象ID（用户ID、服务器ID等）
	Search   string // 搜索关键词（在content中搜索）
}

// SystemLogUsecase 系统日志用例
type SystemLogUsecase struct {
	repo SystemLogRepo
	log  *log.Helper
}

// NewSystemLogUsecase 创建系统日志用例
func NewSystemLogUsecase(repo SystemLogRepo, logger log.Logger) *SystemLogUsecase {
	return &SystemLogUsecase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "biz/admin/log/system")),
	}
}

// 日志类型常量（对应 proxy_system_log 表的 type 字段）
// 这些值必须与原项目保持一致，以便能正确查询已有的日志数据
const (
	LogTypeEmail            int8 = 10 // Email Message
	LogTypeMobile           int8 = 11 // Mobile Message
	LogTypeSubscribe        int8 = 20 // Subscribe
	LogTypeSubscribeTraffic int8 = 21 // Subscribe Traffic
	LogTypeServerTraffic    int8 = 22 // Server Traffic
	LogTypeResetSubscribe   int8 = 23 // Reset Subscribe
	LogTypeLogin            int8 = 30 // Login
	LogTypeRegister         int8 = 31 // Register
	LogTypeBalance          int8 = 32 // Balance
	LogTypeCommission       int8 = 33 // Commission
	LogTypeGift             int8 = 34 // Gift
)

// FilterBalanceLog 过滤余额日志
func (uc *SystemLogUsecase) FilterBalanceLog(ctx context.Context, page, size int32, date string, userID *int64) ([]*ent.ProxySystemLog, int64, error) {
	return uc.repo.FilterSystemLog(ctx, &FilterParams{
		Page:     page,
		Size:     size,
		Type:     LogTypeBalance,
		Date:     date,
		ObjectID: userID,
	})
}

// FilterCommissionLog 过滤佣金日志
func (uc *SystemLogUsecase) FilterCommissionLog(ctx context.Context, page, size int32, date string, userID *int64) ([]*ent.ProxySystemLog, int64, error) {
	return uc.repo.FilterSystemLog(ctx, &FilterParams{
		Page:     page,
		Size:     size,
		Type:     LogTypeCommission,
		Date:     date,
		ObjectID: userID,
	})
}

// FilterEmailLog 过滤邮件日志
func (uc *SystemLogUsecase) FilterEmailLog(ctx context.Context, page, size int32, date, search string) ([]*ent.ProxySystemLog, int64, error) {
	return uc.repo.FilterSystemLog(ctx, &FilterParams{
		Page:     page,
		Size:     size,
		Type:     LogTypeEmail,
		Date:     date,
		Search:   search,
	})
}

// FilterGiftLog 过滤赠送日志
func (uc *SystemLogUsecase) FilterGiftLog(ctx context.Context, page, size int32, date string, userID *int64) ([]*ent.ProxySystemLog, int64, error) {
	return uc.repo.FilterSystemLog(ctx, &FilterParams{
		Page:     page,
		Size:     size,
		Type:     LogTypeGift,
		Date:     date,
		ObjectID: userID,
	})
}

// FilterLoginLog 过滤登录日志
func (uc *SystemLogUsecase) FilterLoginLog(ctx context.Context, page, size int32, date string, userID *int64) ([]*ent.ProxySystemLog, int64, error) {
	return uc.repo.FilterSystemLog(ctx, &FilterParams{
		Page:     page,
		Size:     size,
		Type:     LogTypeLogin,
		Date:     date,
		ObjectID: userID,
	})
}

// GetMessageLogList 获取消息日志列表
func (uc *SystemLogUsecase) GetMessageLogList(ctx context.Context, page, size int32, logType int32, search string) ([]*ent.ProxySystemLog, int64, error) {
	// logType: 1=Email, 2=Mobile
	var messageType int8
	if logType == 1 {
		messageType = LogTypeEmail
	} else if logType == 2 {
		messageType = LogTypeMobile
	} else {
		messageType = LogTypeEmail // 默认Email
	}

	return uc.repo.FilterSystemLog(ctx, &FilterParams{
		Page:     page,
		Size:     size,
		Type:     messageType,
		Search:   search,
	})
}

// FilterMobileLog 过滤手机日志
func (uc *SystemLogUsecase) FilterMobileLog(ctx context.Context, page, size int32, date, search string) ([]*ent.ProxySystemLog, int64, error) {
	return uc.repo.FilterSystemLog(ctx, &FilterParams{
		Page:     page,
		Size:     size,
		Type:     LogTypeMobile,
		Date:     date,
		Search:   search,
	})
}

// FilterRegisterLog 过滤注册日志
func (uc *SystemLogUsecase) FilterRegisterLog(ctx context.Context, page, size int32, date string, userID *int64) ([]*ent.ProxySystemLog, int64, error) {
	return uc.repo.FilterSystemLog(ctx, &FilterParams{
		Page:     page,
		Size:     size,
		Type:     LogTypeRegister,
		Date:     date,
		ObjectID: userID,
	})
}

// FilterServerTrafficLog 过滤服务器流量日志
func (uc *SystemLogUsecase) FilterServerTrafficLog(ctx context.Context, page, size int32, date string, serverID *int64) ([]*ent.ProxySystemLog, int64, error) {
	return uc.repo.FilterSystemLog(ctx, &FilterParams{
		Page:     page,
		Size:     size,
		Type:     LogTypeServerTraffic,
		Date:     date,
		ObjectID: serverID,
	})
}

// FilterSubscribeLog 过滤订阅日志
func (uc *SystemLogUsecase) FilterSubscribeLog(ctx context.Context, page, size int32, date string, userID *int64) ([]*ent.ProxySystemLog, int64, error) {
	return uc.repo.FilterSystemLog(ctx, &FilterParams{
		Page:     page,
		Size:     size,
		Type:     LogTypeSubscribe,
		Date:     date,
		ObjectID: userID,
	})
}

// FilterResetSubscribeLog 过滤重置订阅日志
func (uc *SystemLogUsecase) FilterResetSubscribeLog(ctx context.Context, page, size int32, date string, userID *int64) ([]*ent.ProxySystemLog, int64, error) {
	return uc.repo.FilterSystemLog(ctx, &FilterParams{
		Page:     page,
		Size:     size,
		Type:     LogTypeResetSubscribe,
		Date:     date,
		ObjectID: userID,
	})
}

// FilterUserSubscribeTrafficLog 过滤用户订阅流量日志
func (uc *SystemLogUsecase) FilterUserSubscribeTrafficLog(ctx context.Context, page, size int32, date string, userID, subscribeID *int64) ([]*ent.ProxySystemLog, int64, error) {
	params := &FilterParams{
		Page:     page,
		Size:     size,
		Type:     LogTypeSubscribeTraffic,
		Date:     date,
		ObjectID: userID,
	}

	// 如果指定了subscribeID,在content中搜索
	if subscribeID != nil && *subscribeID > 0 {
		params.Search = `"subscribe_id":`
	}

	return uc.repo.FilterSystemLog(ctx, params)
}
