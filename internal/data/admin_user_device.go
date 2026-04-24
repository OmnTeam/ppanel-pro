package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/user/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserdevice"
	userbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/user"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

type adminUserDeviceRepo struct {
	data   *Data
	logger *log.Helper
}

// NewAdminUserDeviceRepo creates a new admin user device repository
func NewAdminUserDeviceRepo(d *Data, logger log.Logger) userbiz.DeviceRepo {
	return &adminUserDeviceRepo{
		data:   d,
		logger: log.NewHelper(logger),
	}
}

// UpdateUserDevice 更新用户设备
func (r *adminUserDeviceRepo) UpdateUserDevice(ctx context.Context, req *v1.UpdateUserDeviceRequest) error {
	deviceID := req.Id
	if deviceID <= 0 {
		return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	// 查找设备
	device, err := r.data.db.ProxyUserDevice.Query().
		Where(
			proxyuserdevice.IDEQ(deviceID),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrDeviceNotFound)
		}
		r.logger.Errorf("Failed to query device: %v", err)
		return err
	}

	// 根据原项目逻辑，UpdateUserDevice只更新Enabled字段
	err = device.Update().
		SetEnabled(req.Enabled).
		Exec(ctx)

	if err != nil {
		r.logger.Errorf("Failed to update device: %v", err)
		return err
	}

	return nil
}

// DeleteUserDevice 删除用户设备
func (r *adminUserDeviceRepo) DeleteUserDevice(ctx context.Context, deviceID int64) error {
	deletedCount, err := r.data.db.ProxyUserDevice.Delete().
		Where(
			proxyuserdevice.IDEQ(deviceID),
		).
		Exec(ctx)

	if err != nil {
		r.logger.Errorf("Failed to delete device: %v", err)
		return err
	}

	if deletedCount == 0 {
		return nil
	}

	return nil
}

// KickOfflineByUserDevice 踢下线用户设备
func (r *adminUserDeviceRepo) KickOfflineByUserDevice(ctx context.Context, deviceID int64) error {
	// 查找设备
	device, err := r.data.db.ProxyUserDevice.Query().
		Where(
			proxyuserdevice.IDEQ(deviceID),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrDeviceNotFound)
		}
		r.logger.Errorf("Failed to query device: %v", err)
		return err
	}

	if r.data.DeviceManager() != nil {
		r.data.DeviceManager().KickDevice(device.UserID, getStringValue(device.Identifier))
	}

	// 设置设备为离线状态
	err = device.Update().
		SetOnline(false).
		Exec(ctx)

	if err != nil {
		r.logger.Errorf("Failed to update device online status: %v", err)
		return err
	}

	r.logger.Infof("Device %d kicked offline (user_id: %d, identifier: %s)", deviceID, device.UserID, device.Identifier)
	return nil
}
