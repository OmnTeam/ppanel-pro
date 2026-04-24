package document

import (
	"context"
)

// DocumentRepo Public Document数据仓库接口
type DocumentRepo interface {
	// QueryDocumentList 查询文档列表
	QueryDocumentList(ctx context.Context) ([]*DocumentItem, int32, error)

	// QueryDocumentDetail 查询文档详情
	QueryDocumentDetail(ctx context.Context, id int) (*DocumentDetail, error)
}

// DocumentItem 文档项（列表）
type DocumentItem struct {
	ID        int64
	Title     string
	Content   string
	Tags      []string
	Show      bool
	CreatedAt int64
	UpdatedAt int64
}

// DocumentDetail 文档详情
type DocumentDetail struct {
	ID        int64
	Title     string
	Content   string
	Tags      []string
	Show      bool
	CreatedAt int64
	UpdatedAt int64
}

// DocumentUseCase Public Document用例
type DocumentUseCase struct {
	repo DocumentRepo
}

// NewDocumentUseCase 创建Public Document用例
func NewDocumentUseCase(repo DocumentRepo) *DocumentUseCase {
	return &DocumentUseCase{repo: repo}
}

// QueryDocumentList 查询文档列表
func (uc *DocumentUseCase) QueryDocumentList(ctx context.Context) ([]*DocumentItem, int32, error) {
	return uc.repo.QueryDocumentList(ctx)
}

// QueryDocumentDetail 查询文档详情
func (uc *DocumentUseCase) QueryDocumentDetail(ctx context.Context, id int) (*DocumentDetail, error) {
	return uc.repo.QueryDocumentDetail(ctx, id)
}
