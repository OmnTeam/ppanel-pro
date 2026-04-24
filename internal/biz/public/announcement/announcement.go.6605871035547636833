package announcement

import (
	"context"
)

// AnnouncementRepo Public Announcement数据仓库接口
type AnnouncementRepo interface {
	// QueryAnnouncement 查询公告列表
	QueryAnnouncement(ctx context.Context, page, size int32, pinned, popup *bool) ([]*Announcement, int32, error)
}

// Announcement 公告信息
type Announcement struct {
	ID        int64
	Title     string
	Content   string
	Show      bool
	Pinned    bool
	Popup     bool
	CreatedAt int64
	UpdatedAt int64
}

// AnnouncementUseCase Public Announcement用例
type AnnouncementUseCase struct {
	repo AnnouncementRepo
}

// NewAnnouncementUseCase 创建Public Announcement用例
func NewAnnouncementUseCase(repo AnnouncementRepo) *AnnouncementUseCase {
	return &AnnouncementUseCase{repo: repo}
}

// QueryAnnouncement 查询公告列表
func (uc *AnnouncementUseCase) QueryAnnouncement(ctx context.Context, page, size int32, pinned, popup *bool) ([]*Announcement, int32, error) {
	return uc.repo.QueryAnnouncement(ctx, page, size, pinned, popup)
}
