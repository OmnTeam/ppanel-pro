package subscribe

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"

	v1 "github.com/OmnTeam/ppanel-pro/api/admin/subscribe/v1"
	"github.com/OmnTeam/ppanel-pro/ent"
	"github.com/OmnTeam/ppanel-pro/internal/model"
	"github.com/OmnTeam/ppanel-pro/internal/responsecode"
)

const module = "biz/admin/subscribe"

func parseStringID(s string) (int, error) {
	val, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	return int(val), nil
}

func parseStringID64(s string) (int64, error) {
	val, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}
	return val, nil
}

// SubscribeUseCase subscribe use case
type SubscribeUseCase struct {
	repo SubscribeRepo
	log  *log.Helper
}

// NewSubscribeUseCase create subscribe use case
func NewSubscribeUseCase(repo SubscribeRepo, logger log.Logger) *SubscribeUseCase {
	return &SubscribeUseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", module)),
	}
}

// SubscribeRepo subscribe repository interface
type SubscribeRepo interface {
	// Subscribe operations
	CreateSubscribe(ctx context.Context, sub *model.Subscribe) error
	GetSubscribeByID(ctx context.Context, id int) (*ent.ProxySubscribe, error)
	UpdateSubscribe(ctx context.Context, sub *model.Subscribe) error
	DeleteSubscribe(ctx context.Context, id int) error
	GetSubscribeList(ctx context.Context, req *model.SubscribeListParams) ([]*ent.ProxySubscribe, int32, error)
	CheckSubscribeInUse(ctx context.Context, subscribeID int) (bool, error)
	BatchDeleteSubscribe(ctx context.Context, ids []int) error
	GetSubscribeMinSort(ctx context.Context, ids []int) (int64, error)
	BatchUpdateSubscribeSort(ctx context.Context, subscribes []*ent.ProxySubscribe) error

	// Subscribe group operations
	CreateSubscribeGroup(ctx context.Context, group *model.SubscribeGroup) error
	GetSubscribeGroupByID(ctx context.Context, id int) (*ent.ProxySubscribeGroup, error)
	UpdateSubscribeGroup(ctx context.Context, group *model.SubscribeGroup) error
	DeleteSubscribeGroup(ctx context.Context, id int) error
	GetSubscribeGroupList(ctx context.Context) ([]*ent.ProxySubscribeGroup, int32, error)
	BatchDeleteSubscribeGroup(ctx context.Context, ids []int) error

	// User subscription query (for checking if subscribe is in use)
	GetActiveUserSubscriptionCount(ctx context.Context, subscribeID int) (int64, error)
	GetActiveUserSubscriptionCountByIDs(ctx context.Context, subscribeIDs []int64) (map[int64]int64, error)
	ResetAllSubscribeToken(ctx context.Context) error
}

// ==================== Subscribe Operations ====================

// CreateSubscribe create subscribe
func (uc *SubscribeUseCase) CreateSubscribe(ctx context.Context, req *v1.CreateSubscribeRequest) error {
	discountJSON, err := marshalJSON(convertDiscountToModel(req.Discount))
	if err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "Marshal discount failed", "error", err)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}
	trafficLimitJSON, err := marshalJSON(convertTrafficLimitToModel(req.TrafficLimit))
	if err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "Marshal traffic limit failed", "error", err)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	sub := &model.Subscribe{
		Name:              req.Name,
		Language:          req.Language,
		Description:       req.Description,
		UnitPrice:         req.UnitPrice,
		UnitTime:          req.UnitTime,
		Discount:          discountJSON,
		Replacement:       req.Replacement,
		Inventory:         int64(req.Inventory),
		Traffic:           req.Traffic,
		SpeedLimit:        int64(req.SpeedLimit),
		DeviceLimit:       int64(req.DeviceLimit),
		Quota:             int64(req.Quota),
		Nodes:             int64SliceToString(req.Nodes),
		NodeTags:          stringSliceToString(req.NodeTags),
		NodeGroupIDs:      cloneInt64Slice(req.NodeGroupIds),
		NodeGroupID:       req.NodeGroupId,
		TrafficLimit:      trafficLimitJSON,
		Show:              getBoolValue(req.Show, false),
		Sell:              getBoolValue(req.Sell, false),
		Sort:              0,
		DeductionRatio:    int64(req.DeductionRatio),
		AllowDeduction:    getBoolValue(req.AllowDeduction, true),
		ResetCycle:        int64(req.ResetCycle),
		RenewalReset:      getBoolValue(req.RenewalReset, false),
		ShowOriginalPrice: req.ShowOriginalPrice,
	}

	if err := uc.repo.CreateSubscribe(ctx, sub); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "CreateSubscribe failed", "error", err)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// UpdateSubscribe update subscribe
func (uc *SubscribeUseCase) UpdateSubscribe(ctx context.Context, req *v1.UpdateSubscribeRequest) error {
	// Check if subscribe exists
	id := int(req.Id)
	if id <= 0 {
		return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	_, err := uc.repo.GetSubscribeByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			uc.log.WithContext(ctx).Errorw("msg", "UpdateSubscribe subscribe not found", "error", err, "id", req.Id)
			return responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		uc.log.WithContext(ctx).Errorw("msg", "UpdateSubscribe GetSubscribeByID error", "error", err, "id", req.Id)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	discountJSON, err := marshalJSON(convertDiscountToModel(req.Discount))
	if err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "Marshal discount failed", "error", err)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}
	trafficLimitJSON, err := marshalJSON(convertTrafficLimitToModel(req.TrafficLimit))
	if err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "Marshal traffic limit failed", "error", err)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	sub := &model.Subscribe{
		ID:                int64(id),
		Name:              req.Name,
		Language:          req.Language,
		Description:       req.Description,
		UnitPrice:         req.UnitPrice,
		UnitTime:          req.UnitTime,
		Discount:          discountJSON,
		Replacement:       req.Replacement,
		Inventory:         int64(req.Inventory),
		Traffic:           req.Traffic,
		SpeedLimit:        int64(req.SpeedLimit),
		DeviceLimit:       int64(req.DeviceLimit),
		Quota:             int64(req.Quota),
		Nodes:             int64SliceToString(req.Nodes),
		NodeTags:          stringSliceToString(req.NodeTags),
		NodeGroupIDs:      cloneInt64Slice(req.NodeGroupIds),
		NodeGroupID:       req.NodeGroupId,
		TrafficLimit:      trafficLimitJSON,
		Show:              getBoolValue(req.Show, false),
		Sell:              getBoolValue(req.Sell, false),
		Sort:              int64(req.Sort),
		DeductionRatio:    int64(req.DeductionRatio),
		AllowDeduction:    getBoolValue(req.AllowDeduction, true),
		ResetCycle:        int64(req.ResetCycle),
		RenewalReset:      getBoolValue(req.RenewalReset, false),
		ShowOriginalPrice: req.ShowOriginalPrice,
	}

	if err := uc.repo.UpdateSubscribe(ctx, sub); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "UpdateSubscribe failed", "error", err, "id", req.Id)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// DeleteSubscribe delete subscribe
func (uc *SubscribeUseCase) DeleteSubscribe(ctx context.Context, id int) error {
	// Check if subscribe is in use by active user subscriptions
	inUse, err := uc.repo.CheckSubscribeInUse(ctx, id)
	if err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "DeleteSubscribe CheckSubscribeInUse error", "error", err, "id", id)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	if inUse {
		uc.log.WithContext(ctx).Warnw("msg", "DeleteSubscribe subscribe is in use", "id", id)
		return responsecode.NewKratosError(responsecode.ErrSubscribeInUse)
	}

	if err := uc.repo.DeleteSubscribe(ctx, id); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "DeleteSubscribe failed", "error", err, "id", id)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// BatchDeleteSubscribe batch delete subscribes
func (uc *SubscribeUseCase) BatchDeleteSubscribe(ctx context.Context, ids []int) error {
	// Check each subscribe if it's in use
	for _, id := range ids {
		inUse, err := uc.repo.CheckSubscribeInUse(ctx, id)
		if err != nil {
			uc.log.WithContext(ctx).Errorw("msg", "BatchDeleteSubscribe CheckSubscribeInUse error", "error", err, "id", id)
			return responsecode.NewKratosError(responsecode.ErrInternalError)
		}

		if inUse {
			uc.log.WithContext(ctx).Warnw("msg", "BatchDeleteSubscribe subscribe is in use", "id", id)
			return responsecode.NewKratosError(responsecode.ErrSubscribeInUse)
		}
	}

	// Delete all subscribes
	if err := uc.repo.BatchDeleteSubscribe(ctx, ids); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "BatchDeleteSubscribe failed", "error", err, "ids", ids)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// GetSubscribeDetails get subscribe details
func (uc *SubscribeUseCase) GetSubscribeDetails(ctx context.Context, id int) (*v1.SubscribeInfo, error) {
	sub, err := uc.repo.GetSubscribeByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			uc.log.WithContext(ctx).Errorw("msg", "GetSubscribeDetails subscribe not found", "error", err, "id", id)
			return nil, responsecode.NewKratosError(responsecode.ErrSubscribeNotFound)
		}
		uc.log.WithContext(ctx).Errorw("msg", "GetSubscribeDetails failed", "error", err, "id", id)
		return nil, responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return convertSubscribeToProto(sub), nil
}

// GetSubscribeList get subscribe list
func (uc *SubscribeUseCase) GetSubscribeList(ctx context.Context, req *v1.GetSubscribeListRequest) (*v1.GetSubscribeListData, error) {
	params := &model.SubscribeListParams{
		Page:        int(req.Page),
		Size:        int(req.Size),
		Language:    req.Language,
		Search:      req.Search,
		NodeGroupID: req.NodeGroupId,
	}

	list, total, err := uc.repo.GetSubscribeList(ctx, params)
	if err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "GetSubscribeList failed", "error", err, "params", params)
		return nil, responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	// Get subscribe IDs for querying sold counts
	subscribeIDs := make([]int64, 0, len(list))
	for _, sub := range list {
		subscribeIDs = append(subscribeIDs, sub.ID)
	}

	// Get active user subscription counts (sold count)
	soldCounts := make(map[int64]int64)
	if len(subscribeIDs) > 0 {
		soldCounts, err = uc.repo.GetActiveUserSubscriptionCountByIDs(ctx, subscribeIDs)
		if err != nil {
			uc.log.WithContext(ctx).Errorw("msg", "GetSubscribeList GetActiveUserSubscriptionCountByIDs error", "error", err)
			// Don't fail the request, just log the error
		}
	}

	// Convert to proto
	items := make([]*v1.SubscribeItem, 0, len(list))
	for _, sub := range list {
		item := convertSubscribeToProtoItem(sub)
		item.Sold = soldCounts[int64(sub.ID)]
		items = append(items, item)
	}

	return &v1.GetSubscribeListData{
		List:  items,
		Total: total,
	}, nil
}

// SubscribeSort subscribe sort
func (uc *SubscribeUseCase) SubscribeSort(ctx context.Context, req *v1.SubscribeSortRequest) error {
	if len(req.Sort) == 0 {
		return nil
	}

	// Extract IDs
	ids := make([]int, 0, len(req.Sort))
	sortMap := make(map[int64]int64)
	for i, item := range req.Sort {
		if item.Id <= 0 {
			return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
		}
		id := int(item.Id)
		ids = append(ids, id)
		sortMap[int64(id)] = int64(i)
	}

	// Get minimum sort value
	minSort, err := uc.repo.GetSubscribeMinSort(ctx, ids)
	if err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "SubscribeSort GetSubscribeMinSort error", "error", err, "ids", ids)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	// Get subscribes
	idsInt64 := make([]int64, len(ids))
	for i, v := range ids {
		idsInt64[i] = int64(v)
	}
	params := &model.SubscribeListParams{
		Page: 1,
		Size: 9999,
		IDs:  idsInt64,
	}
	subscribes, _, err := uc.repo.GetSubscribeList(ctx, params)
	if err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "SubscribeSort GetSubscribeList error", "error", err, "ids", ids)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	// Update sort values
	for _, sub := range subscribes {
		if newSort, ok := sortMap[sub.ID]; ok {
			sub.Sort = int32(minSort + newSort)
		}
	}

	// Batch update
	if err := uc.repo.BatchUpdateSubscribeSort(ctx, subscribes); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "SubscribeSort BatchUpdateSubscribeSort error", "error", err)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// ==================== Subscribe Group Operations ====================

// CreateSubscribeGroup create subscribe group
func (uc *SubscribeUseCase) CreateSubscribeGroup(ctx context.Context, req *v1.CreateSubscribeGroupRequest) error {
	group := &model.SubscribeGroup{
		Name:                req.Name,
		Description:         req.Description,
		IsExpiredGroup:      req.IsExpiredGroup,
		ExpiredDaysLimit:    req.ExpiredDaysLimit,
		MaxTrafficGBExpired: req.MaxTrafficGbExpired,
		SpeedLimit:          int64(req.SpeedLimit),
	}

	if err := uc.repo.CreateSubscribeGroup(ctx, group); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "CreateSubscribeGroup failed", "error", err)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// UpdateSubscribeGroup update subscribe group
func (uc *SubscribeUseCase) UpdateSubscribeGroup(ctx context.Context, req *v1.UpdateSubscribeGroupRequest) error {
	id := req.Id
	if id <= 0 {
		return responsecode.NewKratosError(responsecode.ErrInvalidParameter)
	}

	group := &model.SubscribeGroup{
		ID:                  id,
		Name:                req.Name,
		Description:         req.Description,
		IsExpiredGroup:      req.IsExpiredGroup,
		ExpiredDaysLimit:    req.ExpiredDaysLimit,
		MaxTrafficGBExpired: req.MaxTrafficGbExpired,
		SpeedLimit:          int64(req.SpeedLimit),
	}

	if err := uc.repo.UpdateSubscribeGroup(ctx, group); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "UpdateSubscribeGroup failed", "error", err, "id", req.Id)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// DeleteSubscribeGroup delete subscribe group
func (uc *SubscribeUseCase) DeleteSubscribeGroup(ctx context.Context, id int) error {
	if err := uc.repo.DeleteSubscribeGroup(ctx, id); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "DeleteSubscribeGroup failed", "error", err, "id", id)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// BatchDeleteSubscribeGroup batch delete subscribe groups
func (uc *SubscribeUseCase) BatchDeleteSubscribeGroup(ctx context.Context, ids []int) error {
	if err := uc.repo.BatchDeleteSubscribeGroup(ctx, ids); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "BatchDeleteSubscribeGroup failed", "error", err, "ids", ids)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	return nil
}

// GetSubscribeGroupList get subscribe group list
func (uc *SubscribeUseCase) GetSubscribeGroupList(ctx context.Context) (*v1.GetSubscribeGroupListData, error) {
	list, total, err := uc.repo.GetSubscribeGroupList(ctx)
	if err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "GetSubscribeGroupList failed", "error", err)
		return nil, responsecode.NewKratosError(responsecode.ErrInternalError)
	}

	// Convert to proto
	groups := make([]*v1.SubscribeGroupInfo, 0, len(list))
	for _, group := range list {
		desc := ""
		if group.Description != nil {
			desc = *group.Description
		}
		groups = append(groups, &v1.SubscribeGroupInfo{
			Id:                  int64(group.ID),
			Name:                group.Name,
			Description:         desc,
			IsExpiredGroup:      group.IsExpiredGroup,
			ExpiredDaysLimit:    derefInt32(group.ExpiredDaysLimit),
			MaxTrafficGbExpired: derefInt32(group.MaxTrafficGBExpired),
			SpeedLimit:          int32(derefInt64(group.SpeedLimit)),
			CreatedAt:           group.CreatedAt.Unix(),
			UpdatedAt:           group.UpdatedAt.Unix(),
		})
	}

	return &v1.GetSubscribeGroupListData{
		List:  groups,
		Total: total,
	}, nil
}

func (uc *SubscribeUseCase) ResetAllSubscribeToken(ctx context.Context) error {
	if err := uc.repo.ResetAllSubscribeToken(ctx); err != nil {
		uc.log.WithContext(ctx).Errorw("msg", "ResetAllSubscribeToken failed", "error", err)
		return responsecode.NewKratosError(responsecode.ErrInternalError)
	}
	return nil
}

// ==================== Helper Functions ====================

// convertDiscountToModel convert proto discount to model discount
func convertDiscountToModel(discounts []*v1.SubscribeDiscount) []model.SubscribeDiscount {
	result := make([]model.SubscribeDiscount, 0, len(discounts))
	for _, d := range discounts {
		result = append(result, model.SubscribeDiscount{
			Quantity: d.Quantity,
			Discount: d.Discount,
		})
	}
	return result
}

// convertDiscountFromJSON convert JSON discount to proto discount
func convertDiscountFromJSON(discountJSON string) []*v1.SubscribeDiscount {
	if discountJSON == "" {
		return nil
	}

	var discounts []model.SubscribeDiscount
	if err := json.Unmarshal([]byte(discountJSON), &discounts); err != nil {
		return nil
	}

	result := make([]*v1.SubscribeDiscount, 0, len(discounts))
	for _, d := range discounts {
		result = append(result, &v1.SubscribeDiscount{
			Quantity: d.Quantity,
			Discount: d.Discount,
		})
	}
	return result
}

func convertTrafficLimitToModel(limits []*v1.TrafficLimit) []model.TrafficLimit {
	result := make([]model.TrafficLimit, 0, len(limits))
	for _, limit := range limits {
		result = append(result, model.TrafficLimit{
			StatType:     limit.StatType,
			StatValue:    limit.StatValue,
			TrafficUsage: limit.TrafficUsage,
			SpeedLimit:   int64(limit.SpeedLimit),
		})
	}
	return result
}

func convertTrafficLimitFromJSON(raw string) []*v1.TrafficLimit {
	if raw == "" {
		return nil
	}

	var limits []model.TrafficLimit
	if err := json.Unmarshal([]byte(raw), &limits); err != nil {
		return nil
	}

	result := make([]*v1.TrafficLimit, 0, len(limits))
	for _, limit := range limits {
		result = append(result, &v1.TrafficLimit{
			StatType:     limit.StatType,
			StatValue:    limit.StatValue,
			TrafficUsage: limit.TrafficUsage,
			SpeedLimit:   int32(int64(limit.SpeedLimit)),
		})
	}
	return result
}

func marshalJSON(v any) (string, error) {
	if v == nil {
		return "", nil
	}
	switch vv := v.(type) {
	case []model.SubscribeDiscount:
		if len(vv) == 0 {
			return "", nil
		}
	case []model.TrafficLimit:
		if len(vv) == 0 {
			return "", nil
		}
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// int64SliceToString convert int64 slice to comma-separated string
func int64SliceToString(slice []int64) string {
	if len(slice) == 0 {
		return ""
	}
	strs := make([]string, 0, len(slice))
	for _, v := range slice {
		strs = append(strs, fmt.Sprintf("%d", v))
	}
	return strings.Join(strs, ",")
}

// stringToInt64Slice convert comma-separated string to int64 slice
func stringToInt64Slice(s string) []int64 {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var val int64
		fmt.Sscanf(p, "%d", &val)
		result = append(result, val)
	}
	return result
}

// stringSliceToString convert string slice to comma-separated string
func stringSliceToString(slice []string) string {
	return strings.Join(slice, ",")
}

// stringToStringSlice convert comma-separated string to string slice
func stringToStringSlice(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

// getBoolValue get bool value from optional bool pointer
func getBoolValue(ptr *bool, defaultValue bool) bool {
	if ptr == nil {
		return defaultValue
	}
	return *ptr
}

func cloneInt64Slice(input []int64) []int64 {
	if input == nil {
		return nil
	}
	out := make([]int64, len(input))
	copy(out, input)
	return out
}

func derefInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}
func derefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

// convertSubscribeToProto convert ent subscribe to proto subscribe info
func convertSubscribeToProto(sub *ent.ProxySubscribe) *v1.SubscribeInfo {
	desc := ""
	if sub.Description != nil {
		desc = *sub.Description
	}
	discount := ""
	if sub.Discount != nil {
		discount = *sub.Discount
	}
	trafficLimit := ""
	if sub.TrafficLimit != nil {
		trafficLimit = *sub.TrafficLimit
	}
	deductionRatio := int64(0)
	if sub.DeductionRatio != nil {
		deductionRatio = int64(*sub.DeductionRatio)
	}
	allowDeduction := sub.AllowDeduction
	resetCycle := int64(0)
	if sub.ResetCycle != nil {
		resetCycle = int64(*sub.ResetCycle)
	}
	renewalReset := sub.RenewalReset

	return &v1.SubscribeInfo{
		Id:                int64(sub.ID),
		Name:              sub.Name,
		Language:          sub.Language,
		Description:       desc,
		UnitPrice:         int64(sub.UnitPrice),
		UnitTime:          sub.UnitTime,
		Discount:          convertDiscountFromJSON(discount),
		Replacement:       int64(sub.Replacement),
		Inventory:         int32(sub.Inventory),
		Traffic:           int64(sub.Traffic),
		SpeedLimit:        int32(sub.SpeedLimit),
		DeviceLimit:       int32(sub.DeviceLimit),
		Quota:             int32(sub.Quota),
		Nodes:             stringToInt64Slice(sub.Nodes),
		NodeTags:          stringToStringSlice(sub.NodeTags),
		NodeGroupIds:      cloneInt64Slice(sub.NodeGroupIds),
		NodeGroupId:       derefInt64(sub.NodeGroupID),
		TrafficLimit:      convertTrafficLimitFromJSON(trafficLimit),
		Show:              sub.Show,
		Sell:              sub.Sell,
		Sort:              int32(sub.Sort),
		DeductionRatio:    int32(deductionRatio),
		AllowDeduction:    allowDeduction,
		ResetCycle:        int32(resetCycle),
		RenewalReset:      renewalReset,
		ShowOriginalPrice: sub.ShowOriginalPrice,
		CreatedAt:         sub.CreatedAt.Unix(),
		UpdatedAt:         sub.UpdatedAt.Unix(),
	}
}

// convertSubscribeToProtoItem convert ent subscribe to proto subscribe item
func convertSubscribeToProtoItem(sub *ent.ProxySubscribe) *v1.SubscribeItem {
	desc := ""
	if sub.Description != nil {
		desc = *sub.Description
	}
	discount := ""
	if sub.Discount != nil {
		discount = *sub.Discount
	}
	trafficLimit := ""
	if sub.TrafficLimit != nil {
		trafficLimit = *sub.TrafficLimit
	}
	deductionRatio := int64(0)
	if sub.DeductionRatio != nil {
		deductionRatio = int64(*sub.DeductionRatio)
	}
	allowDeduction := sub.AllowDeduction
	resetCycle := int64(0)
	if sub.ResetCycle != nil {
		resetCycle = int64(*sub.ResetCycle)
	}
	renewalReset := sub.RenewalReset

	return &v1.SubscribeItem{
		Id:                int64(sub.ID),
		Name:              sub.Name,
		Language:          sub.Language,
		Description:       desc,
		UnitPrice:         int64(sub.UnitPrice),
		UnitTime:          sub.UnitTime,
		Discount:          convertDiscountFromJSON(discount),
		Replacement:       int64(sub.Replacement),
		Inventory:         int32(sub.Inventory),
		Traffic:           int64(sub.Traffic),
		SpeedLimit:        int32(sub.SpeedLimit),
		DeviceLimit:       int32(sub.DeviceLimit),
		Quota:             int32(sub.Quota),
		Nodes:             stringToInt64Slice(sub.Nodes),
		NodeTags:          stringToStringSlice(sub.NodeTags),
		NodeGroupIds:      cloneInt64Slice(sub.NodeGroupIds),
		NodeGroupId:       derefInt64(sub.NodeGroupID),
		TrafficLimit:      convertTrafficLimitFromJSON(trafficLimit),
		Show:              sub.Show,
		Sell:              sub.Sell,
		Sort:              int32(sub.Sort),
		DeductionRatio:    int32(deductionRatio),
		AllowDeduction:    allowDeduction,
		ResetCycle:        int32(resetCycle),
		RenewalReset:      renewalReset,
		ShowOriginalPrice: sub.ShowOriginalPrice,
		CreatedAt:         sub.CreatedAt.Unix(),
		UpdatedAt:         sub.UpdatedAt.Unix(),
		Sold:              0, // Will be set by caller
	}
}
