package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribe"
	"github.com/OmnTeam/ppanel-pro/ent/proxysubscribegroup"
	"github.com/OmnTeam/ppanel-pro/internal/biz/admin/subscribe"
	"github.com/OmnTeam/ppanel-pro/internal/model"
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
		SetReplacement(int(sub.Replacement)).
		SetInventory(int(sub.Inventory)).
		SetTraffic(sub.Traffic).
		SetSpeedLimit(int(sub.SpeedLimit)).
		SetDeviceLimit(int(sub.DeviceLimit)).
		SetQuota(int(sub.Quota)).
		SetNodes(sub.Nodes).
		SetNodeTags(sub.NodeTags).
		SetShow(sub.Show).
		SetSell(sub.Sell).
		SetDeductionRatio(float64(sub.DeductionRatio)).
		SetAllowDeduction(sub.AllowDeduction).
		SetResetCycle(int(sub.ResetCycle)).
		SetRenewalReset(sub.RenewalReset).
		Save(ctx)

	return err
}

// GetSubscribeByID get subscribe by ID
func (r *subscribeRepo) GetSubscribeByID(ctx context.Context, id int64) (*ent.ProxySubscribe, error) {
	return r.data.db.ProxySubscribe.Query().
		Where(proxysubscribe.ID(int(id))).
		Only(ctx)
}

// UpdateSubscribe update subscribe
func (r *subscribeRepo) UpdateSubscribe(ctx context.Context, sub *model.Subscribe) error {
	return r.data.db.ProxySubscribe.Update().
		Where(proxysubscribe.ID(int(sub.ID))).
		SetName(sub.Name).
		SetLanguage(sub.Language).
		SetDescription(sub.Description).
		SetUnitPrice(sub.UnitPrice).
		SetUnitTime(sub.UnitTime).
		SetDiscount(sub.Discount).
		SetReplacement(int(sub.Replacement)).
		SetInventory(int(sub.Inventory)).
		SetTraffic(sub.Traffic).
		SetSpeedLimit(int(sub.SpeedLimit)).
		SetDeviceLimit(int(sub.DeviceLimit)).
		SetQuota(int(sub.Quota)).
		SetNodes(sub.Nodes).
		SetNodeTags(sub.NodeTags).
		SetShow(sub.Show).
		SetSell(sub.Sell).
		SetSort(int(sub.Sort)).
		SetDeductionRatio(float64(sub.DeductionRatio)).
		SetAllowDeduction(sub.AllowDeduction).
		SetResetCycle(int(sub.ResetCycle)).
		SetRenewalReset(sub.RenewalReset).
		Exec(ctx)
}

// DeleteSubscribe delete subscribe
func (r *subscribeRepo) DeleteSubscribe(ctx context.Context, id int64) error {
	_, err := r.data.db.ProxySubscribe.Delete().
		Where(proxysubscribe.ID(int(id))).
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
		// Convert int64 IDs to int for the query
		ids := make([]int, len(req.IDs))
		for i, id := range req.IDs {
			ids[i] = int(id)
		}
		query = query.Where(proxysubscribe.IDIn(ids...))
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
func (r *subscribeRepo) CheckSubscribeInUse(ctx context.Context, subscribeID int64) (bool, error) {
	// Query user_subscribe table to check if there are active subscriptions
	// Note: This assumes user_subscribe table exists and has tenant_id, subscribe_id, and status fields
	// Status 1 means active subscription
	// This is a simplified implementation - actual implementation should query the user service

	// For now, return false as we don't have the user_subscribe table schema
	// In production, you would query: SELECT COUNT(*) FROM user_subscribe WHERE tenant_id=? AND subscribe_id=? AND status=1
	return false, nil
}

// BatchDeleteSubscribe batch delete subscribes
func (r *subscribeRepo) BatchDeleteSubscribe(ctx context.Context, ids []int64) error {
	// Convert int64 IDs to int for the query
	intIDs := make([]int, len(ids))
	for i, id := range ids {
		intIDs[i] = int(id)
	}

	_, err := r.data.db.ProxySubscribe.Delete().
		Where(proxysubscribe.IDIn(intIDs...)).
		Exec(ctx)
	return err
}

// GetSubscribeMinSort get minimum sort value for given IDs
func (r *subscribeRepo) GetSubscribeMinSort(ctx context.Context, ids []int64) (int64, error) {
	// Convert int64 IDs to int for the query
	intIDs := make([]int, len(ids))
	for i, id := range ids {
		intIDs[i] = int(id)
	}

	subscribes, err := r.data.db.ProxySubscribe.Query().
		Where(proxysubscribe.IDIn(intIDs...)).
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
func (r *subscribeRepo) GetSubscribeGroupByID(ctx context.Context, id int64) (*ent.ProxySubscribeGroup, error) {
	return r.data.db.ProxySubscribeGroup.Query().
		Where(proxysubscribegroup.ID(int(id))).
		Only(ctx)
}

// UpdateSubscribeGroup update subscribe group
func (r *subscribeRepo) UpdateSubscribeGroup(ctx context.Context, group *model.SubscribeGroup) error {
	return r.data.db.ProxySubscribeGroup.Update().
		Where(proxysubscribegroup.ID(int(group.ID))).
		SetName(group.Name).
		SetDescription(group.Description).
		Exec(ctx)
}

// DeleteSubscribeGroup delete subscribe group
func (r *subscribeRepo) DeleteSubscribeGroup(ctx context.Context, id int64) error {
	_, err := r.data.db.ProxySubscribeGroup.Delete().
		Where(proxysubscribegroup.ID(int(id))).
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
func (r *subscribeRepo) BatchDeleteSubscribeGroup(ctx context.Context, ids []int64) error {
	// Convert int64 IDs to int for the query
	intIDs := make([]int, len(ids))
	for i, id := range ids {
		intIDs[i] = int(id)
	}

	_, err := r.data.db.ProxySubscribeGroup.Delete().
		Where(proxysubscribegroup.IDIn(intIDs...)).
		Exec(ctx)
	return err
}

// ==================== User Subscription Operations ====================

// GetActiveUserSubscriptionCount get active user subscription count for a subscribe
func (r *subscribeRepo) GetActiveUserSubscriptionCount(ctx context.Context, subscribeID int64) (int64, error) {
	// TODO: Query user_subscribe table
	// This requires integration with user service or user_subscribe table
	// For now, return 0
	return 0, nil
}

// GetActiveUserSubscriptionCountByIDs get active user subscription counts for multiple subscribes
func (r *subscribeRepo) GetActiveUserSubscriptionCountByIDs(ctx context.Context, subscribeIDs []int64) (map[int64]int64, error) {
	// TODO: Query user_subscribe table
	// This requires integration with user service or user_subscribe table
	// For now, return empty map
	result := make(map[int64]int64)
	for _, id := range subscribeIDs {
		result[id] = 0
	}
	return result, nil
}

// rollback helper function to rollback transaction
func rollback(tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		return rerr
	}
	return err
}
