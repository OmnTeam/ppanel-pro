package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribeapplication"
	applicationbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/application"
)

type subscribeApplicationRepo struct {
	data *Data
	log  *log.Helper
}

// NewSubscribeApplicationRepo 创建订阅应用配置仓库
func NewSubscribeApplicationRepo(data *Data, logger log.Logger) applicationbiz.SubscribeApplicationRepo {
	return &subscribeApplicationRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Create 创建订阅应用配置
func (r *subscribeApplicationRepo) Create(ctx context.Context, app *applicationbiz.SubscribeApplication) (*applicationbiz.SubscribeApplication, error) {
	po, err := r.data.db.ProxySubscribeApplication.
		Create().
		SetName(app.Name).
		SetNillableIcon(app.Icon).
		SetNillableDescription(app.Description).
		SetScheme(app.Scheme).
		SetUserAgent(app.UserAgent).
		SetIsDefault(app.IsDefault).
		SetNillableSubscribeTemplate(app.SubscribeTemplate).
		SetOutputFormat(app.OutputFormat).
		SetDownloadLink(app.DownloadLink).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return r.convertToModel(po), nil
}

// Update 更新订阅应用配置
func (r *subscribeApplicationRepo) Update(ctx context.Context, app *applicationbiz.SubscribeApplication) (*applicationbiz.SubscribeApplication, error) {
	// 先查询确保应用配置存在
	existing, err := r.data.db.ProxySubscribeApplication.
		Query().
		Where(
			proxysubscribeapplication.ID(app.ID),
		).
		Only(ctx)

	if err != nil {
		return nil, err
	}

	// 构建更新操作，所有字段都直接设置（包括可选字段）
	updateBuilder := r.data.db.ProxySubscribeApplication.
		UpdateOne(existing).
		SetName(app.Name).
		SetScheme(app.Scheme).
		SetUserAgent(app.UserAgent).
		SetIsDefault(app.IsDefault).
		SetOutputFormat(app.OutputFormat).
		SetDownloadLink(app.DownloadLink).
		SetNillableIcon(app.Icon).
		SetNillableDescription(app.Description).
		SetNillableSubscribeTemplate(app.SubscribeTemplate)

	po, err := updateBuilder.Save(ctx)
	if err != nil {
		return nil, err
	}

	return r.convertToModel(po), nil
}

// FindByID 根据ID查找订阅应用配置
func (r *subscribeApplicationRepo) FindByID(ctx context.Context, id int64) (*applicationbiz.SubscribeApplication, error) {
	po, err := r.data.db.ProxySubscribeApplication.
		Query().
		Where(
			proxysubscribeapplication.ID(id),
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

// List 查询订阅应用配置列表
func (r *subscribeApplicationRepo) List(ctx context.Context) ([]*applicationbiz.SubscribeApplication, error) {
	query := r.data.db.ProxySubscribeApplication.Query()

	// 添加租户过滤
	query = query

	// 按创建时间倒序
	query = query.Order(ent.Desc(proxysubscribeapplication.FieldCreatedAt))

	pos, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	apps := make([]*applicationbiz.SubscribeApplication, 0, len(pos))
	for _, po := range pos {
		apps = append(apps, r.convertToModel(po))
	}

	return apps, nil
}

// Delete 删除订阅应用配置
func (r *subscribeApplicationRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.data.db.ProxySubscribeApplication.
		Delete().
		Where(
			proxysubscribeapplication.ID(id),
		).
		Exec(ctx)
	return err
}

// convertToModel 转换为业务模型
func (r *subscribeApplicationRepo) convertToModel(po *ent.ProxySubscribeApplication) *applicationbiz.SubscribeApplication {
	if po == nil {
		return nil
	}

	return &applicationbiz.SubscribeApplication{
		ID:                po.ID,
		Name:              po.Name,
		Icon:              po.Icon,
		Description:       po.Description,
		Scheme:            po.Scheme,
		UserAgent:         po.UserAgent,
		IsDefault:         po.IsDefault,
		SubscribeTemplate: po.SubscribeTemplate,
		OutputFormat:      po.OutputFormat,
		DownloadLink:      po.DownloadLink,
		CreatedAt:         po.CreatedAt,
		UpdatedAt:         po.UpdatedAt,
	}
}
