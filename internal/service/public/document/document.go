package document

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	v1 "github.com/OmnTeam/npanel-pro/api/public/document/v1"
	documentBiz "github.com/OmnTeam/npanel-pro/internal/biz/public/document"
)

// DocumentService Public Document服务实现
type DocumentService struct {
	v1.UnimplementedPublicDocumentServer
	uc *documentBiz.DocumentUseCase
}

// NewDocumentService 创建Public Document服务
func NewDocumentService(uc *documentBiz.DocumentUseCase) *DocumentService {
	return &DocumentService{uc: uc}
}

// QueryDocumentList 查询文档列表
func (s *DocumentService) QueryDocumentList(ctx context.Context, req *emptypb.Empty) (*v1.DocumentListReply, error) {
	// 调用业务层
	documents, total, err := s.uc.QueryDocumentList(ctx)
	if err != nil {
		return nil, err
	}

	// 转换结果
	list := make([]*v1.Document, 0, len(documents))
	for _, d := range documents {
		list = append(list, &v1.Document{
			Id:        d.ID,
			Title:     d.Title,
			Content:   d.Content,
			Tags:      d.Tags,
			Show:      d.Show,
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
		})
	}

	return &v1.DocumentListReply{
		Total: total,
		List:  list,
	}, nil
}

// QueryDocumentDetail 查询文档详情
func (s *DocumentService) QueryDocumentDetail(ctx context.Context, req *v1.QueryDocumentDetailRequest) (*v1.DocumentDetailReply, error) {
	// 调用业务层
	document, err := s.uc.QueryDocumentDetail(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}

	return &v1.DocumentDetailReply{
		Id:        document.ID,
		Title:     document.Title,
		Content:   document.Content,
		Tags:      document.Tags,
		Show:      document.Show,
		CreatedAt: document.CreatedAt,
		UpdatedAt: document.UpdatedAt,
	}, nil
}
