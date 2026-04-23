package handler

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribe"
	"github.com/OmnTeam/ppanel-pro/ent/proxyusersubscribe"
	logmodel "github.com/OmnTeam/ppanel-pro/internal/model/log"
	queueTypes "github.com/OmnTeam/ppanel-pro/internal/queue/types"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// ResetTrafficHandler handles traffic reset logic for different subscription cycles
// Supports three reset modes:
// - reset_cycle = 1: Reset on 1st of every month
// - reset_cycle = 2: Reset monthly based on subscription start date
// - reset_cycle = 3: Reset yearly based on subscription start date
// 完整复刻原项目逻辑：server-master/queue/logic/traffic/resetTrafficLogic.go
type ResetTrafficHandler struct {
	db     *ent.Client
	rdb    *redis.Client
	queue  *asynq.Client
	logger *log.Helper
}

// Cache and retry configuration constants
const (
	maxRetryAttempts = 3
	retryDelay       = 30 * time.Minute
	lockTimeout      = 5 * time.Minute
)

// Cache keys
var (
	cacheKey      = "reset_traffic_cache"
	retryCountKey = "reset_traffic_retry_count"
	lockKey       = "reset_traffic_lock"
)

// resetTrafficCache stores the last reset time to prevent duplicate processing
type resetTrafficCache struct {
	LastResetTime time.Time `json:"last_reset_time"`
}

// NewResetTrafficHandler 创建流量重置处理器
func NewResetTrafficHandler(db *ent.Client, rdb *redis.Client, queue *asynq.Client, logger log.Logger) *ResetTrafficHandler {
	return &ResetTrafficHandler{
		db:     db,
		rdb:    rdb,
		queue:  queue,
		logger: log.NewHelper(logger),
	}
}

// ProcessTask executes the traffic reset task for all subscription types with enhanced retry mechanism
// 完整复刻原项目 line 58-181
func (h *ResetTrafficHandler) ProcessTask(ctx context.Context, _ *asynq.Task) error {
	var err error
	startTime := time.Now()

	// Get current retry count
	retryCount := h.getRetryCount(ctx)
	h.logger.Infof("[ResetTraffic] Starting task execution, retryCount=%d, startTime=%s",
		retryCount, startTime.Format(time.DateTime))

	// Acquire distributed lock to prevent duplicate execution
	lockAcquired := h.acquireLock(ctx)
	if !lockAcquired {
		h.logger.Info("[ResetTraffic] Another task is already running, skipping execution")
		return nil
	}
	defer h.releaseLock(ctx)

	defer func() {
		if err != nil {
			// Check if error is retryable and within retry limit
			if h.isRetryableError(err) && retryCount < maxRetryAttempts {
				// Increment retry count
				h.setRetryCount(ctx, retryCount+1)

				// Schedule retry with delay
				task := asynq.NewTask(queueTypes.SchedulerResetTraffic, nil)
				_, retryErr := h.queue.Enqueue(task, asynq.ProcessIn(retryDelay))
				if retryErr != nil {
					h.logger.Errorf("[ResetTraffic] Failed to enqueue retry task: %v, retryCount=%d", retryErr, retryCount)
				} else {
					h.logger.Infof("[ResetTraffic] Task failed, retrying in 30 minutes: error=%v, retryCount=%d, maxRetryAttempts=%d",
						err, retryCount+1, maxRetryAttempts)
				}
			} else {
				// Max retries reached or non-retryable error
				if retryCount >= maxRetryAttempts {
					h.logger.Errorf("[ResetTraffic] Max retry attempts reached, giving up: retryCount=%d, maxRetryAttempts=%d, error=%v",
						retryCount, maxRetryAttempts, err)
				} else {
					h.logger.Errorf("[ResetTraffic] Non-retryable error, not retrying: error=%v, retryCount=%d", err, retryCount)
				}
				// Reset retry count for next scheduled task
				h.clearRetryCount(ctx)
			}
		} else {
			// Task completed successfully, reset retry count
			h.clearRetryCount(ctx)
			h.logger.Infof("[ResetTraffic] Task completed successfully, processingTime=%v, retryCount=%d",
				time.Since(startTime), retryCount)
		}
	}()

	// Load last reset time from cache
	var cache resetTrafficCache
	cacheData, cacheErr := h.rdb.Get(ctx, cacheKey).Result()
	if cacheErr != nil {
		if !errors.Is(cacheErr, redis.Nil) {
			h.logger.Errorf("[ResetTraffic] Failed to get cache: %v", cacheErr)
		}
		// Set default value if cache not found
		cache = resetTrafficCache{
			LastResetTime: time.Now().Add(-10 * time.Minute),
		}
		h.logger.Infof("[ResetTraffic] Using default cache value, lastResetTime=%s", cache.LastResetTime.Format(time.DateTime))
	} else {
		// Parse JSON data
		if unmarshalErr := json.Unmarshal([]byte(cacheData), &cache); unmarshalErr != nil {
			h.logger.Errorf("[ResetTraffic] Failed to unmarshal cache: %v", unmarshalErr)
			cache = resetTrafficCache{
				LastResetTime: time.Now().Add(-10 * time.Minute),
			}
		} else {
			h.logger.Infof("[ResetTraffic] Cache loaded successfully, lastResetTime=%s", cache.LastResetTime.Format(time.DateTime))
		}
	}

	// Execute reset operations in order: yearly -> monthly (1st) -> monthly (cycle)
	// 复刻原项目 line 144-161
	err = h.resetYear(ctx)
	if err != nil {
		h.logger.Errorf("[ResetTraffic] Yearly reset failed: %v", err)
		return err
	}

	err = h.reset1st(ctx, cache)
	if err != nil {
		h.logger.Errorf("[ResetTraffic] Monthly 1st reset failed: %v", err)
		return err
	}

	err = h.resetMonth(ctx)
	if err != nil {
		h.logger.Errorf("[ResetTraffic] Monthly cycle reset failed: %v", err)
		return err
	}

	// Update cache with current time after successful processing (复刻原项目 line 163-178)
	updatedCache := resetTrafficCache{
		LastResetTime: startTime,
	}
	cacheDataBytes, marshalErr := json.Marshal(updatedCache)
	if marshalErr != nil {
		h.logger.Errorf("[ResetTraffic] Failed to marshal cache: %v", marshalErr)
	} else {
		updateCacheErr := h.rdb.Set(ctx, cacheKey, cacheDataBytes, 0).Err()
		if updateCacheErr != nil {
			h.logger.Errorf("[ResetTraffic] Failed to update cache: %v", updateCacheErr)
			// Don't return error here as the main task completed successfully
		} else {
			h.logger.Infof("[ResetTraffic] Cache updated successfully, newLastResetTime=%s", startTime.Format(time.DateTime))
		}
	}

	return nil
}

// resetMonth handles monthly cycle reset based on subscription start date
// reset_cycle = 2: Reset monthly based on subscription start date
// 完整复刻原项目 line 185-265
func (h *ResetTrafficHandler) resetMonth(ctx context.Context) error {
	now := time.Now()

	// Get all subscriptions that reset monthly based on start date (reset_cycle=2)
	resetMonthSubs, err := h.db.ProxySubscribe.Query().
		Where(proxysubscribe.ResetCycle(2)).
		All(ctx)
	if err != nil {
		h.logger.Errorf("[ResetTraffic] Failed to query monthly subscriptions: %v", err)
		return err
	}

	if len(resetMonthSubs) == 0 {
		h.logger.Info("[ResetTraffic] No monthly cycle subscriptions found")
		return nil
	}

	var resetMonthSubIds []int64
	for _, sub := range resetMonthSubs {
		resetMonthSubIds = append(resetMonthSubIds, sub.ID)
	}

	// Check if today is the last day of current month
	isLastDayOfMonth := now.AddDate(0, 0, 1).Month() != now.Month()

	// Query users for monthly reset (复刻原项目 line 202-225)
	// 注意：由于Ent不支持TIMESTAMPDIFF，需要使用原生SQL或通过应用层过滤
	// 这里我们先查询所有可能的记录，然后在应用层过滤
	queryBuilder := h.db.ProxyUserSubscribe.Query().
		Where(
			proxyusersubscribe.SubscribeIDIn(resetMonthSubIds...),
			proxyusersubscribe.StatusIn(1, 2), // Only active subscriptions
		)

	// 添加日期匹配条件
	if isLastDayOfMonth {
		// Last day of month: handle subscription start dates >= today
		// 由于Ent限制，这里需要获取所有数据后在应用层过滤
		h.logger.Info("[ResetTraffic] Last day of month detected, will filter in application layer")
	}

	allSubs, err := queryBuilder.All(ctx)
	if err != nil {
		h.logger.Errorf("[ResetTraffic] Failed to query monthly reset users: %v", err)
		return err
	}

	// 应用层过滤：检查expire_time和当前时间的关系
	var monthlyResetSubIDs []int64
	for _, sub := range allSubs {
		if sub.ExpireTime == nil {
			continue
		}

		// 检查expire_time是否在未来至少1个月
		// 原项目SQL: TIMESTAMPDIFF(MONTH, CURDATE(), DATE(expire_time)) >= 1
		// 意思是：expire_time - now >= 1 month
		monthsDiff := h.monthsDiff(*sub.ExpireTime, now)
		if monthsDiff < 1 {
			continue
		}

		expireDay := sub.ExpireTime.Day()
		currentDay := now.Day()
		if isLastDayOfMonth {
			// Last day of month: handle subscription start dates >= today
			if expireDay >= currentDay {
				monthlyResetSubIDs = append(monthlyResetSubIDs, sub.ID)
			}
		} else {
			// Normal case: exact day match
			if expireDay == currentDay {
				monthlyResetSubIDs = append(monthlyResetSubIDs, sub.ID)
			}
		}
	}

	if len(monthlyResetSubIDs) > 0 {
		h.logger.Infof("[ResetTraffic] Found users for monthly reset: count=%d, userIds=%v",
			len(monthlyResetSubIDs), monthlyResetSubIDs)

		// 使用事务更新订阅状态 (复刻原项目 line 232-242)
		err = h.db.TX(ctx, func(tx *ent.Tx) error {
			_, updateErr := tx.ProxyUserSubscribe.Update().
				Where(proxyusersubscribe.IDIn(monthlyResetSubIDs...)).
				SetUpload(0).
				SetDownload(0).
				SetStatus(1).      // Ensure status is active
				ClearFinishedAt(). // finished_at = nil
				Save(ctx)
			return updateErr
		})

		if err != nil {
			h.logger.Errorf("[ResetTraffic] Failed to update monthly reset users: %v", err)
			return err
		}

		// Find user subscriptions for cache clearing and logging (复刻原项目 line 244-251)
		userSubs, err := h.db.ProxyUserSubscribe.Query().
			Where(proxyusersubscribe.IDIn(monthlyResetSubIDs...)).
			All(ctx)
		if err != nil {
			h.logger.Errorf("[ResetTraffic] Failed to find user subscriptions for monthly reset: %v", err)
			return err
		}

		// Clear cache for these subscriptions
		h.clearCache(ctx, userSubs)
		h.logger.Infof("[ResetTraffic] Monthly reset completed: count=%d", len(monthlyResetSubIDs))
	} else {
		h.logger.Info("[ResetTraffic] No users found for monthly reset")
	}

	h.clearSubscribeCacheByIDs(ctx, resetMonthSubIds...)

	h.logger.Info("[ResetTraffic] Monthly reset process completed")
	return nil
}

// reset1st handles reset on 1st of every month
// reset_cycle = 1: Reset on 1st of every month
// 完整复刻原项目 line 269-351
func (h *ResetTrafficHandler) reset1st(ctx context.Context, cache resetTrafficCache) error {
	now := time.Now()

	// Check if we already reset this month using cache (复刻原项目 line 273-278)
	if cache.LastResetTime.Year() == now.Year() && cache.LastResetTime.Month() == now.Month() {
		h.logger.Infof("[ResetTraffic] Already reset this month, skipping 1st reset: lastResetTime=%s, currentTime=%s",
			cache.LastResetTime.Format(time.DateOnly), now.Format(time.DateOnly))
		return nil
	}

	// Only reset if it's the 1st day of the month (复刻原项目 line 281-284)
	if now.Day() != 1 {
		h.logger.Infof("[ResetTraffic] Not 1st day of month, skipping 1st reset: currentDay=%d", now.Day())
		return nil
	}

	// Get all subscriptions that reset on 1st of month (reset_cycle=1)
	reset1stSubs, err := h.db.ProxySubscribe.Query().
		Where(proxysubscribe.ResetCycle(1)).
		All(ctx)
	if err != nil {
		h.logger.Errorf("[ResetTraffic] Failed to query 1st reset subscriptions: %v", err)
		return err
	}

	if len(reset1stSubs) == 0 {
		h.logger.Info("[ResetTraffic] No 1st reset subscriptions found")
		return nil
	}

	var reset1stSubIds []int64
	for _, sub := range reset1stSubs {
		reset1stSubIds = append(reset1stSubIds, sub.ID)
	}

	// Get all active users with these subscriptions (复刻原项目 line 301-309)
	users1stReset, err := h.db.ProxyUserSubscribe.Query().
		Where(
			proxyusersubscribe.SubscribeIDIn(reset1stSubIds...),
			proxyusersubscribe.StatusIn(1, 2), // Only active subscriptions
		).
		All(ctx)

	if err != nil {
		h.logger.Errorf("[ResetTraffic] Failed to query 1st reset users: %v", err)
		return err
	}

	if len(users1stReset) > 0 {
		var userIDs []int64
		for _, user := range users1stReset {
			userIDs = append(userIDs, user.ID)
		}

		h.logger.Infof("[ResetTraffic] Found users for 1st reset: count=%d, userIds=%v",
			len(userIDs), userIDs)

		// Reset upload and download traffic to zero (复刻原项目 line 317-327)
		err = h.db.TX(ctx, func(tx *ent.Tx) error {
			_, updateErr := tx.ProxyUserSubscribe.Update().
				Where(proxyusersubscribe.IDIn(userIDs...)).
				SetUpload(0).
				SetDownload(0).
				SetStatus(1).      // Ensure status is active
				ClearFinishedAt(). // finished_at = nil
				Save(ctx)
			return updateErr
		})

		if err != nil {
			h.logger.Errorf("[ResetTraffic] Failed to update 1st reset users: %v", err)
			return err
		}

		// Clear cache for these subscriptions (复刻原项目 line 336)
		h.clearCache(ctx, users1stReset)
		h.logger.Infof("[ResetTraffic] 1st reset completed: count=%d", len(userIDs))
	} else {
		h.logger.Info("[ResetTraffic] No users found for 1st reset")
	}

	h.clearSubscribeCacheByIDs(ctx, reset1stSubIds...)

	h.logger.Info("[ResetTraffic] 1st reset process completed")
	return nil
}

// resetYear handles yearly reset based on subscription start date anniversary
// reset_cycle = 3: Reset yearly based on subscription start date
// 完整复刻原项目 line 355-441
func (h *ResetTrafficHandler) resetYear(ctx context.Context) error {
	now := time.Now()

	// Get all subscriptions that reset yearly (reset_cycle=3)
	resetYearSubs, err := h.db.ProxySubscribe.Query().
		Where(proxysubscribe.ResetCycle(3)).
		All(ctx)
	if err != nil {
		h.logger.Errorf("[ResetTraffic] Failed to query yearly subscriptions: %v", err)
		return err
	}

	if len(resetYearSubs) == 0 {
		h.logger.Info("[ResetTraffic] No yearly reset subscriptions found")
		return nil
	}

	var resetYearSubIds []int64
	for _, sub := range resetYearSubs {
		resetYearSubIds = append(resetYearSubIds, sub.ID)
	}

	// Check if today is February 28th (handle leap year case)
	isLeapYearCase := now.Month() == 2 && now.Day() == 28

	// Query users for yearly reset (复刻原项目 line 378-395)
	allYearSubs, err := h.db.ProxyUserSubscribe.Query().
		Where(
			proxyusersubscribe.SubscribeIDIn(resetYearSubIds...),
			proxyusersubscribe.StatusIn(1, 2), // Only active subscriptions
		).
		All(ctx)

	if err != nil {
		h.logger.Errorf("[ResetTraffic] Query yearly reset users failed: %v", err)
		return err
	}

	// 应用层过滤：检查月份、日期和年份差异
	var usersYearReset []int64
	for _, sub := range allYearSubs {
		if sub.ExpireTime == nil {
			continue
		}

		// Same month check
		if sub.ExpireTime.Month() != now.Month() {
			continue
		}

		// 检查expire_time是否在未来至少1年
		// 原项目SQL: TIMESTAMPDIFF(YEAR, CURDATE(), DATE(expire_time)) >= 1
		// 意思是：expire_time - now >= 1 year
		yearsDiff := h.yearsDiff(*sub.ExpireTime, now)
		if yearsDiff < 1 {
			continue
		}

		expireDay := sub.ExpireTime.Day()
		if isLeapYearCase {
			// February 28th: handle both Feb 28 and Feb 29 subscriptions
			if expireDay == 28 || expireDay == 29 {
				usersYearReset = append(usersYearReset, sub.ID)
			}
		} else {
			// Normal case: exact day match
			if expireDay == now.Day() {
				usersYearReset = append(usersYearReset, sub.ID)
			}
		}
	}

	if len(usersYearReset) > 0 {
		h.logger.Infof("[ResetTraffic] Found users for yearly reset: count=%d, userIds=%v",
			len(usersYearReset), usersYearReset)

		// Reset upload and download traffic to zero (复刻原项目 line 403-413)
		err = h.db.TX(ctx, func(tx *ent.Tx) error {
			_, updateErr := tx.ProxyUserSubscribe.Update().
				Where(proxyusersubscribe.IDIn(usersYearReset...)).
				SetUpload(0).
				SetDownload(0).
				SetStatus(1).      // Ensure status is active
				ClearFinishedAt(). // finished_at = nil
				Save(ctx)
			return updateErr
		})

		if err != nil {
			h.logger.Errorf("[ResetTraffic] Failed to update yearly reset users: %v", err)
			return err
		}

		// Find user subscriptions for cache clearing and logging (复刻原项目 line 415-422)
		userSubs, err := h.db.ProxyUserSubscribe.Query().
			Where(proxyusersubscribe.IDIn(usersYearReset...)).
			All(ctx)
		if err != nil {
			h.logger.Errorf("[ResetTraffic] Failed to find user subscriptions for yearly reset: %v", err)
			return err
		}

		// Clear cache for these subscriptions
		h.clearCache(ctx, userSubs)
		h.logger.Infof("[ResetTraffic] Yearly reset completed: count=%d", len(usersYearReset))
	} else {
		h.logger.Info("[ResetTraffic] No users found for yearly reset")
	}

	h.clearSubscribeCacheByIDs(ctx, resetYearSubIds...)

	h.logger.Info("[ResetTraffic] Yearly reset process completed")
	return nil
}

// Helper methods for retry mechanism and lock management
// 完整复刻原项目 line 443-507

// getRetryCount retrieves the current retry count from Redis (复刻原项目 line 444-461)
func (h *ResetTrafficHandler) getRetryCount(ctx context.Context) int {
	countStr, err := h.rdb.Get(ctx, retryCountKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0 // No retry count found, start with 0
		}
		h.logger.Errorf("[ResetTraffic] Failed to get retry count: %v", err)
		return 0
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		h.logger.Errorf("[ResetTraffic] Invalid retry count format: value=%s", countStr)
		return 0
	}

	return count
}

// setRetryCount sets the retry count in Redis (复刻原项目 line 464-471)
func (h *ResetTrafficHandler) setRetryCount(ctx context.Context, count int) {
	err := h.rdb.Set(ctx, retryCountKey, count, 24*time.Hour).Err()
	if err != nil {
		h.logger.Errorf("[ResetTraffic] Failed to set retry count: count=%d, error=%v", count, err)
	}
}

// clearRetryCount removes the retry count from Redis (复刻原项目 line 474-479)
func (h *ResetTrafficHandler) clearRetryCount(ctx context.Context) {
	err := h.rdb.Del(ctx, retryCountKey).Err()
	if err != nil {
		h.logger.Errorf("[ResetTraffic] Failed to clear retry count: %v", err)
	}
}

// acquireLock attempts to acquire a distributed lock (复刻原项目 line 482-497)
func (h *ResetTrafficHandler) acquireLock(ctx context.Context) bool {
	result := h.rdb.SetNX(ctx, lockKey, "locked", lockTimeout)
	acquired, err := result.Result()
	if err != nil {
		h.logger.Errorf("[ResetTraffic] Failed to acquire lock: %v", err)
		return false
	}

	if acquired {
		h.logger.Info("[ResetTraffic] Lock acquired successfully")
	} else {
		h.logger.Info("[ResetTraffic] Lock already exists, another task is running")
	}

	return acquired
}

// releaseLock releases the distributed lock (复刻原项目 line 500-507)
func (h *ResetTrafficHandler) releaseLock(ctx context.Context) {
	err := h.rdb.Del(ctx, lockKey).Err()
	if err != nil {
		h.logger.Errorf("[ResetTraffic] Failed to release lock: %v", err)
	} else {
		h.logger.Info("[ResetTraffic] Lock released successfully")
	}
}

// isRetryableError determines if an error is retryable (复刻原项目 line 510-575)
func (h *ResetTrafficHandler) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errorMessage := strings.ToLower(err.Error())

	// Network and connection errors (retryable)
	retryableErrors := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"network",
		"timeout",
		"dial",
		"context deadline exceeded",
		"temporary failure",
		"server error",
		"service unavailable",
		"internal server error",
		"database is locked",
		"too many connections",
		"deadlock",
		"lock wait timeout",
	}

	// Database constraint errors (non-retryable)
	nonRetryableErrors := []string{
		"foreign key constraint",
		"unique constraint",
		"check constraint",
		"not null constraint",
		"invalid input syntax",
		"column does not exist",
		"table does not exist",
		"permission denied",
		"access denied",
		"authentication failed",
		"invalid credentials",
	}

	// Check for non-retryable errors first
	for _, nonRetryable := range nonRetryableErrors {
		if strings.Contains(errorMessage, nonRetryable) {
			h.logger.Infof("[ResetTraffic] Non-retryable error detected: error=%v, pattern=%s", err, nonRetryable)
			return false
		}
	}

	// Check for retryable errors
	for _, retryable := range retryableErrors {
		if strings.Contains(errorMessage, retryable) {
			h.logger.Infof("[ResetTraffic] Retryable error detected: error=%v, pattern=%s", err, retryable)
			return true
		}
	}

	// Default: treat unknown errors as retryable, but log for analysis
	h.logger.Infof("[ResetTraffic] Unknown error type, treating as retryable: error=%v", err)
	return true
}

// clearCache clears the reset traffic cache (复刻原项目 line 578-607)
func (h *ResetTrafficHandler) clearCache(ctx context.Context, list []*ent.ProxyUserSubscribe) {
	if len(list) == 0 {
		return
	}

	subscribeIDs := make([]int64, 0, len(list))

	for _, sub := range list {
		if sub.SubscribeID > 0 {
			subscribeIDs = append(subscribeIDs, sub.SubscribeID)
		}
		h.insertLog(ctx, sub.ID, sub.UserID)
	}

	h.clearSubscribeCacheByIDs(ctx, subscribeIDs...)
}

// insertLog inserts a reset traffic log entry (复刻原项目 line 610-625)
func (h *ResetTrafficHandler) insertLog(ctx context.Context, subID, userID int64) {
	trafficLog := logmodel.ResetSubscribe{
		Type:      logmodel.ResetSubscribeTypeAuto,
		UserId:    userID,
		Timestamp: time.Now().UnixMilli(),
	}
	content, _ := trafficLog.Marshal()

	_, err := h.db.ProxySystemLog.Create().
		SetType(int8(logmodel.TypeResetSubscribe)).
		SetObjectID(subID).
		SetDate(time.Now().Format(time.DateOnly)).
		SetContent(string(content)).
		Save(ctx)

	if err != nil {
		h.logger.Errorf("[ResetTraffic] Failed to create system log for subscription: %v", err)
	}
}

func (h *ResetTrafficHandler) clearSubscribeCacheByIDs(ctx context.Context, subscribeIDs ...int64) {
	_ = ctx
	_ = subscribeIDs
}

// monthsDiff calculates the number of months between two times
// Helper function to replace TIMESTAMPDIFF(MONTH, ...)
func (h *ResetTrafficHandler) monthsDiff(current, expire time.Time) int {
	years := current.Year() - expire.Year()
	months := int(current.Month()) - int(expire.Month())
	return years*12 + months
}

// yearsDiff calculates the number of years between two times
// Helper function to replace TIMESTAMPDIFF(YEAR, ...)
func (h *ResetTrafficHandler) yearsDiff(current, expire time.Time) int {
	return current.Year() - expire.Year()
}
