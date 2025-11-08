package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyauthmethod"
	authmethodbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/authmethod"
)

type adminAuthMethodRepo struct {
	data *Data
	log  *log.Helper
}

// NewAdminAuthMethodRepo 创建认证方法仓储
func NewAdminAuthMethodRepo(data *Data, logger log.Logger) authmethodbiz.AuthMethodRepo {
	return &adminAuthMethodRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// FindByMethod 根据方法名查找
func (r *adminAuthMethodRepo) FindByMethod(ctx context.Context, tenantID int64, method string) (*authmethodbiz.AuthMethod, error) {
	po, err := r.data.db.ProxyAuthMethod.
		Query().
		Where(
			proxyauthmethod.Method(method),
		).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return r.convertToModel(po), nil
}

// Update 更新认证方法
func (r *adminAuthMethodRepo) Update(ctx context.Context, auth *authmethodbiz.AuthMethod) (*authmethodbiz.AuthMethod, error) {
	// 先查找
	existing, err := r.data.db.ProxyAuthMethod.
		Query().
		Where(
			proxyauthmethod.Method(auth.Method),
		).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// 不存在则创建
			po, err := r.data.db.ProxyAuthMethod.
				Create().
				SetMethod(auth.Method).
				SetConfig(auth.Config).
				SetEnabled(auth.Enabled).
				Save(ctx)
			if err != nil {
				return nil, err
			}
			return r.convertToModel(po), nil
		}
		return nil, err
	}

	// 存在则更新
	po, err := r.data.db.ProxyAuthMethod.
		UpdateOne(existing).
		SetConfig(auth.Config).
		SetEnabled(auth.Enabled).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return r.convertToModel(po), nil
}

// FindAll 查找所有认证方法
func (r *adminAuthMethodRepo) FindAll(ctx context.Context, tenantID int64) ([]*authmethodbiz.AuthMethod, error) {
	pos, err := r.data.db.ProxyAuthMethod.
		Query().
		All(ctx)

	if err != nil {
		return nil, err
	}

	result := make([]*authmethodbiz.AuthMethod, 0, len(pos))
	for _, po := range pos {
		result = append(result, r.convertToModel(po))
	}

	return result, nil
}

// convertToModel 转换为业务模型
func (r *adminAuthMethodRepo) convertToModel(po *ent.ProxyAuthMethod) *authmethodbiz.AuthMethod {
	if po == nil {
		return nil
	}

	return &authmethodbiz.AuthMethod{
		ID:        int64(po.ID),
		Method:    po.Method,
		Config:    po.Config,
		Enabled:   po.Enabled,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}
