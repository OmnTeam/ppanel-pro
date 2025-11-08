package data

import (
	"context"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyannouncement"
	announcementBiz "github.com/OmnTeam/ppanel-pro/internal/biz/public/announcement"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

type publicAnnouncementRepo struct {
	data *Data
	log  *log.Helper
}

// NewPublicAnnouncementRepo 创建Public Announcement仓库
func NewPublicAnnouncementRepo(data *Data, logger log.Logger) announcementBiz.AnnouncementRepo {
	return &publicAnnouncementRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// QueryAnnouncement 查询公告列表
func (r *publicAnnouncementRepo) QueryAnnouncement(ctx context.Context, tenantID int64, page, size int32, pinned, popup *bool) ([]*announcementBiz.Announcement, int64, error) {
	// 查询条件: show=true (移除tenant_id过滤)
	query := r.data.db.ProxyAnnouncement.Query().
		Where(
			proxyannouncement.Show(true),
		)

	// pinned过滤
	if pinned != nil {
		query = query.Where(proxyannouncement.Pinned(*pinned))
	}

	// popup过滤
	if popup != nil {
		query = query.Where(proxyannouncement.Popup(*popup))
	}

	// 统计总数
	total, err := query.Count(ctx)
	if err != nil {
		r.log.Errorf("QueryAnnouncement count error: %v", err)
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	// 分页查询
	announcements, err := query.
		Order(ent.Desc(proxyannouncement.FieldPinned), ent.Desc(proxyannouncement.FieldCreatedAt)).
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		All(ctx)

	if err != nil {
		r.log.Errorf("QueryAnnouncement query error: %v", err)
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	result := make([]*announcementBiz.Announcement, 0, len(announcements))
	for _, a := range announcements {
		result = append(result, &announcementBiz.Announcement{
			ID:        a.ID,
			Title:     a.Title,
			Content:   a.Content,
			Show:      a.Show,
			Pinned:    a.Pinned,
			Popup:     a.Popup,
			CreatedAt: a.CreatedAt.UnixMilli(),
			UpdatedAt: a.UpdatedAt.UnixMilli(),
		})
	}

	return result, int64(total), nil
}
