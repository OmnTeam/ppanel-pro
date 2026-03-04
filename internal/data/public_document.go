package data

import (
	"context"
	"strings"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxydocument"
	documentBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/document"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

type publicDocumentRepo struct {
	data *Data
	log  *log.Helper
}

// NewPublicDocumentRepo 创建Public Document仓库
func NewPublicDocumentRepo(data *Data, logger log.Logger) documentBiz.DocumentRepo {
	return &publicDocumentRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// QueryDocumentList 查询文档列表
func (r *publicDocumentRepo) QueryDocumentList(ctx context.Context) ([]*documentBiz.DocumentItem, int64, error) {
	// 查询所有文档（无分页）
	documents, err := r.data.db.ProxyDocument.Query().
		Order(ent.Desc(proxydocument.FieldUpdatedAt)).
		All(ctx)

	if err != nil {
		r.log.Errorf("QueryDocumentList query error: %v", err)
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	total := int64(len(documents))
	result := make([]*documentBiz.DocumentItem, 0, len(documents))

	for _, d := range documents {
		// 处理Tags字段（去重）
		tags := stringMergeAndRemoveDuplicates(d.Tags)

		result = append(result, &documentBiz.DocumentItem{
			ID:        int64(d.ID),
			Title:     d.Title,
			Tags:      tags,
			UpdatedAt: d.UpdatedAt.UnixMilli(),
		})
	}

	return result, total, nil
}

// QueryDocumentDetail 查询文档详情
func (r *publicDocumentRepo) QueryDocumentDetail(ctx context.Context, id int) (*documentBiz.DocumentDetail, error) {
	document, err := r.data.db.ProxyDocument.Query().
		Where(
			proxydocument.ID(int64(id)),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			r.log.Errorf("Document not found: id=%d", id)
			return nil, responsecode.NewKratosError(responsecode.ErrDocumentNotFound)
		}
		r.log.Errorf("QueryDocumentDetail query error: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	// 处理Tags字段（去重）
	tags := stringMergeAndRemoveDuplicates(document.Tags)

	return &documentBiz.DocumentDetail{
		ID:        int64(document.ID),
		Title:     document.Title,
		Content:   document.Content,
		Tags:      tags,
		CreatedAt: document.CreatedAt.UnixMilli(),
		UpdatedAt: document.UpdatedAt.UnixMilli(),
	}, nil
}

// stringMergeAndRemoveDuplicates 合并字符串并去重
func stringMergeAndRemoveDuplicates(str string) []string {
	if str == "" {
		return []string{}
	}

	// 按逗号分割
	parts := strings.Split(str, ",")
	seen := make(map[string]bool)
	result := make([]string, 0)

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" && !seen[trimmed] {
			seen[trimmed] = true
			result = append(result, trimmed)
		}
	}

	return result
}
