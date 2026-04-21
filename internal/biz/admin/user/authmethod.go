package user

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/user/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
)

// AuthMethodRepo 认证方法仓储接口
type AuthMethodRepo interface {
	// CreateUserAuthMethod 创建用户认证方法（或更新已存在的）
	CreateUserAuthMethod(ctx context.Context, req *v1.CreateUserAuthMethodRequest) (int64, error)

	// GetUserAuthMethod 获取用户认证方法列表
	GetUserAuthMethod(ctx context.Context, userID int64) ([]*ent.ProxyUserAuthMethod, error)

	// UpdateUserAuthMethod 更新用户认证方法
	UpdateUserAuthMethod(ctx context.Context, req *v1.UpdateUserAuthMethodRequest) error

	// DeleteUserAuthMethod 删除用户认证方法
	DeleteUserAuthMethod(ctx context.Context, userID int64, authType string) error
}

// AuthMethodUsecase 认证方法用例
type AuthMethodUsecase struct {
	repo   AuthMethodRepo
	logger *log.Helper
}

// NewAuthMethodUsecase 创建认证方法用例
func NewAuthMethodUsecase(repo AuthMethodRepo, logger log.Logger) *AuthMethodUsecase {
	return &AuthMethodUsecase{
		repo:   repo,
		logger: log.NewHelper(logger),
	}
}

// CreateUserAuthMethod 创建用户认证方法
func (uc *AuthMethodUsecase) CreateUserAuthMethod(ctx context.Context, req *v1.CreateUserAuthMethodRequest) (int64, error) {
	return uc.repo.CreateUserAuthMethod(ctx, req)
}

// GetUserAuthMethod 获取用户认证方法
func (uc *AuthMethodUsecase) GetUserAuthMethod(ctx context.Context, userID int64) ([]*ent.ProxyUserAuthMethod, error) {
	return uc.repo.GetUserAuthMethod(ctx, userID)
}

// UpdateUserAuthMethod 更新用户认证方法
func (uc *AuthMethodUsecase) UpdateUserAuthMethod(ctx context.Context, req *v1.UpdateUserAuthMethodRequest) error {
	return uc.repo.UpdateUserAuthMethod(ctx, req)
}

// DeleteUserAuthMethod 删除用户认证方法
func (uc *AuthMethodUsecase) DeleteUserAuthMethod(ctx context.Context, userID int64, authType string) error {
	return uc.repo.DeleteUserAuthMethod(ctx, userID, authType)
}
