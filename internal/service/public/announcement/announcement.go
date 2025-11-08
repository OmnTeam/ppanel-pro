package announcement

import (
	"context"

	v1 "github.com/OmnTeam/ppanel-pro/api/public/announcement/v1"
	announcementBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/announcement"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
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
func (s *AnnouncementService) QueryAnnouncement(ctx context.Context, req *v1.QueryAnnouncementRequest) (*v1.AnnouncementListReply, error) {
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
	announcements, total, err := s.uc.QueryAnnouncement(ctx, 0, req.Page, req.Size, pinned, popup)
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
			Show:      a.Show,
			Pinned:    a.Pinned,
			Popup:     a.Popup,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
		})
	}

	return &v1.AnnouncementListReply{
		Code:    int32(responsecode.AnnouncementQuerySuccess),
		Message: responsecode.CodeMessages[responsecode.AnnouncementQuerySuccess],
		Data: &v1.AnnouncementListData{
			List:  list,
			Total: total,
		},
	}, nil
}
