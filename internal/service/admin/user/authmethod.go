package user

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/user/v1"
	userbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/user"
)

// Helper functions for type conversion
func parseInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

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
	id, err := s.uc.CreateUserAuthMethod(ctx, req)
	if err != nil {
		return nil, err
	}

	return &v1.CreateUserAuthMethodReply{
		Id: strconv.FormatInt(id, 10),
	}, nil
}

// GetUserAuthMethod 获取用户认证方法
func (s *UserAuthMethodService) GetUserAuthMethod(ctx context.Context, req *v1.GetUserAuthMethodRequest) (*v1.GetUserAuthMethodReply, error) {
	methods, err := s.uc.GetUserAuthMethod(ctx, parseInt64(req.UserId), req.AuthType)
	if err != nil {
		return nil, err
	}

	// 转换为Proto消息列表
	protoMethods := make([]*v1.UserAuthMethod, 0, len(methods))
	for _, method := range methods {
		protoMethod := &v1.UserAuthMethod{
			Id:             strconv.FormatInt(int64(method.ID), 10),
			UserId:         strconv.FormatInt(int64(method.UserID), 10),
			AuthType:       method.AuthType,
			AuthIdentifier: method.AuthIdentifier,
			Verified:       method.Verified,
			CreatedAt:      strconv.FormatInt(method.CreatedAt.UnixMilli(), 10),
			UpdatedAt:      strconv.FormatInt(method.UpdatedAt.UnixMilli(), 10),
		}
		protoMethods = append(protoMethods, protoMethod)
	}

	return &v1.GetUserAuthMethodReply{
		List: protoMethods,
	}, nil
}

// UpdateUserAuthMethod 更新用户认证方法
func (s *UserAuthMethodService) UpdateUserAuthMethod(ctx context.Context, req *v1.UpdateUserAuthMethodRequest) (*v1.UpdateUserAuthMethodReply, error) {
	err := s.uc.UpdateUserAuthMethod(ctx, req)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateUserAuthMethodReply{}, nil
}

// DeleteUserAuthMethod 删除用户认证方法
func (s *UserAuthMethodService) DeleteUserAuthMethod(ctx context.Context, req *v1.DeleteUserAuthMethodRequest) (*v1.DeleteUserAuthMethodReply, error) {
	err := s.uc.DeleteUserAuthMethod(ctx, parseInt64(req.UserId), req.AuthType)
	if err != nil {
		return nil, err
	}

	return &v1.DeleteUserAuthMethodReply{}, nil
}
