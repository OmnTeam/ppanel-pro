package data

import (
	"context"
	"strconv"

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
	// Parse user ID
	userID, err := strconv.ParseInt(req.UserId, 10, 64)
	if err != nil {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	existing, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.UserIDEQ(userID),
			proxyuserauthmethod.AuthTypeEQ(req.AuthType),
		).
		Only(ctx)

	if ent.IsNotFound(err) {
		created, createErr := r.data.db.ProxyUserAuthMethod.Create().
			SetUserID(userID).
			SetAuthType(req.AuthType).
			SetAuthIdentifier(req.AuthIdentifier).
			Save(ctx)
		if createErr != nil {
			r.logger.Errorf("Failed to create auth method: %v", createErr)
			return 0, createErr
		}
		return created.ID, nil
	}

	if err != nil {
		r.logger.Errorf("Failed to query auth method: %v", err)
		return 0, err
	}

	updated, updateErr := existing.Update().
		SetAuthIdentifier(req.AuthIdentifier).
		Save(ctx)
	if updateErr != nil {
		r.logger.Errorf("Failed to update auth method: %v", updateErr)
		return 0, updateErr
	}

	return updated.ID, nil
}

// GetUserAuthMethod 获取用户认证方法列表
func (r *adminUserAuthMethodRepo) GetUserAuthMethod(ctx context.Context, userID int64) ([]*ent.ProxyUserAuthMethod, error) {
	query := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.UserIDEQ(userID),
		)

	methods, err := query.All(ctx)
	if err != nil {
		r.logger.Errorf("Failed to query auth methods: %v", err)
		return nil, err
	}

	return methods, nil
}

// UpdateUserAuthMethod 更新用户认证方法
func (r *adminUserAuthMethodRepo) UpdateUserAuthMethod(ctx context.Context, req *v1.UpdateUserAuthMethodRequest) error {
	// Parse user ID
	userID, err := strconv.ParseInt(req.UserId, 10, 64)
	if err != nil {
		return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	// 查找要更新的认证方法
	authMethod, err := r.data.db.ProxyUserAuthMethod.Query().
		Where(
			proxyuserauthmethod.UserIDEQ(userID),
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
			proxyuserauthmethod.UserIDEQ(userID),
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
