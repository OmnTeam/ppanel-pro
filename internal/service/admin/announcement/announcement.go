package announcement

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/npanel-pro/api/admin/announcement/v1"
	announcementbiz "github.com/OmnTeam/npanel-pro/internal/biz/admin/announcement"
	"github.com/OmnTeam/npanel-pro/internal/responsecode"
)

// AnnouncementService 公告服务
type AnnouncementService struct {
	v1.UnimplementedAnnouncementServiceServer

	uc     *announcementbiz.AnnouncementUsecase
	logger *log.Helper
}

// NewAnnouncementService 创建公告服务
func NewAnnouncementService(uc *announcementbiz.AnnouncementUsecase, logger log.Logger) *AnnouncementService {
	return &AnnouncementService{
		uc:     uc,
		logger: log.NewHelper(logger),
	}
}

// CreateAnnouncement 创建公告
func (s *AnnouncementService) CreateAnnouncement(ctx context.Context, req *v1.CreateAnnouncementRequest) (*v1.AnnouncementReply, error) {
	announcement := &announcementbiz.Announcement{
		Title:   req.Title,
		Content: &req.Content,
	}

	result, err := s.uc.CreateAnnouncement(ctx, announcement)
	if err != nil {
		return nil, err
	}

	return &v1.AnnouncementReply{
		Code:    int32(responsecode.AdminCreateAnnouncementSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminCreateAnnouncementSuccess],
		Data: &v1.AnnouncementData{
			Announcement: s.convertToProto(result),
		},
	}, nil
}

// UpdateAnnouncement 更新公告
func (s *AnnouncementService) UpdateAnnouncement(ctx context.Context, req *v1.UpdateAnnouncementRequest) (*v1.AnnouncementReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	announcement := &announcementbiz.Announcement{
		ID:      req.Id,
		Title:   req.Title,
		Content: &req.Content,
	}
	if req.Show != nil {
		announcement.Show = req.Show
	}
	if req.Pinned != nil {
		announcement.Pinned = req.Pinned
	}
	if req.Popup != nil {
		announcement.Popup = req.Popup
	}

	result, err := s.uc.UpdateAnnouncement(ctx, announcement)
	if err != nil {
		return nil, err
	}

	return &v1.AnnouncementReply{
		Code:    int32(responsecode.AdminUpdateAnnouncementSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminUpdateAnnouncementSuccess],
		Data: &v1.AnnouncementData{
			Announcement: s.convertToProto(result),
		},
	}, nil
}

// GetAnnouncement 获取公告详情
func (s *AnnouncementService) GetAnnouncement(ctx context.Context, req *v1.GetAnnouncementRequest) (*v1.AnnouncementReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	announcement, err := s.uc.GetAnnouncement(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.AnnouncementReply{
		Code:    int32(responsecode.AdminGetAnnouncementSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminGetAnnouncementSuccess],
		Data: &v1.AnnouncementData{
			Announcement: s.convertToProto(announcement),
		},
	}, nil
}

// ListAnnouncements 获取公告列表
func (s *AnnouncementService) ListAnnouncements(ctx context.Context, req *v1.ListAnnouncementsRequest) (*v1.ListAnnouncementsReply, error) {
	announcements, total, err := s.uc.ListAnnouncements(ctx, int(req.Page), int(req.Size), req.Search, req.Show, req.Pinned, req.Popup)
	if err != nil {
		return nil, err
	}

	// 转换为protobuf格式
	items := make([]*v1.Announcement, 0, len(announcements))
	for _, announcement := range announcements {
		items = append(items, s.convertToProto(announcement))
	}

	return &v1.ListAnnouncementsReply{
		Code:    int32(responsecode.AdminListAnnouncementsSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminListAnnouncementsSuccess],
		Data: &v1.ListAnnouncementsData{
			List:  items,
			Total: total,
		},
	}, nil
}

// DeleteAnnouncement 删除公告
func (s *AnnouncementService) DeleteAnnouncement(ctx context.Context, req *v1.DeleteAnnouncementRequest) (*v1.DeleteAnnouncementReply, error) {
	if req.Id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	err := s.uc.DeleteAnnouncement(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1.DeleteAnnouncementReply{
		Code:    int32(responsecode.AdminDeleteAnnouncementSuccess),
		Message: responsecode.CodeMessages[responsecode.AdminDeleteAnnouncementSuccess],
		Data: &v1.DeleteAnnouncementData{
			Success: true,
		},
	}, nil
}

// convertToProto 将业务模型转换为protobuf消息
func (s *AnnouncementService) convertToProto(announcement *announcementbiz.Announcement) *v1.Announcement {
	result := &v1.Announcement{
		Id:        announcement.ID,
		Title:     announcement.Title,
		CreatedAt: announcement.CreatedAt.Unix(),
		UpdatedAt: announcement.UpdatedAt.Unix(),
	}

	// 处理可选字段
	if announcement.Content != nil {
		result.Content = *announcement.Content
	}
	if announcement.Show != nil {
		result.Show = announcement.Show
	}
	if announcement.Pinned != nil {
		result.Pinned = announcement.Pinned
	}
	if announcement.Popup != nil {
		result.Popup = announcement.Popup
	}

	return result
}
