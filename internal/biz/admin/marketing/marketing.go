package marketing

import (
	"context"

	"github.com/OmnTeam/npanel-pro/ent"
	"github.com/go-kratos/kratos/v2/log"
)

// MarketingRepo 营销仓库接口
type MarketingRepo interface {
	// CreateBatchSendEmailTask 创建批量发送邮件任务
	CreateBatchSendEmailTask(ctx context.Context, subject, content string, scope uint32,
		registerStartTime, registerEndTime int64, additional string, scheduled int64, interval uint32, limit uint64) error

	// GetBatchSendEmailTaskList 获取批量发送邮件任务列表
	GetBatchSendEmailTaskList(ctx context.Context, page, size int32, scope, status *uint32) ([]*ent.ProxyTask, int32, error)

	// StopBatchSendEmailTask 停止批量发送邮件任务
	StopBatchSendEmailTask(ctx context.Context, id int) error

	// GetPreSendEmailCount 获取预发送邮件数量
	GetPreSendEmailCount(ctx context.Context, scope uint32, registerStartTime, registerEndTime int64) (int64, error)

	// GetBatchSendEmailTaskStatus 获取批量发送邮件任务状态
	GetBatchSendEmailTaskStatus(ctx context.Context, id int) (*ent.ProxyTask, error)

	// CreateQuotaTask 创建配额任务
	CreateQuotaTask(ctx context.Context, subscribers []int, isActive *bool,
		startTime, endTime int64, resetTraffic bool, days uint64, giftType uint32, giftValue uint64) error

	// QueryQuotaTaskPreCount 查询配额任务预计数量
	QueryQuotaTaskPreCount(ctx context.Context, subscribers []int, isActive *bool, startTime, endTime int64) (int64, error)

	// QueryQuotaTaskList 查询配额任务列表
	QueryQuotaTaskList(ctx context.Context, page, size int32, status *uint32) ([]*ent.ProxyTask, int32, error)
}

// MarketingUsecase 营销用例
type MarketingUsecase struct {
	repo MarketingRepo
	log  *log.Helper
}

// NewMarketingUsecase 创建营销用例
func NewMarketingUsecase(repo MarketingRepo, logger log.Logger) *MarketingUsecase {
	return &MarketingUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// ========== Email Task Methods ==========

// CreateBatchSendEmailTask 创建批量发送邮件任务
func (uc *MarketingUsecase) CreateBatchSendEmailTask(ctx context.Context, subject, content string, scope uint32,
	registerStartTime, registerEndTime int64, additional string, scheduled int64, interval uint32, limit uint64) error {
	return uc.repo.CreateBatchSendEmailTask(ctx, subject, content, scope, registerStartTime, registerEndTime, additional, scheduled, interval, limit)
}

// GetBatchSendEmailTaskList 获取批量发送邮件任务列表
func (uc *MarketingUsecase) GetBatchSendEmailTaskList(ctx context.Context, page, size int32, scope, status *uint32) ([]*ent.ProxyTask, int32, error) {
	return uc.repo.GetBatchSendEmailTaskList(ctx, page, size, scope, status)
}

// StopBatchSendEmailTask 停止批量发送邮件任务
func (uc *MarketingUsecase) StopBatchSendEmailTask(ctx context.Context, id int) error {
	return uc.repo.StopBatchSendEmailTask(ctx, id)
}

// GetPreSendEmailCount 获取预发送邮件数量
func (uc *MarketingUsecase) GetPreSendEmailCount(ctx context.Context, scope uint32, registerStartTime, registerEndTime int64) (int64, error) {
	return uc.repo.GetPreSendEmailCount(ctx, scope, registerStartTime, registerEndTime)
}

// GetBatchSendEmailTaskStatus 获取批量发送邮件任务状态
func (uc *MarketingUsecase) GetBatchSendEmailTaskStatus(ctx context.Context, id int) (*ent.ProxyTask, error) {
	return uc.repo.GetBatchSendEmailTaskStatus(ctx, id)
}

// ========== Quota Task Methods ==========

// CreateQuotaTask 创建配额任务
func (uc *MarketingUsecase) CreateQuotaTask(ctx context.Context, subscribers []int, isActive *bool,
	startTime, endTime int64, resetTraffic bool, days uint64, giftType uint32, giftValue uint64) error {
	return uc.repo.CreateQuotaTask(ctx, subscribers, isActive, startTime, endTime, resetTraffic, days, giftType, giftValue)
}

// QueryQuotaTaskPreCount 查询配额任务预计数量
func (uc *MarketingUsecase) QueryQuotaTaskPreCount(ctx context.Context, subscribers []int, isActive *bool, startTime, endTime int64) (int64, error) {
	return uc.repo.QueryQuotaTaskPreCount(ctx, subscribers, isActive, startTime, endTime)
}

// QueryQuotaTaskList 查询配额任务列表
func (uc *MarketingUsecase) QueryQuotaTaskList(ctx context.Context, page, size int32, status *uint32) ([]*ent.ProxyTask, int32, error) {
	return uc.repo.QueryQuotaTaskList(ctx, page, size, status)
}
