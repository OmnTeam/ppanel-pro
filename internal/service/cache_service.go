package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/OmnTeam/npanel-pro/ent"
	"github.com/OmnTeam/npanel-pro/ent/proxyorder"
	"github.com/OmnTeam/npanel-pro/ent/proxyuserauthmethod"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

// CacheService 缓存服务
type CacheService struct {
	redisClient   *redis.Client
	db            *ent.Client
	logger        *log.Helper
	defaultExpiry time.Duration
	shortExpiry   time.Duration
}

// NewCacheService 创建缓存服务
func NewCacheService(redisClient *redis.Client, db *ent.Client, logger log.Logger) *CacheService {
	return &CacheService{
		redisClient:   redisClient,
		db:            db,
		logger:        log.NewHelper(logger),
		defaultExpiry: 24 * time.Hour,  // 24小时
		shortExpiry:   5 * time.Minute, // 5分钟
	}
}

// UserCacheModel 用户缓存模型
type UserCacheModel struct {
	ID          int64            `json:"id"`
	Email       string           `json:"email"`
	Balance     *int64           `json:"balance,omitempty"`
	GiftAmount  *int64           `json:"gift_amount,omitempty"`
	AuthMethods []UserAuthMethod `json:"auth_methods,omitempty"`
}

// UserAuthMethod 用户认证方法
type UserAuthMethod struct {
	UserID         int64  `json:"user_id"`
	AuthType       string `json:"auth_type"`
	AuthIdentifier string `json:"auth_identifier"`
}

// OrderCacheModel 订单缓存模型
type OrderCacheModel struct {
	ID         int64  `json:"id"`
	OrderNo    string `json:"order_no"`
	UserID     int64  `json:"user_id"`
	Status     int8   `json:"status"`
	Amount     int64  `json:"amount"`
	GiftAmount int64  `json:"gift_amount"`
}

// 缓存键前缀
const (
	UserIDKeyPrefix      = "user:%d"
	UserBalanceKeyPrefix = "user_balance:%d"
	UserEmailKeyPrefix   = "user_email:%s"
	OrderNoKeyPrefix     = "order:%s"
	UserOrdersPrefix     = "user_orders:%d:%s"
)

// GetCacheKeys 实现CacheKeyGenerator接口
func (u *UserCacheModel) GetCacheKeys() []string {
	if u == nil {
		return []string{}
	}

	keys := []string{
		fmt.Sprintf(UserIDKeyPrefix, u.ID),
		fmt.Sprintf(UserBalanceKeyPrefix, u.ID),
	}

	// 如果有邮箱，添加邮箱缓存键
	for _, auth := range u.AuthMethods {
		if auth.AuthType == "email" {
			keys = append(keys, fmt.Sprintf(UserEmailKeyPrefix, auth.AuthIdentifier))
			break
		}
	}

	return keys
}

// GetCacheKeys 实现CacheKeyGenerator接口
func (o *OrderCacheModel) GetCacheKeys() []string {
	if o == nil {
		return []string{}
	}

	return []string{
		fmt.Sprintf(OrderNoKeyPrefix, o.OrderNo),
		fmt.Sprintf(UserOrdersPrefix, o.UserID, "pending"),
		fmt.Sprintf(UserOrdersPrefix, o.UserID, "paid"),
		fmt.Sprintf(UserOrdersPrefix, o.UserID, "closed"),
	}
}

// GetUserFromCache 从缓存获取用户信息
func (cs *CacheService) GetUserFromCache(ctx context.Context, userID int64) (*UserCacheModel, error) {
	var user UserCacheModel
	key := fmt.Sprintf(UserIDKeyPrefix, userID)

	value, err := cs.redisClient.Get(ctx, key).Result()
	if err != nil {
		cs.logger.Debugf("Get user from cache failed: userID=%d, error=%v", userID, err)
		return nil, err
	}

	err = json.Unmarshal([]byte(value), &user)
	if err != nil {
		cs.logger.Debugf("Unmarshal user cache failed: userID=%d, error=%v", userID, err)
		return nil, err
	}

	cs.logger.Debugf("Get user from cache success: userID=%d", userID)
	return &user, nil
}

// SetUserCache 设置用户缓存
func (cs *CacheService) SetUserCache(ctx context.Context, user *ent.ProxyUser) error {
	var balance, giftAmount int64
	if user.Balance != nil {
		balance = *user.Balance
	}
	if user.GiftAmount != nil {
		giftAmount = *user.GiftAmount
	}

	userModel := &UserCacheModel{
		ID:         user.ID,
		Balance:    &balance,
		GiftAmount: &giftAmount,
	}

	// 获取用户的认证方法
	authMethods, err := cs.db.ProxyUserAuthMethod.Query().
		Where(proxyuserauthmethod.UserID(user.ID)).
		All(ctx)
	if err != nil {
		cs.logger.Warnf("Failed to get user auth methods for cache: userID=%d, error=%v", user.ID, err)
	} else {
		for _, auth := range authMethods {
			userModel.AuthMethods = append(userModel.AuthMethods, UserAuthMethod{
				UserID:         int64(auth.UserID),
				AuthType:       auth.AuthType,
				AuthIdentifier: auth.AuthIdentifier,
			})

			if auth.AuthType == "email" {
				userModel.Email = auth.AuthIdentifier
			}
		}
	}

	key := fmt.Sprintf(UserIDKeyPrefix, user.ID)
	value, err := json.Marshal(userModel)
	if err != nil {
		cs.logger.Errorf("Marshal user cache failed: userID=%d, error=%v", user.ID, err)
		return err
	}

	err = cs.redisClient.Set(ctx, key, value, cs.defaultExpiry).Err()
	if err != nil {
		cs.logger.Errorf("Set user cache failed: userID=%d, error=%v", user.ID, err)
		return err
	}

	// 设置用户余额缓存
	balanceKey := fmt.Sprintf(UserBalanceKeyPrefix, user.ID)
	balanceData := map[string]interface{}{
		"balance":     user.Balance,
		"gift_amount": user.GiftAmount,
	}
	balanceValue, _ := json.Marshal(balanceData)
	err = cs.redisClient.Set(ctx, balanceKey, balanceValue, cs.shortExpiry).Err()
	if err != nil {
		cs.logger.Warnf("Set user balance cache failed: userID=%d, error=%v", user.ID, err)
	}

	cs.logger.Debugf("Set user cache success: userID=%d", user.ID)
	return nil
}

// ClearUserCache 清理用户相关缓存
func (cs *CacheService) ClearUserCache(ctx context.Context, userID int64) error {
	// 先从数据库获取用户信息，以便生成完整的缓存键
	user, err := cs.db.ProxyUser.Get(ctx, userID)
	if err != nil {
		cs.logger.Warnf("Failed to get user for cache clearing: userID=%d, error=%v", userID, err)
		// 即使获取用户失败，也尝试清理基本缓存键
		userModel := &UserCacheModel{ID: userID}
		return cs.clearModelCache(ctx, userModel)
	}

	var balance, giftAmount int64
	if user.Balance != nil {
		balance = int64(*user.Balance)
	}
	if user.GiftAmount != nil {
		giftAmount = *user.GiftAmount
	}

	userModel := &UserCacheModel{
		ID:         user.ID,
		Balance:    &balance,
		GiftAmount: &giftAmount,
	}

	// 获取认证方法
	authMethods, err := cs.db.ProxyUserAuthMethod.Query().
		Where(proxyuserauthmethod.UserID(user.ID)).
		All(ctx)
	if err == nil {
		for _, auth := range authMethods {
			userModel.AuthMethods = append(userModel.AuthMethods, UserAuthMethod{
				UserID:         int64(auth.UserID),
				AuthType:       auth.AuthType,
				AuthIdentifier: auth.AuthIdentifier,
			})

			if auth.AuthType == "email" {
				userModel.Email = auth.AuthIdentifier
			}
		}
	}

	err = cs.clearModelCache(ctx, userModel)
	if err != nil {
		cs.logger.Errorf("Clear user cache failed: userID=%d, error=%v", userID, err)
		return err
	}

	cs.logger.Debugf("Clear user cache success: userID=%d", userID)
	return nil
}

// GetOrderFromCache 从缓存获取订单信息
func (cs *CacheService) GetOrderFromCache(ctx context.Context, orderNo string) (*OrderCacheModel, error) {
	var order OrderCacheModel
	key := fmt.Sprintf(OrderNoKeyPrefix, orderNo)

	value, err := cs.redisClient.Get(ctx, key).Result()
	if err != nil {
		cs.logger.Debugf("Get order from cache failed: orderNo=%s, error=%v", orderNo, err)
		return nil, err
	}

	err = json.Unmarshal([]byte(value), &order)
	if err != nil {
		cs.logger.Debugf("Unmarshal order cache failed: orderNo=%s, error=%v", orderNo, err)
		return nil, err
	}

	cs.logger.Debugf("Get order from cache success: orderNo=%s", orderNo)
	return &order, nil
}

// SetOrderCache 设置订单缓存
func (cs *CacheService) SetOrderCache(ctx context.Context, order *ent.ProxyOrder) error {
	orderModel := &OrderCacheModel{
		ID:         int64(order.ID),
		OrderNo:    order.OrderNo,
		UserID:     int64(order.UserID),
		Status:     order.Status,
		Amount:     order.Amount,
		GiftAmount: order.GiftAmount,
	}

	key := fmt.Sprintf(OrderNoKeyPrefix, order.OrderNo)
	value, err := json.Marshal(orderModel)
	if err != nil {
		cs.logger.Errorf("Marshal order cache failed: orderNo=%s, error=%v", order.OrderNo, err)
		return err
	}

	err = cs.redisClient.Set(ctx, key, value, cs.shortExpiry).Err()
	if err != nil {
		cs.logger.Errorf("Set order cache failed: orderNo=%s, error=%v", order.OrderNo, err)
		return err
	}

	cs.logger.Debugf("Set order cache success: orderNo=%s", order.OrderNo)
	return nil
}

// ClearOrderCache 清理订单相关缓存
func (cs *CacheService) ClearOrderCache(ctx context.Context, orderNo string) error {
	// 先从缓存获取订单信息
	orderModel, err := cs.GetOrderFromCache(ctx, orderNo)
	if err != nil {
		// 如果缓存中没有，从数据库获取
		order, dbErr := cs.db.ProxyOrder.Query().
			Where(proxyorder.OrderNo(orderNo)).
			Only(ctx)
		if dbErr != nil {
			cs.logger.Warnf("Failed to get order for cache clearing: orderNo=%s, error=%v", orderNo, dbErr)
			// 即使获取订单失败，也尝试清理基本缓存键
			orderModel = &OrderCacheModel{OrderNo: orderNo}
		} else {
			orderModel = &OrderCacheModel{
				ID:      int64(order.ID),
				OrderNo: order.OrderNo,
				UserID:  int64(order.UserID),
				Status:  order.Status,
			}
		}
	}

	err = cs.clearModelCache(ctx, orderModel)
	if err != nil {
		cs.logger.Errorf("Clear order cache failed: orderNo=%s, error=%v", orderNo, err)
		return err
	}

	cs.logger.Debugf("Clear order cache success: orderNo=%s", orderNo)
	return nil
}

// UpdateUserBalanceCache 更新用户余额缓存
func (cs *CacheService) UpdateUserBalanceCache(ctx context.Context, userID int64, balance, giftAmount int64) error {
	// 更新余额缓存
	balanceKey := fmt.Sprintf(UserBalanceKeyPrefix, userID)
	balanceData := map[string]interface{}{
		"balance":     balance,
		"gift_amount": giftAmount,
	}

	balanceValue, _ := json.Marshal(balanceData)
	err := cs.redisClient.Set(ctx, balanceKey, balanceValue, cs.shortExpiry).Err()
	if err != nil {
		cs.logger.Errorf("Update user balance cache failed: userID=%d, error=%v", userID, err)
		return err
	}

	// 更新完整用户缓存中的余额信息
	userModel, err := cs.GetUserFromCache(ctx, userID)
	if err != nil {
		// 如果缓存中没有，从数据库获取并重新设置
		user, dbErr := cs.db.ProxyUser.Get(ctx, userID)
		if dbErr != nil {
			cs.logger.Warnf("Failed to get user for balance cache update: userID=%d, error=%v", userID, dbErr)
			return nil // 不阻止业务流程
		}
		user.Balance = &balance
		user.GiftAmount = &giftAmount
		return cs.SetUserCache(ctx, user)
	}

	// 更新缓存中的余额信息
	userModel.Balance = &balance
	userModel.GiftAmount = &giftAmount

	userKey := fmt.Sprintf(UserIDKeyPrefix, userID)
	userValue, _ := json.Marshal(userModel)
	err = cs.redisClient.Set(ctx, userKey, userValue, cs.defaultExpiry).Err()
	if err != nil {
		cs.logger.Errorf("Update user cache balance failed: userID=%d, error=%v", userID, err)
		return err
	}

	cs.logger.Debugf("Update user balance cache success: userID=%d, balance=%d, giftAmount=%d",
		userID, balance, giftAmount)
	return nil
}

// BatchClearRelatedCache 批量清理相关缓存
func (cs *CacheService) BatchClearRelatedCache(ctx context.Context, userID int64) error {
	// 清理用户相关缓存
	err := cs.ClearUserCache(ctx, userID)
	if err != nil {
		cs.logger.Errorf("Failed to clear user cache in batch: userID=%d, error=%v", userID, err)
		// 继续清理其他缓存，不因为单个失败而停止
	}

	// 可以根据需要添加其他相关缓存清理逻辑
	// 例如：清理用户订阅缓存、订单列表缓存等

	cs.logger.Infof("Batch clear related cache completed: userID=%d", userID)
	return nil
}

// clearModelCache 清理模型缓存
func (cs *CacheService) clearModelCache(ctx context.Context, model interface{}) error {
	var keys []string
	switch m := model.(type) {
	case *UserCacheModel:
		keys = m.GetCacheKeys()
	case *OrderCacheModel:
		keys = m.GetCacheKeys()
	default:
		return nil
	}

	for _, key := range keys {
		err := cs.redisClient.Del(ctx, key).Err()
		if err != nil {
			cs.logger.Warnf("Failed to delete cache key: key=%s, error=%v", key, err)
		}
	}

	return nil
}

// SetCache 设置通用缓存
func (cs *CacheService) SetCache(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return cs.redisClient.Set(ctx, key, data, expiry).Err()
}

// GetCache 获取通用缓存
func (cs *CacheService) GetCache(ctx context.Context, key string, dest interface{}) error {
	value, err := cs.redisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(value), dest)
}

// ClearCache 清理通用缓存
func (cs *CacheService) ClearCache(ctx context.Context, key string) error {
	return cs.redisClient.Del(ctx, key).Err()
}
