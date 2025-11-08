package user

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/user/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
)

// SubscribeRepo 订阅仓储接口
type SubscribeRepo interface {
	// GetUserSubscribe 获取用户订阅列表
	GetUserSubscribe(ctx context.Context, req *v1.GetUserSubscribeRequest) ([]*ent.ProxyUserSubscribe, int64, error)

	// CreateUserSubscribe 创建用户订阅
	CreateUserSubscribe(ctx context.Context, req *v1.CreateUserSubscribeRequest) (int64, error)

	// UpdateUserSubscribe 更新用户订阅
	UpdateUserSubscribe(ctx context.Context, req *v1.UpdateUserSubscribeRequest) error

	// DeleteUserSubscribe 删除用户订阅
	DeleteUserSubscribe(ctx context.Context, userSubscribeID int64) error

	// GetUserSubscribeById 根据ID获取用户订阅详情（包含套餐信息）
	GetUserSubscribeById(ctx context.Context, userSubscribeID int64) (*v1.UserSubscribe, string, error)

	// GetUserSubscribeDevices 获取用户订阅设备列表
	GetUserSubscribeDevices(ctx context.Context, req *v1.GetUserSubscribeDevicesRequest) ([]*ent.ProxyUserDevice, int64, error)

	// GetUserSubscribeLogs 获取用户订阅日志
	GetUserSubscribeLogs(ctx context.Context, req *v1.GetUserSubscribeLogsRequest) ([]*ent.ProxySystemLog, int64, error)

	// GetUserSubscribeResetTrafficLogs 获取用户订阅重置流量日志
	GetUserSubscribeResetTrafficLogs(ctx context.Context, req *v1.GetUserSubscribeResetTrafficLogsRequest) ([]*ent.ProxySystemLog, int64, error)

	// GetUserSubscribeTrafficLogs 获取用户订阅流量日志
	GetUserSubscribeTrafficLogs(ctx context.Context, req *v1.GetUserSubscribeTrafficLogsRequest) ([]*ent.ProxyTrafficLog, int64, error)
}

// SubscribeUsecase 订阅用例
type SubscribeUsecase struct {
	repo   SubscribeRepo
	logger *log.Helper
}

// NewSubscribeUsecase 创建订阅用例
func NewSubscribeUsecase(repo SubscribeRepo, logger log.Logger) *SubscribeUsecase {
	return &SubscribeUsecase{
		repo:   repo,
		logger: log.NewHelper(logger),
	}
}

// GetUserSubscribe 获取用户订阅列表
func (uc *SubscribeUsecase) GetUserSubscribe(ctx context.Context, req *v1.GetUserSubscribeRequest) ([]*ent.ProxyUserSubscribe, int64, error) {
	return uc.repo.GetUserSubscribe(ctx, req)
}

// CreateUserSubscribe 创建用户订阅
func (uc *SubscribeUsecase) CreateUserSubscribe(ctx context.Context, req *v1.CreateUserSubscribeRequest) (int64, error) {
	return uc.repo.CreateUserSubscribe(ctx, req)
}

// UpdateUserSubscribe 更新用户订阅
func (uc *SubscribeUsecase) UpdateUserSubscribe(ctx context.Context, req *v1.UpdateUserSubscribeRequest) error {
	return uc.repo.UpdateUserSubscribe(ctx, req)
}

// DeleteUserSubscribe 删除用户订阅
func (uc *SubscribeUsecase) DeleteUserSubscribe(ctx context.Context, userSubscribeID int64) error {
	return uc.repo.DeleteUserSubscribe(ctx, userSubscribeID)
}

// GetUserSubscribeById 根据ID获取用户订阅详情
func (uc *SubscribeUsecase) GetUserSubscribeById(ctx context.Context, userSubscribeID int64) (*v1.UserSubscribe, string, error) {
	return uc.repo.GetUserSubscribeById(ctx, userSubscribeID)
}

// GetUserSubscribeDevices 获取用户订阅设备列表
func (uc *SubscribeUsecase) GetUserSubscribeDevices(ctx context.Context, req *v1.GetUserSubscribeDevicesRequest) ([]*ent.ProxyUserDevice, int64, error) {
	return uc.repo.GetUserSubscribeDevices(ctx, req)
}

// GetUserSubscribeLogs 获取用户订阅日志
func (uc *SubscribeUsecase) GetUserSubscribeLogs(ctx context.Context, req *v1.GetUserSubscribeLogsRequest) ([]*ent.ProxySystemLog, int64, error) {
	return uc.repo.GetUserSubscribeLogs(ctx, req)
}

// GetUserSubscribeResetTrafficLogs 获取用户订阅重置流量日志
func (uc *SubscribeUsecase) GetUserSubscribeResetTrafficLogs(ctx context.Context, req *v1.GetUserSubscribeResetTrafficLogsRequest) ([]*ent.ProxySystemLog, int64, error) {
	return uc.repo.GetUserSubscribeResetTrafficLogs(ctx, req)
}

// GetUserSubscribeTrafficLogs 获取用户订阅流量日志
func (uc *SubscribeUsecase) GetUserSubscribeTrafficLogs(ctx context.Context, req *v1.GetUserSubscribeTrafficLogsRequest) ([]*ent.ProxyTrafficLog, int64, error) {
	return uc.repo.GetUserSubscribeTrafficLogs(ctx, req)
}
