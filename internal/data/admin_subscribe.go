package data

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribe"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribegroup"
	"github.com/OmnTeam/ppanel-pro/ent/proxyusersubscribe"
	"github.com/OmnTeam/ppanel-pro/internal/biz/admin/subscribe"
	"github.com/OmnTeam/ppanel-pro/internal/model"
	"github.com/OmnTeam/ppanel-pro/pkg/uuidx"
)

const subscribeModule = "data/admin_subscribe"

type subscribeRepo struct {
	data *Data
	log  *log.Helper
}

// NewSubscribeRepo create subscribe repository
func NewSubscribeRepo(data *Data, logger log.Logger) subscribe.SubscribeRepo {
	return &subscribeRepo{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", subscribeModule)),
	}
}

// ==================== Subscribe Operations ====================

// CreateSubscribe create subscribe
func (r *subscribeRepo) CreateSubscribe(ctx context.Context, sub *model.Subscribe) error {
	_, err := r.data.db.ProxySubscribe.Create().
		SetName(sub.Name).
		SetLanguage(sub.Language).
		SetDescription(sub.Description).
		SetUnitPrice(sub.UnitPrice).
		SetUnitTime(sub.UnitTime).
		SetDiscount(sub.Discount).
		SetReplacement(sub.Replacement).
		SetInventory(sub.Inventory).
		SetTraffic(sub.Traffic).
		SetSpeedLimit(sub.SpeedLimit).
		SetDeviceLimit(sub.DeviceLimit).
		SetQuota(sub.Quota).
		SetNodes(sub.Nodes).
		SetNodeTags(sub.NodeTags).
		SetShow(sub.Show).
		SetSell(sub.Sell).
		SetDeductionRatio(sub.DeductionRatio).
		SetAllowDeduction(sub.AllowDeduction).
		SetResetCycle(sub.ResetCycle).
		SetRenewalReset(sub.RenewalReset).
		Save(ctx)

	return err
}

// GetSubscribeByID get subscribe by ID
func (r *subscribeRepo) GetSubscribeByID(ctx context.Context, id int) (*ent.ProxySubscribe, error) {
	return r.data.db.ProxySubscribe.Query().
		Where(proxysubscribe.ID(int64(id))).
		Only(ctx)
}

// UpdateSubscribe update subscribe
func (r *subscribeRepo) UpdateSubscribe(ctx context.Context, sub *model.Subscribe) error {
	return r.data.db.ProxySubscribe.Update().
		Where(proxysubscribe.ID(sub.ID)).
		SetName(sub.Name).
		SetLanguage(sub.Language).
		SetDescription(sub.Description).
		SetUnitPrice(sub.UnitPrice).
		SetUnitTime(sub.UnitTime).
		SetDiscount(sub.Discount).
		SetReplacement(sub.Replacement).
		SetInventory(sub.Inventory).
		SetTraffic(sub.Traffic).
		SetSpeedLimit(sub.SpeedLimit).
		SetDeviceLimit(sub.DeviceLimit).
		SetQuota(sub.Quota).
		SetNodes(sub.Nodes).
		SetNodeTags(sub.NodeTags).
		SetShow(sub.Show).
		SetSell(sub.Sell).
		SetSort(sub.Sort).
		SetDeductionRatio(sub.DeductionRatio).
		SetAllowDeduction(sub.AllowDeduction).
		SetResetCycle(sub.ResetCycle).
		SetRenewalReset(sub.RenewalReset).
		Exec(ctx)
}

// DeleteSubscribe delete subscribe
func (r *subscribeRepo) DeleteSubscribe(ctx context.Context, id int) error {
	_, err := r.data.db.ProxySubscribe.Delete().
		Where(proxysubscribe.ID(int64(id))).
		Exec(ctx)
	return err
}

// GetSubscribeList get subscribe list with pagination and filters
func (r *subscribeRepo) GetSubscribeList(ctx context.Context, req *model.SubscribeListParams) ([]*ent.ProxySubscribe, int64, error) {
	query := r.data.db.ProxySubscribe.Query()

	// Apply filters
	if req.Language != "" {
		query = query.Where(proxysubscribe.Language(req.Language))
	}

	if req.Search != "" {
		query = query.Where(proxysubscribe.NameContains(req.Search))
	}

	if len(req.IDs) > 0 {
		query = query.Where(proxysubscribe.IDIn(req.IDs...))
	}

	// Get total count
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (req.Page - 1) * req.Size
	list, err := query.
		Order(ent.Desc(proxysubscribe.FieldSort)).
		Offset(offset).
		Limit(req.Size).
		All(ctx)

	if err != nil {
		return nil, 0, err
	}

	return list, int64(total), nil
}

// CheckSubscribeInUse check if subscribe is being used by active user subscriptions
func (r *subscribeRepo) CheckSubscribeInUse(ctx context.Context, subscribeID int) (bool, error) {
	// Query user_subscribe table to check if there are active subscriptions
	// Note: This assumes user_subscribe table exists and has subscribe_id and status fields
	// Status 1 means active subscription
	// This is a simplified implementation - actual implementation should query the user service

	// For now, return false as we don't have the user_subscribe table schema
	// In production, you would query: SELECT COUNT(*) FROM user_subscribe WHERE subscribe_id=? AND status=1
	return false, nil
}

// BatchDeleteSubscribe batch delete subscribes
func (r *subscribeRepo) BatchDeleteSubscribe(ctx context.Context, ids []int) error {
	// Convert []int to []int64 for the query
	int64IDs := make([]int64, len(ids))
	for i, id := range ids {
		int64IDs[i] = int64(id)
	}
	_, err := r.data.db.ProxySubscribe.Delete().
		Where(proxysubscribe.IDIn(int64IDs...)).
		Exec(ctx)
	return err
}

// GetSubscribeMinSort get minimum sort value for given IDs
func (r *subscribeRepo) GetSubscribeMinSort(ctx context.Context, ids []int) (int64, error) {
	// Convert []int to []int64 for the query
	int64IDs := make([]int64, len(ids))
	for i, id := range ids {
		int64IDs[i] = int64(id)
	}
	subscribes, err := r.data.db.ProxySubscribe.Query().
		Where(proxysubscribe.IDIn(int64IDs...)).
		Order(ent.Asc(proxysubscribe.FieldSort)).
		Limit(1).
		All(ctx)

	if err != nil {
		return 0, err
	}

	if len(subscribes) == 0 {
		return 0, nil
	}

	return int64(subscribes[0].Sort), nil
}

// BatchUpdateSubscribeSort batch update subscribe sort values
func (r *subscribeRepo) BatchUpdateSubscribeSort(ctx context.Context, subscribes []*ent.ProxySubscribe) error {
	// Use transaction to update all subscribes
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return err
	}

	for _, sub := range subscribes {
		err = tx.ProxySubscribe.UpdateOneID(sub.ID).
			SetSort(sub.Sort).
			Exec(ctx)
		if err != nil {
			return rollback(tx, err)
		}
	}

	return tx.Commit()
}

// ==================== Subscribe Group Operations ====================

// CreateSubscribeGroup create subscribe group
func (r *subscribeRepo) CreateSubscribeGroup(ctx context.Context, group *model.SubscribeGroup) error {
	_, err := r.data.db.ProxySubscribeGroup.Create().
		SetName(group.Name).
		SetDescription(group.Description).
		Save(ctx)

	return err
}

// GetSubscribeGroupByID get subscribe group by ID
func (r *subscribeRepo) GetSubscribeGroupByID(ctx context.Context, id int) (*ent.ProxySubscribeGroup, error) {
	return r.data.db.ProxySubscribeGroup.Query().
		Where(proxysubscribegroup.ID(int64(id))).
		Only(ctx)
}

// UpdateSubscribeGroup update subscribe group
func (r *subscribeRepo) UpdateSubscribeGroup(ctx context.Context, group *model.SubscribeGroup) error {
	return r.data.db.ProxySubscribeGroup.Update().
		Where(proxysubscribegroup.ID(group.ID)).
		SetName(group.Name).
		SetDescription(group.Description).
		Exec(ctx)
}

// DeleteSubscribeGroup delete subscribe group
func (r *subscribeRepo) DeleteSubscribeGroup(ctx context.Context, id int) error {
	_, err := r.data.db.ProxySubscribeGroup.Delete().
		Where(proxysubscribegroup.ID(int64(id))).
		Exec(ctx)
	return err
}

// GetSubscribeGroupList get all subscribe groups (no pagination)
func (r *subscribeRepo) GetSubscribeGroupList(ctx context.Context) ([]*ent.ProxySubscribeGroup, int64, error) {
	list, err := r.data.db.ProxySubscribeGroup.Query().
		All(ctx)

	if err != nil {
		return nil, 0, err
	}

	return list, int64(len(list)), nil
}

// BatchDeleteSubscribeGroup batch delete subscribe groups
func (r *subscribeRepo) BatchDeleteSubscribeGroup(ctx context.Context, ids []int) error {
	// Convert []int to []int64 for the query
	int64IDs := make([]int64, len(ids))
	for i, id := range ids {
		int64IDs[i] = int64(id)
	}
	_, err := r.data.db.ProxySubscribeGroup.Delete().
		Where(proxysubscribegroup.IDIn(int64IDs...)).
		Exec(ctx)
	return err
}

// ==================== User Subscription Operations ====================

// GetActiveUserSubscriptionCount get active user subscription count for a subscribe
func (r *subscribeRepo) GetActiveUserSubscriptionCount(ctx context.Context, subscribeID int) (int64, error) {
	// 查询ProxyUserSubscribe表，统计该订阅套餐的活跃用户数
	// 活跃订阅：status=1（激活状态）
	status := int8(1) // 激活状态
	count, err := r.data.db.ProxyUserSubscribe.Query().
		Where(
			proxyusersubscribe.SubscribeIDEQ(int64(subscribeID)),
			proxyusersubscribe.StatusEQ(status),
		).
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return int64(count), nil
}

// GetActiveUserSubscriptionCountByIDs get active user subscription counts for multiple subscribes
func (r *subscribeRepo) GetActiveUserSubscriptionCountByIDs(ctx context.Context, subscribeIDs []int64) (map[int64]int64, error) {
	// 查询ProxyUserSubscribe表，统计多个订阅套餐的活跃用户数
	// 活跃订阅：status=1（激活状态）
	result := make(map[int64]int64)
	status := int8(1) // 激活状态

	// 为每个订阅套餐ID统计用户数
	for _, id := range subscribeIDs {
		count, err := r.data.db.ProxyUserSubscribe.Query().
			Where(
				proxyusersubscribe.SubscribeIDEQ(id),
				proxyusersubscribe.StatusEQ(status),
			).
			Count(ctx)
		if err != nil {
			return nil, err
		}
		result[id] = int64(count)
	}
	return result, nil
}

func (r *subscribeRepo) ResetAllSubscribeToken(ctx context.Context) error {
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return err
	}

	userSubs, err := tx.ProxyUserSubscribe.Query().
		Where(proxyusersubscribe.StatusIn(1, 2)).
		All(ctx)
	if err != nil {
		return rollback(tx, err)
	}

	nowMillis := time.Now().UnixMilli()
	oldTokens := make(map[int64]string, len(userSubs))
	for _, userSub := range userSubs {
		if userSub.Token != nil {
			oldTokens[userSub.ID] = *userSub.Token
		}
		token := uuidx.SubscribeToken(fmt.Sprintf("%d%d", nowMillis, userSub.ID))
		subscribeUUID := uuid.NewString()
		if _, err := tx.ProxyUserSubscribe.UpdateOneID(userSub.ID).
			SetToken(token).
			SetUUID(subscribeUUID).
			Save(ctx); err != nil {
			return rollback(tx, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// rollback helper function to rollback transaction
func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		return rerr
	}
	return err
}
