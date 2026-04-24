package announcement

import (
	"context"

	v1 "github.com/OmnTeam/ppanel-pro/api/public/announcement/v1"
	announcementBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/announcement"
)

// AnnouncementService Public Announcement服务实现
type AnnouncementService struct {
	v1.UnimplementedAnnouncementServer
	uc *announcementBiz.AnnouncementUseCase
}

// NewAnnouncementService 创建Public Announcement服务
func NewAnnouncementService(uc *announcementBiz.AnnouncementUseCase) *AnnouncementService {
	return &AnnouncementService{uc: uc}
}

// QueryAnnouncement 查询公告列表
func (s *AnnouncementService) QueryAnnouncement(ctx context.Context, req *v1.QueryAnnouncementRequest) (*v1.QueryAnnouncementReply, error) {
	size := req.Size
	if size == 0 {
		size = 15
	}

	// 处理可选参数
	var pinned, popup *bool
	if req.Pinned != nil {
		v := *req.Pinned
		pinned = &v
	}
	if req.Popup != nil {
		v := *req.Popup
		popup = &v
	}

	// 调用业务层
	announcements, total, err := s.uc.QueryAnnouncement(ctx, req.Page, size, pinned, popup)
	if err != nil {
		return nil, err
	}

	// 转换结果
	list := make([]*v1.AnnouncementItem, 0, len(announcements))
	for _, a := range announcements {
		list = append(list, &v1.AnnouncementItem{
			Id:        a.ID,
			Title:     a.Title,
			Content:   a.Content,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
		})
		list[len(list)-1].Show = &a.Show
		list[len(list)-1].Pinned = &a.Pinned
		list[len(list)-1].Popup = &a.Popup
	}

	return &v1.QueryAnnouncementReply{
		Announcements: list,
		Total:         total,
	}, nil
}
