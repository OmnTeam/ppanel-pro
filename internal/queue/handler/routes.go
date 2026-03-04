package handler

import (
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/internal/conf"
	queueTypes "github.com/OmnTeam/ppanel-pro/internal/queue/types"
	"github.com/OmnTeam/ppanel-pro/internal/service"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// RegisterHandlers 注册所有任务处理器
// 所有handler从数据库根据租户ID获取配置，不再依赖全局配置
func RegisterHandlers(mux *asynq.ServeMux, db *ent.Client, rdb *redis.Client, queue *asynq.Client, config *conf.Application, cacheService *service.CacheService, logger log.Logger) {
	// 注册批量邮件任务处理器（从数据库获取配置）
	mux.Handle(queueTypes.ScheduledBatchSendEmail, NewBatchEmailHandler(db, logger))

	// 注册定时检查订阅状态任务处理器（定时任务：检查流量用尽和过期的订阅）
	mux.Handle(queueTypes.SchedulerCheckSubscription, NewCheckSubscriptionHandler(db, queue, logger))

	// 注册定时重置流量任务处理器（定时任务：支持三种重置模式 - 月初/按月/按年）
	mux.Handle(queueTypes.SchedulerResetTraffic, NewResetTrafficHandler(db, rdb, queue, logger))

	// 注册配额任务处理器（从数据库获取配置）
	mux.Handle(queueTypes.ForthwithQuotaTask, NewQuotaTaskHandler(db, logger))

	// 注册立即发送邮件任务处理器（从数据库获取配置）
	mux.Handle(queueTypes.ForthwithSendEmail, NewSendEmailHandler(db, logger))

	// 注册立即发送短信任务处理器（从数据库获取配置）
	mux.Handle(queueTypes.ForthwithSendSms, NewSendSmsHandler(db, logger))

	// 注册延迟关闭订单任务处理器（Portal订单15分钟超时关闭）
	mux.Handle(queueTypes.DeferCloseOrder, NewCloseOrderHandler(db, logger, cacheService))

	// 注册激活订单任务处理器（Portal订单用户创建 + 订单激活）
	mux.Handle(queueTypes.ForthwithActivateOrder, NewActivateOrderHandler(db, rdb, logger))

	// 注册定时获取服务器数据任务处理器（统计服务器和用户流量排行榜）
	mux.Handle(queueTypes.SchedulerTotalServerData, NewServerDataHandler(db, rdb, logger))

	// 注册定时流量统计任务处理器（每日流量统计和排行榜）
	mux.Handle(queueTypes.SchedulerTrafficStat, NewTrafficStatHandler(db, rdb, logger))

	// 注册立即流量统计任务处理器（处理节点服务器上报的实时流量）
	mux.Handle(queueTypes.ForthwithTrafficStatistics, NewTrafficStatisticsHandler(db, logger))
}
