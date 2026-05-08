package announcement

import (
	"context"
	"time"

	"github.com/OmnTeam/npanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

// Announcement 公告业务实体
type Announcement struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   *string   `json:"content"`
	Show      *bool     `json:"show"`
	Pinned    *bool     `json:"pinned"`
	Popup     *bool     `json:"popup"`
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
	FindByID(ctx context.Context, id int64) (*Announcement, error)
	// ListAll 获取公告列表
	ListAll(ctx context.Context, page, size int, search string, show, pinned, popup *bool) ([]*Announcement, int32, error)
	// Delete 删除公告
	Delete(ctx context.Context, id int64) error
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
	info, err := uc.repo.FindByID(ctx, announcement.ID)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("get announcement error: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	if info == nil {
		return nil, responsecode.NewKratosError(responsecode.ErrAnnouncementNotFound)
	}

	// 更新字段
	info.Title = announcement.Title
	info.Content = announcement.Content
	if announcement.Show != nil {
		info.Show = announcement.Show
	}
	if announcement.Pinned != nil {
		info.Pinned = announcement.Pinned
	}
	if announcement.Popup != nil {
		info.Popup = announcement.Popup
	}

	result, err := uc.repo.Update(ctx, info)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("update announcement error: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
	}
	return result, nil
}

// GetAnnouncement 获取单个公告
func (uc *AnnouncementUsecase) GetAnnouncement(ctx context.Context, id int64) (*Announcement, error) {
	announcement, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("get announcement error: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	if announcement == nil {
		return nil, responsecode.NewKratosError(responsecode.ErrAnnouncementNotFound)
	}
	return announcement, nil
}

// ListAnnouncements 获取公告列表
func (uc *AnnouncementUsecase) ListAnnouncements(ctx context.Context, page, size int, search string, show, pinned, popup *bool) ([]*Announcement, int32, error) {
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

	list, total, err := uc.repo.ListAll(ctx, page, size, search, show, pinned, popup)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("list announcements error: %v", err)
		return nil, 0, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	return list, total, nil
}

// DeleteAnnouncement 删除公告
func (uc *AnnouncementUsecase) DeleteAnnouncement(ctx context.Context, id int64) error {
	info, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("get announcement error: %v", err)
		return responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}
	if info == nil {
		return responsecode.NewKratosError(responsecode.ErrAnnouncementNotFound)
	}

	err = uc.repo.Delete(ctx, id)
	if err != nil {
		uc.logger.WithContext(ctx).Errorf("delete announcement error: %v", err)
		return responsecode.NewKratosError(responsecode.ErrDatabaseDelete)
	}
	return nil
}
