package document

import (
	"context"
	"strconv"

	"google.golang.org/protobuf/types/known/emptypb"

	v1 "github.com/OmnTeam/ppanel-pro/api/public/document/v1"
	documentBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/document"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

// Helper functions for type conversion
func parseInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

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
	// 调用业务层
	documents, total, err := s.uc.QueryDocumentList(ctx)
	if err != nil {
		return nil, err
	}

	// 转换结果
	list := make([]*v1.DocumentItem, 0, len(documents))
	for _, d := range documents {
		list = append(list, &v1.DocumentItem{
			Id:        strconv.FormatInt(d.ID, 10),
			Title:     d.Title,
			Tags:      d.Tags,
			UpdatedAt: strconv.FormatInt(d.UpdatedAt, 10),
		})
	}

	return &v1.DocumentListReply{
		Code:    int32(responsecode.DocumentQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.DocumentQuerySuccess],
		Data: &v1.DocumentListData{
			List:  list,
			Total: int32(total),
		},
	}, nil
}

// QueryDocumentDetail 查询文档详情
func (s *DocumentService) QueryDocumentDetail(ctx context.Context, req *v1.QueryDocumentDetailRequest) (*v1.DocumentDetailReply, error) {
	// 调用业务层
	document, err := s.uc.QueryDocumentDetail(ctx, int(parseInt64(req.Id)))
	if err != nil {
		return nil, err
	}

	return &v1.DocumentDetailReply{
		Code:    int32(responsecode.DocumentQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.DocumentQuerySuccess],
		Data: &v1.DocumentDetailData{
			Id:        strconv.FormatInt(document.ID, 10),
			Title:     document.Title,
			Content:   document.Content,
			Tags:      document.Tags,
			CreatedAt: strconv.FormatInt(document.CreatedAt, 10),
			UpdatedAt: strconv.FormatInt(document.UpdatedAt, 10),
		},
	}, nil
}
