package document

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	v1 "github.com/OmnTeam/ppanel-pro/api/public/document/v1"
	documentBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/document"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

// DocumentService Public Document服务实现
type DocumentService struct {
	v1.UnimplementedDocumentServer
	uc *documentBiz.DocumentUseCase
}

// NewDocumentService 创建Public Document服务
func NewDocumentService(uc *documentBiz.DocumentUseCase) *DocumentService {
	return &DocumentService{uc: uc}
}

// QueryDocumentList 查询文档列表
func (s *DocumentService) QueryDocumentList(ctx context.Context, req *emptypb.Empty) (*v1.DocumentListReply, error) {
	// 调用业务层，使用默认租户ID
	documents, total, err := s.uc.QueryDocumentList(ctx, 0)
	if err != nil {
		return nil, err
	}

	// 转换结果
	list := make([]*v1.DocumentItem, 0, len(documents))
	for _, d := range documents {
		list = append(list, &v1.DocumentItem{
			Id:        d.ID,
			Title:     d.Title,
			Tags:      d.Tags,
			UpdatedAt: d.UpdatedAt,
		})
	}

	return &v1.DocumentListReply{
		Code:    int32(responsecode.DocumentQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.DocumentQuerySuccess],
		Data: &v1.DocumentListData{
			List:  list,
			Total: total,
		},
	}, nil
}

// QueryDocumentDetail 查询文档详情
func (s *DocumentService) QueryDocumentDetail(ctx context.Context, req *v1.QueryDocumentDetailRequest) (*v1.DocumentDetailReply, error) {
	// 调用业务层，使用默认租户ID
	document, err := s.uc.QueryDocumentDetail(ctx, 0, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.DocumentDetailReply{
		Code:    int32(responsecode.DocumentQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.DocumentQuerySuccess],
		Data: &v1.DocumentDetailData{
			Id:        document.ID,
			Title:     document.Title,
			Content:   document.Content,
			Tags:      document.Tags,
			CreatedAt: document.CreatedAt,
			UpdatedAt: document.UpdatedAt,
		},
	}, nil
}
