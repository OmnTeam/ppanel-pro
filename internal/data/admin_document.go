package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxydocument"
	documentbiz "github.com/OmnTeam/ppanel-pro/internal/biz/admin/document"
)

type adminDocumentRepo struct {
	data *Data
	log  *log.Helper
}

// NewAdminDocumentRepo 创建管理员文档仓库
func NewAdminDocumentRepo(data *Data, logger log.Logger) documentbiz.DocumentRepo {
	return &adminDocumentRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// Create 创建文档
// 对应原项目 documentModel.Insert
func (r *adminDocumentRepo) Create(ctx context.Context, doc *documentbiz.Document) (*documentbiz.Document, error) {
	po, err := r.data.db.ProxyDocument.
		Create().
		SetTitle(doc.Title).
		SetContent(doc.Content).
		SetTags(doc.Tags).
		SetShow(doc.Show).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return r.convertToModel(po), nil
}

// Update 更新文档
// 对应原项目 documentModel.Update
func (r *adminDocumentRepo) Update(ctx context.Context, doc *documentbiz.Document) error {
	_, err := r.data.db.ProxyDocument.
		UpdateOneID(doc.ID).
		SetTitle(doc.Title).
		SetContent(doc.Content).
		SetTags(doc.Tags).
		SetShow(doc.Show).
		Save(ctx)

	return err
}

// Delete 删除文档
// 对应原项目 documentModel.Delete
func (r *adminDocumentRepo) Delete(ctx context.Context, id int) error {
	_, err := r.data.db.ProxyDocument.
		Delete().
		Where(
			proxydocument.ID(int64(id)),
		).
		Exec(ctx)

	return err
}

// FindByID 根据ID查找文档
// 对应原项目 documentModel.QueryDocumentDetail
func (r *adminDocumentRepo) FindByID(ctx context.Context, id int) (*documentbiz.Document, error) {
	po, err := r.data.db.ProxyDocument.
		Query().
		Where(
			proxydocument.ID(int64(id)),
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

// List 获取文档列表
// 对应原项目 documentModel.QueryDocumentList
func (r *adminDocumentRepo) List(ctx context.Context, page, size int, tag, search string) (int64, []*documentbiz.Document, error) {
	query := r.data.db.ProxyDocument.
		Query()

	// 第38-40行：tag 过滤（使用 FIND_IN_SET）
	if tag != "" {
		// Ent 不直接支持 FIND_IN_SET，使用 Contains 近似
		// 或者使用原始 SQL 查询
		query = query.Where(proxydocument.TagsContains(tag))
	}

	// 第41-43行：search 过滤（标题或内容）
	if search != "" {
		query = query.Where(
			proxydocument.Or(
				proxydocument.TitleContains(search),
				proxydocument.ContentContains(search),
			),
		)
	}

	// 第44行：获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return 0, nil, err
	}

	// 第44行：分页查询
	pos, err := query.
		Order(ent.Desc(proxydocument.FieldUpdatedAt)).
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		All(ctx)

	if err != nil {
		return 0, nil, err
	}

	// 转换为业务模型
	documents := make([]*documentbiz.Document, len(pos))
	for i, po := range pos {
		documents[i] = r.convertToModel(po)
	}

	return int64(total), documents, nil
}

// convertToModel 将数据库实体转换为业务模型
func (r *adminDocumentRepo) convertToModel(po *ent.ProxyDocument) *documentbiz.Document {
	return &documentbiz.Document{
		ID:        int64(po.ID),
		Title:     po.Title,
		Content:   po.Content,
		Tags:      po.Tags,
		Show:      po.Show,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}
