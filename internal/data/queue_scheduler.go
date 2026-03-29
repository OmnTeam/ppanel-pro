package data

import (
	"time"

	queueTypes "github.com/OmnTeam/ppanel-pro/internal/queue/types"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

func newQueueScheduler(redisOpt asynq.RedisConnOpt) *asynq.Scheduler {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		location = time.FixedZone("CST", 8*60*60)
	}
	return asynq.NewScheduler(
		redisOpt,
		&asynq.SchedulerOpts{
			Location: location,
		},
	)
}

func startQueueScheduler(scheduler *asynq.Scheduler, logger log.Logger) error {
	if scheduler == nil {
		return nil
	}

	helper := log.NewHelper(logger)

	checkTask := asynq.NewTask(queueTypes.SchedulerCheckSubscription, nil)
	if _, err := scheduler.Register("@every 60s", checkTask); err != nil {
		helper.Errorf("register check subscription task failed: %v", err)
	}

	resetTrafficTask := asynq.NewTask(queueTypes.SchedulerResetTraffic, nil)
	if _, err := scheduler.Register("30 0 * * *", resetTrafficTask); err != nil {
		helper.Errorf("register reset traffic task failed: %v", err)
	}

	trafficStatTask := asynq.NewTask(queueTypes.SchedulerTrafficStat, nil)
	if _, err := scheduler.Register("0 0 * * *", trafficStatTask, asynq.MaxRetry(3)); err != nil {
		helper.Errorf("register traffic stat task failed: %v", err)
	}

	// 对齐老项目实际代码：01:00 入队的是 ForthwithQuotaTask，而不是 SchedulerExchangeRate。
	quotaTask := asynq.NewTask(queueTypes.ForthwithQuotaTask, nil)
	if _, err := scheduler.Register("0 1 * * *", quotaTask, asynq.MaxRetry(3)); err != nil {
		helper.Errorf("register quota task failed: %v", err)
	}

	recalculateGroupTask := asynq.NewTask(queueTypes.SchedulerRecalculateGroup, nil)
	if _, err := scheduler.Register("@every 6h", recalculateGroupTask, asynq.MaxRetry(2)); err != nil {
		helper.Errorf("register recalculate group task failed: %v", err)
	}

	return scheduler.Run()
}
