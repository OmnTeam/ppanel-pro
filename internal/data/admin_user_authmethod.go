package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/user/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	userbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/user"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

type adminUserAuthMethodRepo struct {
	data   *Data
	logger *log.Helper
}

// NewAdminUserAuthMethodRepo creates a new admin user auth method repository
func NewAdminUserAuthMethodRepo(d *Data, logger log.Logger) userbiz.AuthMethodRepo {
	return &adminUserAuthMethodRepo{
		data:   d,
		logger: log.NewHelper(logger),
	}
}

// CreateUserAuthMethod 创建用户认证方法（或更新已存在的）
func (r *adminUserAuthMethodRepo) CreateUserAuthMethod(ctx context.Context, req *v1.CreateUserAuthMethodRequest) (int64, error) {
	// 查询是否已存在该用户的该类型认证方法
	existing, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.UserIDEQ(int(req.UserId)),
			proxyuserauthmethod.AuthTypeEQ(req.AuthType),
		).
		Only(ctx)

	// 如果不存在，创建新记录
	if ent.IsNotFound(err) {
		created, createErr := r.data.db.ProxyUserAuthMethod.Create().
			SetUserID(int(req.UserId)).
			SetAuthType(req.AuthType).
			SetAuthIdentifier(req.AuthIdentifier).
			SetVerified(req.Verified).
			Save(ctx)
		if createErr != nil {
			r.logger.Errorf("Failed to create auth method: %v", createErr)
			return 0, createErr
		}
		return int64(created.ID), nil
	}

	if err != nil {
		r.logger.Errorf("Failed to query auth method: %v", err)
		return 0, err
	}

	// 如果已存在，更新记录
	updated, updateErr := existing.Update().
		SetAuthIdentifier(req.AuthIdentifier).
		SetVerified(req.Verified).
		Save(ctx)
	if updateErr != nil {
		r.logger.Errorf("Failed to update auth method: %v", updateErr)
		return 0, updateErr
	}

	return int64(updated.ID), nil
}

// GetUserAuthMethod 获取用户认证方法列表
func (r *adminUserAuthMethodRepo) GetUserAuthMethod(ctx context.Context, userID int64, authType string) ([]*ent.ProxyUserAuthMethod, error) {
	query := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.UserIDEQ(int(userID)),
		)

	// 如果指定了认证类型，添加过滤条件
	if authType != "" {
		query = query.Where(proxyuserauthmethod.AuthTypeEQ(authType))
	}

	methods, err := query.All(ctx)
	if err != nil {
		r.logger.Errorf("Failed to query auth methods: %v", err)
		return nil, err
	}

	return methods, nil
}

// UpdateUserAuthMethod 更新用户认证方法
func (r *adminUserAuthMethodRepo) UpdateUserAuthMethod(ctx context.Context, req *v1.UpdateUserAuthMethodRequest) error {
	// 查找要更新的认证方法
	authMethod, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.UserIDEQ(int(req.UserId)),
			proxyuserauthmethod.AuthTypeEQ(req.AuthType),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return responsecode.NewKratosError(responsecode.ErrAuthMethodNotFound)
		}
		r.logger.Errorf("Failed to query auth method: %v", err)
		return err
	}

	// 更新认证标识
	err = authMethod.Update().
		SetAuthIdentifier(req.AuthIdentifier).
		Exec(ctx)

	if err != nil {
		r.logger.Errorf("Failed to update auth method: %v", err)
		return err
	}

	return nil
}

// DeleteUserAuthMethod 删除用户认证方法
func (r *adminUserAuthMethodRepo) DeleteUserAuthMethod(ctx context.Context, userID int64, authType string) error {
	deletedCount, err := r.data.db.ProxyUserAuthMethod.Delete().
		Where(
			proxyuserauthmethod.UserIDEQ(int(userID)),
			proxyuserauthmethod.AuthTypeEQ(authType),
		).
		Exec(ctx)

	if err != nil {
		r.logger.Errorf("Failed to delete auth method: %v", err)
		return err
	}

	if deletedCount == 0 {
		return responsecode.NewKratosError(responsecode.ErrAuthMethodNotFound)
	}

	return nil
}
