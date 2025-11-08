package handler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysystem"
	"github.com/OmnTeam/ppanel-pro/ent/proxyuserauthmethod"
	"github.com/OmnTeam/ppanel-pro/ent/proxyusersubscribe"
	queueTypes "github.com/OmnTeam/ppanel-pro/internal/queue/types"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

// CheckSubscriptionHandler 检查订阅状态处理器
// 定时任务：检查流量用尽和过期的订阅
// ⚠️ 完整复刻原项目逻辑（checkSubscriptionLogic.go）
type CheckSubscriptionHandler struct {
	db     *ent.Client
	queue  *asynq.Client
	logger *log.Helper
}

// NewCheckSubscriptionHandler 创建检查订阅处理器
func NewCheckSubscriptionHandler(db *ent.Client, queue *asynq.Client, logger log.Logger) *CheckSubscriptionHandler {
	return &CheckSubscriptionHandler{
		db:     db,
		queue:  queue,
		logger: log.NewHelper(logger),
	}
}

// ProcessTask 处理任务
// 1. 检查流量用尽的订阅 (upload+download >= traffic)
// 2. 检查过期的订阅 (expire_time < now)
// 复刻原项目逻辑：server-master/queue/logic/subscription/checkSubscriptionLogic.go
func (h *CheckSubscriptionHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	h.logger.Infof("[CheckSubscription] 开始检查订阅: %s", time.Now().Format("2006-01-02 15:04:05"))

	// 1. 检查流量用尽的订阅（复刻原项目 line 30-71）
	if err := h.checkTrafficExceeded(ctx); err != nil {
		h.logger.Errorf("[CheckSubscription] 检查流量失败: %v", err)
	}

	// 2. 检查过期的订阅（复刻原项目 line 75-116）
	if err := h.checkExpired(ctx); err != nil {
		h.logger.Errorf("[CheckSubscription] 检查过期失败: %v", err)
	}

	return nil
}

// checkTrafficExceeded 检查流量用尽的订阅
// 复刻原项目 line 31-71
func (h *CheckSubscriptionHandler) checkTrafficExceeded(ctx context.Context) error {
	h.logger.Info("[CheckSubscription] 开始检查流量用尽订阅")

	// 查询流量用尽的订阅
	// upload + download >= traffic AND status IN (0, 1) AND traffic > 0
	// status: 0=未激活, 1=激活中
	list, err := h.db.ProxyUserSubscribe.Query().
		Where(
			// traffic > 0 表示有流量限制
			proxyusersubscribe.TrafficGT(0),
			// status IN (0, 1)
			proxyusersubscribe.StatusIn(0, 1),
		).
		All(ctx)

	if err != nil {
		h.logger.Errorf("[CheckSubscription] 查询订阅失败: %v", err)
		return err
	}

	// 过滤出流量已用尽的订阅
	var trafficExceededSubs []*ent.ProxyUserSubscribe
	for _, sub := range list {
		if sub.Traffic == nil {
			continue
		}
		used := int64(0)
		if sub.Upload != nil {
			used += *sub.Upload
		}
		if sub.Download != nil {
			used += *sub.Download
		}
		// 已用流量 >= 总流量
		if used >= *sub.Traffic {
			trafficExceededSubs = append(trafficExceededSubs, sub)
		}
	}

	if len(trafficExceededSubs) == 0 {
		h.logger.Info("[CheckSubscription] 没有流量用尽的订阅")
		return nil
	}

	// 收集ID
	var ids []int
	for _, sub := range trafficExceededSubs {
		ids = append(ids, sub.ID)
	}

	// 使用事务更新订阅状态
	err = h.db.TX(ctx, func(tx *ent.Tx) error {
		now := time.Now()
		// 批量更新状态为2（流量用尽）
		_, err := tx.ProxyUserSubscribe.Update().
			Where(proxyusersubscribe.IDIn(ids...)).
			SetStatus(2).       // status=2 表示流量用尽
			SetFinishedAt(now). // 设置finished_at
			Save(ctx)

		if err != nil {
			h.logger.Errorf("[CheckSubscription] 更新订阅状态失败: %v", err)
			return err
		}

		h.logger.Infof("[CheckSubscription] 更新流量用尽订阅状态成功: count=%d, ids=%v", len(ids), ids)
		return nil
	})

	if err != nil {
		return err
	}

	// 发送流量用尽通知（异步）
	if err := h.sendTrafficNotify(ctx, trafficExceededSubs); err != nil {
		h.logger.Warnf("[CheckSubscription] 发送流量通知失败: %v", err)
	}

	// 清理缓存（复刻原项目 line 58-63）
	// TODO: 需要实现用户订阅缓存清理 ClearSubscribeCache(ctx, trafficExceededSubs)
	h.clearServerCache(ctx, trafficExceededSubs)

	return nil
}

// checkExpired 检查过期的订阅
// 复刻原项目 line 76-116
func (h *CheckSubscriptionHandler) checkExpired(ctx context.Context) error {
	h.logger.Info("[CheckSubscription] 开始检查过期订阅")

	// 查询过期的订阅
	// status IN (0, 1) AND expire_time < now AND expire_time != 0 AND finished_at IS NULL
	zeroTime := time.UnixMilli(0)
	now := time.Now()

	list, err := h.db.ProxyUserSubscribe.Query().
		Where(
			// status IN (0, 1)
			proxyusersubscribe.StatusIn(0, 1),
			// expire_time < now
			proxyusersubscribe.ExpireTimeLT(now),
			// expire_time != 0 (排除永久订阅)
			proxyusersubscribe.ExpireTimeNEQ(zeroTime),
			// finished_at IS NULL
			proxyusersubscribe.FinishedAtIsNil(),
		).
		All(ctx)

	if err != nil {
		h.logger.Errorf("[CheckSubscription] 查询过期订阅失败: %v", err)
		return err
	}

	if len(list) == 0 {
		h.logger.Info("[CheckSubscription] 没有过期的订阅")
		return nil
	}

	// 收集ID
	var ids []int
	for _, sub := range list {
		ids = append(ids, sub.ID)
	}

	// 使用事务更新订阅状态
	err = h.db.TX(ctx, func(tx *ent.Tx) error {
		now := time.Now()
		// 批量更新状态为3（已过期）
		_, err := tx.ProxyUserSubscribe.Update().
			Where(proxyusersubscribe.IDIn(ids...)).
			SetStatus(3).       // status=3 表示已过期
			SetFinishedAt(now). // 设置finished_at
			Save(ctx)

		if err != nil {
			h.logger.Errorf("[CheckSubscription] 更新过期订阅状态失败: %v", err)
			return err
		}

		h.logger.Infof("[CheckSubscription] 更新过期订阅状态成功: count=%d, ids=%v", len(ids), ids)
		return nil
	})

	if err != nil {
		return err
	}

	// 发送过期通知（异步）
	if err := h.sendExpiredNotify(ctx, list); err != nil {
		h.logger.Warnf("[CheckSubscription] 发送过期通知失败: %v", err)
	}

	// 清理缓存（复刻原项目 line 101-105）
	// TODO: 需要实现用户订阅缓存清理 ClearSubscribeCache(ctx, list)
	h.clearServerCache(ctx, list)

	return nil
}

// sendTrafficNotify 发送流量用尽通知
// 复刻原项目 line 159-196
func (h *CheckSubscriptionHandler) sendTrafficNotify(ctx context.Context, subs []*ent.ProxyUserSubscribe) error {
	for _, sub := range subs {
		// 1. 查询用户邮箱（复刻原项目 line 166-170）
		method, err := h.db.ProxyUserAuthMethod.Query().
			Where(
				proxyuserauthmethod.UserIDEQ(sub.UserID),
				proxyuserauthmethod.AuthTypeEQ("email"),
			).
			Only(ctx)
		if err != nil {
			h.logger.Warnf("[CheckSubscription] 查询用户邮箱失败: userID=%d, error=%v", sub.UserID, err)
			continue
		}

		// 2. 查询站点配置
		siteConfig, err := h.loadSiteConfig(ctx, 0)
		if err != nil {
			h.logger.Warnf("[CheckSubscription] 查询站点配置失败: error=%v", err)
			// 配置加载失败时使用空字符串，不阻断发送
			siteConfig = map[string]string{"SiteLogo": "", "SiteName": ""}
		}

		// 3. 构建邮件任务payload（复刻原项目 line 171-178）
		payload := queueTypes.SendEmailPayload{
			TenantID: 0,
			Type:     "traffic_exceed", // EmailTypeTrafficExceed
			Email:    method.AuthIdentifier,
			Subject:  "Subscription Traffic Exceed",
			Content: map[string]interface{}{
				"SiteLogo": siteConfig["SiteLogo"],
				"SiteName": siteConfig["SiteName"],
			},
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			h.logger.Warnf("[CheckSubscription] 序列化邮件payload失败: subscribeID=%d, error=%v", sub.ID, err)
			continue
		}

		// 4. 入队邮件发送任务（复刻原项目 line 184-189）
		task := asynq.NewTask(queueTypes.ForthwithSendEmail, payloadBytes, asynq.MaxRetry(3))
		taskInfo, err := h.queue.Enqueue(task)
		if err != nil {
			h.logger.Warnf("[CheckSubscription] 邮件任务入队失败: subscribeID=%d, error=%v", sub.ID, err)
			continue
		}

		h.logger.Infof("[CheckSubscription] 流量通知邮件已入队: subscribeID=%d, userID=%d, email=%s, taskID=%s",
			sub.ID, sub.UserID, method.AuthIdentifier, taskInfo.ID)
	}

	return nil
}

// sendExpiredNotify 发送过期通知
// 复刻原项目 line 119-157
func (h *CheckSubscriptionHandler) sendExpiredNotify(ctx context.Context, subs []*ent.ProxyUserSubscribe) error {
	for _, sub := range subs {
		// 1. 查询用户邮箱（复刻原项目 line 126-130）
		method, err := h.db.ProxyUserAuthMethod.Query().
			Where(
				proxyuserauthmethod.UserIDEQ(sub.UserID),
				proxyuserauthmethod.AuthTypeEQ("email"),
			).
			Only(ctx)
		if err != nil {
			h.logger.Warnf("[CheckSubscription] 查询用户邮箱失败: userID=%d, error=%v", sub.UserID, err)
			continue
		}

		// 2. 查询站点配置
		siteConfig, err := h.loadSiteConfig(ctx, 0)
		if err != nil {
			h.logger.Warnf("[CheckSubscription] 查询站点配置失败: error=%v", err)
			// 配置加载失败时使用空字符串，不阻断发送
			siteConfig = map[string]string{"SiteLogo": "", "SiteName": ""}
		}

		// 3. 格式化过期时间（复刻原项目 line 138）
		expireDate := ""
		if sub.ExpireTime != nil {
			expireDate = sub.ExpireTime.Format("2006-01-02 15:04:05")
		}

		// 4. 构建邮件任务payload（复刻原项目 line 131-139）
		payload := queueTypes.SendEmailPayload{
			TenantID: 0,
			Type:     "expiration", // EmailTypeExpiration
			Email:    method.AuthIdentifier,
			Subject:  "Subscription Expired",
			Content: map[string]interface{}{
				"SiteLogo":   siteConfig["SiteLogo"],
				"SiteName":   siteConfig["SiteName"],
				"ExpireDate": expireDate,
			},
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			h.logger.Warnf("[CheckSubscription] 序列化邮件payload失败: subscribeID=%d, error=%v", sub.ID, err)
			continue
		}

		// 5. 入队邮件发送任务（复刻原项目 line 145-150）
		task := asynq.NewTask(queueTypes.ForthwithSendEmail, payloadBytes, asynq.MaxRetry(3))
		taskInfo, err := h.queue.Enqueue(task)
		if err != nil {
			h.logger.Warnf("[CheckSubscription] 邮件任务入队失败: subscribeID=%d, error=%v", sub.ID, err)
			continue
		}

		h.logger.Infof("[CheckSubscription] 过期通知邮件已入队: subscribeID=%d, userID=%d, email=%s, taskID=%s",
			sub.ID, sub.UserID, method.AuthIdentifier, taskInfo.ID)
	}

	return nil
}

// loadSiteConfig 加载站点配置
// 从proxy_system表读取Site.SiteLogo和Site.SiteName配置
func (h *CheckSubscriptionHandler) loadSiteConfig(ctx context.Context, tenantID int64) (map[string]string, error) {
	result := map[string]string{
		"SiteLogo": "",
		"SiteName": "",
	}

	// 查询Site类别下的所有配置
	configs, err := h.db.ProxySystem.Query().
		Where(
			proxysystem.CategoryEQ("Site"),
			proxysystem.KeyIn("SiteLogo", "SiteName"),
		).
		All(ctx)

	if err != nil {
		return result, err
	}

	// 填充配置值
	for _, config := range configs {
		result[config.Key] = config.Value
	}

	return result, nil
}

// clearServerCache clears server cache by collecting unique subscribe_ids
// 复刻原项目 checkSubscriptionLogic.go line 198-211
func (h *CheckSubscriptionHandler) clearServerCache(ctx context.Context, userSubs []*ent.ProxyUserSubscribe) {
	if len(userSubs) == 0 {
		return
	}

	// Collect unique subscribe_ids
	subs := make(map[int]bool)
	for _, sub := range userSubs {
		if sub.SubscribeID > 0 {
			if _, ok := subs[sub.SubscribeID]; !ok {
				subs[sub.SubscribeID] = true
			}
		}
	}

	// TODO: Clear cache for each subscribe_id (需要缓存模块实现)
	// for subID := range subs {
	//     if err := h.clearSubscriptionCache(ctx, subID); err != nil {
	//         h.logger.Errorf("[CheckSubscription] ClearCache failed: subscribeID=%d, error=%v", subID, err)
	//     }
	// }

	h.logger.Infof("[CheckSubscription] Collected %d unique subscribe_ids for cache clearing", len(subs))
}
