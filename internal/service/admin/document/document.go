package document

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/npanel-pro/api/admin/document/v1"
	documentbiz "github.com/OmnTeam/npanel-pro/internal/biz/admin/document"
	"github.com/OmnTeam/npanel-pro/internal/responsecode"
	"github.com/OmnTeam/npanel-pro/pkg/tool"
)

// DocumentService 文档服务
type DocumentService struct {
	v1.UnimplementedDocumentServiceServer

	uc     *documentbiz.DocumentUsecase
	logger *log.Helper
}

// NewDocumentService 创建文档服务
func NewDocumentService(uc *documentbiz.DocumentUsecase, logger log.Logger) *DocumentService {
	return &DocumentService{
		uc:     uc,
		logger: log.NewHelper(logger),
	}
}

// CreateDocument 创建文档
func (s *DocumentService) CreateDocument(ctx context.Context, req *v1.CreateDocumentRequest) (*v1.CreateDocumentReply, error) {
	// 处理 show 字段
	var show *bool
	if req.Show != nil {
		show = &req.Show.Value
	}

	_, err := s.uc.CreateDocument(ctx, req.Title, req.Content, req.Tags, show)
	if err != nil {
		return nil, err
	}

	return &v1.CreateDocumentReply{
		Code:    int32(responsecode.AdminCreateDocumentSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminCreateDocumentSuccess],
	}, nil
}

// UpdateDocument 更新文档
func (s *DocumentService) UpdateDocument(ctx context.Context, req *v1.UpdateDocumentRequest) (*v1.UpdateDocumentReply, error) {
	// 处理 show 字段
	var show *bool
	if req.Show != nil {
		show = &req.Show.Value
	}

	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	err := s.uc.UpdateDocument(ctx, int(req.Id), req.Title, req.Content, req.Tags, show)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateDocumentReply{
		Code:    int32(responsecode.AdminUpdateDocumentSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminUpdateDocumentSuccess],
	}, nil
}

// DeleteDocument 删除文档
func (s *DocumentService) DeleteDocument(ctx context.Context, req *v1.DeleteDocumentRequest) (*v1.DeleteDocumentReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	err := s.uc.DeleteDocument(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}

	return &v1.DeleteDocumentReply{
		Code:    int32(responsecode.AdminDeleteDocumentSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminDeleteDocumentSuccess],
		Data: &v1.DeleteDocumentData{
			Success: true,
		},
	}, nil
}

// BatchDeleteDocument 批量删除文档
func (s *DocumentService) BatchDeleteDocument(ctx context.Context, req *v1.BatchDeleteDocumentRequest) (*v1.BatchDeleteDocumentReply, error) {
	err := s.uc.BatchDeleteDocument(ctx, convertStringSliceToIntSlice(req.Ids))
	if err != nil {
		return nil, err
	}

	return &v1.BatchDeleteDocumentReply{
		Code:    int32(responsecode.AdminBatchDeleteDocumentSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminBatchDeleteDocumentSuccess],
		Data: &v1.BatchDeleteDocumentData{
			Success: true,
		},
	}, nil
}

// GetDocumentDetail 获取文档详情
func (s *DocumentService) GetDocumentDetail(ctx context.Context, req *v1.GetDocumentDetailRequest) (*v1.GetDocumentDetailReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	doc, err := s.uc.GetDocumentDetail(ctx, int(req.Id))
	if err != nil {
		return nil, err
	}

	// 对应原项目第35-44行：构造响应，转换 tags
	return &v1.GetDocumentDetailReply{
		Code:    int32(responsecode.AdminGetDocumentDetailSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetDocumentDetailSuccess],
		Data: &v1.GetDocumentDetailData{
			Document: &v1.Document{
				Id:        doc.ID,
				Title:     doc.Title,
				Content:   doc.Content,
				Tags:      tool.StringMergeAndRemoveDuplicates(doc.Tags),
				Show:      doc.Show,
				CreatedAt: doc.CreatedAt.UnixMilli(),
				UpdatedAt: doc.UpdatedAt.UnixMilli(),
			},
		},
	}, nil
}

// GetDocumentList 获取文档列表
func (s *DocumentService) GetDocumentList(ctx context.Context, req *v1.GetDocumentListRequest) (*v1.GetDocumentListReply, error) {
	total, data, err := s.uc.GetDocumentList(ctx, int(req.Page), int(req.Size), req.Tag, req.Search)
	if err != nil {
		return nil, err
	}

	// 对应原项目第35-50行：构造响应列表
	list := make([]*v1.Document, 0, len(data))
	for _, doc := range data {
		list = append(list, &v1.Document{
			Id:        doc.ID,
			Title:     doc.Title,
			Content:   doc.Content,
			Tags:      tool.StringMergeAndRemoveDuplicates(doc.Tags),
			Show:      doc.Show,
			CreatedAt: doc.CreatedAt.UnixMilli(),
			UpdatedAt: doc.UpdatedAt.UnixMilli(),
		})
	}

	return &v1.GetDocumentListReply{
		Code:    int32(responsecode.AdminGetDocumentListSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetDocumentListSuccess],
		Data: &v1.GetDocumentListData{
			Total: total,
			List:  list,
		},
	}, nil
}

func convertStringSliceToIntSlice(input []int64) []int {
	if input == nil {
		return nil
	}
	result := make([]int, len(input))
	for i, v := range input {
		result[i] = int(v)
	}
	return result
}
