package ads

import (
	"context"
	"time"

	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
	"github.com/go-kratos/kratos/v2/log"
)

// Ads 广告实体
type Ads struct {
	ID          int64     `json:"id"`
	TenantID    int64     `json:"tenant_id"`
	Title       string    `json:"title"`
	Type        string    `json:"type"`
	Content     string    `json:"content"`
	Description string    `json:"description"`
	TargetURL   string    `json:"target_url"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AdsFilter 广告过滤条件
type AdsFilter struct {
	Search   string `json:"search"`
	Status   *int8  `json:"status"`
	TenantID int64  `json:"tenant_id"`
}

// AdsRepo 广告数据仓库接口
type AdsRepo interface {
	GetAdsListByPage(ctx context.Context, page, size int64, filter AdsFilter) (total int64, list []*Ads, err error)
	GetAdsByID(ctx context.Context, tenantID, id int64) (*Ads, error)
	CreateAds(ctx context.Context, ads *Ads) (*Ads, error)
	UpdateAds(ctx context.Context, ads *Ads) (*Ads, error)
	DeleteAds(ctx context.Context, tenantID, id int64) error
}

// AdsUsecase 广告用例
type AdsUsecase struct {
	repo AdsRepo
	log  *log.Helper
}

// NewAdsUsecase 创建广告用例
func NewAdsUsecase(repo AdsRepo, logger log.Logger) *AdsUsecase {
	return &AdsUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// GetAdsListByPage 分页获取广告列表
func (uc *AdsUsecase) GetAdsListByPage(ctx context.Context, tenantID, page, size int64, filter AdsFilter) (total int64, list []*Ads, err error) {
	filter.TenantID = tenantID

	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	total, list, err = uc.repo.GetAdsListByPage(ctx, page, size, filter)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to get ads list: %v", err)
		return 0, nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	return total, list, nil
}

// GetAdsByID 根据ID获取广告详情
func (uc *AdsUsecase) GetAdsByID(ctx context.Context, tenantID, id int64) (*Ads, error) {
	if id <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	ads, err := uc.repo.GetAdsByID(ctx, tenantID, id)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("Failed to get ads by id %d: %v", id, err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	return ads, nil
}

// CreateAds 创建广告
func (uc *AdsUsecase) CreateAds(ctx context.Context, ads *Ads) (*Ads, error) {
	if ads == nil {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	if ads.Title == "" {
		return nil, responsecode.NewKratosError(responsecode.ErrTitleRequired)
	}

	if ads.Type == "" {
		return nil, responsecode.NewKratosError(responsecode.ErrTypeRequired)
	}

	// 验证时间逻辑
	if !ads.EndTime.IsZero() && !ads.StartTime.IsZero() && ads.EndTime.Before(ads.StartTime) {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidTimeRange)
	}

	result, err := uc.repo.CreateAds(ctx, ads)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("insert ads error: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseInsert)
	}

	uc.log.WithContext(ctx).Infof("Created ads successfully, tenant_id: %d, ads_id: %d", ads.TenantID, result.ID)
	return result, nil
}

// UpdateAds 更新广告
func (uc *AdsUsecase) UpdateAds(ctx context.Context, ads *Ads) (*Ads, error) {
	if ads == nil || ads.ID <= 0 {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	if ads.Title == "" {
		return nil, responsecode.NewKratosError(responsecode.ErrTitleRequired)
	}

	if ads.Type == "" {
		return nil, responsecode.NewKratosError(responsecode.ErrTypeRequired)
	}

	// 验证时间逻辑
	if !ads.EndTime.IsZero() && !ads.StartTime.IsZero() && ads.EndTime.Before(ads.StartTime) {
		return nil, responsecode.NewKratosError(responsecode.ErrInvalidTimeRange)
	}

	// 先查询现有数据
	data, err := uc.repo.GetAdsByID(ctx, ads.TenantID, ads.ID)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("find ads error: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseQuery)
	}

	if data == nil {
		return nil, responsecode.NewKratosError(responsecode.ErrAdsNotFound)
	}

	// 更新字段
	data.Title = ads.Title
	data.Type = ads.Type
	data.Content = ads.Content
	data.Description = ads.Description
	data.TargetURL = ads.TargetURL
	data.StartTime = ads.StartTime
	data.EndTime = ads.EndTime
	data.Status = ads.Status

	result, err := uc.repo.UpdateAds(ctx, data)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("update ads error: %v", err)
		return nil, responsecode.NewKratosError(responsecode.ErrDatabaseUpdate)
	}

	uc.log.WithContext(ctx).Infof("Updated ads successfully, tenant_id: %d, ads_id: %d", ads.TenantID, ads.ID)
	return result, nil
}

// DeleteAds 删除广告
func (uc *AdsUsecase) DeleteAds(ctx context.Context, tenantID, id int64) error {
	if id <= 0 {
		return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	err := uc.repo.DeleteAds(ctx, tenantID, id)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("delete ads error: %v", err)
		return responsecode.NewKratosError(responsecode.ErrDatabaseDelete)
	}

	uc.log.WithContext(ctx).Infof("Deleted ads successfully, tenant_id: %d, ads_id: %d", tenantID, id)
	return nil
}
