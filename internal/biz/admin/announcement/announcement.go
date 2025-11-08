package announcement

import (
	"context"
	"time"

	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

// Announcement 公告业务实体
type Announcement struct {
	ID        int64     `json:"id"`
	TenantID  int64     `json:"tenant_id"`
	Title     string    `json:"title"`
	Content   *string   `json:"content"`
	Show      bool      `json:"show"`
	Pinned    bool      `json:"pinned"`
	Popup     bool      `json:"popup"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AnnouncementRepo 公告数据仓库接口
type AnnouncementRepo interface {
	// Save 保存公告
	Save(ctx context.Context, announcement *Announcement) (*Announcement, error)
	// Update 更新公告
	Update(ctx context.Context, announcement *Announcement) (*Announcement, error)
	// FindByID 根据ID查找公告
	FindByID(ctx context.Context, tenantID, id int64) (*Announcement, error)
	// ListAll 获取公告列表
	ListAll(ctx context.Context, tenantID int64, page, size int64, show, pinned, popup *bool) ([]*Announcement, int64, error)
	// Delete 删除公告
	Delete(ctx context.Context, tenantID, id int64) error
}

// AnnouncementUsecase 公告业务用例
type AnnouncementUsecase struct {
	repo   AnnouncementRepo
	logger *log.Helper
}

// NewAnnouncementUsecase 创建公告业务用例
func NewAnnouncementUsecase(repo AnnouncementRepo, logger log.Logger) *AnnouncementUsecase {
	return &AnnouncementUsecase{
		repo:   repo,
		logger: log.NewHelper(logger),
	}
}

// CreateAnnouncement 创建公告
func (uc *AnnouncementUsecase) CreateAnnouncement(ctx context.Context, announcement *Announcement) (*Announcement, error) {
	result, err := uc.repo.Save(ctx, announcement)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("create announcement failed: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseInsert)
	}
	return result, nil
}

// UpdateAnnouncement 更新公告
func (uc *AnnouncementUsecase) UpdateAnnouncement(ctx context.Context, announcement *Announcement) (*Announcement, error) {
	// 查找现有公告
	info, err := uc.repo.FindByID(ctx, announcement.TenantID, announcement.ID)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("get announcement error: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	// 更新字段
	info.Title = announcement.Title
	info.Content = announcement.Content
	info.Show = announcement.Show
	info.Pinned = announcement.Pinned
	info.Popup = announcement.Popup

	result, err := uc.repo.Update(ctx, info)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("update announcement error: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
	}
	return result, nil
}

// GetAnnouncement 获取单个公告
func (uc *AnnouncementUsecase) GetAnnouncement(ctx context.Context, tenantID, id int64) (*Announcement, error) {
	announcement, err := uc.repo.FindByID(ctx, tenantID, id)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("get announcement error: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	return announcement, nil
}

// ListAnnouncements 获取公告列表
func (uc *AnnouncementUsecase) ListAnnouncements(ctx context.Context, tenantID, page, size int64, show, pinned, popup *bool) ([]*Announcement, int64, error) {
	// 参数验证和默认值
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}

	list, total, err := uc.repo.ListAll(ctx, tenantID, page, size, show, pinned, popup)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("list announcements error: %v", err)
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	return list, total, nil
}

// DeleteAnnouncement 删除公告
func (uc *AnnouncementUsecase) DeleteAnnouncement(ctx context.Context, tenantID, id int64) error {
	err := uc.repo.Delete(ctx, tenantID, id)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("delete announcement error: %v", err)
		return responsecode.NewKratosError(responsecode.ErrDatabaseDelete)
	}
	return nil
}
