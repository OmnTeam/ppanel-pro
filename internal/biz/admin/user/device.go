package user

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/npanel-pro/api/admin/user/v1"
)

// DeviceRepo 设备仓储接口
type DeviceRepo interface {
	// UpdateUserDevice 更新用户设备
	UpdateUserDevice(ctx context.Context, req *v1.UpdateUserDeviceRequest) error

	// DeleteUserDevice 删除用户设备
	DeleteUserDevice(ctx context.Context, deviceID int64) error

	// KickOfflineByUserDevice 踢下线用户设备
	KickOfflineByUserDevice(ctx context.Context, deviceID int64) error
}

// DeviceUsecase 设备用例
type DeviceUsecase struct {
	repo   DeviceRepo
	logger *log.Helper
}

// NewDeviceUsecase 创建设备用例
func NewDeviceUsecase(repo DeviceRepo, logger log.Logger) *DeviceUsecase {
	return &DeviceUsecase{
		repo:   repo,
		logger: log.NewHelper(logger),
	}
}

// UpdateUserDevice 更新用户设备
func (uc *DeviceUsecase) UpdateUserDevice(ctx context.Context, req *v1.UpdateUserDeviceRequest) error {
	return uc.repo.UpdateUserDevice(ctx, req)
}

// DeleteUserDevice 删除用户设备
func (uc *DeviceUsecase) DeleteUserDevice(ctx context.Context, deviceID int64) error {
	return uc.repo.DeleteUserDevice(ctx, deviceID)
}

// KickOfflineByUserDevice 踢下线用户设备
func (uc *DeviceUsecase) KickOfflineByUserDevice(ctx context.Context, deviceID int64) error {
	return uc.repo.KickOfflineByUserDevice(ctx, deviceID)
}
