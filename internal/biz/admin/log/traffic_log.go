package log

import (
	"context"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/go-kratos/kratos/v2/log"
)

// TrafficLogRepo 流量日志仓库接口
type TrafficLogRepo interface {
	// FilterTrafficLogDetails 过滤流量日志详情
	FilterTrafficLogDetails(ctx context.Context, tenantID int64, page, size int32, date string, serverID, userID, subscribeID *int64) ([]*ent.ProxyTrafficLog, int64, error)
}

// TrafficLogUsecase 流量日志用例
type TrafficLogUsecase struct {
	repo TrafficLogRepo
	log  *log.Helper
}

// NewTrafficLogUsecase 创建流量日志用例
func NewTrafficLogUsecase(repo TrafficLogRepo, logger log.Logger) *TrafficLogUsecase {
	return &TrafficLogUsecase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "biz/admin/log/traffic")),
	}
}

// FilterTrafficLogDetails 过滤流量日志详情
func (uc *TrafficLogUsecase) FilterTrafficLogDetails(ctx context.Context, tenantID int64, page, size int32, date string, serverID, userID, subscribeID *int64) ([]*ent.ProxyTrafficLog, int64, error) {
	return uc.repo.FilterTrafficLogDetails(ctx, tenantID, page, size, date, serverID, userID, subscribeID)
}
