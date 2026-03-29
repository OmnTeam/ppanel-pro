package redemption

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxyorder"
	"github.com/OmnTeam/ppanel-pro/ent/proxyredemptioncode"
	"github.com/OmnTeam/ppanel-pro/ent/proxyredemptionrecord"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribe"
	"github.com/OmnTeam/ppanel-pro/ent/proxyusersubscribe"
	"github.com/OmnTeam/ppanel-pro/pkg/tool"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// RedemptionRepo 兑换码仓库接口
type RedemptionRepo interface {
	GetDB() *ent.Client
	GetRedis() *redis.Client
	GetQueue() *asynq.Client
}

// RedemptionUseCase 兑换码用例
type RedemptionUseCase struct {
	repo   RedemptionRepo
	logger *log.Helper
}

// NewRedemptionUseCase 创建兑换码用例
func NewRedemptionUseCase(repo RedemptionRepo, logger log.Logger) *RedemptionUseCase {
	return &RedemptionUseCase{
		repo:   repo,
		logger: log.NewHelper(logger),
	}
}

// RedeemCodeResult 兑换结果
type RedeemCodeResult struct {
	OrderNo string
	Message string
}

// RedeemCode 兑换兑换码
func (uc *RedemptionUseCase) RedeemCode(ctx context.Context, userID int64, code string) (*RedeemCodeResult, error) {
	db := uc.repo.GetDB()
	redis := uc.repo.GetRedis()
	queue := uc.repo.GetQueue()

	// 使用Redis分布式锁防止并发重复兑换
	lockKey := fmt.Sprintf("redemption_lock:%d:%s", userID, code)
	lockSuccess, err := redis.SetNX(ctx, lockKey, "1", 10*time.Second).Result()
	if err != nil {
		return nil, fmt.Errorf("获取锁失败: %w", err)
	}
	if !lockSuccess {
		return nil, fmt.Errorf("兑换进行中，请稍候")
	}
	defer redis.Del(ctx, lockKey)

	// 查询兑换码
	redemptionCode, err := db.ProxyRedemptionCode.Query().
		Where(proxyredemptioncode.CodeEQ(code)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("兑换码不存在")
		}
		return nil, fmt.Errorf("查询兑换码失败: %w", err)
	}

	// 检查兑换码是否启用
	if redemptionCode.Status != 1 {
		return nil, fmt.Errorf("兑换码已禁用")
	}

	// 检查兑换码是否还有剩余次数
	if redemptionCode.TotalCount > 0 && redemptionCode.UsedCount >= redemptionCode.TotalCount {
		return nil, fmt.Errorf("兑换码已用完")
	}

	// 检查用户是否已经兑换过此码
	existingRecord, err := db.ProxyRedemptionRecord.Query().
		Where(
			proxyredemptionrecord.UserIDEQ(userID),
			proxyredemptionrecord.RedemptionCodeIDEQ(redemptionCode.ID),
		).
		First(ctx)

	if err == nil && existingRecord != nil {
		return nil, fmt.Errorf("您已经兑换过此兑换码")
	}

	// 查询订阅套餐
	subscribePlan, err := db.ProxySubscribe.Query().
		Where(proxysubscribe.IDEQ(redemptionCode.SubscribePlan)).
		Only(ctx)

	if err != nil {
		return nil, fmt.Errorf("订阅套餐不存在: %w", err)
	}

	// 检查订阅套餐是否可售
	if !subscribePlan.Sell {
		return nil, fmt.Errorf("订阅套餐不可用")
	}

	// 检查配额限制
	if subscribePlan.Quota > 0 {
		count, err := db.ProxyUserSubscribe.Query().
			Where(
				proxyusersubscribe.UserIDEQ(userID),
				proxyusersubscribe.SubscribeIDEQ(redemptionCode.SubscribePlan),
			).
			Count(ctx)

		if err != nil {
			return nil, fmt.Errorf("检查配额失败: %w", err)
		}

		if int64(count) >= subscribePlan.Quota {
			return nil, fmt.Errorf("订阅配额已达上限")
		}
	}

	// 判断是否首次购买
	isNew := false
	orderCount, err := db.ProxyOrder.Query().
		Where(
			proxyorder.UserIDEQ(userID),
			proxyorder.StatusEQ(2), // 已支付
		).
		Count(ctx)
	if err == nil && orderCount == 0 {
		isNew = true
	}

	// 创建订单
	orderNo := tool.GenerateTradeNo()
	order, err := db.ProxyOrder.Create().
		SetUserID(userID).
		SetOrderNo(orderNo).
		SetType(5). // 兑换类型
		SetQuantity(int64(redemptionCode.Quantity)).
		SetPrice(0).
		SetAmount(0).
		SetDiscount(0).
		SetGiftAmount(0).
		SetCoupon("").
		SetCouponDiscount(0).
		SetPaymentID(0).
		SetMethod("redemption").
		SetFeeAmount(0).
		SetCommission(0).
		SetStatus(2). // 直接设置为已支付
		SetSubscribeID(redemptionCode.SubscribePlan).
		SetIsNew(isNew).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("创建订单失败: %w", err)
	}

	// 缓存兑换码信息到Redis
	cacheKey := fmt.Sprintf("redemption_order:%s", orderNo)
	cacheData := map[string]interface{}{
		"redemption_code_id": redemptionCode.ID,
		"unit_time":          redemptionCode.UnitTime,
		"quantity":           redemptionCode.Quantity,
	}
	jsonData, _ := json.Marshal(cacheData)
	err = redis.Set(ctx, cacheKey, jsonData, 2*time.Hour).Err()
	if err != nil {
		// 删除已创建的订单
		db.ProxyOrder.DeleteOneID(order.ID).Exec(ctx)
		return nil, fmt.Errorf("缓存数据失败: %w", err)
	}

	// 触发队列任务
	payload := map[string]interface{}{
		"order_no": orderNo,
	}
	payloadBytes, _ := json.Marshal(payload)
	task := asynq.NewTask("order:activate", payloadBytes, asynq.MaxRetry(5))
	_, err = queue.EnqueueContext(ctx, task)
	if err != nil {
		// 删除订单和缓存
		redis.Del(ctx, cacheKey)
		db.ProxyOrder.DeleteOneID(order.ID).Exec(ctx)
		return nil, fmt.Errorf("入队任务失败: %w", err)
	}

	uc.logger.Infof("Redemption order created: order_no=%s, user_id=%d", orderNo, userID)

	return &RedeemCodeResult{
		OrderNo: orderNo,
		Message: "兑换成功，正在处理中...",
	}, nil
}
