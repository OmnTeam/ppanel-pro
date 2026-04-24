package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxytask"
	"github.com/OmnTeam/ppanel-pro/ent/proxyusersubscribe"
	logmodel "github.com/OmnTeam/ppanel-pro/internal/model/log"
	taskmodel "github.com/OmnTeam/ppanel-pro/internal/model/task"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// QuotaTaskHandler 配额任务处理器
type QuotaTaskHandler struct {
	db  *ent.Client
	rdb *redis.Client
	log *log.Helper
}

// NewQuotaTaskHandler 创建配额任务处理器
func NewQuotaTaskHandler(db *ent.Client, rdb *redis.Client, logger log.Logger) *QuotaTaskHandler {
	return &QuotaTaskHandler{
		db:  db,
		rdb: rdb,
		log: log.NewHelper(logger),
	}
}

type quotaTaskErrorInfo struct {
	UserSubscribeID int64  `json:"user_subscribe_id"`
	Error           string `json:"error"`
}

// ProcessTask 处理配额任务
// 复刻原项目 queue/logic/task/quotaLogic.go 的可观察行为。
func (h *QuotaTaskHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	taskID, err := h.parseTaskID(ctx, task.Payload())
	if err != nil {
		return err
	}

	taskInfo, err := h.getTaskInfo(ctx, taskID)
	if err != nil {
		return err
	}

	if taskInfo.Status != int8(taskmodel.StatusPending) {
		h.log.WithContext(ctx).Infof("[QuotaTaskHandler.ProcessTask] task already processed, taskID: %d, status: %d",
			taskID, taskInfo.Status)
		return nil
	}

	scope, content, err := h.parseTaskData(ctx, taskInfo)
	if err != nil {
		return err
	}

	subscribes, err := h.getSubscribes(ctx, scope.Objects)
	if err != nil {
		return err
	}

	if err = h.processSubscribes(ctx, taskInfo, subscribes, *content); err != nil {
		return err
	}

	h.log.WithContext(ctx).Infof("[QuotaTaskHandler] Successfully completed quota task %d, processed %d subscriptions",
		taskID, len(subscribes))
	return nil
}

func (h *QuotaTaskHandler) parseTaskData(ctx context.Context, taskInfo *ent.ProxyTask) (*taskmodel.QuotaScope, *taskmodel.QuotaContent, error) {
	scope, err := taskmodel.UnmarshalQuotaScope(taskInfo.Scope)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[QuotaTaskHandler.parseTaskData] unmarshal scope error: %v", err)
		return nil, nil, asynq.SkipRetry
	}

	content, err := taskmodel.UnmarshalQuotaContent(taskInfo.Content)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[QuotaTaskHandler.parseTaskData] unmarshal content error: %v", err)
		return nil, nil, asynq.SkipRetry
	}

	return scope, content, nil
}

func (h *QuotaTaskHandler) getSubscribes(ctx context.Context, subscribeIDs []int64) ([]*ent.ProxyUserSubscribe, error) {
	if len(subscribeIDs) == 0 {
		return []*ent.ProxyUserSubscribe{}, nil
	}

	subscribes, err := h.db.ProxyUserSubscribe.Query().
		Where(proxyusersubscribe.IDIn(subscribeIDs...)).
		All(ctx)
	if err != nil {
		h.log.WithContext(ctx).Errorf("[QuotaTaskHandler.getSubscribes] find subscribes error: %v, subscribers=%v",
			err, subscribeIDs)
		return nil, asynq.SkipRetry
	}
	return subscribes, nil
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

func (h *QuotaTaskHandler) processSubscribes(ctx context.Context, taskInfo *ent.ProxyTask, subscribes []*ent.ProxyUserSubscribe, content taskmodel.QuotaContent) error {
	now := time.Now()
	var errorInfos []quotaTaskErrorInfo

	err := h.db.TX(ctx, func(tx *ent.Tx) error {
		for _, sub := range subscribes {
			if err := h.processSubscription(ctx, tx, sub, content, now, &errorInfos); err != nil {
				return err
			}
		}

		status := int8(taskmodel.StatusCompleted)
		if len(errorInfos) > 0 && len(errorInfos) == len(subscribes) {
			status = int8(taskmodel.StatusFailed)
		}

		errorJSON := ""
		if len(errorInfos) > 0 {
			errBytes, err := json.Marshal(errorInfos)
			if err != nil {
				h.log.WithContext(ctx).Errorf("[QuotaTaskHandler.processSubscribes] marshal errors failed: %v", err)
				return err
			}
			errorJSON = string(errBytes)
		}

		if err := tx.ProxyTask.UpdateOneID(taskInfo.ID).
			SetCurrent(uint32(len(subscribes))).
			SetStatus(status).
			SetErrors(errorJSON).
			Exec(ctx); err != nil {
			h.log.WithContext(ctx).Errorf("[QuotaTaskHandler.processSubscribes] update task status error: %v, taskID=%d",
				err, taskInfo.ID)
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (h *QuotaTaskHandler) processSubscription(ctx context.Context, tx *ent.Tx, sub *ent.ProxyUserSubscribe, content taskmodel.QuotaContent, now time.Time, errors *[]quotaTaskErrorInfo) error {
	if sub == nil {
		*errors = append(*errors, quotaTaskErrorInfo{
			UserSubscribeID: 0,
			Error:           "subscription is nil",
		})
		return nil
	}

	updateBuilder := tx.ProxyUserSubscribe.UpdateOneID(sub.ID)
	updated := false

	if content.Days != 0 {
		newExpireTime := now.AddDate(0, 0, int(content.Days))
		if sub.ExpireTime != nil && sub.ExpireTime.Unix() != 0 && sub.ExpireTime.After(now) {
			newExpireTime = sub.ExpireTime.AddDate(0, 0, int(content.Days))
		}
		updateBuilder.SetExpireTime(newExpireTime)
		updated = true

		if sub.Status == nil || *sub.Status != 1 {
			updateBuilder.SetStatus(1)
		}
	}

	if content.ResetTraffic {
		updateBuilder.SetDownload(0).SetUpload(0)
		updated = true

		if err := h.createResetTrafficLog(ctx, tx, sub.ID, sub.UserID, now); err != nil {
			*errors = append(*errors, quotaTaskErrorInfo{
				UserSubscribeID: sub.ID,
				Error:           "create reset traffic log error: " + err.Error(),
			})
		}
	}

	if content.GiftValue != 0 {
		if err := h.processGift(ctx, tx, sub, content, now, errors); err != nil {
			return err
		}
	}

	if !updated {
		return nil
	}

	if err := updateBuilder.Exec(ctx); err != nil {
		*errors = append(*errors, quotaTaskErrorInfo{
			UserSubscribeID: sub.ID,
			Error:           "update subscription error: " + err.Error(),
		})
	}

	return nil
}

func (h *QuotaTaskHandler) processGift(ctx context.Context, tx *ent.Tx, sub *ent.ProxyUserSubscribe, content taskmodel.QuotaContent, now time.Time, errors *[]quotaTaskErrorInfo) error {
	if content.GiftType != 1 && content.GiftType != 2 {
		*errors = append(*errors, quotaTaskErrorInfo{
			UserSubscribeID: sub.ID,
			Error:           fmt.Sprintf("invalid gift type: %d", content.GiftType),
		})
		return nil
	}

	userInfo, err := tx.ProxyUser.Get(ctx, sub.UserID)
	if err != nil {
		*errors = append(*errors, quotaTaskErrorInfo{
			UserSubscribeID: sub.ID,
			Error:           "find user error: " + err.Error(),
		})
		return nil
	}

	var giftAmount int64
	switch content.GiftType {
	case 1:
		giftAmount = int64(content.GiftValue)
	case 2:
		subscribeInfo, err := tx.ProxySubscribe.Get(ctx, sub.SubscribeID)
		if err != nil {
			*errors = append(*errors, quotaTaskErrorInfo{
				UserSubscribeID: sub.ID,
				Error:           "find subscribe error: " + err.Error(),
			})
			return nil
		}
		if subscribeInfo.UnitPrice > 0 {
			giftAmount = int64(float64(subscribeInfo.UnitPrice) * (float64(content.GiftValue) / 100))
		}
	}

	if giftAmount <= 0 {
		return nil
	}

	currentGiftAmount := int64(0)
	if userInfo.GiftAmount != nil {
		currentGiftAmount = *userInfo.GiftAmount
	}
	newGiftAmount := currentGiftAmount + giftAmount

	if err := tx.ProxyUser.UpdateOneID(userInfo.ID).SetGiftAmount(newGiftAmount).Exec(ctx); err != nil {
		*errors = append(*errors, quotaTaskErrorInfo{
			UserSubscribeID: sub.ID,
			Error:           "update user gift amount error: " + err.Error(),
		})
		return nil
	}

	if err := h.createGiftLog(ctx, tx, sub.ID, userInfo.ID, giftAmount, newGiftAmount, now); err != nil {
		*errors = append(*errors, quotaTaskErrorInfo{
			UserSubscribeID: sub.ID,
			Error:           "create gift log error: " + err.Error(),
		})
		if rollbackErr := tx.ProxyUser.UpdateOneID(userInfo.ID).SetGiftAmount(currentGiftAmount).Exec(ctx); rollbackErr != nil {
			h.log.WithContext(ctx).Warnf("[QuotaTaskHandler.processGift] rollback user gift amount failed: %v, userID=%d",
				rollbackErr, userInfo.ID)
		}
	}

	return nil
}

func (h *QuotaTaskHandler) createGiftLog(ctx context.Context, tx *ent.Tx, subscribeID, userID, amount, balance int64, now time.Time) error {
	giftLog := logmodel.Gift{
		Type:        logmodel.GiftTypeIncrease,
		OrderNo:     "",
		SubscribeId: subscribeID,
		Amount:      amount,
		Balance:     balance,
		Remark:      "Quota task gift",
		Timestamp:   now.UnixMilli(),
	}

	contentJSON, err := giftLog.Marshal()
	if err != nil {
		return err
	}

	_, err = tx.ProxySystemLog.Create().
		SetType(int8(logmodel.TypeGift)).
		SetDate(now.Format(time.DateOnly)).
		SetObjectID(userID).
		SetContent(string(contentJSON)).
		Save(ctx)

	return err
}

func (h *QuotaTaskHandler) createResetTrafficLog(ctx context.Context, tx *ent.Tx, subscribeID, userID int64, now time.Time) error {
	trafficLog := logmodel.ResetSubscribe{
		Type:      logmodel.ResetSubscribeTypeQuota,
		UserId:    userID,
		OrderNo:   "",
		Timestamp: now.UnixMilli(),
	}

	contentJSON, err := trafficLog.Marshal()
	if err != nil {
		return err
	}

	_, err = tx.ProxySystemLog.Create().
		SetType(int8(logmodel.TypeResetSubscribe)).
		SetDate(now.Format(time.DateOnly)).
		SetObjectID(subscribeID).
		SetContent(string(contentJSON)).
		Save(ctx)

	return err
}
