package user

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/user/v1"
	userbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/user"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

// UserDeviceService 用户设备服务
type UserDeviceService struct {
	v1.UnimplementedUserDeviceServiceServer

	uc     *userbiz.DeviceUsecase
	logger *log.Helper
}

// NewUserDeviceService 创建用户设备服务
func NewUserDeviceService(uc *userbiz.DeviceUsecase, logger log.Logger) *UserDeviceService {
	return &UserDeviceService{
		uc:     uc,
		logger: log.NewHelper(logger),
	}
}

// UpdateUserDevice 更新用户设备
func (s *UserDeviceService) UpdateUserDevice(ctx context.Context, req *v1.UpdateUserDeviceRequest) (*v1.UpdateUserDeviceReply, error) {
	err := s.uc.UpdateUserDevice(ctx, req)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateUserDeviceReply{
		Code:    200,
		Message: responsecode.CodeMessages[200],
		Data:    &v1.UpdateUserDeviceData{Success: true},
	}, nil
}

// DeleteUserDevice 删除用户设备
func (s *UserDeviceService) DeleteUserDevice(ctx context.Context, req *v1.DeleteUserDeviceRequest) (*v1.DeleteUserDeviceReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	err := s.uc.DeleteUserDevice(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.DeleteUserDeviceReply{
		Code:    200,
		Message: responsecode.CodeMessages[200],
		Data:    &v1.DeleteUserDeviceData{Success: true},
	}, nil
}

// KickOfflineByUserDevice 踢下线用户设备
func (s *UserDeviceService) KickOfflineByUserDevice(ctx context.Context, req *v1.KickOfflineByUserDeviceRequest) (*v1.KickOfflineByUserDeviceReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	err := s.uc.KickOfflineByUserDevice(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.KickOfflineByUserDeviceReply{
		Code:    200,
		Message: responsecode.CodeMessages[200],
		Data:    &v1.KickOfflineByUserDeviceData{Success: true},
	}, nil
}
