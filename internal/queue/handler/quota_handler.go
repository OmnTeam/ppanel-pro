package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxytask"
	logmodel "github.com/OmnTeam/ppanel-pro/internal/model/log"
	taskmodel "github.com/OmnTeam/ppanel-pro/internal/model/task"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
)

// QuotaTaskHandler 配额任务处理器
type QuotaTaskHandler struct {
	db  *ent.Client
	log *log.Helper
}

// NewQuotaTaskHandler 创建配额任务处理器
// 所有配置从数据库根据租户ID获取
func NewQuotaTaskHandler(db *ent.Client, logger log.Logger) *QuotaTaskHandler {
	return &QuotaTaskHandler{
		db:  db,
		log: log.NewHelper(logger),
	}
}

// ProcessTask 处理配额任务 - 实现 asynq.Handler 接口
func (h *QuotaTaskHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	// 解析任务ID
	taskID, err := h.parseTaskID(ctx, task.Payload())
	if err != nil {
		return err
	}

	// 查询任务信息
	taskInfo, err := h.getTaskInfo(ctx, taskID)
	if err != nil {
		return err
	}

	// 检查任务状态
	if taskInfo.Status != int8(taskmodel.StatusPending) {
		h.log.WithContext(ctx).Infof("[QuotaTaskHandler.ProcessTask] task already processed, taskID: %d, status: %d",
			taskID, taskInfo.Status)
		return nil
	}

	h.log.WithContext(ctx).Infof("[QuotaTaskHandler] Starting to process quota task %d", taskID)

	// 解析scope
	scope, err := taskmodel.UnmarshalQuotaScope(taskInfo.Scope)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[QuotaTaskHandler] Failed to unmarshal scope: %v", err)
		return asynq.SkipRetry
	}

	// 解析content
	content, err := taskmodel.UnmarshalQuotaContent(taskInfo.Content)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[QuotaTaskHandler] Failed to unmarshal content: %v", err)
		return asynq.SkipRetry
	}

	// 获取用户订阅列表
	if len(scope.Objects) == 0 {
		h.log.WithContext(ctx).Errorf("[QuotaTaskHandler] No objects found in scope for task %d", taskID)
		// 标记任务为完成
		_ = h.db.ProxyTask.UpdateOneID(taskID).
			SetStatus(int8(taskmodel.StatusCompleted)).
			Exec(ctx)
		return nil
	}

	// 更新任务状态为进行中
	err = h.db.ProxyTask.UpdateOneID(taskID).
		SetStatus(int8(taskmodel.StatusInProgress)).
		Exec(ctx)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[QuotaTaskHandler] Failed to update task status to InProgress: %v", err)
		return err
	}

	// 处理每个用户订阅
	var count uint64
	var errors []string

	for _, subID := range scope.Objects {
		select {
		case <-ctx.Done():
			h.log.WithContext(ctx).Infof("[QuotaTaskHandler] Worker stopped by context cancellation, taskID: %d", taskID)
			// 更新任务状态
			_ = h.db.ProxyTask.UpdateOneID(taskID).
				SetCurrent(count).
				Exec(ctx)
			return nil
		default:
		}

		// 查询用户订阅
		userSub, err := h.db.ProxyUserSubscribe.Get(ctx, int(subID))
		if err != nil {
			h.log.WithContext(ctx).Errorf("[QuotaTaskHandler] Failed to get user subscribe %d: %v", subID, err)
			errors = append(errors, fmt.Sprintf("subscribe_id:%d, error:%v", subID, err))
			count++
			continue
		}

		// 执行配额操作
		if err := h.processQuotaOperations(ctx, userSub, content); err != nil {
			h.log.WithContext(ctx).Errorf("[QuotaTaskHandler] Failed to process quota for subscribe %d: %v", subID, err)
			errors = append(errors, fmt.Sprintf("subscribe_id:%d, error:%v", subID, err))
		}

		count++

		// 更新任务进度
		errorJSON := ""
		if len(errors) > 0 {
			errBytes, _ := json.Marshal(errors)
			errorJSON = string(errBytes)
		}

		err = h.db.ProxyTask.UpdateOneID(taskID).
			SetCurrent(count).
			SetErrors(errorJSON).
			Exec(ctx)
		if err != nil {
			h.log.WithContext(ctx).Errorf("[QuotaTaskHandler] Failed to update task progress: %v", err)
		}
	}

	// TODO: 清理缓存（复刻原项目 quotaLogic.go:76-102）
	// 如果有赠送金，需要清理用户缓存（因为gift_amount字段改变了）
	// 无论是否有赠送金，都需要清理用户订阅缓存
	//
	// 实现示例：
	// if content.GiftValue != 0 {
	//     // 收集所有受影响的用户ID
	//     userIDs := make(map[int64]bool)
	//     for _, subID := range scope.Objects {
	//         userSub, _ := h.db.ProxyUserSubscribe.Get(ctx, subID)
	//         userIDs[userSub.UserID] = true
	//     }
	//
	//     // 清理用户缓存
	//     for userID := range userIDs {
	//         clearUserCache(ctx, userID)
	//     }
	// }
	//
	// // 清理所有用户订阅缓存
	// for _, subID := range scope.Objects {
	//     clearUserSubscribeCache(ctx, subID)
	// }

	// 标记任务为完成
	err = h.db.ProxyTask.UpdateOneID(taskID).
		SetStatus(int8(taskmodel.StatusCompleted)).
		SetCurrent(count).
		Exec(ctx)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[QuotaTaskHandler] Failed to update task status to Completed: %v", err)
		return err
	}

	h.log.WithContext(ctx).Infof("[QuotaTaskHandler] Successfully completed quota task %d, processed %d/%d subscriptions",
		taskID, count-uint64(len(errors)), count)
	return nil
}

// processQuotaOperations 处理配额操作
func (h *QuotaTaskHandler) processQuotaOperations(ctx context.Context, userSub *ent.ProxyUserSubscribe, content *taskmodel.QuotaContent) error {
	// 1. 延长天数
	if content.Days > 0 {
		newExpireTime := userSub.ExpireTime.Add(time.Duration(content.Days) * 24 * time.Hour)
		if err := h.db.ProxyUserSubscribe.UpdateOneID(userSub.ID).
			SetExpireTime(newExpireTime).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to extend days: %w", err)
		}
		h.log.Infof("[QuotaTaskHandler] Extended %d days for subscribe %d, new expire time: %s",
			content.Days, userSub.ID, newExpireTime.Format(time.DateTime))
	}

	// 2. 重置流量
	if content.ResetTraffic {
		if err := h.db.ProxyUserSubscribe.UpdateOneID(userSub.ID).
			SetUpload(0).
			SetDownload(0).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to reset traffic: %w", err)
		}

		// 记录流量重置日志
		if err := h.logTrafficReset(ctx, 0, int64(userSub.ID), int64(userSub.UserID)); err != nil {
			h.log.Warnf("[QuotaTaskHandler] Failed to log traffic reset: %v", err)
		}

		h.log.Infof("[QuotaTaskHandler] Reset traffic for subscribe %d", userSub.ID)
	}

	// 3. 赠送余额（gift_amount）- 修复：原逻辑错误地赠送流量，应该赠送余额
	// 复刻原项目逻辑：server-master/queue/logic/task/quotaLogic.go:295-357
	if content.GiftValue > 0 {
		// 3.1 查询用户信息
		user, err := h.db.ProxyUser.Get(ctx, userSub.UserID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		var giftAmount int64

		if content.GiftType == 1 { // Fixed: 固定金额（分/cents）
			giftAmount = int64(content.GiftValue)
			h.log.Infof("[QuotaTaskHandler] Gift type: fixed amount, value: %d cents", giftAmount)

		} else if content.GiftType == 2 { // Ratio: 按订阅套餐价格比例
			// 查询订阅套餐信息
			subscribe, err := h.db.ProxySubscribe.Get(ctx, userSub.SubscribeID)
			if err != nil {
				return fmt.Errorf("failed to get subscribe: %w", err)
			}

			// 按unit_price比例计算
			// 注意：UnitPrice是int64类型（非指针），有default(0)
			if subscribe.UnitPrice > 0 {
				giftAmount = int64(float64(subscribe.UnitPrice) * (float64(content.GiftValue) / 100.0))
				h.log.Infof("[QuotaTaskHandler] Gift type: ratio %d%%, unitPrice: %d, calculated amount: %d cents",
					content.GiftValue, subscribe.UnitPrice, giftAmount)
			} else {
				h.log.Warnf("[QuotaTaskHandler] Subscribe unit_price is 0, skip gift")
				return nil
			}
		}

		if giftAmount > 0 {
			// 3.2 更新用户的gift_amount（余额）
			currentGiftAmount := int64(0)
			if user.GiftAmount != nil {
				currentGiftAmount = *user.GiftAmount
			}
			newGiftAmount := currentGiftAmount + giftAmount

			err = h.db.ProxyUser.UpdateOneID(user.ID).
				SetGiftAmount(newGiftAmount).
				Exec(ctx)
			if err != nil {
				return fmt.Errorf("failed to update user gift_amount: %w", err)
			}

			// 3.3 记录gift日志（object_id应该是user_id，不是subscribe_id）
			giftLog := map[string]interface{}{
				"type":         1, // GiftTypeIncrease = 1
				"order_no":     "",
				"subscribe_id": userSub.ID,
				"amount":       giftAmount,
				"balance":      newGiftAmount,
				"remark":       "Quota task gift",
				"timestamp":    time.Now().UnixMilli(),
			}
			giftLogJSON, err := json.Marshal(giftLog)
			if err != nil {
				h.log.Warnf("[QuotaTaskHandler] Failed to marshal gift log: %v", err)
			} else {
				_, err = h.db.ProxySystemLog.Create().
					SetType(int8(34)). // TypeGift = 34
					SetDate(time.Now().Format(time.DateOnly)).
					SetObjectID(int64(user.ID)). // 注意：object_id是user_id，不是subscribe_id
					SetContent(string(giftLogJSON)).
					Save(ctx)

				if err != nil {
					h.log.Warnf("[QuotaTaskHandler] Failed to log gift: %v", err)
				}
			}

			h.log.Infof("[QuotaTaskHandler] Gifted %d cents to user %d (subscribe %d), oldBalance=%d, newBalance=%d",
				giftAmount, user.ID, userSub.ID, currentGiftAmount, newGiftAmount)

			// 3.4 TODO: 清理用户缓存（原项目 quotaLogic.go:76-95）
			// 需要在任务完成后清理所有相关用户的缓存
		}
	}

	return nil
}

// logTrafficReset 记录流量重置日志
func (h *QuotaTaskHandler) logTrafficReset(ctx context.Context, tenantID, subscribeID, userID int64) error {
	logContent := map[string]interface{}{
		"type":         "reset_traffic",
		"subscribe_id": subscribeID,
		"user_id":      userID,
		"time":         time.Now().Unix(),
	}

	contentJSON, err := json.Marshal(logContent)
	if err != nil {
		return err
	}

	_, err = h.db.ProxySystemLog.Create().
		SetType(int8(logmodel.TypeResetSubscribe)).
		SetDate(time.Now().Format(time.DateOnly)).
		SetObjectID(subscribeID).
		SetContent(string(contentJSON)).
		Save(ctx)

	return err
}

// parseTaskID 解析任务ID
func (h *QuotaTaskHandler) parseTaskID(ctx context.Context, payload []byte) (int64, error) {
	if len(payload) == 0 {
		h.log.WithContext(ctx).Error("[QuotaTaskHandler.parseTaskID] empty payload")
		return 0, asynq.SkipRetry
	}

	taskID, err := strconv.ParseInt(string(payload), 10, 64)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[QuotaTaskHandler.parseTaskID] invalid task ID, error: %v, payload: %s",
			err, string(payload))
		return 0, asynq.SkipRetry
	}
	return taskID, nil
}

// getTaskInfo 获取任务信息
func (h *QuotaTaskHandler) getTaskInfo(ctx context.Context, taskID int64) (*ent.ProxyTask, error) {
	taskInfo, err := h.db.ProxyTask.Query().
		Where(proxytask.ID(taskID)).
		Only(ctx)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[QuotaTaskHandler.getTaskInfo] find task error, taskID: %d, error: %v",
			taskID, err)
		return nil, asynq.SkipRetry
	}
	return taskInfo, nil
}
