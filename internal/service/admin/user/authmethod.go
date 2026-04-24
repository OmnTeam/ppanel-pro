package user

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/user/v1"
	userbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/user"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/OmnTeam/ppanel-pro/pkg/phone"
)

// UserAuthMethodService 用户认证方法服务
type UserAuthMethodService struct {
	v1.UnimplementedUserAuthMethodServiceServer

	uc     *userbiz.AuthMethodUsecase
	logger *log.Helper
}

// NewUserAuthMethodService 创建用户认证方法服务
func NewUserAuthMethodService(uc *userbiz.AuthMethodUsecase, logger log.Logger) *UserAuthMethodService {
	return &UserAuthMethodService{
		uc:     uc,
		logger: log.NewHelper(logger),
	}
}

// CreateUserAuthMethod 创建用户认证方法
func (s *UserAuthMethodService) CreateUserAuthMethod(ctx context.Context, req *v1.CreateUserAuthMethodRequest) (*v1.CreateUserAuthMethodReply, error) {
	if _, err := s.uc.CreateUserAuthMethod(ctx, req); err != nil {
		return nil, err
	}

	return &v1.CreateUserAuthMethodReply{
		Code:    200,
		Message: responsecode.CodeMessages[200],
		Data:    &v1.CreateUserAuthMethodData{Success: true},
	}, nil
}

// GetUserAuthMethod 获取用户认证方法
func (s *UserAuthMethodService) GetUserAuthMethod(ctx context.Context, req *v1.GetUserAuthMethodRequest) (*v1.GetUserAuthMethodReply, error) {
	if req.UserId <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	methods, err := s.uc.GetUserAuthMethod(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	// 转换为Proto消息列表
	protoMethods := make([]*v1.UserAuthMethod, 0, len(methods))
	for _, method := range methods {
		protoMethod := &v1.UserAuthMethod{
			AuthType:       method.AuthType,
			AuthIdentifier: method.AuthIdentifier,
			Verified:       method.Verified,
		}
		if method.AuthType == "mobile" {
			protoMethod.AuthIdentifier = phone.FormatToInternational(method.AuthIdentifier)
		}
		protoMethods = append(protoMethods, protoMethod)
	}

	return &v1.GetUserAuthMethodReply{
		Code:    200,
		Message: responsecode.CodeMessages[200],
		Data: &v1.GetUserAuthMethodData{
			AuthMethods: protoMethods,
		},
	}, nil
}

// UpdateUserAuthMethod 更新用户认证方法
func (s *UserAuthMethodService) UpdateUserAuthMethod(ctx context.Context, req *v1.UpdateUserAuthMethodRequest) (*v1.UpdateUserAuthMethodReply, error) {
	err := s.uc.UpdateUserAuthMethod(ctx, req)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateUserAuthMethodReply{
		Code:    200,
		Message: responsecode.CodeMessages[200],
		Data:    &v1.UpdateUserAuthMethodData{Success: true},
	}, nil
}

// DeleteUserAuthMethod 删除用户认证方法
func (s *UserAuthMethodService) DeleteUserAuthMethod(ctx context.Context, req *v1.DeleteUserAuthMethodRequest) (*v1.DeleteUserAuthMethodReply, error) {
	if req.UserId <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	err := s.uc.DeleteUserAuthMethod(ctx, req.UserId, req.AuthType)
	if err != nil {
		return nil, err
	}

	return &v1.DeleteUserAuthMethodReply{
		Code:    200,
		Message: responsecode.CodeMessages[200],
		Data:    &v1.DeleteUserAuthMethodData{Success: true},
	}, nil
}
